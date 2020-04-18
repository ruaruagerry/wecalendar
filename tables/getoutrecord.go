package tables

import "time"

const (
	// GetoutStatusReview 审核中
	GetoutStatusReview = 0
	// GetoutStatusRefused 审核拒绝
	GetoutStatusRefused = 1
	// GetoutStatusSuccess 提现成功
	GetoutStatusSuccess = 2
	// GetoutStatusFailed 提现失败
	GetoutStatusFailed = 3
)

// Getoutrecord 提现记录
type Getoutrecord struct {
	Rid         int64     `xorm:"pk autoincr BIGINT(20) <-" json:"rid"`
	ID          string    `xorm:"id" json:"id"`                   // 用户ID
	Name        string    `xorm:"name" json:"name"`               // 用户昵称
	GetoutMoney int64     `xorm:"getoutmoney" json:"getoutmoney"` // 提现金额(分)
	CreateTime  time.Time `xorm:"createtime" json:"createtime"`   // 创建时间
	ResultTime  time.Time `xorm:"resulttime" json:"resulttime"`   // 处理时间
	Status      int32     `xorm:"status" json:"status"`           // 提现状态
}
