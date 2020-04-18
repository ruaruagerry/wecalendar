/*
 * @Author: your name
 * @Date: 2019-12-27 12:10:07
 * @LastEditTime : 2019-12-27 16:59:55
 * @LastEditors  : Please set LastEditors
 * @Description: In User Settings Edit
 * @FilePath: \weagent\handles\data\data_entrance.go
 */

package data

import (
	"encoding/json"
	"time"
	"weagent/gconst"
	"weagent/pb"
	"weagent/rconst"
	"weagent/server"

	"github.com/garyburd/redigo/redis"
	"github.com/golang/protobuf/proto"
)

type entranceRsp struct {
	YestardayAll   float32 `json:"yestardayall"`   // 昨日全网收益（单位元）
	HistoryAll     float32 `json:"historyall"`     // 历史全网收益（单位元）
	TodayAdNum     int32   `json:"todayadnum"`     // 今日广告总数
	TodayOnlineNum int32   `json:"todayonlinenum"` // 今日在线总数
	TodayAll       float32 `json:"todayall"`       // 今日全网收益（单位元）
}

func entranceHandle(c *server.StupidContext) {
	log := c.Log.WithField("func", "data.entranceHandle")

	httpRsp := pb.HTTPResponse{
		Result: proto.Int32(int32(gconst.UnknownError)),
	}
	defer c.WriteJSONRsp(&httpRsp)

	log.Info("entranceHandle enter, req:", string(c.Body))

	conn := c.RedisConn
	// playerid := c.UserID
	timenow := time.Now()
	timeyesterday := timenow.AddDate(0, 0, -1)
	nowkey := timenow.Format("2006-01-02")
	yestardaykey := timeyesterday.Format("2006-01-02")

	// redis multi get
	conn.Send("MULTI")
	conn.Send("GET", rconst.StringDataDayEarningsPrefix+yestardaykey)
	conn.Send("GET", rconst.StringDataEarnings)
	conn.Send("GET", rconst.StringDataDayAdNumPrefix+nowkey)
	conn.Send("SCARD", rconst.SetUsers)
	conn.Send("GET", rconst.StringDataDayEarningsPrefix+nowkey)
	redisMDArray, err := redis.Values(conn.Do("EXEC"))
	if err != nil {
		httpRsp.Result = proto.Int32(int32(gconst.ErrRedis))
		httpRsp.Msg = proto.String("统一获取缓存操作失败")
		log.Errorf("code:%d msg:%s redisMDArray Values err, err:%s", httpRsp.GetResult(), httpRsp.GetMsg(), err.Error())
		return
	}

	yestardayall, _ := redis.Int64(redisMDArray[0], nil)
	historyall, _ := redis.Int64(redisMDArray[1], nil)
	todayadnum, _ := redis.Int(redisMDArray[2], nil)
	todayonlinenum, _ := redis.Int(redisMDArray[3], nil)
	todayall, _ := redis.Int64(redisMDArray[4], nil)

	// rsp
	rsp := &entranceRsp{
		YestardayAll:   float32(yestardayall) / float32(100),
		HistoryAll:     float32(historyall) / float32(100),
		TodayAdNum:     int32(todayadnum),
		TodayOnlineNum: int32(todayonlinenum),
		TodayAll:       float32(todayall) / float32(100),
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
