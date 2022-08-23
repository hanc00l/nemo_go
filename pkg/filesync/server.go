package filesync

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"net"
	"os"
	"path/filepath"
)

var cwd string

func init() {
	var err error
	cwd, err = os.Getwd()
	if err != nil {
		logging.CLILog.Error(err)
	}
}

type GobConn struct {
	cnRd *bufio.Reader
	cnWt *bufio.Writer
	Dec  *gob.Decoder
	enc  *gob.Encoder
}

// initGobConn 初始化连接
func initGobConn(conn net.Conn) *GobConn {
	gbc := new(GobConn)
	gbc.cnRd = bufio.NewReader(conn)
	gbc.cnWt = bufio.NewWriter(conn)
	gbc.Dec = gob.NewDecoder(gbc.cnRd)
	gbc.enc = gob.NewEncoder(gbc.cnWt)
	return gbc
}

// gobConnWt 发送消息与数据
func (gbc *GobConn) gobConnWt(mg interface{}) error {
	err := gbc.enc.Encode(mg)
	if err != nil {
		return err
	}
	err = gbc.cnWt.Flush()
	return err
}

// StartFileSyncServer 启动文件同步服务监听
func StartFileSyncServer(host, port, authKey string) {
	serverAddr := fmt.Sprintf("%s:%s", host, port)
	srv, err := net.Listen("tcp", serverAddr)
	if err != nil {
		logging.CLILog.Error(err)
		logging.RuntimeLog.Error(err)
		return
	}
	for {
		conn, err := srv.Accept()
		if err != nil {
			logging.CLILog.Error(err)
			continue
		}
		go handleSync(conn, authKey)
	}
}

func handleSync(conn net.Conn, authKey string) {
	defer conn.Close()

	gbc := initGobConn(conn)
	for {
		mg := Message{}
		err := gbc.Dec.Decode(&mg)
		if err != nil {
			logging.CLILog.Error(err)
			return
		}
		// 检查authKey，如果不通过直接返回
		if success := checkSyncAuthKey(authKey, mg.MgAuthKey, gbc); success == false {
			logging.RuntimeLog.Warnf("Invalid auth from %s", conn.RemoteAddr().String())
			logging.CLILog.Warnf("Invalid auth from %s", conn.RemoteAddr().String())
			return
		}
		switch mg.MgType {
		// 请求同步
		case MsgSync:
			logging.RuntimeLog.Infof("file sync from %s", conn.RemoteAddr().String())
			logging.CLILog.Infof("file sync from %s", conn.RemoteAddr().String())
			hdSync(gbc)
		// 请求文件传输
		case MsgTran:
			hdTranFile(&mg, gbc)
		// 结束
		case MsgEnd:
			return
		// 未知消息
		default:
			hdNoType(gbc)
			return
		}
	}
}

// hdSync 处理worker的全部文件同步请求
func hdSync(gbc *GobConn) {
	srcPath, err := filepath.Abs(conf.GetRootPath())
	if err != nil {
		logging.CLILog.Error(err)
		writeErrorMg(err.Error(), gbc)
		return
	}
	//srcPath := "/tmp/test/src"
	fileMd5List, err := Traverse(srcPath)
	if err != nil {
		logging.CLILog.Error(err)
		writeErrorMg(err.Error(), gbc)
		return
	}
	if len(fileMd5List) == 0 {
		writeErrorMg("emtry file list", gbc)
		return
	}
	cr := Message{
		MgStrings: fileMd5List,
		MgType:    MsgMd5List,
		Overwrite: true,
	}
	err = gbc.gobConnWt(cr)
	if err != nil {
		logging.CLILog.Error(err)
		return
	}
}

// hdTranFile 向worker同步一个文件
func hdTranFile(mg *Message, gbc *GobConn) {
	if len(mg.MgString) <= 0 {
		writeErrorMg("no file to transfer", gbc)
		return
	}
	if checkFileIsSyncWhileList(mg.MgString) == false {
		writeErrorMg("invalid file or path to sync", gbc)
		return
	}

	//srcPath := "/tmp/test/src"
	srcPath, err := filepath.Abs(conf.GetRootPath())
	if err != nil {
		logging.CLILog.Error(err)
		writeErrorMg(err.Error(), gbc)
		return
	}
	srcPathFileName := filepath.Join(srcPath, mg.MgString)
	var cr Message
	st, err := os.Stat(srcPathFileName)
	if err != nil {
		writeErrorMg(fmt.Sprintf("read sync file:%s error", err), gbc)
		return
	}
	cr.MgFileMode = st.Mode()
	cr.MgByte, err = os.ReadFile(srcPathFileName)
	if err != nil {
		writeErrorMg(fmt.Sprintf("read sync file:%s error", srcPathFileName), gbc)
		return
	}
	cr.MgType = MsgTranData
	err = gbc.gobConnWt(cr)
	if err != nil {
		logging.CLILog.Error(err)
	}
}

// hdNoType 处理消息类型错误
func hdNoType(gbc *GobConn) {
	writeErrorMg("error, not a recognizable message.", gbc)
}

// writeErrorMg 返回错误信息的消息
func writeErrorMg(message string, gbc *GobConn) {
	var errMsg Message
	errMsg.MgType = MsgError
	errMsg.MgString = message
	sendErr := gbc.gobConnWt(errMsg)
	if sendErr != nil {
		logging.CLILog.Error(sendErr)
	}
}

// checkSyncAuthKey 同步的认证检查
func checkSyncAuthKey(authKey, workerAuthKey string, gbc *GobConn) (success bool) {
	if workerAuthKey == authKey {
		return true
	}
	writeErrorMg("authKey error!", gbc)
	return false
}
