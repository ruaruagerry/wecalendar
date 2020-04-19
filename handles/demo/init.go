package demo

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

	server.RegisterGetHandle("/demon/hello", helloHandle)
}
