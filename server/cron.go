package server

import (
	"github.com/robfig/cron"
	log "github.com/sirupsen/logrus"
)

// CronJob 定时任务
func CronJob() {
	log.Info("Start CronJob")

	cron := cron.New()

	// 12点略微延后一点
	cron.AddFunc("1 0 0 * * ?", func() {
		cronData(pool.Get(), dbConnect)
	})

	cron.Start()
}
