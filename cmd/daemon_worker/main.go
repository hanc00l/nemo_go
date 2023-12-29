package main

import (
	"flag"
	"github.com/hanc00l/nemo_go/pkg/comm"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/filesync"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"os"
	"os/signal"
	"syscall"
)

func parseDaemonWorkerOption() *comm.WorkerDaemonOption {
	option := &comm.WorkerDaemonOption{}
	if conf.RunMode == conf.Debug {
		option.NoFilesync = true
	}
	flag.IntVar(&option.Concurrency, "c", 3, "concurrent number of tasks")
	flag.IntVar(&option.WorkerPerformance, "p", 0, "worker performance,default is autodetect (0:autodetect, 1:high, 2:normal)")
	flag.StringVar(&option.WorkerRunTaskMode, "m", "0", "worker run task mode; 0: all, 1:active, 2:finger, 3:passive, 4:pocscan, 5:custom; run multiple mode separated by \",\"")
	flag.StringVar(&option.TaskWorkspaceGUID, "w", "", "workspace guid for custom task; multiple workspace separated by \",\"")
	flag.StringVar(&option.ManualSyncHost, "mh", "", "manual file sync host address")
	flag.StringVar(&option.ManualSyncPort, "mp", "", "manual file sync port,default is 5002")
	flag.StringVar(&option.ManualSyncAuth, "ma", "", "manual file sync auth key")
	flag.BoolVar(&option.NoFilesync, "nf", option.NoFilesync, "disable file sync")
	flag.BoolVar(&option.NoProxy, "np", false, "disable proxy configuration,include socks5 proxy and socks5forward")
	flag.BoolVar(&option.TLSEnabled, "tls", false, "use TLS for RPC and filesync")
	flag.StringVar(&option.DefaultConfigFile, "f", conf.WorkerDefaultConfigFile, "worker default config file")
	flag.Parse()

	return option
}

func setupCloseHandler() {
	quitSignal := make(chan os.Signal, 1)
	signal.Notify(quitSignal, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-quitSignal
		logging.CLILog.Info("Ctrl+C pressed in Terminal,waiting for worker exit...")
		logging.RuntimeLog.Info("Ctrl+C pressed in Terminal,waiting for worker exit...")
		comm.KillWorker()
		os.Exit(0)
	}()
}

func main() {
	comm.DaemonRunOption = parseDaemonWorkerOption()
	if comm.DaemonRunOption == nil {
		return
	}

	comm.TLSEnabled = comm.DaemonRunOption.TLSEnabled
	filesync.TLSEnabled = comm.DaemonRunOption.TLSEnabled

	if comm.DaemonRunOption.ManualSyncHost != "" && comm.DaemonRunOption.ManualSyncPort != "" && comm.DaemonRunOption.ManualSyncAuth != "" {
		logging.RuntimeLog.Info("start onetime file sync...")
		logging.CLILog.Info("start onetime file sync...")
		filesync.WorkerStartupSync(comm.DaemonRunOption.ManualSyncHost, comm.DaemonRunOption.ManualSyncPort, comm.DaemonRunOption.ManualSyncAuth)
		return
	}
	go comm.StartSaveRuntimeLog(comm.GetWorkerNameBySelf())
	go setupCloseHandler()
	comm.StartWorkerDaemon()
}
