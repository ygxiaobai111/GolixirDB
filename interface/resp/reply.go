package resp

type Reply interface {
	// ToBytes 将向客户端的响应转化为字节码
	ToBytes() []byte
}
