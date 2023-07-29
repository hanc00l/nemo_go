package main

import (
	"flag"
	"fmt"
	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
	beegoContext "github.com/beego/beego/v2/server/web/context"
	"github.com/hanc00l/nemo_go/pkg/comm"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/runner"
	_ "github.com/hanc00l/nemo_go/pkg/web/routers"
	"net/http"
	"time"
)

var UrlFilterWhiteList = []string{"/"}

// StartCronTask 启动定时任务
func StartCronTask() {
	num := runner.StartCronTask()
	logging.CLILog.Infof("cron task total:%d", num)
	logging.RuntimeLog.Infof("cron task total:%d", num)
}

// StartMainTaskDemon 启动任务监控，生成运行任务，分发到队列由worker执行
func StartMainTaskDemon() {
	go runner.StartMainTaskDamon()
}

// StartWebServer 启动web server
func StartWebServer() {
	err := logs.SetLogger("file", `{"filename":"log/access.log"}`)
	if err != nil {
		logging.RuntimeLog.Error(err)
		logging.CLILog.Error(err)
		return
	}
	if conf.RunMode == conf.Release {
		web.InsertFilter("/*", web.BeforeRouter, filterLoginCheck)
	}
	logging.RuntimeLog.Info("nemo server started...")
	logging.CLILog.Info("nemo server started...")
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
	if !ok || len(userRole) == 0 {
		ctx.Redirect(http.StatusFound, "/")
	}
	if workspaceId, ok := ctx.Input.Session("Workspace").(int); !ok || (userRole != "superadmin" && workspaceId <= 0) {
		ctx.Redirect(http.StatusFound, "/")
	}
}

func main() {
	var noFilesync, noRPC bool
	if conf.RunMode == conf.Debug {
		noFilesync = true
	}
	flag.BoolVar(&noFilesync, "nf", noFilesync, "disable file sync")
	flag.BoolVar(&noRPC, "nr", false, "disable rpc")
	flag.Parse()

	if noFilesync == false {
		go comm.StartFileSyncServer()
		go comm.StartFileSyncMonitor()
		time.Sleep(time.Second * 1)
	}
	if noRPC == false {
		go comm.StartRPCServer()
		time.Sleep(time.Second * 1)
	}
	go comm.StartSaveRuntimeLog("server@nemo")
	StartCronTask()
	StartMainTaskDemon()

	time.Sleep(time.Second * 1)
	err := comm.GenerateRSAKey()
	if err != nil {
		logging.CLILog.Error(err)
		logging.RuntimeLog.Error(err)
		return
	}
	StartWebServer()
}
