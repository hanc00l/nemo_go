package db

import (
	"context"
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

func GetClient() (*mongo.Client, error) {
	if len(configString) == 0 {
		getDbConfig()
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := mongo.Connect(options.Client().ApplyURI(configString))
	if err != nil {
		logging.RuntimeLog.Fatal("Failed to connect database")
		return nil, err
	}
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		logging.RuntimeLog.Fatal("Failed to connect database:")
		return nil, err
	}

	return client, nil
}

func CloseClient(client *mongo.Client) {
	if client != nil {
		err := client.Disconnect(context.Background())
		if err != nil {
			return
		}
	}
}
