package phone

import (
	"math/rand"
	"time"
	"weagent/server"
)

func init() {
	droprand = rand.New(rand.NewSource(time.Now().UnixNano()))

	server.RegisterGetHandle("/phone/entrance", entranceHandle)      // 手机验证入口
	server.RegisterPostHandle("/phone/getcode", getcodeHandle)       // 获取手机验证码
	server.RegisterPostHandle("/phone/bind", bindHandle)             // 绑定手机号
	server.RegisterPostHandle("/phone/modifybind", modifyBindHandle) // 修改绑定手机号
}
