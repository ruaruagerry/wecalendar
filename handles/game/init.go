package game

import (
	"math/rand"
	"time"
	"weagent/server"
)

var (
	droprand *rand.Rand
)

func init() {
	droprand = rand.New(rand.NewSource(time.Now().UnixNano()))

	server.RegisterGetHandle("/game/rebirth/use", rebirthUseHandle)    // 使用复活次数
	server.RegisterPostHandle("/game/score/update", scoreUpdateHandle) // 更新玩家分数，返回复活次数
	server.RegisterGetHandle("/game/score/rank", scoreRankHandle)      // 获取玩家分数排行榜
}
