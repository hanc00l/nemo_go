package workerapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/RichardKnop/machinery/v2/log"
	"github.com/hanc00l/nemo_go/pkg/comm"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/ampq"
	"github.com/hanc00l/nemo_go/pkg/utils"
	"github.com/sirupsen/logrus"
	"os"
	"time"

	"github.com/RichardKnop/machinery/v2/backends/result"
	"github.com/RichardKnop/machinery/v2/tasks"
)

var WStatus ampq.WorkerStatus

// taskMaps 定义work执行的任务
var taskMaps = map[string]interface{}{
	"portscan":   PortScan,
	"batchscan":  BatchScan,
	"domainscan": DomainScan,
	"iplocation": IPLocation,
	"fofa":       Fofa,
	"quake":      Quake,
	"hunter":     Hunter,
	"xray":       PocScan,
	"pocsuite":   PocScan,
	"dirsearch":  PocScan,
	"nuclei":	PocScan,
	"icpquery":   ICPQuery,
	"test":       TaskTest,
}

// SucceedTask 任务执行成功的状态的结果
func SucceedTask(msg string) string {
	r := ampq.TaskResult{Status: ampq.SUCCESS, Msg: msg}
	js, _ := json.Marshal(r)
	return string(js)
}

// FailedTask 任务执行失败的状态的结果
func FailedTask(msg string) string {
	r := ampq.TaskResult{Status: ampq.FAILURE, Msg: msg}
	js, _ := json.Marshal(r)
	return string(js)
}

// RevokedTask 任务取消的状态和消息
func RevokedTask(msg string) string {
	r := ampq.TaskResult{Status: ampq.REVOKED, Msg: msg}
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

	server := ampq.GetWorkerAMPQServer()
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

// postTaskHandler 任务完成时处理工作
func postTaskHandler(signature *tasks.Signature) {
	//log.INFO.Println("I am an end of task handler for:", signature.Name)
	server := ampq.GetWorkerAMPQServer()
	r := result.NewAsyncResult(signature, server.GetBackend())
	rr, _ := r.Get(time.Duration(0) * time.Second)
	//更新任务的结果和状态
	var tr ampq.TaskResult
	//检查REVOKED的任务
	if err := json.Unmarshal([]byte(tasks.HumanReadableResults(rr)), &tr); err == nil {
		if tr.Status == ampq.REVOKED {
			UpdateTaskStatus(signature.UUID, ampq.REVOKED, WStatus.WorkerName, tasks.HumanReadableResults(rr))
			return
		}
	}
	UpdateTaskStatus(signature.UUID, r.GetState().State, WStatus.WorkerName, tasks.HumanReadableResults(rr))
}

// preTaskHandler 任务开始前的处理工作
func preTaskHandler(signature *tasks.Signature) {
	//更新任务执行state和worker
	WStatus.Lock()
	WStatus.TaskExecutedNumber++
	WStatus.Unlock()

	var taskStatus comm.TaskStatusArgs
	x := comm.NewXClient()
	err := x.Call(context.Background(), "CheckTask", &signature.UUID, &taskStatus)
	if err != nil {
		return
	}
	if !taskStatus.IsExist {
		return
	}
	//REVOKED的任务：不要更新状态，只更新workName
	if taskStatus.State == ampq.REVOKED {
		UpdateTaskStatus(signature.UUID, "", WStatus.WorkerName, "")
		return
	}
	UpdateTaskStatus(signature.UUID, ampq.STARTED, WStatus.WorkerName, "")
}

// CheckTaskStatus 检查任务状态：是否不存在或取消
func CheckTaskStatus(taskId string) (ok bool, result string, err error) {
	var taskStatus comm.TaskStatusArgs
	x := comm.NewXClient()
	err = x.Call(context.Background(), "CheckTask", &taskId, &taskStatus)
	if err != nil {
		return false, FailedTask(err.Error()), err
	}
	if !taskStatus.IsExist {
		logging.RuntimeLog.Errorf("task not exists: %s", taskId)
		return false, FailedTask("task not exist"), errors.New("task not exist")
	}
	if taskStatus.IsRevoked {
		return false, RevokedTask(""), nil
	}
	return true, "", nil
}

// UpdateTaskStatus 更新任务状态
func UpdateTaskStatus(taskId string, state string, worker string, result string) bool {
	taskStatus := comm.TaskStatusArgs{
		TaskID: taskId,
		State:  state,
		Worker: worker,
		Result: result,
	}
	var updateStatus bool
	x := comm.NewXClient()
	err := x.Call(context.Background(), "UpdateTask", &taskStatus, &updateStatus)
	if err != nil {
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
