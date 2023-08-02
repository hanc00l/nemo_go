package comm

import (
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/ampq"
	"sync"
	"time"
)

type KeepAliveInfo struct {
	WorkerStatus ampq.WorkerStatus
}

type WorkerDaemonManualInfo struct {
	ManualReloadFlag   bool `json:"manual_reload_flag"`
	ManualFileSyncFlag bool `json:"manual_file_sync_flag"`
}

var (
	WorkerStatusMutex sync.Mutex
	WorkerStatus      = make(map[string]*ampq.WorkerStatus)
)

// DoKeepAlive worker请求keepAlive
func DoKeepAlive(ws ampq.WorkerStatus) bool {
	kari := newKeepAliveRequestInfo(ws)
	var replay string

	if err := CallXClient("KeepAlive", &kari, &replay); err != nil {
		logging.RuntimeLog.Errorf("keep alive fail:%v", err)
		logging.CLILog.Errorf("keep alive fail:%v", err)
		return false
	}
	return true
}

// DoDaemonKeepAlive worker请求keepAlive
func DoDaemonKeepAlive() (replay WorkerDaemonManualInfo, err error) {
	if err = CallXClient("KeepDaemonAlive", &WorkerName, &replay); err != nil {
		logging.CLILog.Errorf("keep daemon alive fail:%v", err)
		logging.RuntimeLog.Errorf("keep daemon alive fail:%v", err)
		return
	}
	return
}

// newKeepAliveRequestInfo worker请求的keepAlive数据
func newKeepAliveRequestInfo(ws ampq.WorkerStatus) KeepAliveInfo {
	kai := KeepAliveInfo{
		WorkerStatus: ws,
	}
	kai.WorkerStatus.UpdateTime = time.Now()
	return kai
}
