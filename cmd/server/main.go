package main

import (
	"fmt"
	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/logging"
	_ "github.com/hanc00l/nemo_go/pkg/web/routers"
)

func main() {
	logging.RuntimeLog.Info("Nemo Server started...")

	web.BConfig.WebConfig.Session.SessionOn = true
	web.BConfig.WebConfig.Session.SessionName = "sessionID"
	web.BConfig.CopyRequestBody = true
	logs.SetLogger("file",`{"filename":"log/access.log"}`)


	UrlFilterWhiteList := []string{
		"/",
		"/worker-alive",
		"/upload-screenshot",
		"/upload-icpinfo",
	}
	var FilterLoginCheck = func(ctx *context.Context) {
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

	web.Run(fmt.Sprintf("%s:%d", conf.Nemo.Web.Host, conf.Nemo.Web.Port))
}
