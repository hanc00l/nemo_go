// forked from https://github.com/ren-zc/gosync

package filesync

import (
	"crypto/tls"
	"fmt"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/jacenr/filediff/diff"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// WorkerStartupSync worker在启动时进行文件同步
func WorkerStartupSync(host, port, authKey string) {
	serverAddr := fmt.Sprintf("%s:%s", host, port)
	// 1 连接到server
	var conn net.Conn
	var err error
	if TLSEnabled {
		conn, err = tls.Dial("tcp", serverAddr, &tls.Config{InsecureSkipVerify: true})
	} else {
		conn, err = net.Dial("tcp", serverAddr)
	}
	if err != nil {
		logging.CLILog.Error(err)
		logging.RuntimeLog.Error(err)
		return
	}
	defer conn.Close()

	gbc := initGobConn(conn)
	// 2 发送SYNC请求
	msgSync := Message{MgType: MsgSync, MgAuthKey: authKey}
	encErr := gbc.gobConnWt(msgSync)
	if encErr != nil {
		logging.CLILog.Error(encErr)
		logging.RuntimeLog.Error(encErr)
		return
	}
	// 3 服务器返回信息
	var hostMessage Message
	err = gbc.Dec.Decode(&hostMessage)
	if err != nil {
		logging.CLILog.Error(err)
		logging.RuntimeLog.Error(err)
		return
	}
	if hostMessage.MgType != MsgMd5List {
		logging.CLILog.Errorf("get sync file fail:%s", hostMessage.MgString)
		logging.RuntimeLog.Errorf("get sync file fail:%s", hostMessage.MgString)
		return
	}
	// 4 获取服务器所有文件及md5值,并预处理本地的路径和文件
	transFiles, err := doFileMd5List(&hostMessage)
	if err == nil {
		logging.CLILog.Infof("file needed sync: %d", len(transFiles))
		// 5 同步文件
		for i, file := range transFiles {
			status := doTranFile(file, authKey, gbc)
			logging.CLILog.Infof("%d %s %v", i+1, file, status)

		}
		logging.RuntimeLog.Info("finish file sync")
		logging.CLILog.Info("finish file sync")
	}
	// 6 结束同步
	endMsg := Message{MgType: MsgEnd, MgAuthKey: authKey}
	encErr = gbc.gobConnWt(endMsg)
	if encErr != nil {
		logging.CLILog.Error(encErr)
		logging.RuntimeLog.Error(encErr)
	}
}

// doFileMd5List 读取worker本地文件列表及md5值，并与服务端进行对比，确定需要同步的文件列表
func doFileMd5List(mg *Message) (transFiles []string, err error) {
	var slinkNeedCreat = make(map[string]string)
	var slinkNeedChange = make(map[string]string)
	var needDelete = make([]string, 0)
	var needCreDir = make([]string, 0)

	//srcPath := "/tmp/test/dst"
	var srcPath string
	srcPath, err = filepath.Abs(conf.GetRootPath())
	if err != nil {
		logging.RuntimeLog.Error(err)
		logging.CLILog.Error(err)
		return
	}
	// 遍历本地目标路径失败
	localFilesMd5, err := Traverse(srcPath)
	// DebugInfor(localFilesMd5)
	if err != nil {
		msg := "traverse in worker failure"
		logging.CLILog.Error(msg)
		logging.RuntimeLog.Error(msg)
		return
	}
	sort.Strings(localFilesMd5)
	var diffrm []string
	var diffadd []string
	if len(localFilesMd5) != 0 {
		diffrm, diffadd = diff.DiffOnly(mg.MgStrings, localFilesMd5)
	} else {
		diffrm, diffadd = mg.MgStrings, localFilesMd5
	}
	if len(diffrm) == 0 && len(diffadd) == 0 {
		logging.CLILog.Info("No file need be transfored.")
		return
	}
	//fmt.Println(diffrm)
	//fmt.Println(diffadd)
	// 重组成map
	diffrmM := make(map[string]string)
	diffaddM := make(map[string]string)
	for _, v := range diffrm {
		s := strings.Split(v, ",,")
		if len(s) != 1 {
			diffrmM[s[0]] = s[1]
		}
	}
	for _, v := range diffadd {
		s := strings.Split(v, ",,")
		if len(s) != 1 {
			diffaddM[s[0]] = s[1]
		}
	}
	// 整理
	for k, _ := range diffaddM {
		v2, ok := diffrmM[k]
		if ok {
			if !mg.Overwrite {
				delete(diffrmM, k)
			}
			if mg.Overwrite {
				if strings.HasPrefix(v2, "symbolLink&&") {
					slinkNeedChange[k] = strings.TrimPrefix(v2, "symbolLink&&")
					delete(diffrmM, k)
				}
				needDelete = append(needDelete, k)
			}
		}
		if !ok && mg.Del {
			needDelete = append(needDelete, k)
		}

	}
	for k, v := range diffrmM {
		if strings.HasPrefix(v, "symbolLink&&") {
			slinkNeedCreat[k] = strings.TrimPrefix(v, "symbolLink&&")
			delete(diffrmM, k)
			continue
		}
		if v == "Directory" {
			needCreDir = append(needCreDir, k)
			delete(diffrmM, k)
		}
	}
	// 接收新文件的本地操作
	fErr := os.Chdir(srcPath)
	if fErr != nil {
		logging.CLILog.Error(err)
		logging.RuntimeLog.Error(err)
		return nil, fErr
	}
	defer os.Chdir(cwd)

	err = localOP(slinkNeedCreat, slinkNeedChange, needDelete, needCreDir)
	if err != nil {
		logging.CLILog.Error(err)
		logging.RuntimeLog.Error(err)
		return
	}
	// do request needTrans files
	for k, _ := range diffrmM {
		transFiles = append(transFiles, k)
	}
	//DebugInfor(transFiles)
	sort.Strings(transFiles)

	return
}

// doTranFile worker向server请求同步一个文件
func doTranFile(filePathName, authKey string, gbc *GobConn) bool {
	var err error
	var srcPath string
	//dstPath := "/tmp/test/dst"
	srcPath, err = filepath.Abs(conf.GetRootPath())
	if err != nil {
		logging.CLILog.Error(err)
		logging.RuntimeLog.Error(err)
		return false
	}
	mg := Message{
		MgAuthKey: authKey,
		MgType:    MsgTran,
		MgString:  filePathName,
	}
	err = gbc.gobConnWt(mg)
	if err != nil {
		logging.CLILog.Error(err)
		logging.RuntimeLog.Error(err)
		return false
	}
	var hostMessage Message
	err = gbc.Dec.Decode(&hostMessage)
	if err != nil {
		logging.CLILog.Error(err)
		logging.RuntimeLog.Error(err)
		return false
	}
	if hostMessage.MgType == MsgTranData {
		dstFilePathName := filepath.Join(srcPath, filePathName)
		err = os.WriteFile(dstFilePathName, hostMessage.MgByte, hostMessage.MgFileMode)
		if err != nil {
			logging.CLILog.Error(err)
			logging.RuntimeLog.Error(err)
			return false
		}
		return true
	}
	logging.CLILog.Error(hostMessage.MgString)
	logging.RuntimeLog.Error(hostMessage.MgString)
	return false
}
