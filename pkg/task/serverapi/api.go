package serverapi

import (
	"errors"
	"fmt"
	"github.com/RichardKnop/machinery/v2/tasks"
	"github.com/google/uuid"
	"github.com/hanc00l/nemo_go/pkg/db"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/ampq"
	"time"
)

// NewRunTask 创建一个新执行任务
func NewRunTask(taskName, configJSON, mainTaskId, lastRunTaskId string) (taskId string, err error) {
	topicName := ampq.GetTopicByTaskName(taskName)
	if topicName == "" {
		logging.RuntimeLog.Errorf("task not defined for topic:%s", taskName)
		return "", errors.New("task not defined for topic")
	}
	server := ampq.GetServerTaskAMPQServer(topicName)
	// 延迟5秒后执行：如果不延迟，有可能任务在完成数据库之前执行，从而导致task not exist错误
	eta := time.Now().Add(time.Second * 5)
	taskId = uuid.New().String()
	workerTask := tasks.Signature{
		Name: taskName,
		UUID: taskId,
		ETA:  &eta,
		Args: []tasks.Arg{
			{Name: "taskId", Type: "string", Value: taskId},
			{Name: "mainTaskId", Type: "string", Value: mainTaskId},
			{Name: "configJSON", Type: "string", Value: configJSON},
		},
		//RoutingKey：分发到不同功能的worker队列
		RoutingKey: ampq.GetRoutingKeyByTopic(topicName),
	}
	_, err = server.SendTask(&workerTask)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return "", err
	}
	addTask(taskId, taskName, configJSON, mainTaskId, lastRunTaskId)

	return taskId, nil
}

// RevokeUnexcusedTask 取消一个未开始执行的任务
func RevokeUnexcusedTask(taskId string) (isRevoked bool, err error) {
	task := &db.TaskRun{TaskId: taskId}
	if !task.GetByTaskId() {
		logging.RuntimeLog.Warningf("task not exists when revoked:%s", taskId)
		return false, errors.New("task not exists")
	}
	//检查状态，只有CREATED状态的才能取消
	if task.State == ampq.CREATED {
		updateRevokedTask(taskId)
		logging.RuntimeLog.Infof("task revoked:%s", taskId)
		return true, nil
	}
	return false, nil
}

// addTask 将任务写入到数据库中
func addTask(taskId, taskName, kwArgs, mainTaskId, lastRunTaskId string) {
	taskMain := db.TaskMain{TaskId: mainTaskId}
	if taskMain.GetByTaskId() == false {
		logging.RuntimeLog.Errorf("add new task fail: main task %s not exist", taskId)
		logging.CLILog.Errorf("add new task fail: main task %s not exist", taskId)
		return
	}
	dt := time.Now()
	task := &db.TaskRun{
		TaskId:        taskId,
		TaskName:      taskName,
		KwArgs:        kwArgs,
		State:         ampq.CREATED,
		ReceivedTime:  &dt,
		MainTaskId:    mainTaskId,
		LastRunTaskId: lastRunTaskId,
		WorkspaceId:   taskMain.WorkspaceId,
	}
	//kwargs可能因为target很多导致超过数据库中的字段设计长度，因此作一个长度截取
	const argsLength = 6000
	if len(kwArgs) > argsLength {
		task.KwArgs = fmt.Sprintf("%s...", kwArgs[:argsLength])
	}
	if !task.Add() {
		logging.RuntimeLog.Errorf("add new task fail: %s,%s,%s", taskId, taskName, kwArgs)
	}
}

// updateRevokedTask 更新取消的任务状态
func updateRevokedTask(taskId string) {
	dt := time.Now()
	task := &db.TaskRun{
		TaskId:      taskId,
		State:       ampq.REVOKED,
		RevokedTime: &dt,
	}
	if !task.SaveOrUpdate() {
		logging.RuntimeLog.Errorf("update task:%s,state:%s fail !", taskId, ampq.REVOKED)
	}
}
