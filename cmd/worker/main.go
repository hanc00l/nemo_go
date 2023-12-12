package main

import (
	"flag"
	"fmt"
	"github.com/hanc00l/nemo_go/pkg/comm"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/ampq"
	"github.com/hanc00l/nemo_go/pkg/task/workerapi"
	"github.com/hanc00l/nemo_go/pkg/utils"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	//_ "net/http/pprof"
	"time"
)

type WorkerOption struct {
	Concurrency       int
	WorkerPerformance int
	WorkerTopic       map[string]struct{}
	TLSEnabled        bool
	DefaultConfigFile string
}

func parseWorkerOptions() *WorkerOption {
	option := &WorkerOption{
		WorkerTopic: make(map[string]struct{}),
	}

	var workerRunTaskMode string
	var taskWorkspaceGUID string
	flag.IntVar(&option.Concurrency, "c", 3, "concurrent number of tasks")
	flag.IntVar(&option.WorkerPerformance, "p", 0, "worker performance,default is autodetect (0:autodetect, 1:high, 2:normal)")
	flag.StringVar(&workerRunTaskMode, "m", "0", "worker run task mode; 0: all, 1:active, 2:finger, 3:passive, 4:pocscan, 5:custom; run multiple mode separated by \",\"")
	flag.StringVar(&taskWorkspaceGUID, "w", "", "workspace guid for custom task; multiple workspace separated by \",\"")
	flag.BoolVar(&option.TLSEnabled, "tls", false, "use TLS for RPC and filesync")
	flag.StringVar(&option.DefaultConfigFile, "f", conf.WorkerDefaultConfigFile, "worker default config file")

	flag.Parse()

	if workerRunTaskMode == "0" {
		option.WorkerTopic[ampq.TopicActive] = struct{}{}
		option.WorkerTopic[ampq.TopicFinger] = struct{}{}
		option.WorkerTopic[ampq.TopicPassive] = struct{}{}
		option.WorkerTopic[ampq.TopicPocscan] = struct{}{}
		option.Concurrency = 2
	} else if workerRunTaskMode == "5" {
		for _, workspaceGUID := range strings.Split(taskWorkspaceGUID, ",") {
			guid := strings.TrimSpace(workspaceGUID)
			// GUID的长度为36个字符
			if len(guid) != 36 {
				logging.CLILog.Error("error workspace guid...")
				return nil
			}
			topicName := fmt.Sprintf("%s.%s", ampq.TopicCustom, guid)
			option.WorkerTopic[topicName] = struct{}{}
		}
	} else {
		for _, mode := range strings.Split(workerRunTaskMode, ",") {
			m, err := strconv.Atoi(mode)
			if err != nil {
				logging.CLILog.Error("error worker run task mode...")
				return nil
			}
			if m <= int(ampq.TaskModeDefault) || m > int(ampq.TaskModeCustom) {
				logging.CLILog.Error("error worker run task mode...")
				return nil
			}
			switch ampq.WorkerRunTaskMode(m) {
			case ampq.TaskModeActive:
				option.WorkerTopic[ampq.TopicActive] = struct{}{}
			case ampq.TaskModeFinger:
				option.WorkerTopic[ampq.TopicFinger] = struct{}{}
			case ampq.TaskModePassive:
				option.WorkerTopic[ampq.TopicPassive] = struct{}{}
			case ampq.TaskModePocscan:
				option.WorkerTopic[ampq.TopicPocscan] = struct{}{}
			}
		}
	}
	if len(option.WorkerTopic) == 0 {
		logging.CLILog.Error("error worker run task mode...")
		return nil
	}
	if option.DefaultConfigFile != conf.WorkerDefaultConfigFile {
		conf.WorkerDefaultConfigFile = option.DefaultConfigFile
	}
	return option
}

// keepAlive worker与server的心跳与同步
func keepAlive() {
	time.Sleep(10 * time.Second)
	for {
		workerapi.WStatus.Lock()
		if !comm.DoKeepAlive(&workerapi.WStatus) {
			logging.RuntimeLog.Errorf("keep alive fail")
			logging.CLILog.Error("keep alive fail")
		}
		workerapi.WStatus.Unlock()
		time.Sleep(60 * time.Second)
	}
}

func checkWorkerPerformance(workerPerformance int) {
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

func setupCloseHandler() {
	quitSignal := make(chan os.Signal, 1)
	signal.Notify(quitSignal, os.Interrupt, syscall.SIGTERM)
	<-quitSignal
	logging.CLILog.Info("Ctrl+C pressed in Terminal,waiting for worker exit...")
	logging.RuntimeLog.Info("Ctrl+C pressed in Terminal,waiting for worker exit...")
	os.Exit(0)
}

func initWorkerStatus(option *WorkerOption) {
	workerapi.WStatus.WorkerName = comm.GetWorkerNameBySelf()
	workerapi.WStatus.CreateTime = time.Now()
	workerapi.WStatus.UpdateTime = time.Now()
	workerapi.WStatus.WorkerTopics = utils.SetToString(option.WorkerTopic)
}

func startWorker(option *WorkerOption) {
	for mode := range option.WorkerTopic {
		go func(topicName string, concurrency int) {
			err := workerapi.StartWorker(topicName, concurrency)
			if err != nil {
				logging.CLILog.Error(err.Error())
				logging.RuntimeLog.Fatal(err.Error())
			}
		}(mode, option.Concurrency)
		time.Sleep(1 * time.Second)
	}
}

func main() {
	//pprof
	//if conf.RunMode == conf.Debug {
	//	go func() {
	//		log.Println(http.ListenAndServe("localhost:6060", nil))
	//	}()
	//}
	option := parseWorkerOptions()
	if option == nil {
		return
	}

	comm.TLSEnabled = option.TLSEnabled
	go keepAlive()
	go comm.StartSaveRuntimeLog(comm.GetWorkerNameBySelf())
	checkWorkerPerformance(option.WorkerPerformance)
	initWorkerStatus(option)
	startWorker(option)
	setupCloseHandler()
}
