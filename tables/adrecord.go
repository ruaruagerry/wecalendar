package tables

import "time"

// Adrecord 广告收益记录
type Adrecord struct {
	Rid        int64     `xorm:"pk autoincr BIGINT(20) <-" json:"rid"`
	ID         string    `xorm:"id" json:"id"`                 // 用户ID
	Name       string    `xorm:"name" json:"name"`             // 用户昵称
	Earnings   int64     `xorm:"earnings" json:"earnings"`     // 收益（分）
	AdMoney    int64     `xorm:"admoney" json:"admoney"`       // 当前余额（分）
	CreateTime time.Time `xorm:"createtime" json:"createtime"` // 创建时间
}
