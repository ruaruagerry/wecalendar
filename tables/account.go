package tables

import (
	"time"
)

const (
	// ChannelTypeTest 测试
	ChannelTypeTest = 0
	// ChannelTypeWx 微信
	ChannelTypeWx = 1
)

// Account 玩家
type Account struct {
	ID            int64     `xorm:"id pk autoincr <-"` // 用户ID
	Nick          string    `xorm:"nick"`              // 昵称
	Gender        int32     `xorm:"gender"`            // 性别
	Portrait      string    `xorm:"portrait"`          // 头像
	OpenID        string    `xorm:"open_id"`           // OpenID - 所有channel公用的唯一标志
	UnionID       string    `xorm:"union_id"`          // 微信UnionID
	SessionKey    string    `xorm:"session_key"`       // 微信SessionKey
	CreateTime    time.Time `xorm:"created"`           // 创建时间
	LastLoginTime time.Time `xorm:"updated"`           // 最后登录时间
	Channel       int32     `xorm:"channel"`           // 登陆方式
}
