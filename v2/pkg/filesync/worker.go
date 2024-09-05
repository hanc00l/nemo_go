// forked from https://github.com/ren-zc/gosync

package filesync

import (
	"crypto/tls"
	"fmt"
	"github.com/hanc00l/nemo_go/v2/pkg/conf"
	"github.com/hanc00l/nemo_go/v2/pkg/logging"
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
	var needCreateDir []string
	needCreateDir, _, transFiles = doDiff(mg.MgStrings, localFilesMd5)
	//fmt.Println("need add:", needCreateDir)
	//fmt.Println("need delete:", needDelete)
	//fmt.Println("need transfer:", transFiles)
	// 接收新文件的本地操作
	fErr := os.Chdir(srcPath)
	if fErr != nil {
		logging.CLILog.Error(err)
		logging.RuntimeLog.Error(err)
		return nil, fErr
	}
	defer os.Chdir(cwd)

	err = localOP(nil, nil, nil, needCreateDir)
	if err != nil {
		logging.CLILog.Error(err)
		logging.RuntimeLog.Error(err)
		return
	}
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

// doDiff 对比需要同步的文件列表
func doDiff(src []string, dst []string) (needCreate []string, needDelete []string, needTransfer []string) {
	srcFileMap := make(map[string]struct{})
	srcDifMap := make(map[string]string)
	dstDifMap := make(map[string]string)
	//1，先过滤掉完全相同的条件（文件名,,md5/Directory）,并生成src与dst中不同的项（可能是文件名不同，也可能是md5值不同）
	for _, s := range src {
		srcFileMap[s] = struct{}{}
	}
	for _, d := range dst {
		_, ok := srcFileMap[d]
		if ok {
			// 删除src与dst相同的条目
			delete(srcFileMap, d)
		} else {
			arr := strings.Split(d, ",,")
			if len(arr) == 2 {
				dstDifMap[arr[0]] = arr[1]
			}
		}
	}
	for k := range srcFileMap {
		arr := strings.Split(k, ",,")
		if len(arr) == 2 {
			srcDifMap[arr[0]] = arr[1]
		}
	}
	//2. 检查map中的源的文件名称
	srcFileName := make([]string, 0, len(srcDifMap))
	for s := range srcDifMap {
		srcFileName = append(srcFileName, s)
	}
	for _, s := range srcFileName {
		// 文件名存在于源和目的
		_, ok := dstDifMap[s]
		if ok {
			// md5不一样，需传输
			needTransfer = append(needTransfer, s)
			delete(dstDifMap, s)
		} else {
			//目的中不存在：如果是目录，则需创建；如果不是目录，需要传输
			if srcDifMap[s] == "Directory" {
				needCreate = append(needCreate, s)
			} else {
				needTransfer = append(needTransfer, s)
			}
		}
	}
	//3. 需删除的文件/路径
	for k := range dstDifMap {
		needDelete = append(needDelete, k)
	}

	return
}
