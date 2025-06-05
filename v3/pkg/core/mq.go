package core

import (
	"fmt"
	"github.com/RichardKnop/machinery/v2"
	redisbackend "github.com/RichardKnop/machinery/v2/backends/redis"
	redisbroker "github.com/RichardKnop/machinery/v2/brokers/redis"
	"github.com/RichardKnop/machinery/v2/config"
	eagerlock "github.com/RichardKnop/machinery/v2/locks/eager"
	"github.com/hanc00l/nemo_go/v3/pkg/conf"
	"strings"
)

type WorkerRunTaskMode int

const (
	TaskModeDefault WorkerRunTaskMode = iota
	TaskModeActive
	TaskModeFinger
	TaskModePassive
	TaskModePocscan
	TaskModeStandalone
)

// taskTopicDefineMap 每个task对应的队列名称，以便分配到执行不同任务的worker
var taskTopicDefineMap = map[string]string{
	"nmap":        TopicActive,
	"masscan":     TopicActive,
	"gogo":        TopicActive,
	"subfinder":   TopicPassive,
	"massdns":     TopicPassive,
	"fofa":        TopicPassive,
	"quake":       TopicPassive,
	"hunter":      TopicPassive,
	"whois":       TopicPassive,
	"icp":         TopicPassive,
	"icpPlus":     TopicPassive,
	"qwen":        TopicPassive,
	"kimi":        TopicPassive,
	"deepseek":    TopicPassive,
	"fingerprint": TopicFinger,
	"nuclei":      TopicPocscan,
	"zombie":      TopicPocscan,
	"standalone":  TopicStandalone,
	//test:
	"test": TopicCustom,
}

// taskServer 复用的全局MQ消息队列连接
var taskServerConn = make(map[string]*machinery.Server)

// CustomTaskWorkspaceMap 自定义任务关联的工作空间GUID
var CustomTaskWorkspaceMap = make(map[string]struct{})

// GetServerTaskMQServer 根据server配置文件，获取到消息中心的连接
func GetServerTaskMQServer(topicName string) *machinery.Server {
	if _, ok := taskServerConn[topicName]; !ok {
		mq := conf.GlobalServerConfig().Redis
		taskServerConn[topicName] = startMQServer(mq.Password, mq.Host, mq.Port, topicName, 3)
	}
	return taskServerConn[topicName]
}

// GetWorkerMQServer 根据worker配置文件，获取到消息中心的连接
func GetWorkerMQServer(topicName string, prefetchCount int) *machinery.Server {
	if _, ok := taskServerConn[topicName]; !ok {
		mq := conf.GlobalWorkerConfig().Redis
		taskServerConn[topicName] = startMQServer(mq.Password, mq.Host, mq.Port, topicName, prefetchCount)
	}
	return taskServerConn[topicName]
}

// startMQServer 连接到MQ消息队列服务器
func startMQServer(password, host string, port int, topicName string, prefetchCount int) *machinery.Server {
	mqConfig := fmt.Sprintf("redis://%s:%d", host, port)
	grConfig := fmt.Sprintf("%s:%d", host, port)
	if len(password) > 0 {
		mqConfig = fmt.Sprintf("redis://%s@%s:%d", password, host, port)
		grConfig = fmt.Sprintf("%s@%s:%d", password, host, port)
	}
	routingKey := GetRoutingKeyByTopic(topicName)
	cnf := &config.Config{
		Broker:          mqConfig,
		DefaultQueue:    routingKey,
		ResultBackend:   mqConfig,
		ResultsExpireIn: 300,
		Redis: &config.RedisConfig{
			MaxIdle:                3,
			IdleTimeout:            240,
			ReadTimeout:            15,
			WriteTimeout:           15,
			ConnectTimeout:         15,
			NormalTasksPollPeriod:  1000,
			DelayedTasksPollPeriod: 500,
		},
	}
	broker := redisbroker.NewGR(cnf, []string{grConfig}, 0)
	backend := redisbackend.NewGR(cnf, []string{grConfig}, 0)
	lock := eagerlock.New()
	server := machinery.NewServer(cnf, broker, backend, lock)

	return server
}

func GetTopicByTaskName(taskName string, workspaceGUID string) string {
	if _, ok := CustomTaskWorkspaceMap[workspaceGUID]; ok {
		// custom.1a0ca919-7960-4067-9981-9abcb4eaa735
		return fmt.Sprintf("%s.%s", TopicCustom, workspaceGUID)
	}
	if queueName, ok := taskTopicDefineMap[taskName]; ok {
		// active、finger...
		return queueName
	}
	return ""
}

func GetTopicByMQRoutingKey(routingKey string) string {
	keys := strings.Split(routingKey, ".")
	if len(keys) == 3 {
		// nemo_mq.custom.1a0ca919-7960-4067-9981-9abcb4eaa735
		return fmt.Sprintf("%s.%s", keys[1], keys[2])
	} else if len(keys) == 2 {
		// nemo_mq.active...
		return keys[1]
	}
	return ""
}

func GetRoutingKeyByTopic(topicName string) string {
	return fmt.Sprintf("%s.%s", TopicMQPrefix, topicName)
}
