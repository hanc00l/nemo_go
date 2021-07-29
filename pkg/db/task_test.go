package db

import (
	"testing"
	"time"
)

func TestTask_GetsBySearch(t *testing.T) {
	searchMap := make(map[string]interface{})
	searchMap["task_name"] = "portscan"
	searchMap["state"] = "SUCCESS"

	task := &Task{}
	tasklist,count := task.Gets(searchMap,-1,-1)
	t.Log(count)
	for _,ta := range tasklist{
		t.Log(ta)
	}
}

func TestTask_SaveOrUpdate(t *testing.T) {
	taskID := "b9cd7ecc-ddb0-4160-9c41-75c55ffa212f"
	taskOld := &Task{TaskId: taskID}
	taskOld.GetByTaskId()
	t.Log(taskOld)

	dt := time.Now()
	task := &Task{TaskId: taskID,State: "FAIL",FailedTime: &dt}
	task.SaveOrUpdate()

	taskNew := &Task{TaskId: taskID}
	taskNew.GetByTaskId()
	t.Log(taskNew)
}