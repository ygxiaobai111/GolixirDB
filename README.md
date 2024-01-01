# GolixirDB
High performance concurrent middleware implemented by go

**GolixirDB**是一个用 Go 语言实现的 内存数据库

关键功能:
- 兼容redis协议 可用redis-cli或其他redis客户端进行连接
- 支持 string, set数据结构
- AOF 持久化及 AOF 重写
- 内置集群模式. 集群对客户端是透明的, 可以像使用单机版 GolixirDB 一样使用 GolixirDB 集群
- MSET, MSETNX, DEL, Rename, RenameNX 命令在集群模式下原子性执行, 目前不允许 key 在集群的不同节点上
- 并行引擎, 无需担心操作会阻塞整个服务器.

TODO（大饼）: 
- 自动过期
- 其他数据结构
- 事务

### 支持的操作：
`ping
del
exists
type
rename
renamenx
set
setnx
get
getset
flushdb
select`
## 运行 GolixirDB ：
go run main.go -cf config.yaml

若启动时未设置配置文件路径，则会尝试读取工作目录中的 config.yaml 文件

### 集群模式启动：
配置文件添加
#### 是否启动集群
- clusterMode: `true`
#### 本节点地址
- self: `127.0.0.1:14332`
#### 集群其他节点地址
peers:
- `127.0.0.1:14333`

集群模式中访问任意节点访问集群所有数据
