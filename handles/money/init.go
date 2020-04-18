package money

import (
	"weagent/server"
)

func init() {
	/* 服务器的金币全部都以分为单位 */

	server.RegisterGetHandle("/money/ad/see", adSeeHandle)                // 查看广告上报
	server.RegisterGetHandle("/money/ad/click", adClickHandle)            // 点击广告上报
	server.RegisterPostHandle("/money/ad/record", adRecordHandle)         // 查看广告收益记录
	server.RegisterGetHandle("/money/entrance", entranceHandle)           // 主界面
	server.RegisterPostHandle("/money/getout/apply", getoutApplyHandle)   // 提现申请
	server.RegisterPostHandle("/money/getout/record", getoutRecordHandle) // 查看提现记录
}
