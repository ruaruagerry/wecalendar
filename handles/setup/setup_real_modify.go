package setup

import (
	"encoding/json"
	"weagent/gconst"
	"weagent/pb"
	"weagent/rconst"
	"weagent/server"

	"github.com/golang/protobuf/proto"
	"github.com/gomodule/redigo/redis"
)

type realModifyReq struct {
	RealNick string `json:"realnick"`
	CardCode string `json:"cardcode"`
}

func realModifyHandle(c *server.StupidContext) {
	log := c.Log.WithField("func", "setup.realModifyHandle")

	httpRsp := pb.HTTPResponse{
		Result: proto.Int32(int32(gconst.UnknownError)),
	}
	defer c.WriteJSONRsp(&httpRsp)

	// req
	req := &realModifyReq{}
	if err := json.Unmarshal(c.Body, req); err != nil {
		httpRsp.Result = proto.Int32(int32(gconst.ErrParse))
		httpRsp.Msg = proto.String("请求信息解析失败")
		log.Errorf("code:%d msg:%s json Unmarshal err:%s", httpRsp.GetResult(), httpRsp.GetMsg(), err.Error())
		return
	}

	log.Info("realModifyHandle enter, req:", string(c.Body))

	conn := c.RedisConn
	playerid := c.UserID

	// 检查
	conn.Send("MULTI")
	conn.Send("SETNX", rconst.StringLockRealModifyHandlePrefix+playerid, "1")
	conn.Send("EXPIRE", rconst.StringLockRealModifyHandlePrefix+playerid, gconst.LockTime)
	redisMDArray, err := redis.Values(conn.Do("EXEC"))
	if err != nil {
		httpRsp.Result = proto.Int32(int32(gconst.ErrRedis))
		httpRsp.Msg = proto.String("请求锁获取缓存失败")
		log.Errorf("code:%d msg:%s, GET lock redis data error:(%s)", httpRsp.GetResult(), httpRsp.GetMsg(), err.Error())
		return
	}
	locktag, _ := redis.Int(redisMDArray[0], nil)
	if locktag == 0 {
		httpRsp.Result = proto.Int32(int32(gconst.ErrHTTPTooFast))
		httpRsp.Msg = proto.String("请求过于频繁")
		log.Errorf("code:%d msg:%s, request too fast", httpRsp.GetResult(), httpRsp.GetMsg())
		return
	}

	defer func() {
		conn.Do("DEL", rconst.StringLockRealModifyHandlePrefix+playerid)
	}()

	// 先检测身份证信息是否有效
	valid := checkCardCodeValid(req.CardCode)
	if !valid {
		httpRsp.Result = proto.Int32(int32(gconst.ErrSetupCardCode))
		httpRsp.Msg = proto.String("身份证格式错误")
		log.Errorf("code:%d msg:%s, invalid card code style, cardcode:%s", httpRsp.GetResult(), httpRsp.GetMsg(), req.CardCode)
		return
	}

	// 检测姓名是否有效
	valid = checkNameValid(req.RealNick)
	if !valid {
		httpRsp.Result = proto.Int32(int32(gconst.ErrSetupRealNick))
		httpRsp.Msg = proto.String("姓名格式错误")
		log.Errorf("code:%d msg:%s, invalid name style, name:%s", httpRsp.GetResult(), httpRsp.GetMsg(), req.RealNick)
		return
	}

	// redis multi get
	conn.Send("MULTI")
	conn.Send("HEXISTS", rconst.HashSetupPrefix+playerid, rconst.FieldSetupCardcode)
	conn.Send("SISMEMBER", rconst.SetSetupCardCode, req.CardCode)
	redisMDArray, err = redis.Values(conn.Do("EXEC"))
	if err != nil {
		httpRsp.Result = proto.Int32(int32(gconst.ErrRedis))
		httpRsp.Msg = proto.String("统一获取缓存操作失败")
		log.Errorf("code:%d msg:%s redisMDArray Values err, err:%s", httpRsp.GetResult(), httpRsp.GetMsg(), err.Error())
		return
	}

	hasreal, _ := redis.Bool(redisMDArray[0], nil)
	cardcodeexist, _ := redis.Bool(redisMDArray[1], nil)

	// do something
	if hasreal {
		httpRsp.Result = proto.Int32(int32(gconst.ErrSetupAlreadyRealCheck))
		httpRsp.Msg = proto.String("已进行过实名认证")
		log.Errorf("code:%d msg:%s, already real check", httpRsp.GetResult(), httpRsp.GetMsg())
		return
	}

	if cardcodeexist {
		httpRsp.Result = proto.Int32(int32(gconst.ErrSetupExistCardCode))
		httpRsp.Msg = proto.String("您的实名信息已存在")
		log.Errorf("code:%d msg:%s, exist card code, cardcode:%s", httpRsp.GetResult(), httpRsp.GetMsg(), req.CardCode)
		return
	}

	// redis multi set
	conn.Send("MULTI")
	conn.Send("SADD", rconst.SetSetupCardCode, req.CardCode)
	conn.Send("HSET", rconst.HashSetupPrefix+playerid, rconst.FieldSetupRealNick, req.RealNick)
	conn.Send("HSET", rconst.HashSetupPrefix+playerid, rconst.FieldSetupCardcode, req.CardCode)
	_, err = conn.Do("EXEC")
	if err != nil {
		httpRsp.Result = proto.Int32(int32(gconst.ErrRedis))
		httpRsp.Msg = proto.String("统一存储缓存操作失败")
		log.Errorf("code:%d msg:%s exec err, err:%s", httpRsp.GetResult(), httpRsp.GetMsg(), err.Error())
		return
	}

	// rsp
	httpRsp.Result = proto.Int32(int32(gconst.Success))

	log.Info("realModifyHandle rsp, result:", httpRsp.GetResult())

	return
}
