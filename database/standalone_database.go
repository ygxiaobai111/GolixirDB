package database

import (
	"fmt"
	"github.com/ygxiaobai111/GolixirDB/aof"
	"github.com/ygxiaobai111/GolixirDB/config"
	"github.com/ygxiaobai111/GolixirDB/interface/resp"
	util "github.com/ygxiaobai111/GolixirDB/lib/logger"
	"github.com/ygxiaobai111/GolixirDB/resp/reply"
	"runtime/debug"
	"strconv"
	"strings"
)

// StandaloneDatabase is a set of multiple database set
type StandaloneDatabase struct {
	dbSet      []*DB
	aofHandler *aof.AofHandler
}

// NewStandaloneDatabase creates a redis database,
func NewStandaloneDatabase() *StandaloneDatabase {
	mdb := &StandaloneDatabase{}
	if config.Properties.Databases == 0 {
		config.Properties.Databases = 16
	}
	mdb.dbSet = make([]*DB, config.Properties.Databases)
	for i := range mdb.dbSet {
		singleDB := makeDB()
		singleDB.index = i
		mdb.dbSet[i] = singleDB
	}
	if config.Properties.AppendOnly {
		aofH, err := aof.NewAOFHandler(mdb)
		if err != nil {
			panic(err)
		}
		mdb.aofHandler = aofH
		//将这个函数赋给每个db
		for _, db := range mdb.dbSet {
			/*
				//闭包bug 内部函数引用外部局部变量，变量将逃逸到堆
				//此bug会导致AddAof()的第一个参数都是最后一个db.index的值
				db.addAof = func(line CmdLine) {
					mdb.aofHandler.AddAof(db.index, line)
				}
			*/
			ndb := db
			ndb.addAof = func(line CmdLine) {
				mdb.aofHandler.AddAof(ndb.index, line)
			}
		}

	}
	return mdb
}

// Exec executes command
// parameter `cmdLine` contains command and its arguments, for example: "set key value"
func (mdb *StandaloneDatabase) Exec(c resp.Connection, cmdLine [][]byte) (result resp.Reply) {
	defer func() {
		if err := recover(); err != nil {
			util.LogrusObj.Warn(fmt.Sprintf("error occurs: %v\n%s", err, string(debug.Stack())))
		}
	}()

	cmdName := strings.ToLower(string(cmdLine[0]))
	//当命令为 select 则是选择数据库 单独处理
	if cmdName == "select" {
		if len(cmdLine) != 2 {
			return reply.MakeArgNumErrReply("select")
		}
		return execSelect(c, mdb, cmdLine[1:])
	}
	// normal commands
	dbIndex := c.GetDBIndex()
	selectedDB := mdb.dbSet[dbIndex]
	return selectedDB.Exec(c, cmdLine)
}

// Close graceful shutdown database
func (mdb *StandaloneDatabase) Close() {

}

func (mdb *StandaloneDatabase) AfterClientClose(c resp.Connection) {
}

func execSelect(c resp.Connection, mdb *StandaloneDatabase, args [][]byte) resp.Reply {
	dbIndex, err := strconv.Atoi(string(args[0]))
	if err != nil {
		return reply.MakeErrReply("ERR invalid DB index")
	}
	if dbIndex >= len(mdb.dbSet) {
		return reply.MakeErrReply("ERR DB index is out of range")
	}
	c.SelectDB(dbIndex)
	return reply.MakeOkReply()
}
