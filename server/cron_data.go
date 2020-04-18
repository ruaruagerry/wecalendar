package server

import (
	"time"
	"weagent/rconst"

	"github.com/garyburd/redigo/redis"
	"github.com/go-xorm/xorm"
	log "github.com/sirupsen/logrus"
)

func cronData(conn redis.Conn, db *xorm.Engine) {
	timenow := time.Now()
	timeyesterday := timenow.AddDate(0, 0, -1)
	yestardaykey := timeyesterday.Format("2006-01-02")

	// redis multi get
	conn.Send("MULTI")
	conn.Send("GET", rconst.StringDataDayEarningsPrefix+yestardaykey)
	conn.Send("GET", rconst.StringDataEarnings)
	redisMDArray, err := redis.Values(conn.Do("EXEC"))
	if err != nil {
		log.Errorf("redisMDArray Values err, err:%s", err.Error())
		return
	}

	yestardayall, _ := redis.Int64(redisMDArray[0], nil)
	historyall, _ := redis.Int64(redisMDArray[1], nil)

	// redis multi set
	conn.Send("MULTI")
	conn.Send("SET", rconst.StringDataEarnings, historyall+yestardayall)
	_, err = conn.Do("EXEC")
	if err != nil {
		log.Errorf("set exec err, err:%s", err.Error())
		return
	}

	return
}
