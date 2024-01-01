package database

import (
	"github.com/ygxiaobai111/GolixirDB/interface/resp"
	util "github.com/ygxiaobai111/GolixirDB/lib/logger"
	"github.com/ygxiaobai111/GolixirDB/resp/reply"
)

type EchoDatabase struct {
}

func NewEchoDatabase() *EchoDatabase {
	return &EchoDatabase{}
}

func (e EchoDatabase) Exec(client resp.Connection, args [][]byte) resp.Reply {
	return reply.MakeMultiBulkReply(args)

}

func (e EchoDatabase) AfterClientClose(c resp.Connection) {
	util.LogrusObj.Info("EchoDatabase AfterClientClose")
}

func (e EchoDatabase) Close() {
	util.LogrusObj.Info("EchoDatabase Close")

}
