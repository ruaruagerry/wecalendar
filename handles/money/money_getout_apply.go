package money

import (
	"encoding/json"
	"time"
	"weagent/gconst"
	"weagent/pb"
	"weagent/rconst"
	"weagent/server"
	"weagent/tables"

	"github.com/garyburd/redigo/redis"
	"github.com/golang/protobuf/proto"
)

type getoutApplyReq struct {
	GetoutMoney int64 `json:"getoutmoney"` // 取的钱，单位元
}

func getoutApplyHandle(c *server.StupidContext) {
	log := c.Log.WithField("func", "money.getoutApplyHandle")

	httpRsp := pb.HTTPResponse{
		Result: proto.Int32(int32(gconst.UnknownError)),
	}
	defer c.WriteJSONRsp(&httpRsp)

	// req
	req := &getoutApplyReq{}
	if err := json.Unmarshal(c.Body, req); err != nil {
		httpRsp.Result = proto.Int32(int32(gconst.ErrParse))
		httpRsp.Msg = proto.String("请求信息解析失败")
		log.Errorf("code:%d msg:%s json Unmarshal err:%s", httpRsp.GetResult(), httpRsp.GetMsg(), err.Error())
		return
	}

	log.Info("getoutApplyHandle enter, req:", string(c.Body))

	// 查看提现金额是否合法
	_, has := getoutMoneyConfig[req.GetoutMoney]
	if !has {
		httpRsp.Result = proto.Int32(int32(gconst.ErrMoneyInvalidGetout))
		httpRsp.Msg = proto.String("提现金额错误")
		log.Errorf("code:%d msg:%s getoutmoney is invalid, getoutmoney:%d", httpRsp.GetResult(), httpRsp.GetMsg())
		return
	}
	// 元转成分
	getoutmoney := req.GetoutMoney * 100

	db := c.DbConn
	conn := c.RedisConn
	playerid := c.UserID
	nowtime := time.Now()

	// 检查
	conn.Send("MULTI")
	conn.Send("SETNX", rconst.StringLockMoneyGetoutApplyPrefix+playerid, "1")
	conn.Send("EXPIRE", rconst.StringLockMoneyGetoutApplyPrefix+playerid, gconst.LockTime)
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
		conn.Do("DEL", rconst.StringLockMoneyGetoutApplyPrefix+playerid)
	}()

	// redis multi get
	conn.Send("MULTI")
	conn.Send("HGET", rconst.HashAccountPrefix+playerid, rconst.FieldAccName)
	conn.Send("HGET", rconst.HashMoneyPrefix+playerid, rconst.FieldMoneyNum)
	redisMDArray, err = redis.Values(conn.Do("EXEC"))
	if err != nil {
		httpRsp.Result = proto.Int32(int32(gconst.ErrRedis))
		httpRsp.Msg = proto.String("统一获取缓存操作失败")
		log.Errorf("code:%d msg:%s redisMDArray Values err, err:%s", httpRsp.GetResult(), httpRsp.GetMsg(), err.Error())
		return
	}

	name, _ := redis.String(redisMDArray[0], nil)
	moneynum, _ := redis.Int64(redisMDArray[1], nil)
	if moneynum < getoutmoney {
		httpRsp.Result = proto.Int32(int32(gconst.ErrMoneyNotEnough))
		httpRsp.Msg = proto.String("提现金额不足")
		log.Errorf("code:%d msg:%s money not enough, moneynum:%d getout:%d", httpRsp.GetResult(), httpRsp.GetMsg(), moneynum, getoutmoney)
		return
	}

	moneynum -= getoutmoney

	// 插入提现记录
	getoutrecord := &tables.Getoutrecord{
		ID:          playerid,
		GetoutMoney: getoutmoney,
		CreateTime:  nowtime,
		Status:      tables.GetoutStatusReview,
		Name:        name,
	}
	_, err = db.Insert(getoutrecord)
	if err != nil {
		httpRsp.Result = proto.Int32(int32(gconst.ErrDB))
		httpRsp.Msg = proto.String("提现记录插入失败")
		log.Errorf("code:%d msg:%s getoutrecord insert err, err:%s", httpRsp.GetResult(), httpRsp.GetMsg(), err.Error())
		return
	}

	// redis multi set
	conn.Send("MULTI")
	conn.Send("HSET", rconst.HashMoneyPrefix+playerid, rconst.FieldMoneyNum, moneynum)
	_, err = conn.Do("EXEC")
	if err != nil {
		httpRsp.Result = proto.Int32(int32(gconst.ErrRedis))
		httpRsp.Msg = proto.String("统一存储缓存操作失败")
		log.Errorf("code:%d msg:%s exec err, err:%s", httpRsp.GetResult(), httpRsp.GetMsg(), err.Error())
		return
	}

	// rsp
	httpRsp.Result = proto.Int32(int32(gconst.Success))

	log.Info("getoutApplyHandle rsp, result:", httpRsp.GetResult())

	return
}
