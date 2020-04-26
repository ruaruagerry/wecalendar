package divination

import (
	"encoding/json"
	"strconv"
	"wecalendar/gconst"
	"wecalendar/pb"
	"wecalendar/rconst"
	"wecalendar/server"

	"github.com/golang/protobuf/proto"
	"github.com/gomodule/redigo/redis"
)

type divinationRankItem struct {
	Nickname string `json:"nickname"`
	Portrait string `json:"portrait"`
	Num      int32  `json:"num"`
	Rank     int32  `json:"rank"`
}

type divinationRankRsp struct {
	Ranks []*divinationRankItem `json:"ranks"`
}

// 只拉取前四名
func divinationRankHandle(c *server.StupidContext) {
	log := c.Log.WithField("func", "divination.divinationRankHandle")

	httpRsp := pb.HTTPResponse{
		Result: proto.Int32(int32(gconst.UnknownError)),
	}
	defer c.WriteJSONRsp(&httpRsp)

	log.Info("divinationRankHandle enter")

	conn := c.RedisConn
	// playerid := c.UserID

	// redis multi get
	conn.Send("MULTI")
	conn.Send("HGET", rconst.HashDivinationConfig, rconst.FieldDivinationFirst)
	conn.Send("ZRANGE", rconst.ZSetDivinationRank, 0, 4, "withscores")
	redisMDArray, err := redis.Values(conn.Do("EXEC"))
	if err != nil {
		httpRsp.Result = proto.Int32(int32(gconst.ErrRedis))
		httpRsp.Msg = proto.String("统一获取缓存操作失败")
		log.Errorf("code:%d msg:%s redisMDArray Values err, err:%s", httpRsp.GetResult(), httpRsp.GetMsg(), err.Error())
		return
	}

	first, _ := redis.Bool(redisMDArray[0], nil)
	rankstrs, _ := redis.Strings(redisMDArray[1], nil)

	// do something
	playerids := []string{}
	nums := []int32{}
	for i := range rankstrs {
		if i%2 == 0 {
			playerids = append(playerids, rankstrs[i])
		} else {
			tmpint, err := strconv.Atoi(rankstrs[i])
			if err != nil {
				httpRsp.Result = proto.Int32(int32(gconst.ErrParse))
				httpRsp.Msg = proto.String("吐槽数解析失败")
				log.Errorf("code:%d msg:%s Atoi err, err:%s", httpRsp.GetResult(), httpRsp.GetMsg(), err.Error())
				return
			}

			nums = append(nums, int32(tmpint))
		}
	}

	// 获取玩家数据
	conn.Send("MULTI")
	for _, v := range playerids {
		conn.Send("HGET", rconst.HashAccountPrefix+v, rconst.FieldAccName)
		conn.Send("HGET", rconst.HashAccountPrefix+v, rconst.FieldAccImage)
	}
	redisMDArray, err = redis.Values(conn.Do("EXEC"))
	if err != nil {
		httpRsp.Result = proto.Int32(int32(gconst.ErrRedis))
		httpRsp.Msg = proto.String("统一获取缓存操作失败")
		log.Errorf("code:%d msg:%s redisMDArray Values err, err:%s", httpRsp.GetResult(), httpRsp.GetMsg(), err.Error())
		return
	}

	rsp := &divinationRankRsp{}

	if first {
		num := int32(0)
		if len(nums) > 0 {
			num = nums[0] + droprand.Int31n(10)
		}

		faker := &divinationRankItem{
			Nickname: "夜",
			Portrait: "https://ss0.bdstatic.com/94oJfD_bAAcT8t7mm9GUKT-xh_/timg?image&quality=100&size=b4000_4000&sec=1587483648&di=05526f3c2d17061f3d9bd42fd9afc39a&src=http://pic2.zhimg.com/50/v2-98012e57831cd600529dc58677030eba_hd.jpg",
			Num:      num,
			Rank:     1,
		}

		rsp.Ranks = append(rsp.Ranks, faker)
	}

	for i := range playerids {
		nickname, _ := redis.String(redisMDArray[2*i], nil)
		portrait, _ := redis.String(redisMDArray[2*i+1], nil)
		rank := int32(i + 1)
		if first {
			rank = int32(i + 2)
		}

		tmprank := &divinationRankItem{
			Nickname: nickname,
			Portrait: portrait,
			Num:      nums[i],
			Rank:     rank,
		}

		if len(rsp.Ranks) < 5 {
			rsp.Ranks = append(rsp.Ranks, tmprank)
		}
	}

	// rsp
	data, err := json.Marshal(rsp)
	if err != nil {
		httpRsp.Result = proto.Int32(int32(gconst.ErrParse))
		httpRsp.Msg = proto.String("返回信息marshal解析失败")
		log.Errorf("code:%d msg:%s json marshal err, err:%s", httpRsp.GetResult(), httpRsp.GetMsg(), err.Error())
		return
	}
	httpRsp.Result = proto.Int32(int32(gconst.Success))
	httpRsp.Data = data

	// log.Info("divinationRankHandle rsp, rsp:", string(data))

	return
}
