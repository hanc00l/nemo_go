package workerapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/RichardKnop/machinery/v2/log"
	"github.com/hanc00l/nemo_go/pkg/db"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/asynctask"
	"github.com/hanc00l/nemo_go/pkg/utils"
	"github.com/sirupsen/logrus"
	"os"
	"time"

	"github.com/RichardKnop/machinery/v2/backends/result"
	"github.com/RichardKnop/machinery/v2/tasks"
)

var WStatus asynctask.WorkerStatus

// taskMaps 定义work执行的任务
var taskMaps = map[string]interface{}{
	"portscan":   PortScan,
	"domainscan": DomainScan,
	"iplocation": IPLocation,
	"fofa":       Fofa,
	"xray":       PocScan,
	"pocsuite":   PocScan,
	"icpquery":   ICPQuery,
	"test":       TaskTest,
}

// SucceedTask 任务执行成功的状态的结果
func SucceedTask(msg string) string {
	r := asynctask.TaskResult{Status: asynctask.SUCCESS, Msg: msg}
	js, _ := json.Marshal(r)
	return string(js)
}

// FailedTask 任务执行失败的状态的结果
func FailedTask(msg string) string {
	r := asynctask.TaskResult{Status: asynctask.FAILURE, Msg: msg}
	js, _ := json.Marshal(r)
	return string(js)
}

// RevokedTask 任务取消的状态和消息
func RevokedTask(msg string) string {
	r := asynctask.TaskResult{Status: asynctask.REVOKED, Msg: msg}
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
func StartWorker(concurrency int) error {
	hostIP, _ := utils.GetOutBoundIP()
	if hostIP == "" {
		hostIP, _ = utils.GetClientIp()
	}
	hostName, _ := os.Hostname()
	pid := os.Getpid()
	WStatus.WorkerName = fmt.Sprintf("%s@%s#%d", hostName, hostIP, pid)
	consumerTag := WStatus.WorkerName //"machinery_worker"

	server := asynctask.GetTaskServer()
	err := server.RegisterTasks(taskMaps)
	if err != nil {
		return err
	}
	worker := server.NewWorker(consumerTag, concurrency)
	worker.SetPreTaskHandler(preTaskHandler)
	worker.SetPostTaskHandler(postTaskHandler)
	WStatus.CreateTime = time.Now()
	WStatus.UpdateTime = time.Now()
	//设置machinery的日志和Level
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(logging.GetCustomLoggerFormatter())
	log.Set(logger)
	logging.RuntimeLog.Infof("Starting Worker: %s", WStatus.WorkerName)
	logging.CLILog.Infof("Starting Worker: %s", WStatus.WorkerName)

	return worker.Launch()
}

// CheckIsExistOrRevoked 检查任务是否被删除或取消
func CheckIsExistOrRevoked(taskId string) (isRevoked bool, err error) {
	task := &db.Task{TaskId: taskId}
	if !task.GetByTaskId() {
		logging.RuntimeLog.Errorf("task not exists: %s", taskId)
		return false, errors.New("task not exists")
	}
	if task.State == asynctask.REVOKED {
		return true, nil
	}
	return false, nil
}

// postTaskHandler 任务完成时处理工作
func postTaskHandler(signature *tasks.Signature) {
	//log.INFO.Println("I am an end of task handler for:", signature.Name)
	server := asynctask.GetTaskServer()
	r := result.NewAsyncResult(signature, server.GetBackend())
	rr, _ := r.Get(time.Duration(0) * time.Second)
	//更新任务的结果和状态
	var tr asynctask.TaskResult
	//检查REVOKED的任务
	if err := json.Unmarshal([]byte(tasks.HumanReadableResults(rr)), &tr); err == nil {
		if tr.Status == asynctask.REVOKED {
			asynctask.UpdateTask(signature.UUID, asynctask.REVOKED, WStatus.WorkerName, tasks.HumanReadableResults(rr))
			return
		}
	}
	asynctask.UpdateTask(signature.UUID, r.GetState().State, WStatus.WorkerName, tasks.HumanReadableResults(rr))
}

// preTaskHandler 任务开始前的处理工作
func preTaskHandler(signature *tasks.Signature) {
	//log.INFO.Println("I am a start of task handler for:", signature.Name)
	//更新任务执行state和worker
	WStatus.Lock()
	WStatus.TaskExecutedNumber++
	WStatus.Unlock()
	task := &db.Task{TaskId: signature.UUID}
	if task.GetByTaskId() {
		//REVOKED的任务：不要更新状态，只更新workName
		if task.State == asynctask.REVOKED {
			asynctask.UpdateTask(signature.UUID, "", WStatus.WorkerName, "")
		} else {
			asynctask.UpdateTask(signature.UUID, asynctask.STARTED, WStatus.WorkerName, "")
		}
	}
}

// TaskTest 测试任务
func TaskTest(taskId, r string) (string, error) {
	isRevoked, err := CheckIsExistOrRevoked(taskId)
	if err != nil {
		return FailedTask(err.Error()), err
	}
	if isRevoked {
		return RevokedTask(""), nil
	}

	fmt.Println(taskId)
	fmt.Println("sleep...")
	time.Sleep(time.Second * 5)

	return SucceedTask(r), nil
}
