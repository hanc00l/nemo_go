package runner

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/hanc00l/nemo_go/pkg/comm"
	"github.com/hanc00l/nemo_go/pkg/db"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/notify"
	"github.com/hanc00l/nemo_go/pkg/task/ampq"
	"strings"
	"time"
)

// StartMainTaskDamon MainTask任务的后台监控
func StartMainTaskDamon() {
	comm.MainTaskResult = make(map[string]comm.MainTaskResultMap)
	var err error
	for {
		// 处理已开始的任务
		err = processStartedTask()
		if err != nil {
			logging.CLILog.Error(err)
			logging.RuntimeLog.Error(err)
		}
		// 处理新建的任务
		err = processCreatedTask()
		if err != nil {
			logging.CLILog.Error(err)
			logging.RuntimeLog.Error(err)
		}
		// sleep
		time.Sleep(10 * time.Second)
	}
}

// SaveMainTask 保存一个MainTask到数据库中
func SaveMainTask(taskName, configJSON, cronTaskId string, workspaceId int) (taskId string, err error) {
	taskId = uuid.New().String()
	task := &db.TaskMain{
		TaskId:      taskId,
		TaskName:    taskName,
		KwArgs:      configJSON,
		State:       ampq.CREATED,
		CronTaskId:  cronTaskId,
		WorkspaceId: workspaceId,
	}
	//kwargs可能因为target很多导致超过数据库中的字段设计长度，因此作一个长度截取
	const argsLength = 6000
	if len(task.KwArgs) > argsLength {
		//kwargs可能因为target很多导致超过数据库中的字段设计长度，直接返回错误，不保存任务信息
		err = errors.New(fmt.Sprintf("maintask %s:%s arguments too long:%s", taskName, taskId, task.KwArgs))
		logging.CLILog.Error(err)
		logging.RuntimeLog.Error(fmt.Sprintf("maintask %s:%s arguments too long:%d", taskName, taskId, len(task.KwArgs)))
		return
	}
	if !task.Add() {
		err = errors.New(fmt.Sprintf("save new maintask fail: %s,%s,%s", taskName, taskId, task.KwArgs))
		logging.CLILog.Error(err)
		logging.RuntimeLog.Error(err)
	}
	return
}

// runMainTask 运行一个创建的maintask
func runMainTask(taskName, taskId, kwArgs string, workspaceId int) (err error) {
	var taskRunId string
	if taskName == "portscan" {
		var req PortscanRequestParam
		if err = json.Unmarshal([]byte(kwArgs), &req); err != nil {
			logging.RuntimeLog.Error(err)
			return
		}
		if taskRunId, err = StartPortScanTask(req, taskId, workspaceId); err != nil {
			logging.RuntimeLog.Error(err)
			return
		}
	} else if taskName == "batchscan" {
		var req PortscanRequestParam
		if err = json.Unmarshal([]byte(kwArgs), &req); err != nil {
			logging.RuntimeLog.Error(err)
			return
		}
		if taskRunId, err = StartBatchScanTask(req, taskId, workspaceId); err != nil {
			logging.RuntimeLog.Error(err)
			return
		}
	} else if taskName == "domainscan" {
		var req DomainscanRequestParam
		if err = json.Unmarshal([]byte(kwArgs), &req); err != nil {
			logging.RuntimeLog.Error(err)
			return
		}
		if taskRunId, err = StartDomainScanTask(req, taskId, workspaceId); err != nil {
			logging.RuntimeLog.Error(err)
			return
		}
	} else if taskName == "pocscan" {
		var req PocscanRequestParam
		if err = json.Unmarshal([]byte(kwArgs), &req); err != nil {
			logging.RuntimeLog.Error(err)
			return
		}
		if taskRunId, err = StartPocScanTask(req, taskId, workspaceId); err != nil {
			logging.RuntimeLog.Error(err)
			return
		}
	} else if taskName == "xportscan" || taskName == "xdomainscan" || taskName == "xorgscan" || taskName == "xonlineapi" || taskName == "xonlineapi_custom" {
		var req XScanRequestParam
		if err = json.Unmarshal([]byte(kwArgs), &req); err != nil {
			logging.RuntimeLog.Error(err)
			return
		}
		switch taskName {
		case "xportscan":
			taskRunId, err = StartXPortScanTask(req, taskId, workspaceId)
		case "xdomainscan":
			taskRunId, err = StartXDomainScanTask(req, taskId, workspaceId)
		case "xonlineapi":
			taskRunId, err = StartXOnlineAPIKeywordTask(req, taskId, workspaceId)
		case "xonlineapi_custom":
			taskRunId, err = StartXOnlineAPIKeywordCustomTask(req, taskId, workspaceId)
		case "xorgscan":
			taskRunId, err = StartXOrgScanTask(req, taskId, workspaceId)
		}
		if err != nil {
			logging.RuntimeLog.Error(err)
			return
		}
	} else {
		logging.RuntimeLog.Errorf("invalid task name:%s in %s...", taskName, taskId)
		return
	}
	logging.CLILog.Infof("start %s task:%s,running taskRunId:%s", taskName, taskId, taskRunId)
	logging.RuntimeLog.Debugf("start %s task:%s,running taskRunId:%s", taskName, taskId, taskRunId)

	return
}

// processCreatedTask 处理新建的maintask
func processCreatedTask() (err error) {
	task := db.TaskMain{}
	searchMap := make(map[string]interface{})
	searchMap["state"] = ampq.CREATED
	results, _ := task.Gets(searchMap, -1, -1)
	for _, t := range results {
		comm.MainTaskResultMutex.Lock()
		comm.MainTaskResult[t.TaskId] = comm.MainTaskResultMap{
			IPResult:     make(map[string]map[int]interface{}),
			DomainResult: make(map[string]interface{}),
			VulResult:    make(map[string]map[string]interface{}),
		}
		comm.MainTaskResultMutex.Unlock()
		// 启动任务执行
		if err = runMainTask(t.TaskName, t.TaskId, t.KwArgs, t.WorkspaceId); err != nil {
			logging.RuntimeLog.Error(err)
			return err
		}
		// 更新任务状态
		if updateMainTask(&t, ampq.STARTED, "", "") == false {
			msg := fmt.Sprintf("update maintask state fail:%s", t.TaskId)
			logging.RuntimeLog.Error(msg)
			return errors.New(msg)
		}
	}
	return
}

// processStartedTask 处理正在运行的maintask
func processStartedTask() (err error) {
	task := db.TaskMain{}
	searchMap := make(map[string]interface{})
	searchMap["state"] = ampq.STARTED
	results, _ := task.Gets(searchMap, -1, -1)
	var finishedTask []string
	for _, t := range results {
		// 如果数据库中的任务是STARTED状态，但缓存中有结果则重建（服务端重启有可能导致）
		comm.MainTaskResultMutex.Lock()
		if _, ok := comm.MainTaskResult[t.TaskId]; !ok {
			comm.MainTaskResult[t.TaskId] = comm.MainTaskResultMap{
				IPResult:     make(map[string]map[int]interface{}),
				DomainResult: make(map[string]interface{}),
				VulResult:    make(map[string]map[string]interface{}),
			}
		}
		comm.MainTaskResultMutex.Unlock()
		// 检查子任务runtask
		createdTask, startedTask, totalTask := checkRunTask(t.TaskId)
		updatedProgress := fmt.Sprintf("%d/%d/%d", startedTask, createdTask, totalTask)
		// 任务已完成，需要更改任务状态和任务结果
		var updatedState, updatedResult string
		if totalTask > 0 && createdTask == 0 && startedTask == 0 {
			updatedState = ampq.SUCCESS
			updatedResult = checkMainTaskResult(t.TaskId)
			finishedTask = append(finishedTask, t.TaskId)
		}
		// 如果进度相同则不需要更新
		if updatedProgress == t.ProgressMessage {
			updatedProgress = ""
		}
		// 更新任务
		if updatedState != "" || updatedProgress != "" || updatedResult != "" {
			if updateMainTask(&t, updatedState, updatedProgress, updatedResult) == false {
				msg := fmt.Sprintf("update maintask status fail:%s", t.TaskId)
				logging.RuntimeLog.Error(msg)
				return errors.New(msg)
			}
		}
	}
	// 发送任务通知，从map中移除已完成任务
	comm.MainTaskResultMutex.Lock()
	for _, taskId := range finishedTask {
		message := formatNotifyMessage(taskId)
		go notify.Send(message)
		delete(comm.MainTaskResult, taskId)
	}
	comm.MainTaskResultMutex.Unlock()
	return
}

// formatNotifyMessage 返回发送通知消息内容
func formatNotifyMessage(taskId string) (message string) {
	task := db.TaskMain{TaskId: taskId}
	if !task.GetByTaskId() {
		return
	}
	var sb strings.Builder
	sb.WriteString(task.TaskName)
	sb.WriteString("->")
	sb.WriteString("runtime:")
	sb.WriteString(task.SucceededTime.Sub(*task.StartedTime).Truncate(time.Second).String())
	sb.WriteString(",runtask:")
	sb.WriteString(task.ProgressMessage)
	sb.WriteString("  \n")
	sb.WriteString("target->")
	sb.WriteString(strings.ReplaceAll(ParseTargetFromKwArgs(task.TaskName, task.KwArgs), "\n", ","))
	if len(task.Result) > 0 {
		sb.WriteString("  \nresult->")
		sb.WriteString(task.Result)
	}
	return sb.String()
}

// checkRunTask 根据maintaskId，获取runtask运行情况
func checkRunTask(taskId string) (createdTask, startedTask, totalTask int) {
	taskRun := db.TaskRun{}
	searchMapRun := make(map[string]interface{})
	searchMapRun["main_id"] = taskId
	runTasks, _ := taskRun.Gets(searchMapRun, -1, -1)
	for _, t := range runTasks {
		if t.State == ampq.CREATED {
			createdTask++
		} else if t.State == ampq.STARTED {
			startedTask++
		}
	}
	totalTask = len(runTasks)
	return
}

// checkMainTaskResult 获取maintask的任务结果汇总
func checkMainTaskResult(taskId string) (result string) {
	comm.MainTaskResultMutex.Lock()
	defer comm.MainTaskResultMutex.Unlock()

	if _, ok := comm.MainTaskResult[taskId]; !ok {
		return
	}
	var resultAllString []string
	taskObj := comm.MainTaskResult[taskId]
	if len(taskObj.IPResult) > 0 {
		if taskObj.IPNew > 0 {
			resultAllString = append(resultAllString, fmt.Sprintf("ip:%d(+%d)", len(taskObj.IPResult), taskObj.IPNew))

		} else {
			resultAllString = append(resultAllString, fmt.Sprintf("ip:%d", len(taskObj.IPResult)))
		}
		var portNum int
		for _, ports := range taskObj.IPResult {
			portNum += len(ports)
		}
		if portNum > 0 {
			if taskObj.PortNew > 0 {
				resultAllString = append(resultAllString, fmt.Sprintf("port:%d(+%d)", portNum, taskObj.PortNew))
			} else {
				resultAllString = append(resultAllString, fmt.Sprintf("port:%d", portNum))
			}
		}
	}
	if len(taskObj.DomainResult) > 0 {
		if taskObj.DomainNew > 0 {
			resultAllString = append(resultAllString, fmt.Sprintf("domain:%d(+%d)", len(taskObj.DomainResult), taskObj.DomainNew))
		} else {
			resultAllString = append(resultAllString, fmt.Sprintf("domain:%d", len(taskObj.DomainResult)))
		}
	}
	if len(taskObj.VulResult) > 0 {
		var vulNum int
		for _, vul := range taskObj.VulResult {
			vulNum += len(vul)
		}
		if vulNum > 0 {
			if taskObj.VulnerabilityNew > 0 {
				resultAllString = append(resultAllString, fmt.Sprintf("vulnerability:%d(+%d)", vulNum, taskObj.VulnerabilityNew))
			} else {
				resultAllString = append(resultAllString, fmt.Sprintf("vulnerability:%d", vulNum))
			}
		}
	}
	if taskObj.ScreenShotResult > 0 {
		resultAllString = append(resultAllString, fmt.Sprintf("screenshot:%d", taskObj.ScreenShotResult))
	}
	if len(resultAllString) > 0 {
		result = strings.Join(resultAllString, ",")
	}
	return
}

// updateMainTask 更新数据库中maintask
func updateMainTask(task *db.TaskMain, state string, progress string, result string) bool {
	updateMap := make(map[string]interface{})
	if state != "" {
		updateMap["state"] = state
		if state == ampq.SUCCESS {
			updateMap["succeeded"] = time.Now()
		} else if state == ampq.STARTED {
			updateMap["started"] = time.Now()
		}
	}
	if progress != "" {
		updateMap["progress_message"] = progress
	}
	if result != "" {
		updateMap["result"] = result
	}

	return task.Update(updateMap)
}
