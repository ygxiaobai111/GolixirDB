package handler

/*
 * A tcp.RespHandler implements redis protocol
 */

import (
	"context"
	"github.com/ygxiaobai111/GolixirDB/cluster"
	"github.com/ygxiaobai111/GolixirDB/config"
	"github.com/ygxiaobai111/GolixirDB/database"
	databaseface "github.com/ygxiaobai111/GolixirDB/interface/database"
	util "github.com/ygxiaobai111/GolixirDB/lib/logger"
	"github.com/ygxiaobai111/GolixirDB/lib/sync/atomic"
	"github.com/ygxiaobai111/GolixirDB/resp/connection"
	"github.com/ygxiaobai111/GolixirDB/resp/parser"
	"github.com/ygxiaobai111/GolixirDB/resp/reply"
	"io"
	"net"
	"strings"
	"sync"
)

var (
	unknownErrReplyBytes = []byte("-ERR unknown\r\n") // 未知错误回复的字节表示
)

// RespHandler 实现了tcp.Handler，充当redis处理程序
type RespHandler struct {
	activeConn sync.Map              // 存储活跃的客户端连接
	db         databaseface.Database // 数据库接口
	closing    atomic.Boolean        // 是否拒绝新客户端和新请求的标志
}

// MakeHandler 创建一个RespHandler实例
func MakeHandler() *RespHandler {
	var db databaseface.Database
	//db = database.NewEchoDatabase()  //示例
	//是否开启集群
	if config.Properties.ClusterMode && len(config.Properties.Peers) > 0 && config.Properties.Self != "" {
		db = cluster.MakeClusterDatabase() // 初始化数据库
	} else {
		db = database.NewStandaloneDatabase() // 初始化数据库

	}

	return &RespHandler{
		db: db,
	}
}

// closeClient 关闭客户端连接并进行清理
func (h *RespHandler) closeClient(client *connection.Connection) {
	_ = client.Close()
	h.db.AfterClientClose(client)
	h.activeConn.Delete(client)
}

// Handle 接收并执行redis命令
func (h *RespHandler) Handle(ctx context.Context, conn net.Conn) {
	if h.closing.Get() {
		// 如果处理程序正在关闭，拒绝新连接
		_ = conn.Close()
	}

	client := connection.NewConn(conn) // 创建新的客户端连接
	h.activeConn.Store(client, 1)

	ch := parser.ParseStream(conn) // 解析输入流
	for payload := range ch {
		if payload.Err != nil {
			// 处理连接错误
			if payload.Err == io.EOF ||
				payload.Err == io.ErrUnexpectedEOF ||
				strings.Contains(payload.Err.Error(), "use of closed network connection") {
				// 连接关闭
				h.closeClient(client)
				util.LogrusObj.Info("connection closed: " + client.RemoteAddr().String())
				return
			}
			// 协议错误
			errReply := reply.MakeErrReply(payload.Err.Error())
			err := client.Write(errReply.ToBytes())
			if err != nil {
				h.closeClient(client)
				util.LogrusObj.Info("connection closed: " + client.RemoteAddr().String())
				return
			}
			continue
		}
		if payload.Data == nil {
			// 空负载错误
			util.LogrusObj.Error("empty payload")
			continue
		}
		r, ok := payload.Data.(*reply.MultiBulkReply)
		if !ok {
			// 需要多批量回复
			util.LogrusObj.Error("require multi bulk reply")
			continue
		}
		result := h.db.Exec(client, r.Args) // 执行数据库命令
		if result != nil {
			_ = client.Write(result.ToBytes()) // 返回执行结果
		} else {
			_ = client.Write(unknownErrReplyBytes) // 返回未知错误
		}
	}
}

// Close 停止处理程序
func (h *RespHandler) Close() error {
	util.LogrusObj.Info("handler shutting down...")
	h.closing.Set(true)
	// 待实现：并发等待
	h.activeConn.Range(func(key interface{}, val interface{}) bool {
		client := key.(*connection.Connection)
		_ = client.Close() // 关闭所有客户端连接
		return true
	})
	h.db.Close() // 关闭数据库
	return nil
}
