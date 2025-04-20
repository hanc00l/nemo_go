package workerapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/RichardKnop/machinery/v2/log"
	"github.com/hanc00l/nemo_go/v3/pkg/core"
	"github.com/hanc00l/nemo_go/v3/pkg/logging"
	"github.com/sirupsen/logrus"
	"time"

	"github.com/RichardKnop/machinery/v2/backends/result"
	"github.com/RichardKnop/machinery/v2/tasks"
)

var WStatus core.WorkerStatus

// taskMaps 定义work执行的任务；在添加了对应的任务后，在ampq/api.go中指定任务对应的队列映射：taskTopicDefineMap
var taskMaps = map[string]interface{}{
	"nmap":        PortScan,
	"masscan":     PortScan,
	"gogo":        PortScan,
	"subfinder":   DomainScan,
	"massdns":     DomainScan,
	"fofa":        OnlineAPI,
	"hunter":      OnlineAPI,
	"quake":       OnlineAPI,
	"whois":       QueryData,
	"icp":         QueryData,
	"icpPlus":     ICPPlusScan,
	"fingerprint": Fingerprint,
	"nuclei":      PocScan,
	"qwen":        LLMScan,
	"kimi":        LLMScan,
	"deepseek":    LLMScan,
	//test:
	"test": TaskTest,
}

// SucceedTask 任务执行成功的状态的结果
func SucceedTask(msg string) string {
	r := core.TaskResult{Status: core.SUCCESS, Msg: msg}
	js, _ := json.Marshal(r)
	return string(js)
}

// FailedTask 任务执行失败的状态的结果
func FailedTask(msg string) string {
	r := core.TaskResult{Status: core.FAILURE, Msg: msg}
	js, _ := json.Marshal(r)
	return string(js)
}

// ParseConfig 解析任务执行的参数
func ParseConfig(configJSON string, config interface{}) (err error) {
	err = json.Unmarshal([]byte(configJSON), &config)
	if err != nil {
		logging.RuntimeLog.Error(err)
	}
	return
}

// StartWorker 启动worker
func StartWorker(topicName string, concurrency int) error {
	server := core.GetWorkerMQServer(topicName, concurrency)
	err := server.RegisterTasks(taskMaps)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return err
	}

	worker := server.NewWorker(WStatus.WorkerName, concurrency)
	worker.SetPreTaskHandler(preTaskHandler)
	worker.SetPostTaskHandler(postTaskHandler)
	//设置machinery的日志和Level
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(logging.GetCustomLoggerFormatter())
	log.Set(logger)
	logging.RuntimeLog.Infof("starting worker: %s", WStatus.WorkerName)
	logging.CLILog.Infof("starting worker: %s", WStatus.WorkerName)

	return worker.Launch()
}

// postTaskHandler 任务完成时处理工作
func postTaskHandler(signature *tasks.Signature) {
	//更新任务执行state和worker
	WStatus.Lock()
	WStatus.TaskStartedNumber--
	WStatus.Unlock()

	//log.INFO.Println("I am an end of task handler for:", signature.ProfileName)
	server := core.GetWorkerMQServer(core.GetTopicByMQRoutingKey(signature.RoutingKey), 3)
	r := result.NewAsyncResult(signature, server.GetBackend())
	rr, _ := r.Get(time.Duration(0) * time.Second)
	UpdateTaskStatus(signature.UUID, r.GetState().State, WStatus.WorkerName, tasks.HumanReadableResults(rr))
}

// preTaskHandler 任务开始前的处理工作
func preTaskHandler(signature *tasks.Signature) {
	//更新任务执行state和worker
	WStatus.Lock()
	WStatus.TaskExecutedNumber++
	WStatus.TaskStartedNumber++
	WStatus.Unlock()

	var taskStatus core.TaskStatusArgs
	var checkTaskArgs = core.CheckTaskArgs{
		TaskID: signature.UUID,
	}
	if err := core.CallXClient("CheckTask", &checkTaskArgs, &taskStatus); err != nil {
		return
	}
	if !taskStatus.IsExist {
		return
	}
	UpdateTaskStatus(signature.UUID, core.STARTED, WStatus.WorkerName, "")
}

// CheckTaskStatus 检查任务状态：是否不存在或取消
func CheckTaskStatus(taskId string, mainTaskId string) (ok bool, result string, err error) {
	var taskStatus core.TaskStatusArgs
	var taskStatusArgs = core.CheckTaskArgs{
		TaskID:     taskId,
		MainTaskID: mainTaskId,
	}
	if err = core.CallXClient("CheckTask", &taskStatusArgs, &taskStatus); err != nil {
		logging.RuntimeLog.Error(err)
		return false, FailedTask(err.Error()), err
	}
	if !taskStatus.IsExist {
		logging.RuntimeLog.Warningf("task not exists: %s", taskId)
		return false, FailedTask("task not exist"), errors.New("task not exist")
	}
	if taskStatus.IsFinished {
		logging.RuntimeLog.Warningf("task has finished: %s", taskId)
		return false, SucceedTask(""), errors.New("task has finished")
	}
	return true, "", nil
}

// UpdateTaskStatus 更新任务状态
func UpdateTaskStatus(taskId string, state string, worker string, result string) bool {
	taskStatus := core.TaskStatusArgs{
		TaskID: taskId,
		State:  state,
		Worker: worker,
		Result: result,
	}
	var updateStatus bool
	if err := core.CallXClient("UpdateTask", &taskStatus, &updateStatus); err != nil {
		logging.RuntimeLog.Error(err)
		return false
	}
	return updateStatus
}

// TaskTest 测试任务
func TaskTest(taskId, r string) (string, error) {
	fmt.Println(taskId)
	fmt.Println("sleep...")
	time.Sleep(time.Second * 5)
	return SucceedTask(r), nil
}

// NopTask 空任务
func NopTask(taskId, r string) (string, error) {
	return SucceedTask(r), nil
}
