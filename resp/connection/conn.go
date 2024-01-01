package connection

import (
	"github.com/ygxiaobai111/GolixirDB/lib/sync/wait"
	"net"
	"sync"
	"time"
)

// Connection 表示一个连接
type Connection struct {
	conn net.Conn // 网络连接
	// 等待直到回复完成
	waitingReply wait.Wait
	// 发送响应时的锁
	mu sync.Mutex
	// 选定的数据库
	selectedDB int
}

// NewConn 创建一个新的Connection实例
func NewConn(conn net.Conn) *Connection {
	return &Connection{
		conn: conn,
	}
}

// RemoteAddr 返回远程网络地址
func (c *Connection) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

// Close 与客户端断开连接
func (c *Connection) Close() error {
	c.waitingReply.WaitWithTimeout(10 * time.Second)
	_ = c.conn.Close()
	return nil
}

// Write 通过tcp连接向客户端发送响应
func (c *Connection) Write(b []byte) error {
	if len(b) == 0 {
		return nil
	}
	c.mu.Lock()
	c.waitingReply.Add(1)
	defer func() {
		c.waitingReply.Done()
		c.mu.Unlock()
	}()

	_, err := c.conn.Write(b)
	return err
}

// GetDBIndex 返回选定的数据库索引
func (c *Connection) GetDBIndex() int {
	return c.selectedDB
}

// SelectDB 选择一个数据库
func (c *Connection) SelectDB(dbNum int) {
	c.selectedDB = dbNum
}
