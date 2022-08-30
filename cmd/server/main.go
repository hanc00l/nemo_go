package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
	beegoContext "github.com/beego/beego/v2/server/web/context"
	"github.com/hanc00l/nemo_go/pkg/comm"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/filesync"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/runner"
	_ "github.com/hanc00l/nemo_go/pkg/web/routers"
	"github.com/smallnest/rpcx/protocol"
	"github.com/smallnest/rpcx/server"
	"time"
)

// startWebServer 启动web server
func startWebServer() {
	logs.SetLogger("file", `{"filename":"log/access.log"}`)

	UrlFilterWhiteList := []string{
		"/",
	}
	var FilterLoginCheck = func(ctx *beegoContext.Context) {
		for _, url := range UrlFilterWhiteList {
			if ctx.Request.RequestURI == url {
				return
			}
		}
		if _, ok := ctx.Input.Session("IsLogin").(bool); !ok {
			ctx.Redirect(302, "/")
		}
	}
	if conf.RunMode == conf.Release {
		web.InsertFilter("/*", web.BeforeRouter, FilterLoginCheck)
	}

	logging.RuntimeLog.Info("Nemo Server started...")
	addr := fmt.Sprintf("%s:%d", conf.GlobalServerConfig().Web.Host, conf.GlobalServerConfig().Web.Port)
	web.Run(addr)

}

// startRPCServer 启动RPC server
func startRPCServer() {
	rpc := conf.GlobalServerConfig().Rpc
	logging.RuntimeLog.Infof("rpc server running on tcp@%s:%d...", rpc.Host, rpc.Port)
	logging.CLILog.Infof("rpc server running on tcp@%s:%d...", rpc.Host, rpc.Port)

	s := server.NewServer()
	s.Register(new(comm.Service), "")
	s.AuthFunc = auth
	s.Serve("tcp", fmt.Sprintf("%s:%d", rpc.Host, rpc.Port))
}

// startCronTask 启动定时任务
func startCronTask() {
	num := runner.StartCronTask()
	logging.CLILog.Infof("cron task total:%d", num)
	logging.RuntimeLog.Infof("cron task total:%d", num)
}

// auth RPC调用认证
func auth(ctx context.Context, req *protocol.Message, token string) error {
	if token == conf.GlobalServerConfig().Rpc.AuthKey {
		return nil
	}

	return errors.New("invalid token")
}

// startFileSyncServer 启动文件同步服务
func startFileSyncServer() {
	fileSyncServer := conf.GlobalServerConfig().FileSync
	logging.RuntimeLog.Infof("filesync server running on tcp@%s:%d...", fileSyncServer.Host, fileSyncServer.Port)
	logging.CLILog.Infof("filesync server running on tcp@%s:%d...", fileSyncServer.Host, fileSyncServer.Port)

	filesync.StartFileSyncServer(fileSyncServer.Host, fmt.Sprintf("%d", fileSyncServer.Port), fileSyncServer.AuthKey)
}

// startFileSyncMonitor server文件变化检测并同步worker
func startFileSyncMonitor() {
	w := filesync.NewNotifyFile()
	w.WatchDir()
	for {
		select {
		case fileName := <-w.ChNeedWorkerSync:
			logging.CLILog.Infof("monitor file changed:%s", fileName)
			// 设置worker同步标志
			comm.WorkerStatusMutex.Lock()
			for k := range comm.WorkerStatus {
				comm.WorkerStatus[k].ManualFileSyncFlag = true
			}
			comm.WorkerStatusMutex.Unlock()
		}
	}
}

func main() {
	var noFilesync bool
	flag.BoolVar(&noFilesync, "nf", false, "disable file sync")
	flag.Parse()

	go startRPCServer()
	if noFilesync == false {
		time.Sleep(time.Second * 1)
		go startFileSyncServer()
		go startFileSyncMonitor()
	}
	time.Sleep(time.Second * 1)
	startCronTask()
	startWebServer()
}
