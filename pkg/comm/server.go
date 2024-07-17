package comm

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/db"
	"github.com/hanc00l/nemo_go/pkg/es"
	"github.com/hanc00l/nemo_go/pkg/filesync"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/utils"
	"github.com/smallnest/rpcx/protocol"
	"github.com/smallnest/rpcx/server"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

var (
	RsaPrivateKeyText []byte
	RsaPublicKeyText  []byte
)

// StartRPCServer 启动RPC server
func StartRPCServer() {
	rpc := conf.GlobalServerConfig().Rpc
	logging.RuntimeLog.Infof("start rpc server running on tcp@%s:%d...", rpc.Host, rpc.Port)
	logging.CLILog.Infof("start rpc server running on tcp@%s:%d...", rpc.Host, rpc.Port)

	var s *server.Server
	if TLSEnabled {
		cert, err := tls.LoadX509KeyPair(TLSCertFile, TLSKeyFile)
		if err != nil {
			logging.RuntimeLog.Infof("load tls cert fail:%s", err)
			logging.CLILog.Infof("load tls cert fail:%s", err)
			return
		}
		configs := &tls.Config{Certificates: []tls.Certificate{cert}}
		s = server.NewServer(server.WithTLSConfig(configs))
	} else {
		s = server.NewServer()
	}
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
	logging.RuntimeLog.Infof("start filesync server running on tcp@%s:%d...", fileSyncServer.Host, fileSyncServer.Port)
	logging.CLILog.Infof("start filesync server running on tcp@%s:%d...", fileSyncServer.Host, fileSyncServer.Port)

	filesync.StartFileSyncServer(fileSyncServer.Host, fmt.Sprintf("%d", fileSyncServer.Port), fileSyncServer.AuthKey)
}

// StartFileSyncMonitor server文件变化检测并同步worker
func StartFileSyncMonitor() {
	srcPath, err := filepath.Abs(conf.GetRootPath())
	if err != nil {
		logging.RuntimeLog.Error(err)
		logging.CLILog.Error(err)
		return
	}
	w := filesync.NewNotifyFile()
	w.WatchDir(srcPath)
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

// GenerateRSAKey 生成web的RSA公、私钥
func GenerateRSAKey() (err error) {
	if err, RsaPublicKeyText, RsaPrivateKeyText = utils.GenerateRSAKey(2048); err != nil {
		return
	}
	rsaPublicKeyTextJS := bytes.ReplaceAll(RsaPublicKeyText, []byte("\n"), []byte(""))
	// 将rsa的公钥写入到前端的js中：
	// 读取js文件：
	var oldJSText []byte
	webLoginJSFile := filepath.Join(conf.GetRootPath(), "web/static/js/server/login.js")
	if oldJSText, err = os.ReadFile(webLoginJSFile); err != nil {
		return
	}
	// 正则替换原来的key
	pubKeyJS := regexp.MustCompile("(const pubKey = )'(.*?)'")
	var b bytes.Buffer
	b.Write([]byte("$1'"))
	b.Write(rsaPublicKeyTextJS)
	b.Write([]byte("'"))
	newJSText := pubKeyJS.ReplaceAll(oldJSText, b.Bytes())
	// 保存至js文件中
	err = os.WriteFile(webLoginJSFile, newJSText, 0666)

	return
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

func StartSyncElasticAssets() {
	// 同步elastic assets
	if !es.CheckElasticConn() {
		logging.RuntimeLog.Warningf("start sync elastic assets failed, elastic client is nil")
		logging.CLILog.Warningf("start sync elastic assets failed:elastic client is nil")
		return
	}
	logging.RuntimeLog.Info("start sync elastic assets...")
	logging.CLILog.Info("start sync elastic assets...")
	conf.ElasticSyncAssetsChan = make(chan conf.ElasticSyncAssetsArgs, conf.ElasticSyncAssetsChanMax)
	for {
		select {
		case assets := <-conf.ElasticSyncAssetsChan:
			if assets.SyncAssetsType == conf.SyncAssetsTypeIP {
				var ips []db.Ip
				err := json.Unmarshal(assets.Contents, &ips)
				if err != nil {
					logging.RuntimeLog.Errorf("unmarshal ip assets failed:%s", err)
					continue
				}
				if len(ips) == 0 {
					logging.RuntimeLog.Errorf("sync elastic ip assets is empty")
					continue
				}
				time.Sleep(3 * time.Second)
				if !es.SyncIpAssets(ips[0].WorkspaceId, ips) {
					logging.RuntimeLog.Errorf("sync elastic ip assets failed,ip count:%d", len(ips))
				}
			} else if assets.SyncAssetsType == conf.SyncAssetsTypeDomain {
				var domains []db.Domain
				err := json.Unmarshal(assets.Contents, &domains)
				if err != nil {
					logging.RuntimeLog.Errorf("unmarshal domain assets failed:%s", err)
					continue
				}
				if len(domains) == 0 {
					logging.RuntimeLog.Errorf("sync elastic domain assets is empty")
					continue
				}
				time.Sleep(3 * time.Second)
				if !es.SyncDomainAssets(domains[0].WorkspaceId, domains) {
					logging.RuntimeLog.Errorf("sync elastic domain assets failed,domain count:%d", len(domains))
				}
			}
		}
	}
}
