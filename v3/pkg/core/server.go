package core

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/hanc00l/nemo_go/v3/pkg/conf"
	"github.com/hanc00l/nemo_go/v3/pkg/logging"
	"github.com/smallnest/rpcx/protocol"
	"github.com/smallnest/rpcx/server"
)

var (
	RsaPrivateKeyText []byte
	RsaPublicKeyText  []byte
)

// StartServiceServer 启动RPC server
func StartServiceServer() {
	rpc := conf.GlobalServerConfig().Service
	logging.RuntimeLog.Infof("start rpc server running on tcp@%s:%d...", rpc.Host, rpc.Port)
	logging.CLILog.Infof("start rpc server running on tcp@%s:%d...", rpc.Host, rpc.Port)

	var s *server.Server
	cert, err := tls.LoadX509KeyPair(TLSCertFile, TLSKeyFile)
	if err != nil {
		logging.RuntimeLog.Infof("load tls cert fail:%s", err)
		logging.CLILog.Infof("load tls cert fail:%s", err)
		return
	}
	configs := &tls.Config{Certificates: []tls.Certificate{cert}}
	s = server.NewServer(server.WithTLSConfig(configs))
	err = s.Register(new(Service), "")
	if err != nil {
		logging.RuntimeLog.Error(err)
		logging.CLILog.Error(err)
		return
	}
	s.AuthFunc = auth
	err = s.Serve("tcp", fmt.Sprintf("%s:%d", rpc.Host, rpc.Port))
	if err != nil {
		logging.RuntimeLog.Error(err)
		logging.CLILog.Error(err)
	}
}

// auth RPC调用认证
func auth(ctx context.Context, req *protocol.Message, token string) error {
	if token == conf.GlobalServerConfig().Service.AuthKey {
		return nil
	}

	return errors.New("invalid token")
}

func StartSaveRuntimeLog(source string) {
	logging.RuntimeLogChan = make(chan []byte, logging.RuntimeLogChanMax)
	for {
		select {
		case msg := <-logging.RuntimeLogChan:
			resultArgs := RuntimeLogArgs{
				Source:     source,
				LogMessage: msg,
			}
			var result string
			err := CallXClient("SaveRuntimeLog", &resultArgs, &result)
			if err != nil {
				logging.CLILog.Error(err)
			}
		}
	}
}

func StartRedisReverseProxy() {
	// 建立redis反向代理连接的通道，worker通过认证和使用redis tunnel进行通道，避免直接暴露redis服务
	// redisAddr 为真实的redis地址
	redisAddr := fmt.Sprintf("%s:%d", conf.GlobalServerConfig().Redis.Host, conf.GlobalServerConfig().Redis.Port)
	// listenAddr 为redis tunnel监听地址，用于接收redis tunnel客户端的连接
	listenAddr := fmt.Sprintf("%s:%d", conf.GlobalServerConfig().Service.Host, conf.GlobalServerConfig().RedisTunnel.Port)
	reverseServer := NewRedisReverseServer(redisAddr, listenAddr, conf.GlobalServerConfig().RedisTunnel.AuthKey)
	reverseServer.Start()
}
