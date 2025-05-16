package core

import (
	"errors"
	"flag"
	"fmt"
	"github.com/hanc00l/nemo_go/v3/pkg/conf"
	"github.com/hanc00l/nemo_go/v3/pkg/logging"
	"github.com/hanc00l/nemo_go/v3/pkg/resource"
	"github.com/hanc00l/nemo_go/v3/pkg/utils"
	"github.com/jessevdk/go-flags"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var cmd *exec.Cmd
var WorkerName string
var DaemonRunOption *WorkerOption
var WorkerRunOption *WorkerOption

func PrepareWorkerConfig(opts *WorkerOption) bool {
	conf.WorkerDefaultConfigFile = opts.ConfigFile
	if opts.ConfigFile != "" {
		if err := conf.GlobalWorkerConfig().ReloadConfig(); err != nil {
			logging.RuntimeLog.Errorf("load local config file:%s error:%s", conf.WorkerDefaultConfigFile, err.Error())
			return false
		}
	} else {
		if !PrepareWorkerServiceOptions(opts) {
			return false
		}
		var workerConfig conf.Worker
		// CallXClient需要用到service host,port,auth，这里先设置使用本地参数
		workerConfig.Service.Host = opts.ServiceHost
		workerConfig.Service.Port = opts.ServicePort
		workerConfig.Service.AuthKey = opts.ServiceAuth
		conf.GlobalWorkerConfig().SetWorkerConfig(&workerConfig)
		err := CallXClient("LoadWorkerConfig", nil, &workerConfig)
		if err != nil {
			logging.RuntimeLog.Errorf("load config from server:%s error:%s", opts.ServiceHost, err.Error())
			return false
		}
		// 设置worker的运行选项
		conf.GlobalWorkerConfig().SetWorkerConfig(&workerConfig)
		// 使用本地配置参数避免被服务器配置覆盖
		workerConfig.Service.Host = opts.ServiceHost
		workerConfig.Service.Port = opts.ServicePort
		workerConfig.Service.AuthKey = opts.ServiceAuth
	}

	return true
}

func PrepareWorkerServiceOptions(opts *WorkerOption) bool {
	var err error
	if opts.ServiceHost == "" {
		opts.ServiceHost = os.Getenv(EnvServiceHost)
		if os.Getenv(EnvServicePort) != "" {
			if opts.ServicePort, err = strconv.Atoi(os.Getenv(EnvServicePort)); err != nil {
				logging.CLILog.Error("error service port...")
				return false
			}
		}
	}
	if opts.ServiceAuth == "" {
		opts.ServiceAuth = os.Getenv(EnvServiceAuth)
	}
	//check service host,port,auth
	if opts.ServicePort == 0 || opts.ServiceHost == "" || opts.ServiceAuth == "" {
		logging.CLILog.Error("error service host or port or auth...")
		return false
	}
	//
	return true
}

func PrepareWorkerOptions() *WorkerOption {
	var opts WorkerOption
	parser := flags.NewParser(&opts, flags.Default)
	if _, err := parser.Parse(); err != nil {
		if !errors.Is(err.(*flags.Error).Type, flag.ErrHelp) {
			return nil
		}
	}
	conf.NoProxyByCmd = opts.NoProxy
	conf.Ipv6Support = opts.IpV6Support
	//
	if !PrepareWorkerConfig(&opts) {
		return nil
	}
	opts.WorkerTopic = make(map[string]struct{})
	if opts.WorkerRunTaskMode == "0" {
		opts.WorkerTopic[TopicActive] = struct{}{}
		opts.WorkerTopic[TopicFinger] = struct{}{}
		opts.WorkerTopic[TopicPassive] = struct{}{}
		opts.WorkerTopic[TopicPocscan] = struct{}{}
	} else {
		for _, mode := range strings.Split(opts.WorkerRunTaskMode, ",") {
			m, err := strconv.Atoi(mode)
			if err != nil {
				logging.CLILog.Error("error worker run task mode...")
				return nil
			}
			if m <= int(TaskModeDefault) || m > int(TaskModeStandalone) {
				logging.CLILog.Error("error worker run task mode...")
				return nil
			}
			switch WorkerRunTaskMode(m) {
			case TaskModeActive:
				opts.WorkerTopic[TopicActive] = struct{}{}
			case TaskModeFinger:
				opts.WorkerTopic[TopicFinger] = struct{}{}
			case TaskModePassive:
				opts.WorkerTopic[TopicPassive] = struct{}{}
			case TaskModePocscan:
				opts.WorkerTopic[TopicPocscan] = struct{}{}
			default:
				logging.CLILog.Error("error worker run task mode...")
				return nil
			}
		}
	}
	if len(opts.WorkerTopic) == 0 {
		logging.CLILog.Error("error worker run task mode...")
		return nil
	}

	return &opts
}

func ReloadWorkerRunEnv() bool {
	// 增加误删除检查
	var skipRemoveResource bool
	workerBin := utils.GetThirdpartyBinNameByPlatform(utils.Worker)
	serverBin := utils.GetThirdpartyBinNameByPlatform(utils.Server)
	workerPathName, err1 := filepath.Abs(filepath.Join(conf.GetRootPath(), workerBin))
	serverPathName, err2 := filepath.Abs(filepath.Join(conf.GetRootPath(), serverBin))
	if err1 == nil && err2 == nil && utils.CheckFileExist(serverPathName) && utils.CheckFileExist(workerPathName) {
		skipRemoveResource = true
		logging.CLILog.Info("Wanning: worker与server文件同时存在，为防止误删除文件，将不执行清空旧文件操作")
		logging.RuntimeLog.Info("Wanning: worker与server文件同时存在，为防止误删除文件，将不执行清空旧文件操作")
	}
	if !skipRemoveResource {
		// 清除旧的worker可执行文件及thirdparty目录
		err := os.RemoveAll(workerPathName)
		if err != nil {
			logging.RuntimeLog.Errorf(err.Error())
			return false
		}
		thirdpartyPath, err := filepath.Abs(filepath.Join(conf.GetRootPath(), "thirdparty"))
		if err != nil {
			logging.RuntimeLog.Error(err)
			return false
		}
		err = os.RemoveAll(thirdpartyPath)
		if err != nil {
			logging.RuntimeLog.Errorf(err.Error())
			return false
		}
	}
	// 下载worker可执行文件
	return PrepareWorkerRunEnv()
}

func PrepareWorkerRunEnv() bool {
	// 准备日志目录
	logPath, err := filepath.Abs(filepath.Join(conf.GetRootPath(), "log"))
	if err != nil {
		logging.RuntimeLog.Error(err)
		return false
	}
	if !utils.MakePath(logPath) {
		logging.RuntimeLog.Error("创建日志文件路径失败")
		return false
	}
	// 准备worker可执行文件
	workerBin := utils.GetThirdpartyBinNameByPlatform(utils.Worker)
	workerPathName, err := filepath.Abs(filepath.Join(conf.GetRootPath(), workerBin))
	if err != nil {
		logging.RuntimeLog.Error(err)
		return false
	}
	if utils.CheckFileExist(workerPathName) {
		return true
	}
	var re []RequiredResource
	re = append(re, RequiredResource{
		Category: resource.WorkerCategory,
		Name:     workerBin,
	})
	err = CheckRequiredResource(re, false)
	if err != nil {
		logging.RuntimeLog.Errorf("获取资源失败:%s", err.Error())
		return false
	}
	return true
}

func SyncConfigAndDataFileResource() error {
	var re []RequiredResource
	for category, res := range resource.Resources {
		for n, r := range res {
			if r.Type == resource.Dir || r.Type == resource.DataFile || r.Type == resource.ConfigFile {
				re = append(re, RequiredResource{
					Category: category,
					Name:     n,
				})
			}
		}
	}
	// 同步：对已存在的资源进行更新
	return CheckRequiredResource(re, true)
}

func CheckWorkerPerformance(workerPerformance int) {
	switch workerPerformance {
	case 0:
		cpuNumber, err1 := cpu.Counts(true)
		memInfo, err2 := mem.VirtualMemory()
		if err1 != nil || err2 != nil {
			break
		}
		if cpuNumber >= 4 && memInfo.Total >= 4*1024*1024*1024 {
			conf.WorkerPerformanceMode = conf.HighPerformance
		}
	case 1:
		conf.WorkerPerformanceMode = conf.HighPerformance
	}
	return
}

func SetupCloseHandler() {
	quitSignal := make(chan os.Signal, 1)
	signal.Notify(quitSignal, os.Interrupt, syscall.SIGTERM)
	<-quitSignal
	logging.CLILog.Info("Ctrl+C pressed in Terminal,waiting for worker exit...")
	logging.RuntimeLog.Info("Ctrl+C pressed in Terminal,waiting for worker exit...")
	os.Exit(0)
}

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
			DaemonStartWorker()
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
				DaemonStartWorker()
			}
		}
	}
}

// StartWorkerDaemon 启动worker的daemon
func StartWorkerDaemon() {
	if success := DaemonStartWorker(); success == false {
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
			// 更新运行参数
			DaemonRunOption.ServiceHost = replay.WorkerRunOption.ServiceHost
			DaemonRunOption.ServicePort = replay.WorkerRunOption.ServicePort
			DaemonRunOption.ServiceAuth = replay.WorkerRunOption.ServiceAuth
			DaemonRunOption.NoProxy = replay.WorkerRunOption.NoProxy
			DaemonRunOption.NoRedisProxy = replay.WorkerRunOption.NoRedisProxy
			DaemonRunOption.IpV6Support = replay.WorkerRunOption.IpV6Support
			DaemonRunOption.Concurrency = replay.WorkerRunOption.Concurrency
			DaemonRunOption.WorkerPerformance = replay.WorkerRunOption.WorkerPerformance
			DaemonRunOption.ConfigFile = replay.WorkerRunOption.ConfigFile
			DaemonRunOption.WorkerRunTaskMode = replay.WorkerRunOption.WorkerRunTaskMode
			//更新运行参数后，强制重启worker
			replay.ManualReloadFlag = true
		}
		// 同步配置和poc文件
		if replay.ManualConfigAndPocSyncFlag {
			logging.CLILog.Info("开始同步配置、Poc文件...")
			logging.RuntimeLog.Info("开始同步配置、Poc文件...")
			err = SyncConfigAndDataFileResource()
			if err != nil {
				logging.RuntimeLog.Errorf("同步配置文件、Poc失败：%v", err)
			}
		}
		// 收到server的手动重启worker命令，执行停止worker、文件同步、重启worker
		if replay.ManualReloadFlag || replay.ManualInitEnvFlag {
			if KillWorker() {
				if replay.ManualInitEnvFlag {
					logging.CLILog.Info("开始重置worker环境...")
					logging.RuntimeLog.Info("开始重置worker环境...")
					// 强制更新：删除worker和thirdparty目录，重新请求worker二进制文件
					ReloadWorkerRunEnv()
				}
				// 重样报启动worker
				DaemonStartWorker()
			}
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

// DaemonStartWorker 启动worker进程
func DaemonStartWorker() bool {
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
	}
	if DaemonRunOption.ServiceHost != "" {
		cmdArgs = append(cmdArgs, "--service", DaemonRunOption.ServiceHost)
	}
	if DaemonRunOption.ServicePort != 5001 {
		cmdArgs = append(cmdArgs, "--port", fmt.Sprintf("%d", DaemonRunOption.ServicePort))
	}
	if DaemonRunOption.ServiceAuth != "" {
		cmdArgs = append(cmdArgs, "--auth", DaemonRunOption.ServiceAuth)
	}
	if DaemonRunOption.ConfigFile != "" {
		cmdArgs = append(cmdArgs, "-f", DaemonRunOption.ConfigFile)
	}
	if DaemonRunOption.NoProxy {
		cmdArgs = append(cmdArgs, "--no-proxy")
	}
	if DaemonRunOption.NoRedisProxy {
		cmdArgs = append(cmdArgs, "--no-redis-proxy")
	}
	if DaemonRunOption.IpV6Support {
		cmdArgs = append(cmdArgs, "--ipv6")
	}
	cmd = exec.Command(workerPathName, cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err = cmd.Start(); err != nil {
		logging.CLILog.Infof("start worker fail: %v", err.Error())
		logging.RuntimeLog.Infof("start worker fail: %v", err.Error())
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
