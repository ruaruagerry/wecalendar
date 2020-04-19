package rconst

// PhoneMsg 手机验证码
type PhoneMsg struct {
	Phone string `json:"phone"`
	Code  string `json:"code"`
}

const (
	// StringPhoneGetCodeTagPrefix 获取验证码标志
	StringPhoneGetCodeTagPrefix = "wecalendar:phone:codetag:"
	// StringPhoneCodePrefix 验证码信息
	StringPhoneCodePrefix = "wecalendar:phone:code:"
	// SetPhoneHasBinded 已绑定的手机号
	SetPhoneHasBinded = "wecalendar:phone:hasbinded"
)
