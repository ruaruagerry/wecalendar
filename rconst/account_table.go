package rconst

const (
	// HashAccountPrefix account hash table + playerid
	HashAccountPrefix = "weagent:acc:"
	// FieldAccUserID user id
	FieldAccUserID = "id"
	// FieldAccName account name
	FieldAccName = "nickname"
	// FieldAccGender 性别
	FieldAccGender = "gender"
	// FieldAccImage 头像
	FieldAccImage = "avatarurl"
	// FieldAccOpenID account openID
	FieldAccOpenID = "openid"
	// FieldAccUnionID account unionID
	FieldAccUnionID = "unionid"
	// FiledAccCreateTime 创建时间
	FiledAccCreateTime = "createtime"
	// FiledAccLoginTime 最后登录时间
	FiledAccLoginTime = "logintime"
	// FieldAccPhone 手机号
	FieldAccPhone = "phone"
	// FieldAccChannel 登陆方式
	FieldAccChannel = "channel"

	// HashAccountOpenIDPrefix + openid
	HashAccountOpenIDPrefix = "weagent:acc:openid:"
	// FieldAccOpenIDUserID user id
	FieldAccOpenIDUserID = "id"
	// FieldAccOpenIDOpenID account openID
	FieldAccOpenIDOpenID = "openid"
	// FieldAccOpenIDPhone 手机号
	FieldAccOpenIDPhone = "phone"

	// SetUsers 登录过的用户
	SetUsers = "weagent:users"
)
