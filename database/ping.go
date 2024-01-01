package database

import (
	"github.com/ygxiaobai111/GolixirDB/interface/resp"
	"github.com/ygxiaobai111/GolixirDB/resp/reply"
)

func Ping(db *DB, args [][]byte) resp.Reply {
	return &reply.PongReply{}
}

// 注册ping命令
func init() {
	RegisterCommand("ping", Ping, -1)
}
