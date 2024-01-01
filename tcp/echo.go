package tcp

import (
	// 引入所需的包
	"bufio"
	"context"
	util "github.com/ygxiaobai111/GolixirDB/lib/logger"
	"github.com/ygxiaobai111/GolixirDB/lib/sync/atomic"
	"github.com/ygxiaobai111/GolixirDB/lib/sync/wait"
	"io"
	"net"
	"sync"
	"time"
)

// EchoHandler 是一个用于测试的 echo 服务器处理器
type EchoHandler struct {
	activeConn sync.Map // 存储活跃连接的映射

	closing atomic.Boolean // 使用自定义原子包来监测服务器是否正在关闭
}

// MakeEchoHandler 创建并返回一个 EchoHandler 实例
func MakeHandler() *EchoHandler {
	return &EchoHandler{}
}

// EchoClient 是 EchoHandler 的客户端，用于测试
type EchoClient struct {
	Conn    net.Conn  // 客户端连接
	Waiting wait.Wait // 用于控制等待和超时的 Wait 实例
}

// Close 关闭连接
func (c *EchoClient) Close() error {
	c.Waiting.WaitWithTimeout(10 * time.Second) // 最长等待时间
	c.Conn.Close()                              // 关闭连接
	return nil
}

// Handle 是处理 TCP 连接的主要逻辑
func (h *EchoHandler) Handle(ctx context.Context, conn net.Conn) {
	// 如果服务器正在关闭，则拒绝新的连接
	if h.closing.Get() {
		_ = conn.Close()
	}

	// 创建一个新的 EchoClient 实例，并存储到活跃连接映射中
	client := &EchoClient{
		Conn: conn,
	}
	h.activeConn.Store(client, struct{}{})

	// 创建一个读取器来读取客户端发送的数据
	reader := bufio.NewReader(conn)
	for {
		// 读取客户端发送的消息以'\n'为结束符，可能出现 EOF、超时或服务器提前关闭等情况
		msg, err := reader.ReadString('\n')
		if err != nil {
			// 处理读取错误
			if err == io.EOF {
				util.LogrusObj.Info("connection close")
				h.activeConn.Delete(client) //将用户从map用户组去除
			} else {
				util.LogrusObj.Warn(err)
			}
			return
		}

		// 将接收到的消息回显给客户端
		client.Waiting.Add(1)
		b := []byte(msg)
		_, _ = conn.Write(b)
		client.Waiting.Done()
	}
}

// Close 停止 echo 处理器
func (h *EchoHandler) Close() error {
	util.LogrusObj.Info("handler shutting down...")
	h.closing.Set(true) // 设置关闭状态

	// 遍历并关闭所有活跃的客户端连接
	h.activeConn.Range(func(key interface{}, val interface{}) bool {
		client := key.(*EchoClient)
		_ = client.Close()
		return true
	})
	return nil
}
