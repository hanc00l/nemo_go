package comm

import (
	"fmt"
	"github.com/hanc00l/nemo_go/v2/pkg/conf"
	"github.com/hanc00l/nemo_go/v2/pkg/filesync"
	"github.com/hanc00l/nemo_go/v2/pkg/logging"
	"github.com/hanc00l/nemo_go/v2/pkg/utils"
	"github.com/shirou/gopsutil/v3/process"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type WorkerOption struct {
	Concurrency       int                 `json:"concurrency" form:"concurrency"`
	WorkerPerformance int                 `json:"worker_performance" form:"worker_performance"`
	WorkerRunTaskMode string              `json:"worker_run_task_mode" form:"worker_run_task_mode"`
	TaskWorkspaceGUID string              `json:"task_workspace_guid" form:"task_workspace_guid"`
	WorkerTopic       map[string]struct{} `json:"-"`
	TLSEnabled        bool                `json:"-"`
	DefaultConfigFile string              `json:"default_config_file" form:"default_config_file"`
	NoProxy           bool                `json:"no_proxy" form:"no_proxy"`
}

type WorkerDaemonOption struct {
	Concurrency       int
	WorkerPerformance int
	NoFilesync        bool
	NoProxy           bool
	WorkerRunTaskMode string
	TaskWorkspaceGUID string
	ManualSyncHost    string
	ManualSyncPort    string
	ManualSyncAuth    string
	TLSEnabled        bool
	DefaultConfigFile string
}

var cmd *exec.Cmd
var WorkerName string
var DaemonRunOption *WorkerDaemonOption
var WorkerRunOption *WorkerOption

// WatchWorkerProcess worker进程状态监控
func WatchWorkerProcess() {
	if cmd == nil {
		return
	}
	// 检查worker进程是否存在
	p, err := process.NewProcess(int32(cmd.Process.Pid))
	if err != nil {
		logging.RuntimeLog.Warning("detected worker process not exist")
		logging.CLILog.Warning("detected worker process not exist")
		if KillWorker() {
			StartWorker()
		}
		return
	}
	// 获取worker进程状态
	status, err1 := p.Status()
	if err1 != nil {
		logging.CLILog.Error(err)
		logging.RuntimeLog.Error(err)
		return
	}
	//fmt.Println(status)
	for _, s := range status {
		// 如果发现进程挂掉了，重启worker
		if s == process.Zombie {
			logging.RuntimeLog.Warning("detected worker zombie status")
			logging.CLILog.Warning("detected worker zombie status")
			if KillWorker() {
				StartWorker()
			}
		}
	}
}

// StartWorkerDaemon 启动worker的daemon
func StartWorkerDaemon() {
	fileSyncServer := conf.GlobalWorkerConfig().FileSync
	if !DaemonRunOption.NoFilesync {
		logging.CLILog.Info("start file sync...")
		filesync.WorkerStartupSync(fileSyncServer.Host, fmt.Sprintf("%d", fileSyncServer.Port), fileSyncServer.AuthKey)
	}
	if success := StartWorker(); success == false {
		return
	}
	for {
		time.Sleep(15 * time.Second)
		replay, err := DoDaemonKeepAlive()
		if err != nil {
			logging.RuntimeLog.Error("daemon keep alive fail")
			logging.CLILog.Error("daemon keep alive fail")
			continue
		}
		WatchWorkerProcess()
		// 收到server更新运行参数的命令
		if replay.ManualUpdateOptionFlag && replay.WorkerRunOption != nil {
			DaemonRunOption.NoProxy = replay.WorkerRunOption.NoProxy
			DaemonRunOption.Concurrency = replay.WorkerRunOption.Concurrency
			DaemonRunOption.WorkerPerformance = replay.WorkerRunOption.WorkerPerformance
			DaemonRunOption.DefaultConfigFile = replay.WorkerRunOption.DefaultConfigFile
			DaemonRunOption.WorkerRunTaskMode = replay.WorkerRunOption.WorkerRunTaskMode
			DaemonRunOption.TaskWorkspaceGUID = replay.WorkerRunOption.TaskWorkspaceGUID
			//更新运行参数后，强制重启worker
			replay.ManualReloadFlag = true
		}
		// 收到server的手动重启worker命令，执行停止worker、文件同步、重启worker
		if replay.ManualReloadFlag {
			if KillWorker() {
				if !DaemonRunOption.NoFilesync {
					logging.CLILog.Info("manual reload to start file sync...")
					logging.RuntimeLog.Info("manual reload to start file sync...")
					filesync.WorkerStartupSync(fileSyncServer.Host, fmt.Sprintf("%d", fileSyncServer.Port), fileSyncServer.AuthKey)
				}
				StartWorker()
			}
			// 忽略文件同步（如果有）
			continue
		}
		if !DaemonRunOption.NoFilesync && replay.ManualFileSyncFlag {
			logging.CLILog.Info("manual start file sync...")
			logging.RuntimeLog.Info("manual start file sync...")
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
		logging.RuntimeLog.Warningf(msg)
		logging.CLILog.Warningf(msg)
	}
	msg := fmt.Sprintf("kill worker ok,pid:%d", cmd.Process.Pid)
	logging.RuntimeLog.Info(msg)
	logging.CLILog.Info(msg)

	cmd = nil
	return true
}

// StartWorker 启动worker进程
func StartWorker() bool {
	workerBin := utils.GetThirdpartyBinNameByPlatform(utils.Worker)
	//绝对路径
	workerPathName, err := filepath.Abs(filepath.Join(conf.GetRootPath(), workerBin))
	if err != nil {
		logging.RuntimeLog.Error(err)
		logging.CLILog.Error(err)
		return false
	}
	var cmdArgs = []string{
		"-c", fmt.Sprintf("%d", DaemonRunOption.Concurrency),
		"-p", fmt.Sprintf("%d", DaemonRunOption.WorkerPerformance),
		"-m", DaemonRunOption.WorkerRunTaskMode,
		"-f", DaemonRunOption.DefaultConfigFile,
	}
	if DaemonRunOption.TaskWorkspaceGUID != "" {
		cmdArgs = append(cmdArgs, "-w", DaemonRunOption.TaskWorkspaceGUID)
	}
	if DaemonRunOption.TLSEnabled {
		cmdArgs = append(cmdArgs, "-tls")
	}
	if DaemonRunOption.NoProxy {
		cmdArgs = append(cmdArgs, "-np")
	}
	cmd = exec.Command(workerPathName, cmdArgs...)
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
