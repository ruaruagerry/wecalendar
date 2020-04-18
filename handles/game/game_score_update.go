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

type scoreUpdateReq struct {
	Score int32 `json:"score"`
}

type scoreUpdateRsp struct {
	Num int32 `json:"num"`
}

func scoreUpdateHandle(c *server.StupidContext) {
	log := c.Log.WithField("func", "game.scoreUpdateHandle")

	httpRsp := pb.HTTPResponse{
		Result: proto.Int32(int32(gconst.UnknownError)),
	}
	defer c.WriteJSONRsp(&httpRsp)

	// req
	req := &scoreUpdateReq{}
	if err := json.Unmarshal(c.Body, req); err != nil {
		httpRsp.Result = proto.Int32(int32(gconst.ErrParse))
		httpRsp.Msg = proto.String("请求信息解析失败")
		log.Errorf("code:%d msg:%s json Unmarshal err:%s", httpRsp.GetResult(), httpRsp.GetMsg(), err.Error())
		return
	}

	log.Info("scoreUpdateHandle enter, req:", string(c.Body))

	conn := c.RedisConn
	playerid := c.UserID

	// 检查
	conn.Send("MULTI")
	conn.Send("SETNX", rconst.StringLockGameScoreUpdatePrefix+playerid, "1")
	conn.Send("EXPIRE", rconst.StringLockGameScoreUpdatePrefix+playerid, gconst.LockTime)
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
		conn.Do("DEL", rconst.StringLockGameScoreUpdatePrefix+playerid)
	}()

	// redis multi get
	conn.Send("MULTI")
	conn.Send("ZSCORE", rconst.ZSetGameRank, playerid)
	conn.Send("GET", rconst.StringGameRebirthNumPrefix+playerid)
	redisMDArray, err = redis.Values(conn.Do("EXEC"))
	if err != nil {
		httpRsp.Result = proto.Int32(int32(gconst.ErrRedis))
		httpRsp.Msg = proto.String("统一获取缓存操作失败")
		log.Errorf("code:%d msg:%s redisMDArray Values err, err:%s", httpRsp.GetResult(), httpRsp.GetMsg(), err.Error())
		return
	}

	curscore, _ := redis.Int(redisMDArray[0], nil)
	rebirthnum, err := redis.Int(redisMDArray[1], nil)

	// do something
	isupdaterebirth := false
	if err != nil {
		isupdaterebirth = true
		// 初次生成
		rebirthnum = gconst.GameRebirthNumConfig
	}

	isupdatescore := false
	if req.Score > int32(curscore) {
		isupdatescore = true
	}

	// redis multi set
	conn.Send("MULTI")
	if isupdaterebirth {
		conn.Send("SETEX", rconst.StringGameRebirthNumPrefix+playerid, gfunc.TomorrowZeroRemain(), rebirthnum)
	}
	if isupdatescore {
		conn.Send("ZADD", rconst.ZSetGameRank, req.Score, playerid)
	}
	_, err = conn.Do("EXEC")
	if err != nil {
		httpRsp.Result = proto.Int32(int32(gconst.ErrRedis))
		httpRsp.Msg = proto.String("统一存储缓存操作失败")
		log.Errorf("code:%d msg:%s exec err, err:%s", httpRsp.GetResult(), httpRsp.GetMsg(), err.Error())
		return
	}

	// rsp
	rsp := &scoreUpdateRsp{
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

	log.Info("scoreUpdateHandle rsp, rsp:", string(data))

	return
}
