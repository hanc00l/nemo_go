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

	TopicActive  = "active"
	TopicFinger  = "finger"
	TopicPassive = "passive"
	TopicPocscan = "pocscan"
	TopicCustom  = "custom"

	TopicMQPrefix = "nemo_mq"
)

type TaskResult struct {
	Status string `json:"status"`
	Msg    string `json:"msg"`
}

type WorkerStatus struct {
	sync.Mutex             `json:"-"`
	WorkerName             string    `json:"worker_name"`
	WorkerTopics           string    `json:"worker_topic"`
	CreateTime             time.Time `json:"create_time"`
	UpdateTime             time.Time `json:"update_time"`
	TaskExecutedNumber     int       `json:"task_number"`
	ManualReloadFlag       bool      `json:"manual_reload_flag"`
	ManualFileSyncFlag     bool      `json:"manual_file_sync_flag"`
	WorkerDaemonUpdateTime time.Time `json:"worker_daemon_update_time"`
}

type WorkerRunTaskMode int

const (
	TaskModeDefault WorkerRunTaskMode = iota
	TaskModeActive
	TaskModeFinger
	TaskModePassive
	TaskModePocscan
	TaskModeCustom
)

// taskTopicDefineMap 每个task对应的队列名称，以便分配到执行不同任务的worker
var taskTopicDefineMap = map[string]string{
	"portscan":          TopicActive,
	"batchscan":         TopicActive,
	"domainscan":        TopicActive,
	"subfinder":         TopicPassive,
	"subdomainbrute":    TopicPassive,
	"subdomaincrawler":  TopicActive,
	"iplocation":        TopicPassive,
	"fofa":              TopicPassive,
	"quake":             TopicPassive,
	"hunter":            TopicPassive,
	"xray":              TopicPocscan,
	"dirsearch":         TopicPocscan,
	"nuclei":            TopicPocscan,
	"goby":              TopicPocscan,
	"icpquery":          TopicPassive,
	"whoisquery":        TopicPassive,
	"fingerprint":       TopicFinger,
	"xportscan":         TopicActive,
	"xonlineapi":        TopicPassive,
	"xfofa":             TopicPassive,
	"xquake":            TopicPassive,
	"xhunter":           TopicPassive,
	"xdomainscan":       TopicPassive,
	"xsubfinder":        TopicPassive,
	"xsubdomainbrute":   TopicPassive,
	"xsubdomaincralwer": TopicActive,
	"xfingerprint":      TopicFinger,
	"xxray":             TopicPocscan,
	"xnuclei":           TopicPocscan,
	"xgoby":             TopicPocscan,
	"xorgscan":          TopicActive,
	//test:
	"test": TopicCustom,
}

// taskServer 复用的全局AMQP连接
var taskServerConn = make(map[string]*machinery.Server)

// CustomTaskWorkspaceMap 自定义任务关联的工作空间GUID
var CustomTaskWorkspaceMap = make(map[string]struct{})

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
	routingKey := GetRoutingKeyByTopic(topicName)
	cnf := &config.Config{
		Broker:          amqpConfig,
		DefaultQueue:    routingKey,
		ResultBackend:   amqpConfig,
		ResultsExpireIn: 300,
		AMQP: &config.AMQPConfig{
			Exchange:      "nemo_mq_exchange",
			ExchangeType:  "topic",
			BindingKey:    routingKey,
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
