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

// NewTask 创建一个新执行任务
func NewTask(taskName string, configJSON string) (taskId string, err error) {
	server := ampq.GetServerTaskAMPQSrever()
	// 延迟5秒后执行
	eta := time.Now().Add(time.Second * 5)
	taskId = uuid.New().String()
	workerTask := tasks.Signature{
		Name: taskName,
		UUID: taskId,
		ETA:  &eta,
		Args: []tasks.Arg{
			{Name: "taskId", Type: "string", Value: taskId},
			{Name: "configJSON", Type: "string", Value: configJSON},
		},
	}
	_, err = server.SendTask(&workerTask)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return "", err
	}
	addTask(taskId, taskName, configJSON)

	return taskId, nil
}

// RevokeUnexcusedTask 取消一个未开始执行的任务
func RevokeUnexcusedTask(taskId string) (isRevoked bool, err error) {
	task := &db.Task{TaskId: taskId}
	if !task.GetByTaskId() {
		logging.RuntimeLog.Errorf("Task not exists when revoked: %s", taskId)
		return false, errors.New("task not exists")
	}
	//检查状态，只有CREATED状态的才能取消
	if task.State == ampq.CREATED {
		updateRevokedTask(taskId)
		logging.RuntimeLog.Infof("Task revoked: %s", taskId)
		return true, nil
	}
	return false, nil
}

// addTask 将任务写入到数据库中
func addTask(taskId, taskName, kwArgs string) {
	dt := time.Now()
	task := &db.Task{
		TaskId:       taskId,
		TaskName:     taskName,
		KwArgs:       kwArgs,
		State:        ampq.CREATED,
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

// updateRevokedTask 更新取消的任务状态
func updateRevokedTask(taskId string) {
	dt := time.Now()
	task := &db.Task{
		TaskId:      taskId,
		State:       ampq.REVOKED,
		RevokedTime: &dt,
	}
	if !task.SaveOrUpdate() {
		logging.RuntimeLog.Errorf("Update task:%s,state:%s fail !", taskId, ampq.REVOKED)
	}
}
