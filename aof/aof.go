package aof

import (
	"github.com/ygxiaobai111/GolixirDB/config"
	databaseface "github.com/ygxiaobai111/GolixirDB/interface/database"
	util "github.com/ygxiaobai111/GolixirDB/lib/logger"
	"github.com/ygxiaobai111/GolixirDB/lib/utils"
	"github.com/ygxiaobai111/GolixirDB/resp/connection"
	"github.com/ygxiaobai111/GolixirDB/resp/parser"
	"github.com/ygxiaobai111/GolixirDB/resp/reply"
	"io"
	"os"
	"strconv"
)

/*
追加方式进行持久化
*/

type CmdLine = [][]byte

const (
	aofQueueSize = 1 << 16
)

type payload struct {
	cmdLine CmdLine
	dbIndex int
}
type AofHandler struct {
	db          databaseface.Database
	aofChan     chan *payload //缓存区
	aofFile     *os.File
	aofFilename string //持久化文件名
	currentDB   int    //哪一个数据库
}

func NewAOFHandler(db databaseface.Database) (*AofHandler, error) {
	handler := &AofHandler{}
	handler.aofFilename = config.Properties.AppendFilename
	handler.db = db
	handler.LoadAof()
	//进行追加、读写操作
	aofFile, err := os.OpenFile(handler.aofFilename, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	handler.aofFile = aofFile
	handler.aofChan = make(chan *payload, aofQueueSize)
	go func() {
		handler.handleAof()
	}()
	return handler, nil
}
func (handler *AofHandler) AddAof(dbIndex int, cmdLine CmdLine) {
	if config.Properties.AppendOnly && handler.aofChan != nil {
		handler.aofChan <- &payload{
			cmdLine: cmdLine,
			dbIndex: dbIndex,
		}
	}
}

// handleAof 命令写入文件
func (handler *AofHandler) handleAof() {
	// serialized execution
	handler.currentDB = 0
	for p := range handler.aofChan {
		if p.dbIndex != handler.currentDB {
			// select db
			data := reply.MakeMultiBulkReply(utils.ToCmdLine("SELECT", strconv.Itoa(p.dbIndex))).ToBytes()
			_, err := handler.aofFile.Write(data)
			if err != nil {
				util.LogrusObj.Warn(err)
				continue // skip this command
			}
			handler.currentDB = p.dbIndex
		}
		data := reply.MakeMultiBulkReply(p.cmdLine).ToBytes()
		_, err := handler.aofFile.Write(data)
		if err != nil {
			util.LogrusObj.Warn(err)
		}
	}
}

// 将本地数据读取
func (handler *AofHandler) LoadAof() {

	file, err := os.Open(handler.aofFilename)
	if err != nil {
		util.LogrusObj.Warn(err)
		return
	}
	defer file.Close()
	//读取并解析
	ch := parser.ParseStream(file)
	fakeConn := &connection.Connection{} // only used for save dbIndex
	for p := range ch {
		if p.Err != nil {
			if p.Err == io.EOF {
				break
			}
			util.LogrusObj.Error("parse error: " + p.Err.Error())
			continue
		}
		if p.Data == nil {
			util.LogrusObj.Error("empty payload")
			continue
		}
		r, ok := p.Data.(*reply.MultiBulkReply)
		if !ok {
			util.LogrusObj.Error("require multi bulk reply")
			continue
		}
		//执行
		ret := handler.db.Exec(fakeConn, r.Args)
		if reply.IsErrorReply(ret) {
			util.LogrusObj.Error("exec err", err)
		}
	}
}
