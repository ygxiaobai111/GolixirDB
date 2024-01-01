package cluster

import (
	"github.com/ygxiaobai111/GolixirDB/interface/resp"
	"github.com/ygxiaobai111/GolixirDB/resp/reply"
)

// FlushDB 广播给所有节点进行清空数据库信息
func flushDB(cluster *ClusterDatabase, c resp.Connection, args [][]byte) resp.Reply {
	replies := cluster.broadcast(c, args)
	var errReply reply.ErrorReply
	for _, v := range replies {
		if reply.IsErrorReply(v) {
			errReply = v.(reply.ErrorReply)
			break
		}
	}
	if errReply == nil {
		return &reply.OkReply{}
	}
	return reply.MakeErrReply("error occurs: " + errReply.Error())
}
