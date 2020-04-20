package divination

import (
	"encoding/json"
	"time"
	"wecalendar/gconst"
	"wecalendar/pb"
	"wecalendar/rconst"
	"wecalendar/server"

	"github.com/golang/protobuf/proto"
	"github.com/gomodule/redigo/redis"
)

type divinationGetRsp struct {
	Content      string `json:"content"`
	PlayerID     string `json:"playerid"`
	DivinationID int64  `json:"divinationid"`
	NickName     string `json:"nickname"`
	Portrait     string `json:"portrait"`
	Time         string `json:"time"`
}

func divinationGetHandle(c *server.StupidContext) {
	log := c.Log.WithField("func", "divination.divinationGetHandle")

	httpRsp := pb.HTTPResponse{
		Result: proto.Int32(int32(gconst.UnknownError)),
	}
	defer c.WriteJSONRsp(&httpRsp)

	log.Info("helloHandle enter")

	conn := c.RedisConn
	nowtime := time.Now()
	nowdata := nowtime.Format("2006-01-02")

	// redis multi get
	conn.Send("MULTI")
	conn.Send("ZRANGE", rconst.ZSetDivinationRecordPrefix+nowdata, 0, -1)
	redisMDArray, err := redis.Values(conn.Do("EXEC"))
	if err != nil {
		httpRsp.Result = proto.Int32(int32(gconst.ErrRedis))
		httpRsp.Msg = proto.String("统一获取缓存操作失败")
		log.Errorf("code:%d msg:%s redisMDArray Values err, err:%s", httpRsp.GetResult(), httpRsp.GetMsg(), err.Error())
		return
	}

	alldivinationum, _ := redis.Ints(redisMDArray[0], nil)
	if len(alldivinationum) == 0 {
		httpRsp.Result = proto.Int32(int32(gconst.ErrNoDivination))
		httpRsp.Msg = proto.String("当日没有吐槽")
		log.Errorf("code:%d msg:%s not divination", httpRsp.GetResult(), httpRsp.GetMsg())
		return
	}

	divinationid := droprand.Int31n(int32(len(alldivinationum)))

	// do something
	// 获取吐槽信息
	conn.Send("MULTI")
	conn.Send("HGET", rconst.HashDivinationPrefix+nowdata, divinationid)
	redisMDArray, err = redis.Values(conn.Do("EXEC"))
	if err != nil {
		httpRsp.Result = proto.Int32(int32(gconst.ErrRedis))
		httpRsp.Msg = proto.String("吐槽信息统一获取缓存操作失败")
		log.Errorf("code:%d msg:%s redisMDArray Values err, err:%s", httpRsp.GetResult(), httpRsp.GetMsg(), err.Error())
		return
	}

	divinationbyte, _ := redis.Bytes(redisMDArray[0], nil)

	divination := &rconst.Divination{}
	err = json.Unmarshal(divinationbyte, divination)
	if err != nil {
		httpRsp.Result = proto.Int32(int32(gconst.ErrParse))
		httpRsp.Msg = proto.String("吐槽解析失败")
		log.Errorf("code:%d msg:%s databyte unmarshal err, err:%s", httpRsp.GetResult(), httpRsp.GetMsg(), err.Error())
		return
	}

	// 获取玩家信息
	conn.Send("MULTI")
	conn.Send("HGET", rconst.HashAccountPrefix+divination.PlayerID, rconst.FieldAccName)
	conn.Send("HGET", rconst.HashAccountPrefix+divination.PlayerID, rconst.FieldAccImage)
	redisMDArray, err = redis.Values(conn.Do("EXEC"))
	if err != nil {
		httpRsp.Result = proto.Int32(int32(gconst.ErrRedis))
		httpRsp.Msg = proto.String("玩家信息统一获取缓存操作失败")
		log.Errorf("code:%d msg:%s redisMDArray Values err, err:%s", httpRsp.GetResult(), httpRsp.GetMsg(), err.Error())
		return
	}

	nickname, _ := redis.String(redisMDArray[0], nil)
	portrait, _ := redis.String(redisMDArray[1], nil)

	// rsp
	rsp := &divinationGetRsp{
		Content:      divination.Content,
		PlayerID:     divination.PlayerID,
		DivinationID: divination.DivinationID,
		NickName:     nickname,
		Portrait:     portrait,
		Time:         time.Unix(divination.Time, 0).Format("2006-01-02 15:04:05"),
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

	log.Info("divinationGetHandle rsp, rsp:", string(data))

	return
}
