package comm

import (
	"fmt"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/ampq"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"sync"
	"time"
)

type KeepAliveWorkerInfo struct {
	WorkerStatus ampq.WorkerStatus
}

type KeepAliveDaemonInfo struct {
	ManualReloadFlag       bool
	ManualFileSyncFlag     bool
	ManualUpdateOptionFlag bool
	WorkerRunOption        *WorkerOption
}

var (
	WorkerStatusMutex sync.Mutex
	WorkerStatus      = make(map[string]*ampq.WorkerStatus)
)

// DoKeepAlive worker请求keepAlive
func DoKeepAlive(ws *ampq.WorkerStatus) bool {
	kari := &KeepAliveWorkerInfo{
		WorkerStatus: *ws,
	}
	kari.WorkerStatus.UpdateTime = time.Now()
	if cpuLoad, err := load.Avg(); err == nil {
		kari.WorkerStatus.CPULoad = fmt.Sprintf("%.2f%%", cpuLoad.Load1)
	}
	if memLoad, err := mem.VirtualMemory(); err == nil {
		kari.WorkerStatus.MemUsed = fmt.Sprintf("%.2f%c/%.2f%c", float64(memLoad.Used)/1024/1024/1024, 'G', float64(memLoad.Total)/1024/1024/1024, 'G')
	}

	var replay string
	if err := CallXClient("KeepAlive", kari, &replay); err != nil {
		logging.RuntimeLog.Errorf("keep alive fail:%v", err)
		logging.CLILog.Errorf("keep alive fail:%v", err)
		return false
	}
	return true
}

// DoDaemonKeepAlive worker请求keepAlive
func DoDaemonKeepAlive() (replay KeepAliveDaemonInfo, err error) {
	if err = CallXClient("KeepDaemonAlive", &WorkerName, &replay); err != nil {
		logging.CLILog.Errorf("keep daemon alive fail:%v", err)
		logging.RuntimeLog.Errorf("keep daemon alive fail:%v", err)
	}
	return
}
