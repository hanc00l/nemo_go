package runner

import (
	"github.com/google/uuid"
	"github.com/hanc00l/nemo_go/pkg/db"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/robfig/cron/v3"
	"sync"
	"time"
)

type CronTaskJob struct {
	TaskId string
}

var (
	jobEntriesMux sync.Mutex
	jobEntries    = make(map[string]cron.EntryID)
	cronApp       = cron.New()
)

// SaveCronTask 保存定时任务
func SaveCronTask(taskName, kwArgs, cronRule, comment string) (taskId string) {
	tc := db.TaskCron{
		TaskId:   uuid.New().String(),
		TaskName: taskName,
		KwArgs:   kwArgs,
		CronRule: cronRule,
		Status:   "enable",
		Comment:  comment,
	}
	if !tc.Add() {
		logging.RuntimeLog.Errorf("save cron task to databse fail:%s", tc.TaskId)
		return ""
	}
	AddCronTask(tc.TaskId, tc.CronRule)
	return tc.TaskId
}

// ChangeTaskCronStatus 禁用或启用定时任务
func ChangeTaskCronStatus(taskId, status string) bool {
	ct := db.TaskCron{TaskId: taskId}
	if !ct.GetByTaskId() {
		logging.CLILog.Errorf("cron task:%s not exist...", taskId)
		logging.RuntimeLog.Errorf("cron task:%s not exist...", taskId)
		return false
	}
	updateMap := make(map[string]interface{})
	updateMap["status"] = status
	if !ct.Update(updateMap) {
		return false
	}
	if status == "disable" {
		DeleteCronTask(taskId)
	} else if status == "enable" {
		AddCronTask(ct.TaskId, ct.CronRule)
	}
	return true
}

// RunOnceTaskCron 立即执行一次任务
func RunOnceTaskCron(taskId string) bool {
	job := CronTaskJob{TaskId: taskId}
	job.Run()
	return true
}

// DeleteCronTask 移除一个定时任务
func DeleteCronTask(taskId string) {
	jobEntriesMux.Lock()
	defer jobEntriesMux.Unlock()

	if _, ok := jobEntries[taskId]; ok {
		cronApp.Remove(jobEntries[taskId])
		delete(jobEntries, taskId)
	}
}

// AddCronTask 增加一个定时任务
func AddCronTask(taskId string, cronRule string) {
	jobEntriesMux.Lock()
	defer jobEntriesMux.Unlock()
	//避免重复填加任务
	if _, ok := jobEntries[taskId]; ok {
		return
	}
	eid, err := cronApp.AddJob(cronRule, CronTaskJob{TaskId: taskId})
	if err != nil {
		logging.RuntimeLog.Errorf("add cron job err:%s", err.Error())
		return
	}
	jobEntries[taskId] = eid
}

// GetCronTaskNextRunDatetime  获取任务的下一次执行时间
func GetCronTaskNextRunDatetime(taskId string) (jobExist bool, nextRunDatetime time.Time) {
	jobEntriesMux.Lock()
	defer jobEntriesMux.Unlock()

	if _, ok := jobEntries[taskId]; !ok {
		return false, time.Now()
	}
	return true, cronApp.Entry(jobEntries[taskId]).Next
}

// Run 当定时任务启动时，创建任务执行并发送到消息队列中
func (j CronTaskJob) Run() {
	ct := db.TaskCron{TaskId: j.TaskId}
	if !ct.GetByTaskId() {
		logging.CLILog.Errorf("cron task:%s not exist...", j.TaskId)
		logging.RuntimeLog.Errorf("cron task:%s not exist...", j.TaskId)
		return
	}
	taskId, err := SaveMainTask(ct.TaskName, ct.KwArgs, ct.TaskId)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return
	}
	logging.CLILog.Infof("start cron task:%s,main taskId:%s", j.TaskId, taskId)
	// 更新数据库
	updateMap := make(map[string]interface{})
	updateMap["lastrun_datetime"] = time.Now()
	updateMap["run_count"] = ct.RunCount + 1
	ct.Update(updateMap)
}

// StartCronTask 启动定时任务守护和调度
func StartCronTask() (cronTaskNum int) {
	ct := db.TaskCron{}
	searchMap := make(map[string]interface{})
	searchMap["status"] = "enable"
	allTasks, _ := ct.Gets(searchMap, -1, -1)
	jobEntriesMux.Lock()
	defer jobEntriesMux.Unlock()

	for _, t := range allTasks {
		eId, err := cronApp.AddJob(t.CronRule, CronTaskJob{TaskId: t.TaskId})
		if err != nil {
			logging.CLILog.Error(err)
			logging.RuntimeLog.Error(err)
			continue
		}
		jobEntries[t.TaskId] = eId
		cronTaskNum++
	}
	cronApp.Start()
	return
}

func getCronTaskList() {
	for _, e := range cronApp.Entries() {
		logging.CLILog.Info(e)
	}
	for _, eid := range jobEntries {
		logging.CLILog.Info(eid)
	}
}
