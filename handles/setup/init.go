package setup

import (
	"weagent/server"
)

func init() {
	server.RegisterGetHandle("/setup/real/get", realGetHandle)        // 获取实名认证信息
	server.RegisterPostHandle("/setup/real/modify", realModifyHandle) // 修改实名认证信息
}
