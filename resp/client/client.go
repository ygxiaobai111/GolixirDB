package client

import (
	// 引入所需的包
	"github.com/ygxiaobai111/GolixirDB/interface/resp"
	util "github.com/ygxiaobai111/GolixirDB/lib/logger"
	"github.com/ygxiaobai111/GolixirDB/lib/sync/wait"
	"github.com/ygxiaobai111/GolixirDB/resp/parser"
	"github.com/ygxiaobai111/GolixirDB/resp/reply"
	"net"
	"runtime/debug"
	"sync"
	"time"
)

// Client 是一个基于管道模式的 Redis 客户端
type Client struct {
	conn        net.Conn      // 网络连接
	pendingReqs chan *request // 等待发送的请求队列
	waitingReqs chan *request // 等待响应的请求队列
	ticker      *time.Ticker  // 定时器，用于心跳检测
	addr        string        // 服务器地址

	working *sync.WaitGroup // 用于跟踪未完成请求（包括等待和正在处理的请求）
}

// request 是发送到 Redis 服务器的消息
type request struct {
	id        uint64
	args      [][]byte   // 请求参数
	reply     resp.Reply // 响应
	heartbeat bool       // 是否为心跳检测请求
	waiting   *wait.Wait // 等待响应
	err       error      // 请求过程中的错误
}

// 一些常量定义
const (
	chanSize = 256             // 队列大小
	maxWait  = 3 * time.Second // 最大等待时间
)

// MakeClient 创建一个新的客户端实例
func MakeClient(addr string) (*Client, error) {
	conn, err := net.Dial("tcp", addr) // 建立 TCP 连接
	if err != nil {
		return nil, err
	}
	return &Client{
		addr:        addr,
		conn:        conn,
		pendingReqs: make(chan *request, chanSize),
		waitingReqs: make(chan *request, chanSize),
		working:     &sync.WaitGroup{},
	}, nil
}

// Start 启动客户端的异步协程
func (client *Client) Start() {
	client.ticker = time.NewTicker(10 * time.Second) // 设置心跳检测间隔
	go client.handleWrite()                          // 处理写操作的协程
	go func() {
		err := client.handleRead() // 处理读操作的协程
		if err != nil {
			util.LogrusObj.Error(err)
		}
	}()
	go client.heartbeat() // 心跳检测协程
}

// Close 停止异步协程并关闭连接
func (client *Client) Close() {
	// 关闭操作和资源清理
	client.ticker.Stop()
	close(client.pendingReqs)
	client.working.Wait()
	_ = client.conn.Close()
	close(client.waitingReqs)
}

// handleConnectionError 处理连接错误
func (client *Client) handleConnectionError(err error) error {
	// 连接错误处理，重连逻辑
	err1 := client.conn.Close()
	if err1 != nil {
		if opErr, ok := err1.(*net.OpError); ok {
			if opErr.Err.Error() != "use of closed network connection" {
				return err1
			}
		} else {
			return err1
		}
	}
	conn, err1 := net.Dial("tcp", client.addr)
	if err1 != nil {
		util.LogrusObj.Error(err1)
		return err1
	}
	client.conn = conn
	go func() {
		_ = client.handleRead()
	}()
	return nil
}

// heartbeat 心跳检测逻辑
func (client *Client) heartbeat() {
	for range client.ticker.C {
		client.doHeartbeat()
	}
}

// handleWrite 处理写操作
func (client *Client) handleWrite() {
	for req := range client.pendingReqs {
		client.doRequest(req)
	}
}

// Send 向 Redis 服务器发送请求
func (client *Client) Send(args [][]byte) resp.Reply {
	// 发送请求并处理响应
	request := &request{
		args:      args,
		heartbeat: false,
		waiting:   &wait.Wait{},
	}
	request.waiting.Add(1)
	client.working.Add(1)
	defer client.working.Done()
	client.pendingReqs <- request
	timeout := request.waiting.WaitWithTimeout(maxWait)
	if timeout {
		return reply.MakeErrReply("server time out")
	}
	if request.err != nil {
		return reply.MakeErrReply("request failed")
	}
	return request.reply
}

func (client *Client) doHeartbeat() {
	// 执行心跳检测请求
	request := &request{
		args:      [][]byte{[]byte("PING")},
		heartbeat: true,
		waiting:   &wait.Wait{},
	}
	request.waiting.Add(1)
	client.working.Add(1)
	defer client.working.Done()
	client.pendingReqs <- request
	request.waiting.WaitWithTimeout(maxWait)
}

func (client *Client) doRequest(req *request) {
	// 执行实际的请求逻辑
	if req == nil || len(req.args) == 0 {
		return
	}
	re := reply.MakeMultiBulkReply(req.args)
	bytes := re.ToBytes()
	_, err := client.conn.Write(bytes)
	i := 0
	for err != nil && i < 3 {
		err = client.handleConnectionError(err)
		if err == nil {
			_, err = client.conn.Write(bytes)
		}
		i++
	}
	if err == nil {
		client.waitingReqs <- req
	} else {
		req.err = err
		req.waiting.Done()
	}
}

func (client *Client) finishRequest(reply resp.Reply) {
	// 完成请求并处理响应
	defer func() {
		if err := recover(); err != nil {
			debug.PrintStack()
			util.LogrusObj.Error(err)
		}
	}()
	request := <-client.waitingReqs
	if request == nil {
		return
	}
	request.reply = reply
	if request.waiting != nil {
		request.waiting.Done()
	}
}

func (client *Client) handleRead() error {
	// 处理读操作，接收服务器响应
	ch := parser.ParseStream(client.conn)
	for payload := range ch {
		if payload.Err != nil {
			client.finishRequest(reply.MakeErrReply(payload.Err.Error()))
			continue
		}
		client.finishRequest(payload.Data)
	}
	return nil
}
