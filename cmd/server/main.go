package main

import (
	"flag"
	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
	beegoContext "github.com/beego/beego/v2/server/web/context"
	"github.com/hanc00l/nemo_go/pkg/cert"
	"github.com/hanc00l/nemo_go/pkg/comm"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/filesync"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/ampq"
	"github.com/hanc00l/nemo_go/pkg/task/custom"
	"github.com/hanc00l/nemo_go/pkg/task/runner"
	"github.com/hanc00l/nemo_go/pkg/utils"
	_ "github.com/hanc00l/nemo_go/pkg/web/routers"
	"net/http"
	"path/filepath"
	"time"
)

type ServerOption struct {
	NoFilesync  bool
	NoRPC       bool
	TLSEnabled  bool
	TLSCertFile string
	TLSKeyFile  string
}

var UrlFilterWhiteList = []string{"/"}

func parseServerOption() *ServerOption {
	option := &ServerOption{}
	if conf.RunMode == conf.Debug {
		option.NoFilesync = true
	}
	flag.BoolVar(&option.NoFilesync, "nf", option.NoFilesync, "disable file sync")
	flag.BoolVar(&option.NoRPC, "nr", false, "disable rpc")
	flag.BoolVar(&option.TLSEnabled, "tls", false, "use TLS for web、RPC and filesync")
	flag.StringVar(&option.TLSKeyFile, "key", "server.key", "TLS private key file")
	flag.StringVar(&option.TLSCertFile, "cert", "server.crt", "TLS cert file")
	flag.Parse()

	return option
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

// StartWebServer 启动web server
func StartWebServer(option *ServerOption) {
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
	if option.TLSEnabled {
		web.BConfig.Listen.EnableHTTP = false
		web.BConfig.Listen.EnableHTTPS = true
		web.BConfig.Listen.HTTPSCertFile = option.TLSCertFile
		web.BConfig.Listen.HTTPSKeyFile = option.TLSKeyFile
		web.BConfig.Listen.HTTPSAddr = conf.GlobalServerConfig().Web.Host
		web.BConfig.Listen.HTTPSPort = conf.GlobalServerConfig().Web.Port
	} else {
		web.BConfig.Listen.EnableHTTP = true
		web.BConfig.Listen.EnableHTTPS = false
		web.BConfig.Listen.HTTPAddr = conf.GlobalServerConfig().Web.Host
		web.BConfig.Listen.HTTPPort = conf.GlobalServerConfig().Web.Port
	}
	web.Run()
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

func loadCustomTaskWorkspace() {
	ampq.CustomTaskWorkspaceMap = custom.LoadCustomTaskWorkspace()
}

func main() {
	option := parseServerOption()
	if option == nil {
		return
	}

	if option.TLSEnabled {
		if !utils.CheckFileExist(filepath.Join(conf.GetRootPath(), option.TLSCertFile)) || !utils.CheckFileExist(filepath.Join(conf.GetRootPath(), option.TLSKeyFile)) {
			if err := cert.GenerateSelfSignedCert(option.TLSCertFile, option.TLSKeyFile); err != nil {
				logging.CLILog.Error(err)
				return
			}
			logging.CLILog.Info("generate selfsigned cert...")
		}
	}
	if !option.NoFilesync {
		filesync.TLSEnabled = option.TLSEnabled
		filesync.TLSCertFile = option.TLSCertFile
		filesync.TLSKeyFile = option.TLSKeyFile
		go comm.StartFileSyncServer()
		go comm.StartFileSyncMonitor()
		time.Sleep(time.Second * 1)
	}
	if !option.NoRPC {
		comm.TLSEnabled = option.TLSEnabled
		comm.TLSCertFile = option.TLSCertFile
		comm.TLSKeyFile = option.TLSKeyFile
		go comm.StartRPCServer()
		time.Sleep(time.Second * 1)
	}
	go comm.StartSaveRuntimeLog("server@nemo")
	loadCustomTaskWorkspace()
	StartCronTask()
	StartMainTaskDemon()
	time.Sleep(time.Second * 1)

	err := comm.GenerateRSAKey()
	if err != nil {
		logging.CLILog.Error(err)
		logging.RuntimeLog.Error(err)
		return
	}
	StartWebServer(option)
}
