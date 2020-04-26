package divination

import (
	"encoding/json"
	"time"
	"wecalendar/gconst"
	"wecalendar/gfunc"
	"wecalendar/pb"
	"wecalendar/rconst"
	"wecalendar/server"

	"github.com/golang/protobuf/proto"
	"github.com/gomodule/redigo/redis"
)

type divinationPublicReq struct {
	Content string `json:"content"`
	Noname  bool   `json:"noname"`
}

func divinationPublicHandle(c *server.StupidContext) {
	log := c.Log.WithField("func", "divination.divinationPublicHandle")

	httpRsp := pb.HTTPResponse{
		Result: proto.Int32(int32(gconst.UnknownError)),
	}
	defer c.WriteJSONRsp(&httpRsp)

	// req
	req := &divinationPublicReq{}
	if err := json.Unmarshal(c.Body, req); err != nil {
		httpRsp.Result = proto.Int32(int32(gconst.ErrParse))
		httpRsp.Msg = proto.String("请求信息解析失败")
		log.Errorf("code:%d msg:%s json Unmarshal err:%s", httpRsp.GetResult(), httpRsp.GetMsg(), err.Error())
		return
	}

	// log.Info("divinationPublicHandle enter, req:", string(c.Body))

	conn := c.RedisConn
	playerid := c.UserID
	nowtime := time.Now()
	nowdata := nowtime.Format("2006-01-02")

	// 字数判断（去除标点符号和空格）
	content := contentFilter(req.Content)
	if len(content) < gconst.MinDivinationLen {
		httpRsp.Result = proto.Int32(int32(gconst.ErrContentLenNotEnough))
		httpRsp.Msg = proto.String("有效字符不足")
		log.Errorf("code:%d msg:%s content len not enough, content:%s", httpRsp.GetResult(), httpRsp.GetMsg(), content)
		return
	}

	// 关键字过滤
	issen, _ := gfunc.ReplaceSensitiveWord(req.Content)
	if issen {
		httpRsp.Result = proto.Int32(int32(gconst.ErrContentSensitive))
		httpRsp.Msg = proto.String("吐槽包含敏感词")
		log.Errorf("code:%d msg:%s content is sensitive", httpRsp.GetResult(), httpRsp.GetMsg())
		return
	}

	// redis multi get
	conn.Send("MULTI")
	conn.Send("GET", rconst.StringDivinationID)
	redisMDArray, err := redis.Values(conn.Do("EXEC"))
	if err != nil {
		httpRsp.Result = proto.Int32(int32(gconst.ErrRedis))
		httpRsp.Msg = proto.String("统一获取缓存操作失败")
		log.Errorf("code:%d msg:%s redisMDArray Values err, err:%s", httpRsp.GetResult(), httpRsp.GetMsg(), err.Error())
		return
	}

	// do something
	divinationid, _ := redis.Int64(redisMDArray[0], nil)
	divinationid++

	// redis multi set
	conn.Send("MULTI")
	data := rconst.Divination{
		PlayerID:     playerid,
		DivinationID: divinationid,
		Time:         nowtime.Unix(),
		Content:      req.Content,
		Noname:       req.Noname,
	}
	databyte, err := json.Marshal(data)
	if err != nil {
		httpRsp.Result = proto.Int32(int32(gconst.ErrParse))
		httpRsp.Msg = proto.String("吐槽解析错误")
		log.Errorf("code:%d msg:%s marshal err, err:%s", httpRsp.GetResult(), httpRsp.GetMsg(), err.Error())
		return
	}
	conn.Send("HSET", rconst.HashDivinationPrefix+nowdata, divinationid, databyte)
	conn.Send("ZADD", rconst.ZSetDivinationRecordPrefix+nowdata, nowtime.Unix(), divinationid)
	conn.Send("SETEX", rconst.StringDivinationID, gfunc.TomorrowZeroRemain(), divinationid)
	conn.Send("ZINCRBY", rconst.ZSetDivinationRank, 1, playerid)
	conn.Send("EXPIRE", rconst.ZSetDivinationRank, gfunc.TomorrowZeroRemain())
	_, err = conn.Do("EXEC")
	if err != nil {
		httpRsp.Result = proto.Int32(int32(gconst.ErrRedis))
		httpRsp.Msg = proto.String("统一存储缓存操作失败")
		log.Errorf("code:%d msg:%s exec err, err:%s", httpRsp.GetResult(), httpRsp.GetMsg(), err.Error())
		return
	}

	httpRsp.Result = proto.Int32(int32(gconst.Success))

	log.Info("divinationPublicHandle rsp, result:", httpRsp.GetResult())

	return
}
