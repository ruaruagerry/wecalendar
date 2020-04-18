package auth

import "weagent/server"

func init() {
	server.RegisterPostHandleNoUserID("/auth/wxlogin", wxLoginHandle)     // 微信登陆
	server.RegisterPostHandleNoUserID("/auth/testlogin", testLoginHandle) // 测试登陆
}
