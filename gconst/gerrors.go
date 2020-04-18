package gconst

// Error 错误
type Error int32

/* 返回客户端错误码 */
/* 区间：0-10000 */
const (
	// Success 成功 0
	Success = Error(iota)

	// UnknownError 未知错误 1
	UnknownError = Error(1)
	// ErrParam 参数错误 2
	ErrParam = Error(2)
	// ErrParamNil 请求参数为空 3
	ErrParamNil = Error(3)
	// ErrParse 解析失败 4
	ErrParse = Error(4)
	// ErrDB 数据库操作失败 5
	ErrDB = Error(5)
	// ErrRedis 缓存操作失败 6
	ErrRedis = Error(6)
	// ErrTableConfig 表配置错误 7
	ErrTableConfig = Error(7)
	// ErrTokenEmpty token is empty 8
	ErrTokenEmpty = Error(8)
	// ErrTokenDecrypt token decrypt failed 9
	ErrTokenDecrypt = Error(9)
	// ErrTokenFormat token format is invalid 10
	ErrTokenFormat = Error(10)
	// ErrTokenExpired token expired 11
	ErrTokenExpired = Error(11)
	// ErrCreateUUID 生成uuid失败
	ErrCreateUUID = Error(12)
	// ErrHTTP http请求失败
	ErrHTTP = Error(13)
	// ErrHTTPTooFast http请求太快
	ErrHTTPTooFast = Error(14)

	/* auth 100-199 */
	// ErrAuthGetWexinUserInfo 获取微信用户信息失败
	ErrAuthGetWexinUserInfo = Error(100)
	// ErrAuthGetWeixinPlusInfo 获取微信额外信息失败
	ErrAuthGetWeixinPlusInfo = Error(101)
	// ErrAuthNotFindOpenID 没找到用户Openid
	ErrAuthNotFindOpenID = Error(102)

	/* setup 1000-1099 */
	// ErrSetupCardCode 身份证格式错误
	ErrSetupCardCode = Error(1000)
	// ErrSetupRealNick 身份证名称格式错误
	ErrSetupRealNick = Error(1001)
	// ErrSetupAlreadyRealCheck 已进行过实名认证
	ErrSetupAlreadyRealCheck = Error(1002)
	// ErrSetupExistCardCode 实名信息已存在
	ErrSetupExistCardCode = Error(1003)

	/* phone 1100-1199 */
	// ErrPhoneFormat 手机号格式错误
	ErrPhoneFormat = Error(1100)
	// ErrPhoneGetCodeFast 获取验证码过快
	ErrPhoneGetCodeFast = Error(1101)
	// ErrPhoneSendMsg 发送验证码失败
	ErrPhoneSendMsg = Error(1102)
	// ErrPhoneAlreadyBind 已绑定手机号
	ErrPhoneAlreadyBind = Error(1103)
	// ErrPhoneCode 验证码错误
	ErrPhoneCode = Error(1104)
	// ErrPhoneHasBinded 手机号码已被绑定
	ErrPhoneHasBinded = Error(1105)

	/* money 1200-1299 */
	// ErrMoneyInvalidGetout 提现金额错误
	ErrMoneyInvalidGetout = Error(1200)
	// ErrMoneyNotEnough 提现金额不足
	ErrMoneyNotEnough = Error(1201)

	/* game 1300-1399 */
	// ErrRebirtNumNotEnough 复活次数不足
	ErrRebirtNumNotEnough = Error(1300)
)

var errMsg = map[Error]string{
	UnknownError:    "未知错误",
	ErrParam:        "参数错误",
	ErrParamNil:     "请求参数为空",
	ErrParse:        "解析失败",
	ErrDB:           "数据库操作失败",
	ErrRedis:        "缓存操作失败",
	ErrTableConfig:  "表配置错误",
	ErrTokenEmpty:   "token为空",
	ErrTokenDecrypt: "token解析失败",
	ErrTokenFormat:  "token格式错误",
	ErrTokenExpired: "token已过期",
	ErrCreateUUID:   "生成uuid失败",
	ErrHTTP:         "http请求失败",
	ErrHTTPTooFast:  "http请求太快",
}

// String 获得错误码描述信息
func (e Error) String() string {
	v, ok := errMsg[e]
	if !ok {
		return "未定义错误描述"
	}

	return v
}
