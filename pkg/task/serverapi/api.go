package serverapi

import (
	"errors"
	"github.com/RichardKnop/machinery/v2/tasks"
	"github.com/google/uuid"
	"github.com/hanc00l/nemo_go/pkg/db"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/asynctask"
	"time"
)

// NewTask 创建一个新执行任务
func NewTask(taskName string, configJSON string) (taskId string, err error) {
	server := asynctask.GetTaskServer()
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
	asynctask.AddTask(taskId, taskName, configJSON)

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
	if task.State == asynctask.CREATED {
		asynctask.UpdateTask(taskId, asynctask.REVOKED, "", "")
		logging.RuntimeLog.Infof("Task revoked: %s", taskId)
		return true, nil
	}
	return false, nil
}
