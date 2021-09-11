package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
	beegoContext "github.com/beego/beego/v2/server/web/context"
	"github.com/hanc00l/nemo_go/pkg/comm"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/logging"
	_ "github.com/hanc00l/nemo_go/pkg/web/routers"
	"github.com/smallnest/rpcx/protocol"
	"github.com/smallnest/rpcx/server"
	"time"
)

// startWebServer 启动web server
func startWebServer() {
	web.BConfig.WebConfig.Session.SessionOn = true
	web.BConfig.WebConfig.Session.SessionName = "sessionID"
	web.BConfig.CopyRequestBody = true
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

// auth RPC调用认证
func auth(ctx context.Context, req *protocol.Message, token string) error {
	if token == conf.GlobalServerConfig().Rpc.AuthKey {
		return nil
	}

	return errors.New("invalid token")
}

func main() {
	go startRPCServer()
	time.Sleep(time.Second*1)
	startWebServer()
}
