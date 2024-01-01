package cluster

import (
	"github.com/ygxiaobai111/GolixirDB/interface/resp"
	"github.com/ygxiaobai111/GolixirDB/resp/reply"
)

func Rename(cluster *ClusterDatabase, c resp.Connection, args [][]byte) resp.Reply {
	if len(args) != 3 {
		return reply.MakeErrReply("ERR wrong number of arguments for 'rename' command")
	}

	src := string(args[1])
	dest := string(args[2])
	//原本key的ip
	srcPeer := cluster.peerPicker.PickNode(src)
	// 目标key的ip
	destPeer := cluster.peerPicker.PickNode(dest)
	//槽位必须在一个节点
	if srcPeer != destPeer {
		return reply.MakeErrReply("ERR rename must within one slot in cluster mode")
	}
	return cluster.relay(srcPeer, c, args)
}
