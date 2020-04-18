package auth

import (
	"encoding/json"
	"fmt"
	"time"
	"weagent/gconst"
	"weagent/pb"
	"weagent/rconst"
	"weagent/server"
	"weagent/tables"

	"github.com/golang/protobuf/proto"
	"github.com/gomodule/redigo/redis"
)

type loginReq struct {
	Code          string `json:"code"`
	EncryptedData string `json:"encrypteddata"`
	Iv            string `json:"iv"`
}

type clientInfo struct {
	LatestVersion string `json:"latestversion"`
}

type loginUserInfo struct {
	ID        string `json:"id"`
	NickName  string `json:"nickname"`
	Gender    int32  `json:"gender"`
	AvatarURL string `json:"avatarurl"`
}

type loginRsp struct {
	Token      string         `json:"token"`
	UserInfo   *loginUserInfo `json:"userinfo"`
	ClientInfo *clientInfo    `json:"clientinfo"`
}

func wxLoginHandle(c *server.StupidContext) {
	log := c.Log.WithField("func", "auth.loginHandle")

	httpRsp := pb.HTTPResponse{
		Result: proto.Int32(int32(gconst.UnknownError)),
	}
	defer c.WriteJSONRsp(&httpRsp)

	// req
	req := &loginReq{}
	if err := json.Unmarshal(c.Body, req); err != nil {
		httpRsp.Result = proto.Int32(int32(gconst.ErrParse))
		httpRsp.Msg = proto.String("请求信息解析失败")
		log.Errorf("code:%d msg:%s json Unmarshal err:%s", httpRsp.GetResult(), httpRsp.GetMsg(), err.Error())
		return
	}

	log.Info("loginHandle enter, req:", string(c.Body))

	db := c.DbConn
	conn := c.RedisConn
	nowtime := time.Now()

	// redis multi get
	conn.Send("MULTI")
	conn.Send("HGET", rconst.HashClient, rconst.FieldClientLastestVersion)
	redisMDArray, err := redis.Values(conn.Do("EXEC"))
	if err != nil {
		httpRsp.Result = proto.Int32(int32(gconst.ErrRedis))
		httpRsp.Msg = proto.String("统一获取缓存操作失败")
		log.Errorf("code:%d msg:%s redisMDArray Values err, err:%s", httpRsp.GetResult(), httpRsp.GetMsg(), err.Error())
		return
	}

	clientlatestversion, _ := redis.String(redisMDArray[0], nil)

	// account处理
	loadAccessTokenReply, err := WeixinGetUserInfo(req.Code)
	if err != nil {
		httpRsp.Result = proto.Int32(int32(gconst.ErrAuthGetWexinUserInfo))
		httpRsp.Msg = proto.String("获取微信用户信息失败")
		log.Errorf("code:%d msg:%s WeixinGetUserInfo err:%s", httpRsp.GetResult(), httpRsp.GetMsg(), err.Error())
		return
	}

	userinfo, err := GetWeiXinPlusUserInfo(loadAccessTokenReply.SessionKey, req.EncryptedData, req.Iv)
	if err != nil {
		httpRsp.Result = proto.Int32(int32(gconst.ErrAuthGetWeixinPlusInfo))
		httpRsp.Msg = proto.String("获取微信额外信息失败")
		log.Errorf("code:%d msg:%s GetWeiXinPlusUserInfo err:%s", httpRsp.GetResult(), httpRsp.GetMsg(), err.Error())
		return
	}

	if userinfo.OpenID == "" {
		httpRsp.Result = proto.Int32(int32(gconst.ErrAuthNotFindOpenID))
		httpRsp.Msg = proto.String("没找到用户Openid")
		log.Errorf("code:%d msg:%s WeixinGetUserInfo err:%s", httpRsp.GetResult(), httpRsp.GetMsg(), err.Error())
		return
	}

	// db操作
	row := &tables.Account{OpenID: userinfo.OpenID}
	_, err = db.Get(row)
	if err != nil {
		httpRsp.Result = proto.Int32(int32(gconst.ErrDB))
		httpRsp.Msg = proto.String("查询用户信息失败")
		log.Errorf("code:%d msg:%s db where err:%s", httpRsp.GetResult(), httpRsp.GetMsg(), err.Error())
		return
	}
	log.Infof("openid:%s, row:%v", userinfo.OpenID, row)

	if row.ID != 0 {
		row.Nick = userinfo.NickName
		row.Gender = userinfo.Gender
		row.Portrait = userinfo.AvatarURL
		row.OpenID = userinfo.OpenID
		row.UnionID = userinfo.UnionID
		row.SessionKey = loadAccessTokenReply.SessionKey
		row.LastLoginTime = nowtime
		_, err := db.Where("open_id = ?", userinfo.OpenID).Update(row)
		if err != nil {
			httpRsp.Result = proto.Int32(int32(gconst.ErrDB))
			httpRsp.Msg = proto.String("更新用户信息失败")
			log.Errorf("code:%d msg:%s db update err:%s", httpRsp.GetResult(), httpRsp.GetMsg(), err.Error())
			return
		}
	} else {
		row = &tables.Account{
			Nick:          userinfo.NickName,
			Gender:        userinfo.Gender,
			Portrait:      userinfo.AvatarURL,
			OpenID:        userinfo.OpenID,
			UnionID:       userinfo.UnionID,
			SessionKey:    loadAccessTokenReply.SessionKey,
			CreateTime:    nowtime,
			LastLoginTime: nowtime,
			Channel:       tables.ChannelTypeWx,
		}
		_, err := db.Insert(row)
		if err != nil {
			httpRsp.Result = proto.Int32(int32(gconst.ErrParse))
			httpRsp.Msg = proto.String("插入用户信息失败")
			log.Errorf("code:%d msg:%s db insert err:%s", httpRsp.GetResult(), httpRsp.GetMsg(), err.Error())
			return
		}
	}
	playerid := fmt.Sprintf("%d", row.ID)

	// do something

	// redis multi set
	conn.Send("MULTI")
	conn.Send("HMSET", rconst.HashAccountPrefix+playerid,
		rconst.FieldAccUserID, row.ID,
		rconst.FieldAccName, row.Nick,
		rconst.FieldAccImage, row.Portrait,
		rconst.FieldAccGender, row.Gender,
		rconst.FieldAccOpenID, row.OpenID,
		rconst.FieldAccUnionID, row.UnionID,
		rconst.FieldAccChannel, tables.ChannelTypeWx)
	conn.Send("SADD", rconst.SetUsers, playerid)
	if row.OpenID != "" {
		conn.Send("HMSET", rconst.HashAccountPrefix+row.OpenID,
			rconst.FieldAccUserID, row.ID,
			rconst.FieldAccUnionID, row.UnionID)
	}
	_, err = conn.Do("EXEC")
	if err != nil {
		httpRsp.Result = proto.Int32(int32(gconst.ErrRedis))
		httpRsp.Msg = proto.String("统一存储缓存操作失败")
		log.Errorf("code:%d msg:%s exec err, err:%s", httpRsp.GetResult(), httpRsp.GetMsg(), err.Error())
		return
	}

	// 生成token， 根据目前客户端的约定需要设置到header上
	token := server.GenTK(playerid)

	// rsp
	rspuserinfo := &loginUserInfo{
		ID:        playerid,
		NickName:  row.Nick,
		Gender:    row.Gender,
		AvatarURL: row.Portrait,
	}
	rspclientinfo := &clientInfo{
		LatestVersion: clientlatestversion,
	}
	rsp := &loginRsp{
		Token:      token,
		UserInfo:   rspuserinfo,
		ClientInfo: rspclientinfo,
	}
	data, err := json.Marshal(rsp)
	if err != nil {
		httpRsp.Result = proto.Int32(int32(gconst.ErrParse))
		httpRsp.Msg = proto.String("返回信息marshal解析失败")
		log.Errorf("code:%d msg:%s json marshal err, err:%s", httpRsp.GetResult(), httpRsp.GetMsg(), err.Error())
		return
	}
	httpRsp.Result = proto.Int32(int32(gconst.Success))
	httpRsp.Data = data

	log.Info("loginHandle rsp, rsp:", string(data))

	return
}
