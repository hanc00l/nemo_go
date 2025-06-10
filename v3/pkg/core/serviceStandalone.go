package core

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hanc00l/nemo_go/v3/pkg/db"
	"github.com/hanc00l/nemo_go/v3/pkg/logging"
	"github.com/hanc00l/nemo_go/v3/pkg/task/execute"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/bson"
	"time"
)

type StandaloneRequestArgs struct {
	TaskId string
	Worker string
	Status string
	Result string
}

func (s *Service) RequestStandaloneTask(ctx context.Context, args *string, replay *execute.ExecutorTaskInfo) error {
	if len(*args) == 0 {
		logging.RuntimeLog.Error("任务请求参数缺少worker名称")
		return fmt.Errorf("缺少worker名称")
	}
	// redis分布式锁, 防止多个worker同时执行同一个任务
	redisClient, err := GetRedisClient()
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		return err
	}
	defer func(client *redis.Client) {
		_ = CloseRedisClient(client)
	}(redisClient)
	// 尝试获取锁
	lock := NewRedisLock(globalStandaloneTaskLock, 10*time.Second, redisClient)
	acquired, err := lock.TryLock()
	if err != nil {
		logging.RuntimeLog.Error("获取分布式锁失败:", err.Error())
		return fmt.Errorf("获取分布式锁失败")
	}
	if !acquired {
		// 未获取到锁
		logging.RuntimeLog.Warn("未能获得分布式锁")
		return fmt.Errorf("获取分布式锁失败")
	}
	defer func() {
		// 释放锁
		if unlockErr := lock.Unlock(); unlockErr != nil {
			logging.RuntimeLog.Error("释放锁失败:", unlockErr.Error())
		}
	}()
	dbClient, err := db.GetClient()
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		return err
	}
	defer db.CloseClient(dbClient)
	// 从executorTask中获取一个待执行的standalone任务
	executorTasks, err := db.NewExecutorTask(dbClient).Find(bson.M{"executor": "standalone", "status": CREATED}, 1, 1)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		return err
	}
	// 没有待执行的standalone任务
	if len(executorTasks) == 0 {
		return nil
	}
	// 获取执行任务的主任务
	mainTaskInfo, err := db.NewMainTask(dbClient).GetByTaskId(executorTasks[0].MainTaskId)
	if err != nil {
		logging.RuntimeLog.Errorf("获取执行任务的主任务失败:%s", err.Error())
		return fmt.Errorf("获取执行任务的主任务失败")
	}
	taskInfo := execute.ExecutorTaskInfo{
		Executor:  executorTasks[0].Executor,
		TaskId:    executorTasks[0].TaskId,
		PreTaskId: executorTasks[0].PreTaskId,
		MainTaskInfo: execute.MainTaskInfo{
			Target:        executorTasks[0].Target,
			ExcludeTarget: executorTasks[0].ExcludeTarget,
			WorkspaceId:   executorTasks[0].WorkspaceId,
			MainTaskId:    executorTasks[0].MainTaskId,
			IsProxy:       mainTaskInfo.IsProxy,
			OrgId:         mainTaskInfo.OrgId,
		}}
	err = json.Unmarshal([]byte(mainTaskInfo.Args), &taskInfo.MainTaskInfo.ExecutorConfig)
	if err != nil {
		logging.RuntimeLog.Errorf("解析执行任务的配置失败:%s", err.Error())
		return fmt.Errorf("解析执行任务的配置失败")
	}
	*replay = taskInfo
	// 更新任务状态和参数
	taskStatus := TaskStatusArgs{
		TaskID: taskInfo.TaskId,
		State:  STARTED,
		Worker: *args,
	}
	var updateStatus bool
	err = s.UpdateTask(ctx, &taskStatus, &updateStatus)
	if err != nil {
		logging.RuntimeLog.Errorf("更新执行任务状态失败:%s", err.Error())
		return fmt.Errorf("更新执行任务状态失败")
	}
	if !updateStatus {
		logging.RuntimeLog.Error("更新执行任务状态失败")
		return fmt.Errorf("更新执行任务状态失败")
	}

	return nil
}

func (s *Service) FinishStandaloneTask(ctx context.Context, args *StandaloneRequestArgs, replay *string) error {
	if args == nil {
		return fmt.Errorf("请求参数为空")
	}
	if len(args.TaskId) == 0 {
		return fmt.Errorf("任务ID为空")
	}
	if len(args.Status) == 0 {
		return fmt.Errorf("任务状态为空")
	}
	if len(args.Worker) == 0 {
		return fmt.Errorf("worker名称为空")
	}
	client, err := db.GetClient()
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		return err
	}
	defer db.CloseClient(client)
	// 更新任务状态和参数
	taskStatus := TaskStatusArgs{
		TaskID: args.TaskId,
		State:  args.Status,
		Worker: args.Worker,
		Result: args.Result,
	}
	var updateStatus bool
	err = s.UpdateTask(ctx, &taskStatus, &updateStatus)
	if err != nil {
		logging.RuntimeLog.Errorf("更新执行任务状态失败:%s", err.Error())
		return fmt.Errorf("更新执行任务状态失败")
	}
	if !updateStatus {
		logging.RuntimeLog.Error("更新执行任务状态失败")
		return fmt.Errorf("更新执行任务状态失败")
	}
	msg := "ok"
	replay = &msg
	return nil
}
