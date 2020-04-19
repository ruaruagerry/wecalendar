/*
 * @Author: your name
 * @Date: 2019-12-27 12:09:29
 * @LastEditTime : 2019-12-27 16:45:24
 * @LastEditors  : Please set LastEditors
 * @Description: In User Settings Edit
 * @FilePath: \wecalendar\handles\data\init.go
 */

package data

import (
	"wecalendar/server"
)

func init() {
	server.RegisterGetHandle("/data/entrance", entranceHandle) // 分红主界面
}
