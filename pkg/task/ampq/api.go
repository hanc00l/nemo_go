package ampq

import (
	"fmt"
	"github.com/RichardKnop/machinery/v2"
	amqpbackend "github.com/RichardKnop/machinery/v2/backends/amqp"
	amqpbroker "github.com/RichardKnop/machinery/v2/brokers/amqp"
	"github.com/RichardKnop/machinery/v2/config"
	eagerlock "github.com/RichardKnop/machinery/v2/locks/eager"
	"github.com/RichardKnop/machinery/v2/tasks"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"strings"
	"sync"
	"time"
)

const (
	CREATED  string = "CREATED"           //任务创建，但还没有开始执行
	REVOKED  string = "REVOKED"           //任务被取消执行
	STARTED  string = tasks.StateStarted  //任务在执行中
	SUCCESS  string = tasks.StateSuccess  //任务执行完成，结果为SUCCESS
	FAILURE  string = tasks.StateFailure  //任务执行完成，结果为FAILURE
	RECEIVED string = tasks.StateReceived //未使用
	PENDING  string = tasks.StatePending  //未使用
	RETRY    string = tasks.StateRetry    //未使用

	TopicActive     = "active"
	TopicFinger     = "finger"
	TopicPassive    = "passive"
	TopicCustom     = "custom"
	TopicDefaultAll = "*"

	TopicMQPrefix = "nemo_mq"
)

type WorkerRunTaskMode int

const (
	TaskModeDefault WorkerRunTaskMode = iota
	TaskModeActive
	TaskModeFinger
	TaskModePassive
	TaskModeCustom
)

type TaskResult struct {
	Status string `json:"status"`
	Msg    string `json:"msg"`
}

type WorkerStatus struct {
	sync.Mutex             `json:"-"`
	WorkerName             string    `json:"worker_name"`
	WorkerRunTaskMode      string    `json:"worker_mode"`
	CreateTime             time.Time `json:"create_time"`
	UpdateTime             time.Time `json:"update_time"`
	TaskExecutedNumber     int       `json:"task_number"`
	ManualReloadFlag       bool      `json:"manual_reload_flag"`
	ManualFileSyncFlag     bool      `json:"manual_file_sync_flag"`
	WorkerDaemonUpdateTime time.Time `json:"worker_daemon_update_time"`
}

// taskTopicDefineMap 每个task对应的队列名称，以便分配到执行不同任务的worker
var taskTopicDefineMap = map[string]string{
	"portscan":    TopicActive,
	"batchscan":   TopicActive,
	"domainscan":  TopicActive,
	"iplocation":  TopicPassive,
	"fofa":        TopicPassive,
	"quake":       TopicPassive,
	"hunter":      TopicPassive,
	"xray":        TopicActive,
	"dirsearch":   TopicActive,
	"nuclei":      TopicActive,
	"goby":        TopicActive,
	"icpquery":    TopicPassive,
	"whoisquery":  TopicPassive,
	"fingerprint": TopicFinger,
	//xscan:
	"xonlineapi":   TopicPassive,
	"xportscan":    TopicActive,
	"xdomainscan":  TopicActive,
	"xfingerprint": TopicFinger,
	"xxray":        TopicActive,
	"xnuclei":      TopicActive,
	"xgoby":        TopicActive,
	"xorgscan":     TopicActive,
	//test:
	"test": TopicCustom,
}

// workerTopicDefineMap 每个不同工作模式的worker，对应队列
var workerTopicDefineMap = map[WorkerRunTaskMode]string{
	TaskModeDefault: TopicDefaultAll,
	TaskModeActive:  TopicActive,
	TaskModeFinger:  TopicFinger,
	TaskModePassive: TopicPassive,
	TaskModeCustom:  TopicCustom,
}

// taskServer 复用的全局AMQP连接
var taskServerConn = make(map[string]*machinery.Server)

// GetServerTaskAMPQServer 根据server配置文件，获取到消息中心的连接
func GetServerTaskAMPQServer(topicName string) *machinery.Server {
	if _, ok := taskServerConn[topicName]; !ok {
		rabbitmq := conf.GlobalServerConfig().Rabbitmq
		taskServerConn[topicName] = startAMQPServer(rabbitmq.Username, rabbitmq.Password, rabbitmq.Host, rabbitmq.Port, topicName, 3)
	}
	return taskServerConn[topicName]
}

// GetWorkerAMPQServer 根据worker配置文件，获取到消息中心的连接
func GetWorkerAMPQServer(topicName string, prefetchCount int) *machinery.Server {
	if _, ok := taskServerConn[topicName]; !ok {
		rabbitmq := conf.GlobalWorkerConfig().Rabbitmq
		taskServerConn[topicName] = startAMQPServer(rabbitmq.Username, rabbitmq.Password, rabbitmq.Host, rabbitmq.Port, topicName, prefetchCount)
	}
	return taskServerConn[topicName]
}

// startAMQPServer 连接到AMQP消息队列服务器
func startAMQPServer(username, password, host string, port int, topicName string, prefetchCount int) *machinery.Server {
	amqpConfig := fmt.Sprintf("amqp://%s:%s@%s:%d/", username, password, host, port)
	queueName := fmt.Sprintf("%s.%s", TopicMQPrefix, topicName)
	cnf := &config.Config{
		Broker:          amqpConfig,
		DefaultQueue:    queueName,
		ResultBackend:   amqpConfig,
		ResultsExpireIn: 300,
		AMQP: &config.AMQPConfig{
			Exchange:      "nemo_mq_exchange",
			ExchangeType:  "topic",
			BindingKey:    queueName,
			PrefetchCount: prefetchCount,
		},
	}
	// Create server instance
	broker := amqpbroker.New(cnf)
	backend := amqpbackend.New(cnf)
	lock := eagerlock.New()
	server := machinery.NewServer(cnf, broker, backend, lock)

	return server
}

func GetTopicByTaskName(taskName string) string {
	if queueName, ok := taskTopicDefineMap[taskName]; ok {
		return queueName
	}
	return ""
}

func GetTopicByWorkerMode(workerMode WorkerRunTaskMode) string {
	return workerTopicDefineMap[workerMode]
}

func GetTopicByMQRoutingKey(routingKey string) string {
	keys := strings.Split(routingKey, ".")
	if len(keys) >= 2 {
		return keys[1]
	}
	return ""
}

func GetRoutingKeyByTopic(topicName string) string {
	return fmt.Sprintf("%s.%s", TopicMQPrefix, topicName)
}
