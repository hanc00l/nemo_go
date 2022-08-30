package main

import (
	"flag"
	"github.com/hanc00l/nemo_go/pkg/comm"
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
	var concurrency int
	var noFilesync bool
	flag.IntVar(&concurrency, "c", 3, "concurrent number of tasks")
	flag.BoolVar(&noFilesync, "nf", false, "disable file sync")
	flag.Parse()

	go SetupCloseHandler()
	comm.StartWorkerDaemon(concurrency, noFilesync)
}
