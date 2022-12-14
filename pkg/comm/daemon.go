package comm

import (
	"fmt"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/filesync"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/utils"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

var cmd *exec.Cmd
var WorkerName string

// StartWorkerDaemon 启动worker的daemon
func StartWorkerDaemon(concurrency int, workerPerformance int, noFilesync bool) {
	// 1、worker与server文件同步
	fileSyncServer := conf.GlobalWorkerConfig().FileSync
	if noFilesync == false {
		logging.CLILog.Info("开始文件同步...")
		filesync.WorkerStartupSync(fileSyncServer.Host, fmt.Sprintf("%d", fileSyncServer.Port), fileSyncServer.AuthKey)
	}
	// 2、启动worker
	if success := StartWorker(concurrency, workerPerformance); success == false {
		return
	}
	// 3、心跳并接收命令
	for {
		time.Sleep(15 * time.Second)
		replay, err := DoDaemonKeepAlive()
		if err != nil {
			logging.RuntimeLog.Error("deamon keep alive fail")
			logging.CLILog.Error("deamon keep alive fail")
			continue
		}
		// 收到server的手动重启worker命令，执行停止worker、文件同步、重启worker
		if replay.ManualReloadFlag {
			if KillWorker() {
				//1、同步文件
				if noFilesync == false {
					logging.CLILog.Info("开始文件同步...")
					filesync.WorkerStartupSync(fileSyncServer.Host, fmt.Sprintf("%d", fileSyncServer.Port), fileSyncServer.AuthKey)
				}
				//2、重新启动worker
				StartWorker(concurrency, workerPerformance)
			}
			// 忽略文件同步（如果有）
			continue
		}
		if noFilesync == false && replay.ManualFileSyncFlag {
			//同步文件
			logging.CLILog.Info("开始文件同步...")
			filesync.WorkerStartupSync(fileSyncServer.Host, fmt.Sprintf("%d", fileSyncServer.Port), fileSyncServer.AuthKey)
		}
	}
}

// KillWorker 停止当前worker进程
func KillWorker() bool {
	if cmd == nil {
		return true
	}
	err := cmd.Process.Kill()
	if err != nil {
		logging.CLILog.Errorf("kill worker fail,pid:%d,%v", cmd.Process.Pid, err)
		return false
	}
	if err = cmd.Wait(); err != nil {
		logging.CLILog.Infof("kill worker pid:%d,%v", cmd.Process.Pid, err)
	}
	logging.CLILog.Infof("kill worker ok,pid:%d", cmd.Process.Pid)
	logging.RuntimeLog.Infof("kill worker ok,pid:%d", cmd.Process.Pid)

	cmd = nil
	return true
}

// StartWorker 启动worker进程
func StartWorker(concurrency int, workerPerformance int) bool {
	workerBin := "worker_darwin_amd64"
	if runtime.GOOS == "linux" {
		workerBin = "worker_linux_amd64"
	} else if runtime.GOOS == "windows" {
		workerBin = "worker_windows_amd64.exe"
	}
	//绝对路径
	workerPathName, err := filepath.Abs(filepath.Join(conf.GetRootPath(), workerBin))
	if err != nil {
		logging.CLILog.Error(err.Error())
		return false
	}
	cmd = exec.Command(workerPathName, "-c", fmt.Sprintf("%d", concurrency), "-p", fmt.Sprintf("%d", workerPerformance))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err = cmd.Start(); err != nil {
		logging.CLILog.Infof("start worker fail: %v", err)
		logging.RuntimeLog.Infof("start worker fail: %v", err)
		return false
	}
	hostIP, _ := utils.GetOutBoundIP()
	if hostIP == "" {
		hostIP, _ = utils.GetClientIp()
	}

	hostName, _ := os.Hostname()
	WorkerName = fmt.Sprintf("%s@%s#%d", hostName, hostIP, cmd.Process.Pid)

	logging.CLILog.Infof("start worker pid: %d", cmd.Process.Pid)
	logging.RuntimeLog.Infof("start worker pid: %d", cmd.Process.Pid)
	return true
}
