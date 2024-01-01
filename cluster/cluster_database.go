// Package cluster 提供了一个对客户端透明的服务器端集群。你可以连接到集群中的任何一个节点来访问集群中的所有数据。
package cluster

import (
	// 导入所需的包
	"context"
	"fmt"
	pool "github.com/jolestar/go-commons-pool/v2"
	"github.com/ygxiaobai111/GolixirDB/config"
	"github.com/ygxiaobai111/GolixirDB/database"
	databaseface "github.com/ygxiaobai111/GolixirDB/interface/database"
	"github.com/ygxiaobai111/GolixirDB/interface/resp"
	"github.com/ygxiaobai111/GolixirDB/lib/consistenthash"
	util "github.com/ygxiaobai111/GolixirDB/lib/logger"
	"github.com/ygxiaobai111/GolixirDB/resp/reply"

	"runtime/debug"
	"strings"
)

// ClusterDatabase 代表一个 godis 集群的节点
// 它持有部分数据并协调其他节点完成事务
type ClusterDatabase struct {
	self string // 当前节点的标识

	nodes          []string                    // 集群中所有节点的列表
	peerPicker     *consistenthash.NodeMap     // 用于一致性哈希的节点选择器
	peerConnection map[string]*pool.ObjectPool // 对等节点的连接池
	db             databaseface.Database       // 数据库实例
}

// MakeClusterDatabase 创建并启动一个集群节点
func MakeClusterDatabase() *ClusterDatabase {
	cluster := &ClusterDatabase{
		self: config.Properties.Self,

		db:             database.NewStandaloneDatabase(),
		peerPicker:     consistenthash.NewNodeMap(nil),
		peerConnection: make(map[string]*pool.ObjectPool),
	}
	nodes := make([]string, 0, len(config.Properties.Peers)+1)
	for _, peer := range config.Properties.Peers {
		nodes = append(nodes, peer)
	}
	nodes = append(nodes, config.Properties.Self)
	cluster.peerPicker.AddNode(nodes...)
	ctx := context.Background()
	for _, peer := range config.Properties.Peers {
		//对兄弟节点新建连接
		cluster.peerConnection[peer] = pool.NewObjectPoolWithDefaultConfig(ctx, &connectionFactory{
			Peer: peer,
		})
	}
	cluster.nodes = nodes
	return cluster
}

// CmdFunc 代表一个命令的处理函数
type CmdFunc func(cluster *ClusterDatabase, c resp.Connection, cmdAndArgs [][]byte) resp.Reply

// Close 关闭当前的集群节点
func (cluster *ClusterDatabase) Close() {
	cluster.db.Close()
}

var router = makeRouter()

// Exec 在集群上执行命令
func (cluster *ClusterDatabase) Exec(c resp.Connection, cmdLine [][]byte) (result resp.Reply) {
	defer func() {
		if err := recover(); err != nil {
			util.LogrusObj.Warn(fmt.Sprintf("error occurs: %v\n%s", err, string(debug.Stack())))
			result = &reply.UnknownErrReply{}
		}
	}()
	cmdName := strings.ToLower(string(cmdLine[0]))
	cmdFunc, ok := router[cmdName]
	if !ok {
		return reply.MakeErrReply("ERR unknown command '" + cmdName + "', or not supported in cluster mode")
	}
	result = cmdFunc(cluster, c, cmdLine)
	return
}

// AfterClientClose 在客户端关闭连接后进行一些清理工作
func (cluster *ClusterDatabase) AfterClientClose(c resp.Connection) {
	cluster.db.AfterClientClose(c)
}
