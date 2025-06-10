package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/RichardKnop/machinery/v2/tasks"
	"github.com/google/uuid"
	"github.com/hanc00l/nemo_go/v3/pkg/db"
	"github.com/hanc00l/nemo_go/v3/pkg/logging"
	"github.com/hanc00l/nemo_go/v3/pkg/task/execute"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/bson"
	"math/rand"
	"strings"
	"time"
)

const (
	CREATED string = "CREATED" //任务创建，但还没有开始执行
	STARTED string = "STARTED" //任务在执行中
	SUCCESS string = "SUCCESS" //任务执行完成，结果为SUCCESS
	FAILURE string = "FAILURE" //任务执行完成，结果为FAILURE

	TopicActive     = "active"
	TopicFinger     = "finger"
	TopicPassive    = "passive"
	TopicPocscan    = "pocscan"
	TopicCustom     = "custom"
	TopicStandalone = "standalone"

	TopicMQPrefix = "nemo_mq"
)

var (
	globalMainTaskLock       string = "main_task_lock"
	globalMainTaskUpdateTime string = "main_task_update_time"
	globalStandaloneTaskLock string = "standalone_task_lock"
)

// StartMainTaskDamon MainTask任务的后台监控
func StartMainTaskDamon() {
	const (
		BaseInterval = 10 * time.Second // 基础固定间隔
		MaxJitter    = 10 * time.Second // 最大随机增加量
	)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	redisClient, err := GetRedisClient()
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		return
	}
	defer func(client *redis.Client) {
		_ = CloseRedisClient(client)
	}(redisClient)

	for {
		// 随机睡眠
		jitter := time.Duration(r.Int63n(int64(MaxJitter) + 1))
		sleepTime := BaseInterval + jitter
		time.Sleep(sleepTime)
		// 尝试获取锁
		lock := NewRedisLock(globalMainTaskLock, BaseInterval, redisClient)
		acquired, err := lock.TryLock()
		if err != nil {
			logging.RuntimeLog.Error("获取分布式锁失败:", err.Error())
			continue
		}
		if !acquired {
			// 未获取到锁
			logging.RuntimeLog.Warn("maintask未能获得分布式锁, sleep...")
			continue
		}
		mainTaskUpdateTime, err := getTimeFromRedis(redisClient, globalMainTaskUpdateTime)
		// 如果有多个service实例，为了避免重复处理，设置一个时间间隔，防止多个实例同时处理任务
		if errors.Is(err, redis.Nil) || time.Now().Sub(mainTaskUpdateTime).Seconds() >= BaseInterval.Seconds() {
			// 处理已开始的任务
			processStartedTask()
			// 处理新建的任务
			processCreatedTask()
			// 检查worker状态
			checkWorkerStatus()
			// 存储更新时间
			err = storeTimeToRedis(redisClient, globalMainTaskUpdateTime, time.Now())
			if err != nil {
				logging.RuntimeLog.Error("更新maintask的更新时间失败:", err.Error())
			}
			// 释放锁
			if unlockErr := lock.Unlock(); unlockErr != nil {
				logging.RuntimeLog.Error("释放maintask锁失败:", unlockErr.Error())
			}
		} else {
			if err != nil {
				logging.RuntimeLog.Error("获取maintask的更新时间失败:", err.Error())
			}
		}
	}
}

func processStartedTask() {
	client, err := db.GetClient()
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		return
	}
	defer db.CloseClient(client)

	mainTaskDocs, err := db.NewMainTask(client).Find(bson.M{db.Status: STARTED, db.Cron: false}, 0, 0)
	if err != nil {
		logging.RuntimeLog.Error("获取maintask失败:", err.Error())
		return
	}
	for _, task := range mainTaskDocs {
		var status, progress, result string
		var progressRate float64
		// 检查子任务的执行情况：
		createdTask, startedTask, totalTask := checkExecutorTask(task.TaskId)
		progress = fmt.Sprintf("%d/%d/%d", startedTask, createdTask, totalTask)
		if totalTask > 0 && createdTask == 0 && startedTask == 0 {
			// 任务执行完成
			status = SUCCESS
			progressRate = 1.0
			//全部任务完成，将任务的结果同步到全局资产库
			result = SyncTaskAsset(task.WorkspaceId, task.TaskId)
			// 任务执行完成，发送消息
			workspace := db.NewWorkspace(client)
			wDoc, _ := workspace.GetByWorkspaceId(task.WorkspaceId)
			if len(wDoc.NotifyId) > 0 {
				_ = Notify(wDoc.NotifyId, NotifyData{
					TaskName: task.TaskName,
					Target:   task.Target,
					Runtime:  fmt.Sprintf("%s", time.Now().Sub(*task.StartTime)),
					Result:   result,
				})
			}
		} else {
			// 任务执行中或未执行
			status = task.Status
			progressRate = computeMainTaskProgressRate(task.TaskId, task.Args)
		}
		// 更新任务状态
		if status != task.Status || progress != task.Progress || progressRate != task.ProgressRate {
			if updateMainTask(task.Id.Hex(), status, progress, progressRate, result) == false {
				logging.RuntimeLog.Errorf("更新maintask：%s 任务状态失败:", task.TaskId)
				continue
			}
		}
	}

	return
}

func processCreatedTask() {
	client, err := db.GetClient()
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		return
	}
	defer db.CloseClient(client)
	mainTask := db.NewMainTask(client)
	mainTaskDocs, err := mainTask.Find(bson.M{db.Status: CREATED, db.Cron: false}, 0, 0)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		return
	}
	for _, doc := range mainTaskDocs {
		var exeConfig execute.ExecutorConfig
		err = json.Unmarshal([]byte(doc.Args), &exeConfig)
		if err != nil {
			logging.RuntimeLog.Error(err.Error())
			return
		}
		mainTaskInfo := execute.MainTaskInfo{
			Target:          doc.Target,
			ExcludeTarget:   doc.ExcludeTarget,
			ExecutorConfig:  exeConfig,
			OrgId:           doc.OrgId,
			WorkspaceId:     doc.WorkspaceId,
			MainTaskId:      doc.TaskId,
			IsProxy:         doc.IsProxy,
			TargetSliceType: doc.TargetSliceType,
			TargetSliceNum:  doc.TargetSliceNum,
		}

		if err = processExecutorTask(mainTaskInfo); err != nil {
			return
		}
		// 任务启动，更新状态
		var isSuccess bool
		isSuccess, err = mainTask.Update(doc.Id.Hex(), bson.M{db.Status: STARTED, db.StartTime: time.Now()})
		if err != nil {
			logging.RuntimeLog.Error("更新maintask状态失败:", err.Error())
			continue
		}
		if !isSuccess {
			logging.RuntimeLog.Errorf("更新maintask失败， docId:%s, err:%v", doc.Id, err)
			continue
		}
	}

	return
}

func processExecutorTask(mainTaskInfo execute.MainTaskInfo) (err error) {
	f := func(executor string, mainTaskInfo execute.MainTaskInfo) (err error) {
		executorTaskInfo := execute.ExecutorTaskInfo{
			MainTaskInfo: mainTaskInfo,
			Executor:     executor,
			TaskId:       uuid.New().String(),
		}
		err = newExecutorTask(executorTaskInfo)
		if err != nil {
			logging.RuntimeLog.Errorf("创建executor任务失败,mainTaskInfo:%v, err:%v", executorTaskInfo, err)
			return
		}
		return nil
	}
	var succeedTask int
	// ip、域名、OnlineAPI、Standalone任务，是top任务，可以直接开始、并行开始的
	if len(mainTaskInfo.ExecutorConfig.Standalone) > 0 {
		ipSlice := NewTaskSlice()
		ipSlice.IpSliceNumber = mainTaskInfo.TargetSliceNum
		ipSlice.TaskMode = mainTaskInfo.TargetSliceType
		ipSlice.IpTarget = strings.Split(mainTaskInfo.Target, ",")
		targets, _ := ipSlice.DoIpSlice()
		for _, target := range targets {
			mti := mainTaskInfo
			mti.Target = target
			for executor, _ := range mainTaskInfo.ExecutorConfig.Standalone {
				if err = f(executor, mti); err != nil {
					return err
				}
				succeedTask++
			}
		}
	}
	// LLMAPI任务，是top任务，可以直接开始、并行开始的
	// 注意：如果有LLMAPI任务，则域名和onlineapi只会在LLMAPI任务中执行
	if len(mainTaskInfo.ExecutorConfig.LLMAPI) > 0 {
		for executor, _ := range mainTaskInfo.ExecutorConfig.LLMAPI {
			if err = f(executor, mainTaskInfo); err != nil {
				return err
			}
			succeedTask++
		}
	} else {
		if len(mainTaskInfo.ExecutorConfig.DomainScan) > 0 {
			// 域名扫描任务，按行拆分目标，并发执行
			var targets []string
			// 按行拆分目标，并发执行
			if mainTaskInfo.TargetSliceType == SliceByLine {
				targets = strings.Split(mainTaskInfo.Target, ",")
			} else {
				targets = []string{mainTaskInfo.Target}
			}
			for _, target := range targets {
				mti := mainTaskInfo
				mti.Target = target
				for executor, _ := range mainTaskInfo.ExecutorConfig.DomainScan {
					if err = f(executor, mti); err != nil {
						return err
					}
					succeedTask++
				}
			}
		}
		if len(mainTaskInfo.ExecutorConfig.OnlineAPI) > 0 {
			for executor, _ := range mainTaskInfo.ExecutorConfig.OnlineAPI {
				var targets []string
				// 按行拆分目标，并发执行
				if mainTaskInfo.TargetSliceType == SliceByLine {
					targets = strings.Split(mainTaskInfo.Target, ",")
				} else {
					targets = []string{mainTaskInfo.Target}
				}
				for _, target := range targets {
					mti := mainTaskInfo
					mti.Target = target
					if err = f(executor, mti); err != nil {
						return err
					}
					// 特殊情况，icp、whois任务，不计入成功任务数
					if executor == "icp" || executor == "whois" {
						continue
					}
					succeedTask++
				}
			}
		}
	}
	if len(mainTaskInfo.ExecutorConfig.PortScan) > 0 {
		ipSlice := NewTaskSlice()
		ipSlice.IpSliceNumber = mainTaskInfo.TargetSliceNum
		ipSlice.TaskMode = mainTaskInfo.TargetSliceType
		ipSlice.IpTarget = strings.Split(mainTaskInfo.Target, ",")
		targets, _ := ipSlice.DoIpSlice()
		for _, target := range targets {
			mti := mainTaskInfo
			mti.Target = target
			for executor, _ := range mainTaskInfo.ExecutorConfig.PortScan {
				if err = f(executor, mti); err != nil {
					return err
				}
				succeedTask++
			}
		}
	}
	// fingerprint、pocscan任务，是需要等前面的任务执行完成后再开始的；或者前面没有任务，直接开始执行
	if succeedTask == 0 {
		if len(mainTaskInfo.ExecutorConfig.FingerPrint) > 0 {
			// fingerprint不区分executor，由执行行时根据任务配置决定
			if err = f(execute.FingerPrint, mainTaskInfo); err != nil {
				return err
			}
		}
		if len(mainTaskInfo.ExecutorConfig.PocScan) > 0 {
			for executor, _ := range mainTaskInfo.ExecutorConfig.PocScan {
				if err = f(executor, mainTaskInfo); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func computeTaskCategoryRate(exeConfig execute.ExecutorConfig) (categoryRate map[string]float64) {
	categoryRate = make(map[string]float64)
	taskCategoryTotal := 0
	if len(exeConfig.LLMAPI) > 0 {
		taskCategoryTotal++
	}
	if len(exeConfig.PortScan) > 0 {
		taskCategoryTotal++
	}
	if len(exeConfig.DomainScan) > 0 {
		taskCategoryTotal++
	}
	if len(exeConfig.OnlineAPI) > 0 {
		taskCategoryTotal++
	}
	if len(exeConfig.FingerPrint) > 0 {
		taskCategoryTotal++
	}
	if len(exeConfig.PocScan) > 0 {
		taskCategoryTotal++
	}
	taskCategoryRate := 1.0 / float64(taskCategoryTotal)
	if len(exeConfig.FingerPrint) > 0 {
		categoryRate[execute.FingerPrint] = taskCategoryRate
	} else {
		categoryRate[execute.FingerPrint] = 0.0
	}
	if len(exeConfig.PocScan) > 0 {
		categoryRate[execute.PocScan] = taskCategoryRate
	} else {
		categoryRate[execute.PocScan] = 0.0
	}
	if len(exeConfig.PortScan) > 0 {
		categoryRate[execute.PortScan] = taskCategoryRate
	} else {
		categoryRate[execute.PortScan] = 0.0
	}
	if len(exeConfig.DomainScan) > 0 {
		categoryRate[execute.DomainScan] = taskCategoryRate
	} else {
		categoryRate[execute.DomainScan] = 0.0
	}
	if len(exeConfig.OnlineAPI) > 0 {
		categoryRate[execute.OnlineAPI] = taskCategoryRate
	} else {
		categoryRate[execute.OnlineAPI] = 0.0
	}
	if len(exeConfig.LLMAPI) > 0 {
		categoryRate[execute.LLMAPI] = taskCategoryRate
	} else {
		categoryRate[execute.LLMAPI] = 0
	}
	if len(exeConfig.Standalone) > 0 {
		categoryRate[execute.Standalone] = 1
	} else {
		categoryRate[execute.Standalone] = 0
	}

	return
}

func computeMainTaskProgressRate(mainTaskId string, args string) (result float64) {
	// 聚合每个executor的任务状态
	client, err := db.GetClient()
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		return
	}
	defer db.CloseClient(client)
	executeTask := db.NewExecutorTask(client)
	executorTaskResult, err := executeTask.AggregateMainTaskProgress(mainTaskId)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		return
	}
	if executorTaskResult == nil {
		return
	}
	// 根据executor名称，统计不同分类的任务的不同状态的数量
	total := make(map[string]map[string]int)
	total[execute.PortScan] = make(map[string]int)
	total[execute.DomainScan] = make(map[string]int)
	total[execute.OnlineAPI] = make(map[string]int)
	total[execute.FingerPrint] = make(map[string]int)
	total[execute.PocScan] = make(map[string]int)
	total[execute.Standalone] = make(map[string]int)
	total[execute.LLMAPI] = make(map[string]int)
	for _, r := range executorTaskResult {
		for _, s := range r.Statuses {
			switch r.Executor {
			case "masscan", "nmap", "gogo":
				total[execute.PortScan][s.Status] += s.Count
			case "subfinder", "massdns":
				total[execute.DomainScan][s.Status] += s.Count
			case "fofa", "hunter", "quake", "icp", "whois":
				total[execute.OnlineAPI][s.Status] += s.Count
			case "fingerprint":
				total[execute.FingerPrint][s.Status] += s.Count
			case "nuclei", "zombie":
				total[execute.PocScan][s.Status] += s.Count
			case "qwen", "kimi", "deepseek", "icpPlus":
				total[execute.LLMAPI][s.Status] += s.Count
			case "standalone":
				total[execute.Standalone][s.Status] += s.Count
			}
		}
	}
	// 根据分类比例，计算每个分类的任务比例
	var exeConfig execute.ExecutorConfig
	err = json.Unmarshal([]byte(args), &exeConfig)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		return
	}
	taskCategoryRateScore := computeTaskCategoryRate(exeConfig)
	// 计算总进度
	for category, status := range total {
		var created, started, finished int
		if v, ok := status[CREATED]; ok {
			created = v
		}
		if v, ok := status[STARTED]; ok {
			started = v
		}
		if v, ok := status[SUCCESS]; ok {
			finished += v
		}
		if v, ok := status[FAILURE]; ok {
			finished += v
		}
		if created+started+finished == 0 {
			continue
		}
		result += (float64)(finished) / (float64)(created+started+finished) * taskCategoryRateScore[category]
	}

	return
}

func newExecutorTask(executorTaskInfo execute.ExecutorTaskInfo) (err error) {
	// 检查目标中是否有黑名单
	blackTarget, normalTarget := checkTargetForBlacklist(executorTaskInfo.Target, executorTaskInfo.WorkspaceId)
	// 黑名单处理
	if len(blackTarget) > 0 {
		logging.RuntimeLog.Warnf("匹配到黑名单记录: %s, skip", strings.Join(blackTarget, ","))
	}
	// 正常目标处理
	executorTaskInfo.Target = strings.Join(normalTarget, ",")
	//　standalone任务不送入消息队列，只写入数据库；其他任务送入消息队列
	if executorTaskInfo.Executor != "standalone" {
		if err = sendExecutorTaskToMq(executorTaskInfo); err != nil {
			return
		}
	}
	if err = addExecutorTaskToDb(executorTaskInfo); err != nil {
		return
	}
	return nil
}

func checkTargetForBlacklist(target string, workspaceId string) (blacklist, normal []string) {
	blc := NewBlacklist()
	blc.LoadBlacklist(workspaceId)
	// 检查目标中是否有黑名单
	for _, t := range strings.Split(target, ",") {
		targetStripped := strings.TrimSpace(t)
		if targetStripped == "" {
			continue
		}
		hostPort := strings.Split(targetStripped, ":")
		if len(hostPort) == 0 {
			continue
		}
		host := hostPort[0]
		if blc.IsHostBlocked(host) {
			blacklist = append(blacklist, targetStripped)
		} else {
			normal = append(normal, targetStripped)
		}
	}
	return
}

func sendExecutorTaskToMq(executorTaskInfo execute.ExecutorTaskInfo) (err error) {
	topicName := GetTopicByTaskName(executorTaskInfo.Executor, executorTaskInfo.WorkspaceId)
	if topicName == "" {
		msg := fmt.Sprintf("任务没有配置topic:%s", executorTaskInfo.Executor)
		logging.RuntimeLog.Error(msg)
		return errors.New(msg)
	}
	configJSON, err := json.Marshal(executorTaskInfo)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		return err
	}
	server := GetServerTaskMQServer(topicName)
	// 延迟5秒后执行：如果不延迟，有可能任务在完成数据库之前执行，从而导致task not exist错误
	eta := time.Now().Add(time.Second * 5)
	workerTask := tasks.Signature{
		Name: executorTaskInfo.Executor,
		UUID: executorTaskInfo.TaskId,
		ETA:  &eta,
		Args: []tasks.Arg{
			{Name: "configJSON", Type: "string", Value: string(configJSON)},
		},
		//RoutingKey：分发到不同功能的worker队列
		RoutingKey: GetRoutingKeyByTopic(topicName),
	}
	_, err = server.SendTask(&workerTask)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return err
	}

	return nil
}

func addExecutorTaskToDb(executorTaskInfo execute.ExecutorTaskInfo) error {
	doc := db.ExecuteTaskDocument{
		WorkspaceId:   executorTaskInfo.WorkspaceId,
		TaskId:        executorTaskInfo.TaskId,
		MainTaskId:    executorTaskInfo.MainTaskId,
		PreTaskId:     executorTaskInfo.PreTaskId,
		Executor:      executorTaskInfo.Executor,
		Target:        executorTaskInfo.Target,
		ExcludeTarget: executorTaskInfo.ExcludeTarget,
		Status:        CREATED,
	}
	if argsData, err := json.Marshal(executorTaskInfo.ExecutorConfig); err != nil {
		logging.RuntimeLog.Error(err.Error())
		return err
	} else {
		doc.Args = string(argsData)
	}
	mongoClient, err := db.GetClient()
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		return err
	}
	defer db.CloseClient(mongoClient)

	isSuccess, err := db.NewExecutorTask(mongoClient).Insert(doc)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		return err
	}
	if !isSuccess {
		logging.RuntimeLog.Errorf("生成子任务失败, doc:%v, err:%v", doc, err)
		return errors.New("生成子任务失败")
	}
	return nil
}

func checkExecutorTask(mainTaskId string) (createdTask, startedTask, totalTask int) {
	mongoClient, err := db.GetClient()
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		return
	}
	defer db.CloseClient(mongoClient)

	runTasks, err := db.NewExecutorTask(mongoClient).Find(bson.M{db.MainTaskId: mainTaskId}, 0, 0)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		return
	}
	for _, t := range runTasks {
		if t.Status == CREATED {
			createdTask++
		} else if t.Status == STARTED {
			startedTask++
		}
	}
	totalTask = len(runTasks)
	return
}

func updateMainTask(id string, state string, progress string, progressRate float64, result string) bool {
	update := bson.M{}
	if state != "" {
		update[db.Status] = state
		if state == SUCCESS || state == FAILURE {
			update[db.EndTime] = time.Now()
		}
	}
	if progress != "" {
		update[db.Progress] = progress
	}
	if result != "" {
		update[db.Result] = result
	}
	if progressRate != 0 {
		update[db.ProgressRate] = progressRate
	}
	client, err := db.GetClient()
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		return false
	}
	defer db.CloseClient(client)

	isSuccess, err := db.NewMainTask(client).Update(id, update)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		return false
	}
	return isSuccess
}
