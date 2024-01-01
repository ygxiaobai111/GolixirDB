package database

import (
	"strings" // 引入字符串处理包
)

// cmdTable 是一个映射，用于存储所有注册的命令
var cmdTable = make(map[string]*command)

// command 结构体定义了一个数据库命令
type command struct {
	executor ExecFunc // 命令执行函数
	arity    int      // 允许的参数数量，arity < 0 表示其为可变参数但是(len(args) >= -arity )
}

// RegisterCommand 注册一个新命令
// arity 表示允许的命令参数数量，arity < 0 表示 len(args) >= -arity。
// 例如：`get`命令的arity为2，`mget`命令的arity为-2
func RegisterCommand(name string, executor ExecFunc, arity int) {
	name = strings.ToLower(name) // 将命令名转换为小写
	cmdTable[name] = &command{
		executor: executor, // 设置执行函数
		arity:    arity,    // 设置参数数量
	}
}
