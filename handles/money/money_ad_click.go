package money

import (
	"weagent/gconst"
	"weagent/pb"
	"weagent/rconst"
	"weagent/server"

	"github.com/garyburd/redigo/redis"
	"github.com/golang/protobuf/proto"
)

func adClickHandle(c *server.StupidContext) {
	log := c.Log.WithField("func", "money.adClickHandle")

	httpRsp := pb.HTTPResponse{
		Result: proto.Int32(int32(gconst.UnknownError)),
	}
	defer c.WriteJSONRsp(&httpRsp)

	log.Info("adClickHandle enter")

	conn := c.RedisConn
	playerid := c.UserID

	// 检查
	conn.Send("MULTI")
	conn.Send("SETNX", rconst.StringLockMoneyAdClickPrefix+playerid, "1")
	conn.Send("EXPIRE", rconst.StringLockMoneyAdClickPrefix+playerid, gconst.LockTime)
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
		conn.Do("DEL", rconst.StringLockMoneyAdClickPrefix+playerid)
	}()

	// redis multi set
	conn.Send("MULTI")
	conn.Send("HINCRBY", rconst.HashMoneyAdClickNum, playerid, 1)
	_, err = conn.Do("EXEC")
	if err != nil {
		httpRsp.Result = proto.Int32(int32(gconst.ErrRedis))
		httpRsp.Msg = proto.String("统一存储缓存操作失败")
		log.Errorf("code:%d msg:%s exec err, err:%s", httpRsp.GetResult(), httpRsp.GetMsg(), err.Error())
		return
	}

	// rsp
	httpRsp.Result = proto.Int32(int32(gconst.Success))

	log.Info("adClickHandle rsp, result:", httpRsp.GetResult())

	return
}
