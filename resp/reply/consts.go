package reply

// PongReply 是用于返回 +PONG 的结构体
type PongReply struct{}

var pongBytes = []byte("+PONG\r\n")

// ToBytes 方法用于将 PongReply 序列化为字节序列
func (r *PongReply) ToBytes() []byte {
	return pongBytes
}

// OkReply 是用于返回 +OK 的结构体
type OkReply struct{}

var okBytes = []byte("+OK\r\n")

// ToBytes 方法用于将 OkReply 序列化为字节序列
func (r *OkReply) ToBytes() []byte {
	return okBytes
}

var theOkReply = new(OkReply)

// MakeOkReply 用于创建一个 OkReply 实例
func MakeOkReply() *OkReply {
	return theOkReply
}

var nullBulkBytes = []byte("$-1\r\n")

// NullBulkReply 用于表示空的批量回复
type NullBulkReply struct{}

// ToBytes 方法用于将 NullBulkReply 序列化为字节序列
func (r *NullBulkReply) ToBytes() []byte {
	return nullBulkBytes
}

// MakeNullBulkReply 用于创建一个 NullBulkReply 实例
func MakeNullBulkReply() *NullBulkReply {
	return &NullBulkReply{}
}

var emptyMultiBulkBytes = []byte("*0\r\n")

// EmptyMultiBulkReply 用于表示空的多条批量回复
type EmptyMultiBulkReply struct{}

// ToBytes 方法用于将 EmptyMultiBulkReply 序列化为字节序列
func (r *EmptyMultiBulkReply) ToBytes() []byte {
	return emptyMultiBulkBytes
}

// NoReply 用于不返回任何内容的场景，例如 subscribe 命令
type NoReply struct{}

var noBytes = []byte("")

// ToBytes 方法用于将 NoReply 序列化为字节序列
func (r *NoReply) ToBytes() []byte {
	return noBytes
}
