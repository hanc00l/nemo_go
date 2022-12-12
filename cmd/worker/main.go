package main

import (
	"flag"
	"github.com/hanc00l/nemo_go/pkg/comm"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/workerapi"
	"log"
	"net/http"
	_ "net/http/pprof"
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
	flag.IntVar(&concurrency, "c", 3, "concurrent number of tasks")
	flag.Parse()

	if conf.RunMode == conf.Debug {
		go func() {
			log.Println(http.ListenAndServe("localhost:6060", nil))
		}()
	}
	
	go keepAlive()

	err := workerapi.StartWorker(concurrency)
	if err != nil {
		logging.CLILog.Error(err.Error())
		logging.RuntimeLog.Fatal(err.Error())
	}
}
