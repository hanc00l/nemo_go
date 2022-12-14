package main

import (
	"flag"
	"github.com/hanc00l/nemo_go/pkg/comm"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/workerapi"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	//_ "net/http/pprof"
	"time"
)

// keepAlive worker与server的心跳与同步
func keepAlive() {
	time.Sleep(10 * time.Second)
	for {
		if !comm.DoKeepAlive(workerapi.WStatus) {
			logging.RuntimeLog.Errorf("keep alive fail")
			logging.CLILog.Error("keep alive fail")
		}
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

func main() {
	var concurrency int
	var workerPerformance int
	flag.IntVar(&concurrency, "c", 3, "concurrent number of tasks")
	flag.IntVar(&workerPerformance, "p", 0, "worker performance,default is autodetect (0: autodetect,1:high,2:normal)")
	flag.Parse()

	//if conf.RunMode == conf.Debug {
	//	go func() {
	//		log.Println(http.ListenAndServe("localhost:6060", nil))
	//	}()
	//}
	checkWorkerPerformance(workerPerformance)

	go keepAlive()

	err := workerapi.StartWorker(concurrency)
	if err != nil {
		logging.CLILog.Error(err.Error())
		logging.RuntimeLog.Fatal(err.Error())
	}
}
