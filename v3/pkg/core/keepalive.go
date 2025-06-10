package core

import (
	"fmt"
	"github.com/hanc00l/nemo_go/v3/pkg/logging"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"time"
)

type KeepAliveDaemonInfo struct {
	ManualStopFlag             bool // 停止
	ManualReloadFlag           bool // 重启
	ManualConfigAndPocSyncFlag bool // 同步已存在的配置、Poc
	ManualInitEnvFlag          bool // 重置
	ManualUpdateOptionFlag     bool // 更新启动参数
	WorkerRunOption            *WorkerOption
}

// DoKeepAlive worker请求keepAlive
func DoKeepAlive(ws *WorkerStatus) bool {
	ws.UpdateTime = time.Now()
	if cpuLoad, err := load.Avg(); err == nil {
		ws.CPULoad = fmt.Sprintf("%.2f%%", cpuLoad.Load1)
	}
	if memLoad, err := mem.VirtualMemory(); err == nil {
		ws.MemUsed = fmt.Sprintf("%.2f%c/%.2f%c", float64(memLoad.Used)/1024/1024/1024, 'G', float64(memLoad.Total)/1024/1024/1024, 'G')
	}

	var replay string
	if err := CallXClient("KeepAlive", ws, &replay); err != nil {
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
