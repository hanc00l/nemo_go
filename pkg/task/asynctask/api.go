package asynctask

import (
	"fmt"
	"github.com/RichardKnop/machinery/v2"
	amqpbackend "github.com/RichardKnop/machinery/v2/backends/amqp"
	amqpbroker "github.com/RichardKnop/machinery/v2/brokers/amqp"
	"github.com/RichardKnop/machinery/v2/config"
	eagerlock "github.com/RichardKnop/machinery/v2/locks/eager"
	"github.com/RichardKnop/machinery/v2/tasks"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/db"
	"github.com/hanc00l/nemo_go/pkg/logging"
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
	sync.Mutex         `json:"-"`
	WorkerName         string    `json:"worker_name"'`
	CreateTime         time.Time `json:"create_time"`
	UpdateTime         time.Time `json:"update_time"`
	TaskExecutedNumber int       `json:"task_number"`
}

// GetTaskServer 获取到消息中心的连接
func GetTaskServer() (taskServer *machinery.Server) {
	return StartAMQPServer()
}

// StartAMQPServer 连接到AMQP消息队列服务器
func StartAMQPServer() *machinery.Server {
	amqpConfig := fmt.Sprintf("amqp://%s:%s@%s:%d/", conf.Nemo.Rabbitmq.Username,
		conf.Nemo.Rabbitmq.Password,
		conf.Nemo.Rabbitmq.Host,
		conf.Nemo.Rabbitmq.Port)
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

// AddTask 将任务写入到数据库中
func AddTask(taskId, taskName, kwArgs string) {
	dt := time.Now()
	task := &db.Task{
		TaskId:       taskId,
		TaskName:     taskName,
		KwArgs:       kwArgs,
		State:        CREATED,
		ReceivedTime: &dt,
	}
	//kwargs可能因为target很多导致超过数据库中的字段设计长度，因此作一个长度截取
	const argsLength = 2000
	if len(kwArgs) > argsLength {
		task.KwArgs = fmt.Sprintf("%s...", kwArgs[:argsLength])
	}
	if !task.Add() {
		logging.RuntimeLog.Errorf("Add new task fail: %s,%s,%s", taskId, taskName, kwArgs)
	}
}

// UpdateTask 更新任务状态、结果到数据库中
func UpdateTask(taskId, state, worker, result string) {
	dt := time.Now()
	task := &db.Task{
		TaskId: taskId,
		State:  state,
		Worker: worker,
		Result: result,
	}
	switch state {
	case SUCCESS:
		task.SucceededTime = &dt
	case FAILURE:
		task.FailedTime = &dt
	case REVOKED:
		task.RevokedTime = &dt
	case STARTED:
		task.StartedTime = &dt
	case RETRY:
		task.RetriedTime = &dt
	case RECEIVED:
		task.ReceivedTime = &dt
	}
	if !task.SaveOrUpdate() {
		logging.RuntimeLog.Errorf("Update task:%s,state:%s fail !", taskId, state)
	}
}
