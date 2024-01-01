package cluster

/*
连接工厂实现
*/
import (
	"context"
	"errors"
	"github.com/jolestar/go-commons-pool/v2"
	"github.com/ygxiaobai111/GolixirDB/resp/client"
)

type connectionFactory struct {
	Peer string //目标地址
}

// MakeObject 新建连接所作操作
func (f *connectionFactory) MakeObject(ctx context.Context) (*pool.PooledObject, error) {
	c, err := client.MakeClient(f.Peer)
	if err != nil {
		return nil, err
	}
	c.Start()
	return pool.NewPooledObject(c), nil
}

// DestroyObject 关闭连接所需操作
func (f *connectionFactory) DestroyObject(ctx context.Context, object *pool.PooledObject) error {
	//断言确定类型
	c, ok := object.Object.(*client.Client)
	if !ok {
		return errors.New("type mismatch")
	}
	c.Close()
	return nil
}

func (f *connectionFactory) ValidateObject(ctx context.Context, object *pool.PooledObject) bool {
	// do validate
	return true
}

func (f *connectionFactory) ActivateObject(ctx context.Context, object *pool.PooledObject) error {
	// do activate
	return nil
}

func (f *connectionFactory) PassivateObject(ctx context.Context, object *pool.PooledObject) error {
	// do passivate
	return nil
}
