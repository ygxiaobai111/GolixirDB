package reply

import (
	"bytes"
	"github.com/ygxiaobai111/GolixirDB/interface/resp"
	"strconv"
)

var (
	nullBulkReplyBytes = []byte("$-1") // 空的批量回复的字节表示

	CRLF = "\r\n"
)

/* ---- Bulk Reply ---- */

// BulkReply 存储二进制安全字符串
type BulkReply struct {
	Arg []byte
}

// MakeBulkReply 创建BulkReply实例
func MakeBulkReply(arg []byte) *BulkReply {
	return &BulkReply{
		Arg: arg,
	}
}

// ToBytes 将BulkReply序列化为redis响应
func (r *BulkReply) ToBytes() []byte {
	if len(r.Arg) == 0 {
		return nullBulkReplyBytes
	}
	return []byte("$" + strconv.Itoa(len(r.Arg)) + CRLF + string(r.Arg) + CRLF)
}

/* ---- Multi Bulk Reply ---- */

// MultiBulkReply 存储字符串列表
type MultiBulkReply struct {
	Args [][]byte
}

// MakeMultiBulkReply 创建MultiBulkReply实例
func MakeMultiBulkReply(args [][]byte) *MultiBulkReply {
	return &MultiBulkReply{
		Args: args,
	}
}

// ToBytes 将MultiBulkReply序列化为lixir响应
func (r *MultiBulkReply) ToBytes() []byte {
	argLen := len(r.Args)
	var buf bytes.Buffer
	buf.WriteString("*" + strconv.Itoa(argLen) + CRLF)
	for _, arg := range r.Args {
		if arg == nil {
			buf.WriteString("$-1" + CRLF)
		} else {
			buf.WriteString("$" + strconv.Itoa(len(arg)) + CRLF + string(arg) + CRLF)
		}
	}
	return buf.Bytes()
}

/* ---- Status Reply ---- */

// StatusReply 存储一个简单的状态字符串
type StatusReply struct {
	Status string
}

// MakeStatusReply 创建StatusReply实例
func MakeStatusReply(status string) *StatusReply {
	return &StatusReply{
		Status: status,
	}
}

// ToBytes 将StatusReply序列化为lixir响应
func (r *StatusReply) ToBytes() []byte {
	return []byte("+" + r.Status + CRLF)
}

/* ---- Int Reply ---- */

// IntReply 存储一个int64数字
type IntReply struct {
	Code int64
}

// MakeIntReply 创建IntReply实例
func MakeIntReply(code int64) *IntReply {
	return &IntReply{
		Code: code,
	}
}

// ToBytes 将IntReply序列化为lixir响应
func (r *IntReply) ToBytes() []byte {
	return []byte(":" + strconv.FormatInt(r.Code, 10) + CRLF)
}

/* ---- Error Reply ---- */

// ErrorReply 是一个错误类型，同时也是lixir.Reply的实现
type ErrorReply interface {
	Error() string
	ToBytes() []byte
}

// StandardErrReply 代表处理错误
type StandardErrReply struct {
	Status string
}

// ToBytes 将StandardErrReply序列化为lixir响应
func (r *StandardErrReply) ToBytes() []byte {
	return []byte("-" + r.Status + CRLF)
}

func (r *StandardErrReply) Error() string {
	return r.Status
}

// MakeErrReply 创建StandardErrReply实例
func MakeErrReply(status string) *StandardErrReply {
	return &StandardErrReply{
		Status: status,
	}
}

// IsErrorReply 判断给定的回复是否为错误
func IsErrorReply(reply resp.Reply) bool {
	return reply.ToBytes()[0] == '-'
}
