package money

import (
	"encoding/json"
	"weagent/gconst"
	"weagent/gfunc"
	"weagent/pb"
	"weagent/rconst"
	"weagent/server"

	"github.com/garyburd/redigo/redis"
	"github.com/golang/protobuf/proto"
)

type entranceRsp struct {
	Total       float32 `json:"total"`       // 总收益
	Money       float32 `json:"money"`       // 当前账户余额
	GetoutTotal float32 `json:"getouttotal"` // 总提现金额
	RemainSee   int32   `json:"remainsee"`   // 剩余观看广告次数
}

func entranceHandle(c *server.StupidContext) {
	log := c.Log.WithField("func", "money.entranceHandle")

	httpRsp := pb.HTTPResponse{
		Result: proto.Int32(int32(gconst.UnknownError)),
	}
	defer c.WriteJSONRsp(&httpRsp)

	log.Info("entranceHandle enter")

	conn := c.RedisConn
	playerid := c.UserID

	// redis multi get
	conn.Send("MULTI")
	conn.Send("HGET", rconst.HashMoneyPrefix+playerid, rconst.FieldMoneyTotal)
	conn.Send("HGET", rconst.HashMoneyPrefix+playerid, rconst.FieldMoneyNum)
	conn.Send("HGET", rconst.HashMoneyPrefix+playerid, rconst.FieldMoneyGetout)
	conn.Send("EXISTS", rconst.StringMoneyRemainSessNumPrefix+playerid)
	conn.Send("GET", rconst.StringMoneyRemainSessNumPrefix+playerid)
	conn.Send("HGET", rconst.HashMoneyConfig, rconst.FieldMoneyConfigMaxAdNum)
	redisMDArray, err := redis.Values(conn.Do("EXEC"))
	if err != nil {
		httpRsp.Result = proto.Int32(int32(gconst.ErrRedis))
		httpRsp.Msg = proto.String("统一获取缓存操作失败")
		log.Errorf("code:%d msg:%s redisMDArray Values err, err:%s", httpRsp.GetResult(), httpRsp.GetMsg(), err.Error())
		return
	}

	total, _ := redis.Int(redisMDArray[0], nil)
	money, _ := redis.Int(redisMDArray[1], nil)
	getouttotal, _ := redis.Int(redisMDArray[2], nil)
	existremain, _ := redis.Bool(redisMDArray[3], nil)
	remainseenum, _ := redis.Int(redisMDArray[4], nil)
	adseenum, _ := redis.Int(redisMDArray[5], nil)

	// redis multi set
	conn.Send("MULTI")
	if !existremain {
		remainseenum = adseenum
		conn.Send("SETEX", rconst.StringMoneyRemainSessNumPrefix+playerid, gfunc.TomorrowZeroRemain(), remainseenum)
	}
	_, err = conn.Do("EXEC")
	if err != nil {
		httpRsp.Result = proto.Int32(int32(gconst.ErrRedis))
		httpRsp.Msg = proto.String("统一存储缓存操作失败")
		log.Errorf("code:%d msg:%s exec err, err:%s", httpRsp.GetResult(), httpRsp.GetMsg(), err.Error())
		return
	}

	// rsp
	rsp := &entranceRsp{
		Total:       float32(total) / float32(100),
		Money:       float32(money) / float32(100),
		GetoutTotal: float32(getouttotal) / float32(100),
		RemainSee:   int32(remainseenum),
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
