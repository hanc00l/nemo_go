package es

import (
	"context"
	"crypto/tls"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"net/http"
)

var globalClient *elasticsearch.TypedClient

// GetElasticConfig 获取es配置
func GetElasticConfig() elasticsearch.Config {
	transport := http.DefaultTransport
	tlsClientConfig := &tls.Config{InsecureSkipVerify: true}
	transport.(*http.Transport).TLSClientConfig = tlsClientConfig
	esConfig := elasticsearch.Config{
		Transport: transport,
	}
	serverEsCfg := conf.GlobalServerConfig().Elastic
	esConfig.Addresses = []string{serverEsCfg.Url}
	esConfig.Username = serverEsCfg.Username
	esConfig.Password = serverEsCfg.Password

	return esConfig
}

// GetTypedClient 获取es连接对象
func GetTypedClient() *elasticsearch.TypedClient {
	if globalClient == nil {
		var err error
		globalClient, err = elasticsearch.NewTypedClient(GetElasticConfig())
		if err != nil {
			logging.RuntimeLog.Errorf("Error creating the client: %s", err)
		}
	}
	return globalClient
}

// CheckElasticConn 检查ES连接是否正常
func CheckElasticConn() bool {
	cfg := GetElasticConfig()
	if cfg.Addresses == nil || len(cfg.Addresses) == 0 || cfg.Username == "" || cfg.Password == "" {
		return false
	}
	client := GetTypedClient()
	if client == nil {
		return false
	}
	_, err := globalClient.Info().Do(context.Background())
	if err != nil {
		logging.RuntimeLog.Errorf("Error getting response: %s", err)
		return false
	}

	return true
}
