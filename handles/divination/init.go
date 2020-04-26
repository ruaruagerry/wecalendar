package divination

import (
	"math/rand"
	"time"
	"wecalendar/server"
)

var (
	droprand *rand.Rand
)

func init() {
	droprand = rand.New(rand.NewSource(time.Now().UnixNano()))

	server.RegisterPostHandle("/divination/public", divinationPublicHandle) // 发布吐槽
	server.RegisterGetHandle("/divination/get", divinationGetHandle)        // 拉取吐槽
	server.RegisterGetHandle("/divination/rank", divinationRankHandle)      // 获取玩家吐槽排行榜
	server.RegisterGetHandle("/divination/best", divinationBestHandle)      // 最佳吐槽
}
