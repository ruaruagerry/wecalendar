package game

import (
	"encoding/json"
	"fmt"
	"weagent/gconst"
	"weagent/pb"
	"weagent/rconst"
	"weagent/server"

	"github.com/golang/protobuf/proto"
	"github.com/gomodule/redigo/redis"
)

type playerInfo struct {
	ID       string `json:"id"`
	Nick     string `json:"nick"`
	Portrait string `json:"portrait"`
	Score    int32  `json:"score"`
}

type scoreRankRsp struct {
	MyRank  int32         `json:"myrank"`  // 我的排名
	MyScore int32         `json:"score"`   // 我的分数
	Players []*playerInfo `json:"players"` // 前十的排名信息
}

// 只显示前十名信息
func scoreRankHandle(c *server.StupidContext) {
	log := c.Log.WithField("func", "game.scoreRankHandle")

	httpRsp := pb.HTTPResponse{
		Result: proto.Int32(int32(gconst.UnknownError)),
	}
	defer c.WriteJSONRsp(&httpRsp)

	log.Info("scoreRankHandle enter, req:", string(c.Body))

	conn := c.RedisConn
	playerid := c.UserID

	// redis multi get
	conn.Send("MULTI")
	conn.Send("ZREVRANGE", rconst.ZSetGameRank, 0, gconst.RankMaxIndexConfig, "withscores")
	conn.Send("ZREVRANK", rconst.ZSetGameRank, playerid)
	conn.Send("ZSCORE", rconst.ZSetGameRank, playerid)
	redisMDArray, err := redis.Values(conn.Do("EXEC"))
	if err != nil {
		httpRsp.Result = proto.Int32(int32(gconst.ErrRedis))
		httpRsp.Msg = proto.String("统一获取缓存操作失败")
		log.Errorf("code:%d msg:%s redisMDArray Values err, err:%s", httpRsp.GetResult(), httpRsp.GetMsg(), err.Error())
		return
	}

	// do something
	rankinfo, _ := redis.Ints(redisMDArray[0], nil)
	myrank, err := redis.Int(redisMDArray[1], nil)
	if err != nil {
		myrank = -1
	}
	myscore, _ := redis.Int(redisMDArray[2], nil)

	// 获取玩家信息
	players := []*playerInfo{}
	conn.Send("MULTI")
	for i, v := range rankinfo {
		if i%2 == 0 {
			playeridstr := fmt.Sprintf("%d", v)

			tmpplayer := &playerInfo{
				ID: playeridstr,
			}
			players = append(players, tmpplayer)

			conn.Send("HGET", rconst.HashAccountPrefix+playeridstr, rconst.FieldAccName)
			conn.Send("HGET", rconst.HashAccountPrefix+playeridstr, rconst.FieldAccImage)
		} else {
			players[i/2].Score = int32(v)
		}
	}
	redisMDArray, err = redis.Values(conn.Do("EXEC"))
	if err != nil {
		httpRsp.Result = proto.Int32(int32(gconst.ErrRedis))
		httpRsp.Msg = proto.String("获取玩家信息失败")
		log.Errorf("code:%d msg:%s redisMDArray Values err, err:%s", httpRsp.GetResult(), httpRsp.GetMsg(), err.Error())
		return
	}

	for i := range players {
		nick, _ := redis.String(redisMDArray[i*2], nil)
		portrait, _ := redis.String(redisMDArray[i*2+1], nil)

		players[i].Nick = nick
		players[i].Portrait = portrait
	}

	// rsp
	rsp := &scoreRankRsp{
		MyRank:  int32(myrank),
		MyScore: int32(myscore),
		Players: players,
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

	log.Info("scoreRankHandle rsp, rsp:", string(data))

	return
}
