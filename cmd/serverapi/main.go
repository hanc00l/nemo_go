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
	ctrl "github.com/hanc00l/nemo_go/pkg/web/controllers"
	_ "github.com/hanc00l/nemo_go/pkg/webapi/routers"
	"net/http"
	"time"
)

var UrlFilterWhiteList = []string{
	"/v1/login/captcha",
	"/v1/login/login",
}

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

// StartWebAPIServer 启动webapi
func StartWebAPIServer() {
	err := logs.SetLogger("file", `{"filename":"log/access.log"}`)
	if err != nil {
		logging.RuntimeLog.Error(err)
		logging.CLILog.Error(err)
		return
	}
	if conf.RunMode == conf.Debug {
		web.BConfig.WebConfig.DirectoryIndex = true
		web.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
	}
	if conf.RunMode == conf.Release {
		web.InsertFilter("/*", web.BeforeRouter, filterLoginCheck)
	}
	logging.RuntimeLog.Info("nemo API server started...")
	logging.CLILog.Info("nemo API server started...")
	addr := fmt.Sprintf("%s:%d", conf.GlobalServerConfig().WebAPI.Host, conf.GlobalServerConfig().WebAPI.Port)
	web.Run(addr)
}

// filterLoginCheck 全局的token验证
func filterLoginCheck(ctx *beegoContext.Context) {
	for _, url := range UrlFilterWhiteList {
		if ctx.Request.RequestURI == url {
			return
		}
	}
	// 检查token
	tokenString := ctx.Input.Header("Authorization")
	if len(tokenString) == 0 {
		ctx.Redirect(http.StatusFound, "/")
	}
	if jwtData := ctrl.ValidToken(ctrl.GetTokenValueFromHeader(tokenString)); jwtData == nil {
		ctx.Redirect(http.StatusFound, "/")
	}
}

func main() {
	var noFilesync, noRPC bool
	flag.BoolVar(&noFilesync, "nf", false, "disable file sync")
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
	StartCronTask()
	StartMainTaskDemon()

	time.Sleep(time.Second * 1)
	StartWebAPIServer()
}
