package main

import (
	"flag"
	"github.com/hanc00l/nemo_go/pkg/comm"
	"github.com/hanc00l/nemo_go/pkg/filesync"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"os"
	"os/signal"
	"syscall"
)

func SetupCloseHandler() {
	quitSignal := make(chan os.Signal, 1)
	signal.Notify(quitSignal, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-quitSignal
		logging.CLILog.Warn("Ctrl+C pressed in Terminal,waiting for worker exit...")
		logging.RuntimeLog.Warn("Ctrl+C pressed in Terminal,waiting for worker exit...")
		comm.KillWorker()
		os.Exit(0)
	}()
}

func main() {
	var concurrency, workerPerformance int
	var noFilesync bool
	flag.IntVar(&concurrency, "c", 3, "concurrent number of tasks")
	flag.IntVar(&workerPerformance, "p", 0, "worker performance,default is autodetect (0: autodetect,1:high,2:normal)")
	flag.BoolVar(&noFilesync, "nf", false, "disable file sync")
	// manual sync for worker by command line
	var manualSyncHost, manualSyncPort, manualSyncAuth string
	flag.StringVar(&manualSyncHost, "mh", "", "manual file sync host address")
	flag.StringVar(&manualSyncPort, "mp", "", "manual file sync port,default is 5002")
	flag.StringVar(&manualSyncAuth, "ma", "", "manual file sync auth key")
	flag.Parse()

	if manualSyncHost != "" && manualSyncPort != "" && manualSyncAuth != "" {
		logging.CLILog.Info("start to file sync...")
		filesync.WorkerStartupSync(manualSyncHost, manualSyncPort, manualSyncAuth)
		return
	}

	go SetupCloseHandler()
	comm.StartWorkerDaemon(concurrency, workerPerformance, noFilesync)
}
