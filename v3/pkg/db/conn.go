package db

import (
	"context"
	"fmt"
	"github.com/hanc00l/nemo_go/v3/pkg/conf"
	"github.com/hanc00l/nemo_go/v3/pkg/logging"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
	"strconv"
	"time"
)

var configString string

func getDbConfig() {
	dbConfig := conf.GlobalServerConfig().Database
	if dbConfig.Username != "" && dbConfig.Password != "" {
		configString = "mongodb://" + dbConfig.Username + ":" + dbConfig.Password + "@" + dbConfig.Host + ":" + strconv.Itoa(dbConfig.Port) + "/" + dbConfig.Dbname
	} else {
		configString = "mongodb://" + dbConfig.Host + ":" + strconv.Itoa(dbConfig.Port) + "/" + dbConfig.Dbname
	}
}

func GetClientWithRetry(maxRetries int, retryInterval time.Duration) (*mongo.Client, error) {
	if len(configString) == 0 {
		getDbConfig()
	}

	var client *mongo.Client
	var err error

	// 重试逻辑
	for i := 0; i < maxRetries; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		// 尝试连接
		client, err = mongo.Connect(options.Client().ApplyURI(configString))
		if err != nil {
			cancel()
			logging.RuntimeLog.Errorf("Attempt %d/%d: Failed to connect to database: %v", i+1, maxRetries, err)
			time.Sleep(retryInterval)
			continue
		}

		// 尝试 Ping
		err = client.Ping(ctx, readpref.Primary())
		cancel() // 确保在Ping后取消context
		if err != nil {
			_ = client.Disconnect(context.Background()) // 关闭无效连接
			logging.RuntimeLog.Errorf("Attempt %d/%d: Failed to ping database: %v", i+1, maxRetries, err)
			time.Sleep(retryInterval)
			continue
		}

		// 连接成功
		return client, nil
	}

	// 所有重试都失败
	return nil, fmt.Errorf("failed to connect to MongoDB after %d attempts", maxRetries)
}

func GetClient() (*mongo.Client, error) {
	return GetClientWithRetry(5, 5*time.Second)
}

func CloseClient(client *mongo.Client) {
	if client != nil {
		err := client.Disconnect(context.Background())
		if err != nil {
			return
		}
	}
}
