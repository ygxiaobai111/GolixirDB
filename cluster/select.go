package cluster

import "github.com/ygxiaobai111/GolixirDB/interface/resp"

// execSelect 本地执行即可
func execSelect(cluster *ClusterDatabase, c resp.Connection, cmdAndArgs [][]byte) resp.Reply {
	return cluster.db.Exec(c, cmdAndArgs)
}
