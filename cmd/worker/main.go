package main

import (
	"flag"
	"github.com/hanc00l/nemo_go/pkg/comm"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/ampq"
	"github.com/hanc00l/nemo_go/pkg/task/workerapi"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"os"
	"os/signal"
	"syscall"

	//_ "net/http/pprof"
	"time"
)

// keepAlive worker与server的心跳与同步
func keepAlive() {
	time.Sleep(10 * time.Second)
	for {
		workerapi.WStatus.Lock()
		if !comm.DoKeepAlive(workerapi.WStatus) {
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
		if err1 != nil {
			break
		}
		memInfo, err2 := mem.VirtualMemory()
		if err2 != nil {
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
	//go func() {
	<-quitSignal
	logging.CLILog.Info("Ctrl+C pressed in Terminal,waiting for worker exit...")
	logging.RuntimeLog.Info("Ctrl+C pressed in Terminal,waiting for worker exit...")
	os.Exit(0)
	//}()
}

func initWorkerStatus(workerRunTaskMode int) {
	workerapi.WStatus.WorkerName = comm.GetWorkerNameBySelf()
	workerapi.WStatus.CreateTime = time.Now()
	workerapi.WStatus.UpdateTime = time.Now()
	if workerRunTaskMode == 0 {
		workerapi.WStatus.WorkerRunTaskMode = "default"
	} else {
		workerapi.WStatus.WorkerRunTaskMode = ampq.GetTopicByWorkerMode(ampq.WorkerRunTaskMode(workerRunTaskMode))
	}
}

func startWorker(workerRunTaskMode, concurrency int) {
	var modeList []int
	if workerRunTaskMode == int(ampq.TaskModeDefault) {
		modeList = append(modeList, int(ampq.TaskModeActive))
		modeList = append(modeList, int(ampq.TaskModeFinger))
		modeList = append(modeList, int(ampq.TaskModePassive))
		// 固定每个worker的最大任务并发量为3*2
		concurrency = 2
	} else {
		modeList = append(modeList, workerRunTaskMode)
	}
	for _, mode := range modeList {
		go func(m int) {
			err := workerapi.StartWorker(m, concurrency)
			if err != nil {
				logging.CLILog.Error(err.Error())
				logging.RuntimeLog.Fatal(err.Error())
			}
		}(mode)
		time.Sleep(1 * time.Second)
	}
}

func main() {
	var concurrency int
	var workerPerformance int
	var workerRunTaskMode int
	flag.IntVar(&concurrency, "c", 3, "concurrent number of tasks")
	flag.IntVar(&workerPerformance, "p", 0, "worker performance,default is autodetect (0:autodetect, 1:high, 2:normal)")
	flag.IntVar(&workerRunTaskMode, "m", 0, "worker run task mode,default is 0 (0: all, 1:active scan, 2:fingerprint, 3:passive collect, 4:custom)")
	flag.Parse()
	//pprof
	//if conf.RunMode == conf.Debug {
	//	go func() {
	//		log.Println(http.ListenAndServe("localhost:6060", nil))
	//	}()
	//}
	if workerRunTaskMode < int(ampq.TaskModeDefault) || workerRunTaskMode > int(ampq.TaskModeCustom) {
		logging.CLILog.Error("error workerRunTaskMode,exit...")
		return
	}

	go keepAlive()
	go comm.StartSaveRuntimeLog(comm.GetWorkerNameBySelf())

	checkWorkerPerformance(workerPerformance)
	initWorkerStatus(workerRunTaskMode)
	startWorker(workerRunTaskMode, concurrency)
	setupCloseHandler()
}
