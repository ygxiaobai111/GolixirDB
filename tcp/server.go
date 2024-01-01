package tcp

/**
 * A tcp server
 */

import (
	"context"
	"fmt"
	"github.com/ygxiaobai111/GolixirDB/interface/tcp"
	util "github.com/ygxiaobai111/GolixirDB/lib/logger"

	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// Config stores tcp server properties
type Config struct {
	Address string
}

// ListenAndServeWithSignal 绑定端口并处理请求，阻塞直到接收到停止信号
func ListenAndServeWithSignal(cfg *Config, handler tcp.Handler) error {
	// 创建一个用于关闭服务器的通道
	closeChan := make(chan struct{})
	// 创建一个用于接收系统信号的通道
	sigCh := make(chan os.Signal)
	// 注册需要监听的系统信号，包括 SIGHUP, SIGQUIT, SIGTERM, SIGINT
	signal.Notify(sigCh, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)

	// 启动一个新的协程来监听系统信号
	go func() {
		// 阻塞等待信号
		sig := <-sigCh
		// 判断接收到的信号类型，如果是停止信号之一，则向 closeChan 发送消息，用于通知服务器关闭
		switch sig {
		case syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			closeChan <- struct{}{}
		}
	}()

	listener, err := net.Listen("tcp", cfg.Address)
	if err != nil {

		return err
	}

	util.LogrusObj.Info(fmt.Sprintf("bind: %s, start listening...", cfg.Address))
	// 调用 ListenAndServe 函数来处理监听和请求，同时传入 closeChan 以便可以基于信号关闭服务器
	ListenAndServe(listener, handler, closeChan)
	return nil
}

// ListenAndServe 监听端口和处理请求，阻塞直到关闭
func ListenAndServe(listener net.Listener, handler tcp.Handler, closeChan <-chan struct{}) {
	// listen signal
	go func() {
		<-closeChan
		util.LogrusObj.Info("shutting down...")
		_ = listener.Close() // listener.Accept() will return err immediately
		_ = handler.Close()  // close connections
	}()

	// listen port
	defer func() {
		// close during unexpected error
		_ = listener.Close()
		_ = handler.Close()
	}()
	ctx := context.Background()
	var waitDone sync.WaitGroup
	for {
		conn, err := listener.Accept()
		if err != nil {
			break
		}
		// handle
		util.LogrusObj.Info("accept link")
		waitDone.Add(1)
		go func() {
			defer func() {
				//业务完成减一
				waitDone.Done()
			}()
			//实际业务
			handler.Handle(ctx, conn)
		}()
	}
	//等待所有用户完成服务
	waitDone.Wait()

}
