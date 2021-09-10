package main

import (
	"flag"
	"github.com/hanc00l/nemo_go/pkg/comm"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/workerapi"
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

func main() {
	var concurrency int
	flag.IntVar(&concurrency, "c", 2, "concurrent number of tasks")
	flag.Parse()

	go keepAlive()

	err := workerapi.StartWorker(concurrency)
	if err != nil {
		logging.CLILog.Error(err.Error())
		logging.RuntimeLog.Fatal(err.Error())
	}
}
