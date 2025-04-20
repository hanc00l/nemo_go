package main

import (
	"errors"
	"flag"
	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
	beegoContext "github.com/beego/beego/v2/server/web/context"
	"github.com/hanc00l/nemo_go/v3/pkg/cert"
	"github.com/hanc00l/nemo_go/v3/pkg/conf"
	"github.com/hanc00l/nemo_go/v3/pkg/core"
	"github.com/hanc00l/nemo_go/v3/pkg/logging"
	"github.com/hanc00l/nemo_go/v3/pkg/utils"
	"github.com/hanc00l/nemo_go/v3/pkg/web/routers"
	"github.com/jessevdk/go-flags"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

func parseServerOption() *core.ServerOption {
	opts := core.ServerOption{}
	parser := flags.NewParser(&opts, flags.Default)
	if _, err := parser.Parse(); err != nil {
		if !errors.Is(err.(*flags.Error).Type, flag.ErrHelp) {
			return nil
		}
	}
	if !utils.CheckFileExist(filepath.Join(conf.GetRootPath(), opts.TLSCertFile)) || !utils.CheckFileExist(filepath.Join(conf.GetRootPath(), opts.TLSKeyFile)) {
		if err := cert.GenerateSelfSignedCert(opts.TLSCertFile, opts.TLSKeyFile); err != nil {
			logging.CLILog.Error(err)
			return nil
		}
		logging.CLILog.Info("generate selfsigned cert...")
	}
	core.TLSCertFile = opts.TLSCertFile
	core.TLSKeyFile = opts.TLSKeyFile

	return &opts
}

// filterLoginCheck 全局的登录验证
func filterLoginCheck(ctx *beegoContext.Context) {
	if ctx.Request.RequestURI == "/" {
		return
	}
	// 检查用户是否登录（检查登录成功后的session:User、UserRole、Workspace
	if user, ok := ctx.Input.Session("User").(string); !ok || len(user) == 0 {
		ctx.Redirect(http.StatusFound, "/")
	}
	userRole, ok := ctx.Input.Session("UserRole").(string)
	if !ok || len(userRole) == 0 {
		ctx.Redirect(http.StatusFound, "/")
	}
	if workspaceId, ok := ctx.Input.Session("Workspace").(string); !ok || len(workspaceId) == 0 {
		ctx.Redirect(http.StatusFound, "/")
	}
}

// StartWebServer 启动web server
func StartWebServer(option *core.ServerOption) {
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
	web.BConfig.Listen.EnableHTTP = false
	web.BConfig.Listen.EnableHTTPS = true
	web.BConfig.Listen.HTTPSCertFile = option.TLSCertFile
	web.BConfig.Listen.HTTPSKeyFile = option.TLSKeyFile
	web.BConfig.Listen.HTTPSAddr = conf.GlobalServerConfig().Web.Host
	web.BConfig.Listen.HTTPSPort = conf.GlobalServerConfig().Web.Port
	routers.InitRouter()
	web.Run()
}

func setupCloseHandler() {
	quitSignal := make(chan os.Signal, 1)
	signal.Notify(quitSignal, os.Interrupt, syscall.SIGTERM)
	<-quitSignal
	logging.CLILog.Info("Ctrl+C pressed in Terminal,waiting for exit...")
	logging.RuntimeLog.Info("Ctrl+C pressed in Terminal,waiting for exit...")
	os.Exit(0)
}

func main() {
	option := parseServerOption()
	if option == nil {
		return
	}
	err := conf.GlobalServerConfig().ReloadConfig()
	if err != nil {
		logging.CLILog.Error(err)
		return
	}

	if option.Service {
		go core.StartServiceServer()
		go core.StartMainTaskDamon()
		time.Sleep(time.Second * 1)
	}
	if option.RedisTunnel {
		go core.StartRedisReverseProxy()
		time.Sleep(time.Second * 1)
	}
	if option.Cron {
		core.StartCronTaskDamon()
		time.Sleep(time.Second * 1)
	}
	go core.StartSaveRuntimeLog(core.GetWorkerNameBySelf())
	if option.Web {
		// web是阻塞的，直接主线程中运行
		StartWebServer(option)
	} else {
		setupCloseHandler()
	}
}
