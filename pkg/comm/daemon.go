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
	"time"
)

var cmd *exec.Cmd
var WorkerName string

// StartWorkerDaemon 启动worker的daemon
func StartWorkerDaemon(workerRunTaskMode, concurrency, workerPerformance int, noFilesync bool) {
	// 1、worker与server文件同步
	fileSyncServer := conf.GlobalWorkerConfig().FileSync
	if noFilesync == false {
		logging.CLILog.Info("start file sync...")
		logging.RuntimeLog.Info("start file sync...")
		filesync.WorkerStartupSync(fileSyncServer.Host, fmt.Sprintf("%d", fileSyncServer.Port), fileSyncServer.AuthKey)
	}
	// 2、启动worker
	if success := StartWorker(workerRunTaskMode, concurrency, workerPerformance); success == false {
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
					logging.CLILog.Info("manual reload to start file sync...")
					logging.RuntimeLog.Info("manual reload to start file sync...")
					filesync.WorkerStartupSync(fileSyncServer.Host, fmt.Sprintf("%d", fileSyncServer.Port), fileSyncServer.AuthKey)
				}
				//2、重新启动worker
				StartWorker(workerRunTaskMode, concurrency, workerPerformance)
			}
			// 忽略文件同步（如果有）
			continue
		}
		if noFilesync == false && replay.ManualFileSyncFlag {
			//同步文件
			logging.CLILog.Info("manual reload to start file sync...")
			logging.RuntimeLog.Info("manual reload to start file sync...")
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
		msg := fmt.Sprintf("kill worker fail,pid:%d,%v", cmd.Process.Pid, err)
		logging.RuntimeLog.Error(msg)
		logging.CLILog.Error(msg)
		return false
	}
	if err = cmd.Wait(); err != nil {
		msg := fmt.Sprintf("kill worker pid:%d,%v", cmd.Process.Pid, err)
		logging.RuntimeLog.Error(msg)
		logging.CLILog.Error(msg)
	}
	msg := fmt.Sprintf("kill worker ok,pid:%d", cmd.Process.Pid)
	logging.RuntimeLog.Info(msg)
	logging.CLILog.Info(msg)

	cmd = nil
	return true
}

// StartWorker 启动worker进程
func StartWorker(workerRunTaskMode, concurrency, workerPerformance int) bool {
	workerBin := utils.GetThirdpartyBinNameByPlatform(utils.Worker)
	//绝对路径
	workerPathName, err := filepath.Abs(filepath.Join(conf.GetRootPath(), workerBin))
	if err != nil {
		logging.RuntimeLog.Error(err)
		logging.CLILog.Error(err)
		return false
	}
	cmd = exec.Command(workerPathName, "-c", fmt.Sprintf("%d", concurrency), "-p", fmt.Sprintf("%d", workerPerformance), "-m", fmt.Sprintf("%d", workerRunTaskMode))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err = cmd.Start(); err != nil {
		logging.CLILog.Infof("start worker fail: %v", err)
		logging.RuntimeLog.Infof("start worker fail: %v", err)
		return false
	}
	WorkerName = GetWorkerNameByDaemon()
	logging.CLILog.Infof("start worker pid: %d", cmd.Process.Pid)
	logging.RuntimeLog.Infof("start worker pid: %d", cmd.Process.Pid)
	return true
}

func GetWorkerNameByDaemon() string {
	return getWorkerName(cmd.Process.Pid)
}

func GetWorkerNameBySelf() string {
	return getWorkerName(os.Getpid())
}

func getWorkerName(pid int) string {
	hostIP, _ := utils.GetOutBoundIP()
	if hostIP == "" {
		hostIP, _ = utils.GetClientIp()
	}
	hostName, _ := os.Hostname()

	return fmt.Sprintf("%s@%s#%d", hostName, hostIP, pid)
}
