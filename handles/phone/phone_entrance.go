package phone

import (
	"encoding/json"
	"weagent/gconst"
	"weagent/pb"
	"weagent/rconst"
	"weagent/server"

	"github.com/golang/protobuf/proto"
	"github.com/gomodule/redigo/redis"
)

type entranceRsp struct {
	Remain int32  `json:"remain"`
	Phone  string `json:"phone"`
}

func entranceHandle(c *server.StupidContext) {
	log := c.Log.WithField("func", "phone.entranceHandle")

	httpRsp := pb.HTTPResponse{
		Result: proto.Int32(int32(gconst.UnknownError)),
	}
	defer c.WriteJSONRsp(&httpRsp)

	log.Info("entranceHandle enter")

	conn := c.RedisConn
	playerid := c.UserID

	// redis multi get
	conn.Send("MULTI")
	conn.Send("TTL", rconst.StringPhoneGetCodeTagPrefix+playerid)
	conn.Send("HGET", rconst.HashAccountPrefix+playerid, rconst.FieldAccPhone)
	redisMDArray, err := redis.Values(conn.Do("EXEC"))
	if err != nil {
		httpRsp.Result = proto.Int32(int32(gconst.ErrRedis))
		httpRsp.Msg = proto.String("统一获取缓存操作失败")
		log.Errorf("code:%d msg:%s redisMDArray Values err, err:%s", httpRsp.GetResult(), httpRsp.GetMsg(), err.Error())
		return
	}

	remain, _ := redis.Int(redisMDArray[0], nil)
	phone, _ := redis.String(redisMDArray[1], nil)

	// do something
	if remain < 0 {
		remain = 0
	}

	// rsp
	rsp := &entranceRsp{
		Remain: int32(remain),
		Phone:  phone,
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

	log.Info("entranceHandle rsp, rsp:", string(data))

	return
}
