package tcp

import (
	"context"
	"net"
)

// Handler 业务
type Handler interface {
	Handle(ctx context.Context, conn net.Conn)
	Close() error
}
