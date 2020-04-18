package tables

import "time"

// Webaccount 后台玩家
type Webaccount struct {
	ID         int64     `xorm:"id pk autoincr <-"`                    // 用户ID
	Account    string    `xorm:"varchar(25) notnull unique 'account'"` // 账号
	Password   string    `xorm:"password"`                             // 密码
	Role       []string  `xorm:"role"`                                 // 角色
	Nick       string    `xorm:"nick"`                                 // 昵称
	Gender     int32     `xorm:"gender"`                               // 性别(注:账号服返回的性别字段为sex)
	Portrait   string    `xorm:"portrait"`                             // 头像
	CreateTime time.Time `xorm:"createtime created"`                   // 创建时间
}
