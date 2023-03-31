package comm

import (
	"context"
	"errors"
	"fmt"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/filesync"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/smallnest/rpcx/protocol"
	"github.com/smallnest/rpcx/server"
)

// StartRPCServer 启动RPC server
func StartRPCServer() {
	rpc := conf.GlobalServerConfig().Rpc
	logging.RuntimeLog.Infof("rpc server running on tcp@%s:%d...", rpc.Host, rpc.Port)
	logging.CLILog.Infof("rpc server running on tcp@%s:%d...", rpc.Host, rpc.Port)

	s := server.NewServer()
	err := s.Register(new(Service), "")
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

// auth RPC调用认证
func auth(ctx context.Context, req *protocol.Message, token string) error {
	if token == conf.GlobalServerConfig().Rpc.AuthKey {
		return nil
	}

	return errors.New("invalid token")
}

// StartFileSyncServer 启动文件同步服务
func StartFileSyncServer() {
	fileSyncServer := conf.GlobalServerConfig().FileSync
	logging.RuntimeLog.Infof("filesync server running on tcp@%s:%d...", fileSyncServer.Host, fileSyncServer.Port)
	logging.CLILog.Infof("filesync server running on tcp@%s:%d...", fileSyncServer.Host, fileSyncServer.Port)

	filesync.StartFileSyncServer(fileSyncServer.Host, fmt.Sprintf("%d", fileSyncServer.Port), fileSyncServer.AuthKey)
}

// StartFileSyncMonitor server文件变化检测并同步worker
func StartFileSyncMonitor() {
	w := filesync.NewNotifyFile()
	w.WatchDir()
	for {
		select {
		case fileName := <-w.ChNeedWorkerSync:
			logging.CLILog.Infof("monitor file changed:%s", fileName)
			// 设置worker同步标志
			WorkerStatusMutex.Lock()
			for k := range WorkerStatus {
				WorkerStatus[k].ManualFileSyncFlag = true
			}
			WorkerStatusMutex.Unlock()
		}
	}
}
