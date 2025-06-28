package core

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/hanc00l/nemo_go/v3/pkg/db"
	"github.com/hanc00l/nemo_go/v3/pkg/logging"
	"github.com/hanc00l/nemo_go/v3/pkg/task/execute"
	"github.com/redis/go-redis/v9"
	"github.com/robfig/cron/v3"
	"go.mongodb.org/mongo-driver/v2/bson"
	"time"
)

type CronTaskJob struct {
	TaskId string
}

var (
	globalCronTaskLock       = "cron_task_lock"
	globalCronTackUpdateFlag = "cron_task_update_flag"
)

var (
	cronApp = cron.New()
)

// RunOnceTaskCron 立即执行一次任务
func RunOnceTaskCron(taskId string) bool {
	job := CronTaskJob{TaskId: taskId}
	job.Run()
	return true
}

func SetCronTaskUpdateFlag(flag string) (err error) {
	redisClient, err := GetRedisClient()
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		return err
	}
	defer func(client *redis.Client) {
		_ = CloseRedisClient(client)
	}(redisClient)
	// 尝试获取锁
	lock := NewRedisLock(globalCronTaskLock, 5*time.Second, redisClient)
	acquired, err := lock.TryLock()
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		return err
	}
	if !acquired {
		// 未获取到锁
		logging.RuntimeLog.Error("cron task lock not acquired,.")
		return err
	}
	defer lock.Unlock()

	// 更新任务状态
	err = redisClient.Set(context.Background(), globalCronTackUpdateFlag, flag, 0).Err()
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		return err
	}

	return nil
}

func ReloadCronTask() {
	redisClient, err := GetRedisClient()
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		return
	}
	defer func(client *redis.Client) {
		_ = CloseRedisClient(client)
	}(redisClient)

	// 尝试获取锁
	lock := NewRedisLock(globalCronTaskLock, 5*time.Second, redisClient)
	acquired, err := lock.TryLock()
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		return
	}
	if !acquired {
		// 未获取到锁
		logging.RuntimeLog.Error("cron task lock not acquired,.")
		return
	}
	defer lock.Unlock()
	// 清空旧的任务
	cronApp.Stop()
	for _, entry := range cronApp.Entries() {
		cronApp.Remove(entry.ID)
	}
	// 重新加载任务
	err = LoadCronTask()
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		return
	}
	// 更新任务状态
	err = redisClient.Set(context.Background(), globalCronTackUpdateFlag, "false", 0).Err()
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		return
	}
	cronApp.Start()

	logging.RuntimeLog.Info("start cron task daemon...")
	logging.CLILog.Info("start cron task daemon...")
}

// Run 当定时任务启动时，创建任务执行并发送到消息队列中
func (j CronTaskJob) Run() {
	mongoClient, err := db.GetClient()
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		return
	}
	defer db.CloseClient(mongoClient)

	mainTask := db.NewMainTask(mongoClient)
	docCron, err := mainTask.GetByTaskId(j.TaskId)
	if err != nil {
		logging.RuntimeLog.Error(fmt.Sprintf("获取计划任务失败:%s err:%s", j.TaskId, err.Error()))
		return
	}
	mainTaskDoc := db.MainTaskDocument{
		WorkspaceId:     docCron.WorkspaceId,
		TaskId:          uuid.New().String(),
		TaskName:        docCron.TaskName, //+ "-" + time.Now().Format("060102"),
		Description:     fmt.Sprintf("（来自定时任务:%s）", docCron.TaskName),
		Target:          docCron.Target,
		ExcludeTarget:   docCron.ExcludeTarget,
		ProfileName:     docCron.ProfileName,
		OrgId:           docCron.OrgId,
		IsProxy:         docCron.IsProxy,
		Args:            docCron.Args,
		TargetSliceType: docCron.TargetSliceType,
		TargetSliceNum:  docCron.TargetSliceNum,
		IsCron:          false,
		Status:          CREATED,
	}
	isSuccess, err := mainTask.Insert(mainTaskDoc)
	if err != nil {
		logging.RuntimeLog.Error(fmt.Sprintf("生成计划任务的主任务失败:%s err:%s", j.TaskId, err.Error()))
		return
	}
	if !isSuccess {
		logging.RuntimeLog.Error(fmt.Sprintf("生成计划任务的主任务不成功:%s", j.TaskId))
		return
	}
	logging.RuntimeLog.Infof("开始计划任务:%s,主任务Id:%s", j.TaskId, mainTaskDoc.TaskId)
	// 更新任务状态
	updateDoc := bson.M{"cronTaskInfo.lastRun": time.Now(), "cronTaskInfo.runCount": docCron.CronTaskInfo.CronRunCount + 1}
	// 检查如果任务参数的OnlineAPIConfig中设置了SearchStartTime，则更新到新的任务中
	var config execute.ExecutorConfig
	err = json.Unmarshal([]byte(docCron.Args), &config)
	if err != nil {
		logging.RuntimeLog.Error(fmt.Sprintf("反序列化计划任务参数失败:%s", err.Error()))
	} else {
		if config.OnlineAPI != nil {
			// 遍历OnlineAPIConfig，如果有设置SearchStartTime，则更新到新的任务中
			needUpdate := false
			for executorName, onlineApiConfig := range config.OnlineAPI {
				if onlineApiConfig.SearchStartTime != "" {
					needUpdate = true
					onlineApiConfig.SearchStartTime = time.Now().Format("2006-01-02")
					config.OnlineAPI[executorName] = onlineApiConfig
				}
			}
			if needUpdate {
				newArgs, err := json.Marshal(config)
				if err != nil {
					logging.RuntimeLog.Error(fmt.Sprintf("反序列化计划任务参数失败:%s", err.Error()))
				} else {
					updateDoc["args"] = newArgs
				}
			}
		}
	}
	isSuccess, err = mainTask.Update(docCron.Id.Hex(), updateDoc)
	if err != nil {
		logging.RuntimeLog.Error(fmt.Sprintf("更新计划任务到主任务失败:%s err:%s", j.TaskId, err.Error()))
		return
	}
	if !isSuccess {
		logging.RuntimeLog.Error(fmt.Sprintf("更新计划任务到主任务不成功:%s", j.TaskId))
		return
	}
}

func LoadCronTask() (err error) {
	// 从数据库中获取所有定时任务
	mongoClient, err := db.GetClient()
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		return err
	}
	defer db.CloseClient(mongoClient)

	mainTask := db.NewMainTask(mongoClient)
	mainTaskDocs, err := mainTask.Find(bson.M{db.Cron: true, "cronTaskInfo.enabled": true}, 0, 0)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		return err
	}
	for _, doc := range mainTaskDocs {
		_, err = cronApp.AddJob(doc.CronTaskInfo.CronExpr, CronTaskJob{TaskId: doc.TaskId})
		if err != nil {
			logging.RuntimeLog.Error(err)
			return err
		}
	}
	return nil
}

// StartCronTaskDamon 启动定时任务守护和调度
func StartCronTaskDamon() {
	ReloadCronTask()
	// 通过redis全局共享变量，定时检查是否需要重新加载任务
	go func() {
		redisClient, err := GetRedisClient()
		defer func(client *redis.Client) {
			_ = CloseRedisClient(client)
		}(redisClient)

		for {
			time.Sleep(time.Minute)

			if err != nil {
				logging.RuntimeLog.Error(err.Error())
				continue
			}

			val, err := redisClient.Get(context.Background(), globalCronTackUpdateFlag).Result()
			if err != nil {
				logging.RuntimeLog.Error(err.Error())
			} else if val == "true" {
				logging.RuntimeLog.Info("reload cron task...")
				ReloadCronTask()
			}
		}
	}()
}
