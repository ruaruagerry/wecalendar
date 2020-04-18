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

type realGetRsp struct {
	RealNick string `json:"realnick"`
	CardCode string `json:"cardcode"`
}

func realGetHandle(c *server.StupidContext) {
	log := c.Log.WithField("func", "setup.realGetHandle")

	httpRsp := pb.HTTPResponse{
		Result: proto.Int32(int32(gconst.UnknownError)),
	}
	defer c.WriteJSONRsp(&httpRsp)

	log.Info("realGetHandle enter")

	conn := c.RedisConn
	playerid := c.UserID

	// redis multi get
	conn.Send("MULTI")
	conn.Send("HGET", rconst.HashSetupPrefix+playerid, rconst.FieldSetupRealNick)
	conn.Send("HGET", rconst.HashSetupPrefix+playerid, rconst.FieldSetupCardcode)
	redisMDArray, err := redis.Values(conn.Do("EXEC"))
	if err != nil {
		httpRsp.Result = proto.Int32(int32(gconst.ErrRedis))
		httpRsp.Msg = proto.String("统一获取缓存操作失败")
		log.Errorf("code:%d msg:%s redisMDArray Values err, err:%s", httpRsp.GetResult(), httpRsp.GetMsg(), err.Error())
		return
	}

	realnick, _ := redis.String(redisMDArray[0], nil)
	cardcode, _ := redis.String(redisMDArray[1], nil)

	// rsp
	rsp := &realGetRsp{
		RealNick: realnick,
		CardCode: cardcode,
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

	log.Info("realGetHandle rsp, rsp:", string(data))

	return
}
