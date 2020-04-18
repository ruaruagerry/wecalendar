package phone

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"regexp"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/dysmsapi"
)

var (
	droprand *rand.Rand
)

const (
	// getCodeInterval 获取验证码间隔
	getCodeInterval = 60
	// codeLiveTime 验证码有效时间（秒）
	codeLiveTime = 3 * 60
)

// getValidCode 获取验证码
func getValidCode() string {
	vcode := fmt.Sprintf("%06v", droprand.Int31n(1000000))
	return vcode
}

// verifyPhonFormat 验证手机号格式
func verifyPhonFormat(mobileNum string) bool {
	regular := "^((13[0-9])|(14[5,7])|(15[0-3,5-9])|(17[0,3,5-8])|(18[0-9])|166|198|199|(147))\\d{8}$"

	reg := regexp.MustCompile(regular)
	return reg.MatchString(mobileNum)
}

// sendPhoneMsg 发送手机短信
func sendPhoneMsg(phoneno string, code string) error {
	client, err := dysmsapi.NewClientWithAccessKey("cn-hangzhou", "1", "1")
	if err != nil {
		return fmt.Errorf("sendPhoneMsg NewClientWithAccessKey err:%v", err.Error())
	}

	request := dysmsapi.CreateSendSmsRequest()
	request.Scheme = "https"

	param := make(map[string]interface{})
	param["code"] = code
	buf, err := json.Marshal(param)
	if err != nil {
		return fmt.Errorf("Marshal err, err:%s", err.Error())
	}

	request.PhoneNumbers = phoneno
	request.SignName = "1"
	request.TemplateCode = "1"
	request.TemplateParam = string(buf)

	_, err = client.SendSms(request)
	if err != nil {
		return fmt.Errorf("SendSms err, err:%s", err.Error())
	}

	return nil
}
