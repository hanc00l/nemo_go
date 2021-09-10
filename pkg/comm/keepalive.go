package comm

import (
	"context"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/ampq"
	"github.com/hanc00l/nemo_go/pkg/utils"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type KeepAliveInfo struct {
	WorkerStatus ampq.WorkerStatus
	CustomFiles  map[string]string
}

var WorkerStatusMutex sync.Mutex
var WorkerStatus = make(map[string]*ampq.WorkerStatus )

// asynFiles 需要同步的自定义配置文件
var asynFiles = []string{
	"thirdparty/custom/honeypot.txt",
	"thirdparty/custom/iplocation-custom.txt",
	"thirdparty/custom/iplocation-custom-B.txt",
	"thirdparty/custom/iplocation-custom-C.txt",
	"thirdparty/custom/services-custom.txt",
	"thirdparty/icp/icp.cache",
}

// DoKeepAlive worker请求keepAlive
func DoKeepAlive(ws ampq.WorkerStatus) bool {
	kari := newKeepAliveRequestInfo(ws)
	syncMap := make(map[string]string)

	x := NewXClient()
	err := x.Call(context.Background(), "KeepAlive", &kari, &syncMap)
	if err != nil {
		logging.RuntimeLog.Errorf("keep alive fail:%v", err)
		return false
	}
	// 自定义配置文件的同步
	syncCustomFiles(syncMap)
	return true
}

// newKeepAliveRequestInfo worker请求的keepAlive数据
func newKeepAliveRequestInfo(ws ampq.WorkerStatus) KeepAliveInfo {
	kai := KeepAliveInfo{
		WorkerStatus: ws,
		CustomFiles:  make(map[string]string),
	}
	kai.WorkerStatus.UpdateTime = time.Now()

	for _, file := range asynFiles {
		var txt string
		content, err := os.ReadFile(filepath.Join(conf.GetRootPath(), file))
		if err == nil {
			txt = string(content)
		}
		kai.CustomFiles[file] = utils.MD5(txt)
	}
	return kai
}

// newKeepAliveResponseInfo server响应的keepAlive数据
func newKeepAliveResponseInfo(req map[string]string) map[string]string {
	syncCustomFiles := make(map[string]string)
	for _, file := range asynFiles {
		if _, ok := req[file]; !ok || req[file] == "" {
			continue
		}
		content, err := os.ReadFile(filepath.Join(conf.GetRootPath(), file))
		if err != nil || len(content) == 0 {
			logging.RuntimeLog.Errorf("load custom file %s fail", file)
			continue
		}
		fileHash := utils.MD5(string(content))
		if fileHash == req[file] {
			continue
		}
		syncCustomFiles[file] = string(content)
	}
	return syncCustomFiles
}

// syncCustomFiles 同步自定义配置文件
func syncCustomFiles(cf map[string]string) {
	for _, file := range asynFiles {
		if _, ok := cf[file]; !ok || cf[file] == "" {
			continue
		}
		err := os.WriteFile(filepath.Join(conf.GetRootPath(), file), []byte(cf[file]), 0666)
		if err != nil {
			logging.RuntimeLog.Error(err)
		}
	}
}
