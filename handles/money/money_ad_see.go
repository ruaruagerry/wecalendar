package money

import (
	"time"
	"weagent/gconst"
	"weagent/gfunc"
	"weagent/pb"
	"weagent/rconst"
	"weagent/server"
	"weagent/tables"

	"github.com/garyburd/redigo/redis"
	"github.com/golang/protobuf/proto"
)

func adSeeHandle(c *server.StupidContext) {
	log := c.Log.WithField("func", "money.adSeeHandle")

	httpRsp := pb.HTTPResponse{
		Result: proto.Int32(int32(gconst.UnknownError)),
	}
	defer c.WriteJSONRsp(&httpRsp)

	log.Info("adSeeHandle enter")

	conn := c.RedisConn
	db := c.DbConn
	playerid := c.UserID
	timenow := time.Now()
	nowkey := timenow.Format("2006-01-02")

	// 检查
	conn.Send("MULTI")
	conn.Send("SETNX", rconst.StringLockMoneyAdSeePrefix+playerid, "1")
	conn.Send("EXPIRE", rconst.StringLockMoneyAdSeePrefix+playerid, gconst.LockTime)
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
		conn.Do("DEL", rconst.StringLockMoneyAdSeePrefix+playerid)
	}()

	// redis multi get
	conn.Send("MULTI")
	conn.Send("HGET", rconst.HashMoneyConfig, rconst.FieldMoneyConfigMaxAdNum)
	conn.Send("HGET", rconst.HashMoneyConfig, rconst.FieldMoneyConfigSeeEarnings)
	conn.Send("HGET", rconst.HashMoneyAdSeeNum, playerid)
	conn.Send("HGET", rconst.HashMoneyPrefix+playerid, rconst.FieldMoneyNum)
	conn.Send("HGET", rconst.HashMoneyPrefix+playerid, rconst.FieldMoneyTotal)
	conn.Send("EXISTS", rconst.StringMoneyRemainSessNumPrefix+playerid)
	conn.Send("GET", rconst.StringMoneyRemainSessNumPrefix+playerid)
	conn.Send("HGET", rconst.HashAccountPrefix+playerid, rconst.FieldAccName)
	conn.Send("GET", rconst.StringDataDayEarningsPrefix+nowkey)
	conn.Send("GET", rconst.StringDataDayAdNumPrefix+nowkey)
	redisMDArray, err = redis.Values(conn.Do("EXEC"))
	if err != nil {
		httpRsp.Result = proto.Int32(int32(gconst.ErrRedis))
		httpRsp.Msg = proto.String("统一获取缓存操作失败")
		log.Errorf("code:%d msg:%s redisMDArray Values err, err:%s", httpRsp.GetResult(), httpRsp.GetMsg(), err.Error())
		return
	}

	maxadnum, _ := redis.Int(redisMDArray[0], nil)
	seeearnings, _ := redis.Int(redisMDArray[1], nil)
	adseenum, _ := redis.Int(redisMDArray[2], nil)
	moneynum, _ := redis.Int(redisMDArray[3], nil)
	totalnum, _ := redis.Int(redisMDArray[4], nil)
	existremain, _ := redis.Bool(redisMDArray[5], nil)
	remainseenum, _ := redis.Int(redisMDArray[6], nil)
	name, _ := redis.String(redisMDArray[7], nil)
	todayall, _ := redis.Int(redisMDArray[8], nil)
	todayadnum, _ := redis.Int(redisMDArray[9], nil)

	// do something
	if !existremain {
		remainseenum = maxadnum
	}

	// 超出最大广告收益次数，不做收益计算
	todayadnum++
	adseenum++
	if remainseenum > 0 {
		remainseenum--
		moneynum += seeearnings
		totalnum += seeearnings
		todayall += seeearnings

		// 插入收益记录
		go func() {
			adrecord := &tables.Adrecord{
				ID:         playerid,
				Earnings:   int64(seeearnings),
				Name:       name,
				AdMoney:    int64(moneynum),
				CreateTime: timenow,
			}
			_, err = db.Insert(adrecord)
			if err != nil {
				httpRsp.Result = proto.Int32(int32(gconst.ErrDB))
				httpRsp.Msg = proto.String("收益记录插入失败")
				log.Errorf("code:%d msg:%s adrecord insert err, err:%s", httpRsp.GetResult(), httpRsp.GetMsg(), err.Error())
				return
			}
		}()
	}

	// redis multi set
	conn.Send("MULTI")
	conn.Send("HSET", rconst.HashMoneyAdSeeNum, playerid, adseenum)
	conn.Send("HSET", rconst.HashMoneyPrefix+playerid, rconst.FieldMoneyNum, moneynum)
	conn.Send("HSET", rconst.HashMoneyPrefix+playerid, rconst.FieldMoneyTotal, totalnum)
	conn.Send("SETEX", rconst.StringMoneyRemainSessNumPrefix+playerid, gfunc.TomorrowZeroRemain(), remainseenum)
	conn.Send("SETEX", rconst.StringDataDayEarningsPrefix+nowkey, 2*24*3600, todayall)
	conn.Send("SETEX", rconst.StringDataDayAdNumPrefix+nowkey, 2*24*3600, todayadnum)
	_, err = conn.Do("EXEC")
	if err != nil {
		httpRsp.Result = proto.Int32(int32(gconst.ErrRedis))
		httpRsp.Msg = proto.String("统一存储缓存操作失败")
		log.Errorf("code:%d msg:%s exec err, err:%s", httpRsp.GetResult(), httpRsp.GetMsg(), err.Error())
		return
	}

	// rsp
	httpRsp.Result = proto.Int32(int32(gconst.Success))

	log.Info("adSeeHandle rsp, result:", httpRsp.GetResult())

	return
}
