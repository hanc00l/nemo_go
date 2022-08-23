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
)

type TaskResult struct {
	Status string `json:"status"`
	Msg    string `json:"msg"`
}

type WorkerStatus struct {
	sync.Mutex             `json:"-"`
	WorkerName             string    `json:"worker_name"`
	CreateTime             time.Time `json:"create_time"`
	UpdateTime             time.Time `json:"update_time"`
	TaskExecutedNumber     int       `json:"task_number"`
	ManualReloadFlag       bool      `json:"manual_reload_flag"`
	ManualFileSyncFlag     bool      `json:"manual_file_sync_flag"`
	WorkerDaemonUpdateTime time.Time `json:"worker_daemon_update_time"`
}

// GetServerTaskAMPQSrever 根据server配置文件，获取到消息中心的连接
func GetServerTaskAMPQSrever() (taskServer *machinery.Server) {
	rabbitmq := conf.GlobalServerConfig().Rabbitmq
	return startAMQPServer(rabbitmq.Username, rabbitmq.Password, rabbitmq.Host, rabbitmq.Port)
}

// GetWorkerAMPQServer 根据worker配置文件，获取到消息中心的连接
func GetWorkerAMPQServer() (taskServer *machinery.Server) {
	rabbitmq := conf.GlobalWorkerConfig().Rabbitmq
	return startAMQPServer(rabbitmq.Username, rabbitmq.Password, rabbitmq.Host, rabbitmq.Port)
}

// startAMQPServer 连接到AMQP消息队列服务器
func startAMQPServer(username, password, host string, port int) *machinery.Server {
	amqpConfig := fmt.Sprintf("amqp://%s:%s@%s:%d/", username, password, host, port)
	cnf := &config.Config{
		Broker:          amqpConfig,
		DefaultQueue:    "machinery_tasks",
		ResultBackend:   amqpConfig,
		ResultsExpireIn: 3600,
		AMQP: &config.AMQPConfig{
			Exchange:      "machinery_exchange",
			ExchangeType:  "direct",
			BindingKey:    "machinery_task",
			PrefetchCount: 3,
		},
	}
	// Create server instance
	broker := amqpbroker.New(cnf)
	backend := amqpbackend.New(cnf)
	lock := eagerlock.New()
	server := machinery.NewServer(cnf, broker, backend, lock)

	return server
}
