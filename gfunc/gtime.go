package gfunc

import "time"

// TomorrowZeroRemain 到明天凌晨剩余过期时间
func TomorrowZeroRemain() int64 {
	now := time.Now().Unix()
	nowtime := time.Now().Format("2006-01-02")
	t, _ := time.ParseInLocation("2006-01-02 15:04:05", nowtime+" 23:59:59", time.Local)
	// 第二天凌晨
	tomts := t.Unix() + 1

	return tomts - now
}
