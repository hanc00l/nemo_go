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
	"net/http"
	"time"
)

var UrlFilterWhiteList = []string{"/"}

// startWebServer 启动web server
func startWebServer() {
	err := logs.SetLogger("file", `{"filename":"log/access.log"}`)
	if err != nil {
		logging.CLILog.Error(err)
		return
	}
	if conf.RunMode == conf.Release {
		web.InsertFilter("/*", web.BeforeRouter, filterLoginCheck)
	}
	logging.RuntimeLog.Info("Nemo Server started...")
	addr := fmt.Sprintf("%s:%d", conf.GlobalServerConfig().Web.Host, conf.GlobalServerConfig().Web.Port)
	web.Run(addr)

}

// filterLoginCheck 全局的登录验证
func filterLoginCheck(ctx *beegoContext.Context) {
	for _, url := range UrlFilterWhiteList {
		if ctx.Request.RequestURI == url {
			return
		}
	}
	// 检查用户是否登录（检查登录成功后的session:User、UserRole、Workspace
	if user, ok := ctx.Input.Session("User").(string); !ok || len(user) == 0 {
		ctx.Redirect(http.StatusFound, "/")
	}
	userRole, ok := ctx.Input.Session("UserRole").(string)
	{
		if !ok || len(userRole) == 0 {
			ctx.Redirect(http.StatusFound, "/")
		}
	}
	if workspaceId, ok := ctx.Input.Session("Workspace").(int); !ok || (userRole != "superadmin" && workspaceId <= 0) {
		ctx.Redirect(http.StatusFound, "/")
	}
}

// startRPCServer 启动RPC server
func startRPCServer() {
	rpc := conf.GlobalServerConfig().Rpc
	logging.RuntimeLog.Infof("rpc server running on tcp@%s:%d...", rpc.Host, rpc.Port)
	logging.CLILog.Infof("rpc server running on tcp@%s:%d...", rpc.Host, rpc.Port)

	s := server.NewServer()
	err := s.Register(new(comm.Service), "")
	if err != nil {
		logging.RuntimeLog.Error(err)
		logging.CLILog.Error(err)
		return
	}
	s.AuthFunc = auth
	err = s.Serve("tcp", fmt.Sprintf("%s:%d", rpc.Host, rpc.Port))
	if err != nil {
		if err != nil {
			logging.RuntimeLog.Error(err)
			logging.CLILog.Error(err)
		}
		return
	}
}

// startCronTask 启动定时任务
func startCronTask() {
	num := runner.StartCronTask()
	logging.CLILog.Infof("cron task total:%d", num)
	logging.RuntimeLog.Infof("cron task total:%d", num)
}

func startMainTaskDemon() {
	go runner.StartMainTaskDamon()
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

	if noFilesync == false {
		go startFileSyncServer()
		go startFileSyncMonitor()
	}
	time.Sleep(time.Second * 1)
	go startRPCServer()
	time.Sleep(time.Second * 1)
	startCronTask()
	time.Sleep(time.Second * 1)
	startMainTaskDemon()

	startWebServer()
}
