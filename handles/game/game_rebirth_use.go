package game

import (
	"encoding/json"
	"weagent/gconst"
	"weagent/gfunc"
	"weagent/pb"
	"weagent/rconst"
	"weagent/server"

	"github.com/golang/protobuf/proto"
	"github.com/gomodule/redigo/redis"
)

type rebirthUseRsp struct {
	Num int32 `json:"num"` // 剩余复活次数
}

func rebirthUseHandle(c *server.StupidContext) {
	log := c.Log.WithField("func", "game.rebirthUseHandle")

	httpRsp := pb.HTTPResponse{
		Result: proto.Int32(int32(gconst.UnknownError)),
	}
	defer c.WriteJSONRsp(&httpRsp)

	log.Info("rebirthUseHandle enter")

	conn := c.RedisConn
	playerid := c.UserID

	// 检查
	conn.Send("MULTI")
	conn.Send("SETNX", rconst.StringLockGameRebirthUsePrefix+playerid, "1")
	conn.Send("EXPIRE", rconst.StringLockGameRebirthUsePrefix+playerid, gconst.LockTime)
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
		conn.Do("DEL", rconst.StringLockGameRebirthUsePrefix+playerid)
	}()

	// redis multi get
	conn.Send("MULTI")
	conn.Send("GET", rconst.StringGameRebirthNumPrefix+playerid)
	redisMDArray, err = redis.Values(conn.Do("EXEC"))
	if err != nil {
		httpRsp.Result = proto.Int32(int32(gconst.ErrRedis))
		httpRsp.Msg = proto.String("统一获取缓存操作失败")
		log.Errorf("code:%d msg:%s redisMDArray Values err, err:%s", httpRsp.GetResult(), httpRsp.GetMsg(), err.Error())
		return
	}

	rebirthnum, _ := redis.Int(redisMDArray[0], nil)

	// do something
	rebirthnum--
	if rebirthnum < 0 {
		httpRsp.Result = proto.Int32(int32(gconst.ErrRebirtNumNotEnough))
		httpRsp.Msg = proto.String("复活次数不足")
		log.Errorf("code:%d msg:%s rebirth num not enough, rebirth:%d", httpRsp.GetResult(), httpRsp.GetMsg(), rebirthnum)
		return
	}

	// redis multi set
	conn.Send("MULTI")
	conn.Send("SETEX", rconst.StringGameRebirthNumPrefix+playerid, gfunc.TomorrowZeroRemain(), rebirthnum)
	_, err = conn.Do("EXEC")
	if err != nil {
		httpRsp.Result = proto.Int32(int32(gconst.ErrRedis))
		httpRsp.Msg = proto.String("统一存储缓存操作失败")
		log.Errorf("code:%d msg:%s exec err, err:%s", httpRsp.GetResult(), httpRsp.GetMsg(), err.Error())
		return
	}

	// rsp
	rsp := &rebirthUseRsp{
		Num: int32(rebirthnum),
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

	log.Info("rebirthUseHandle rsp, rsp:", string(data))

	return
}
