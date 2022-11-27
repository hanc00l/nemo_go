package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/db"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/ampq"
	"github.com/hanc00l/nemo_go/pkg/task/runner"
	"github.com/hanc00l/nemo_go/pkg/task/serverapi"
	"github.com/hanc00l/nemo_go/pkg/utils"
	"os"
	"path"
	"strings"
	"time"
)

type TaskController struct {
	BaseController
}

type taskRequestParam struct {
	DatableRequestParam
	State        string `form:"task_state"`
	Name         string `form:"task_name"`
	KwArgs       string `form:"task_args"`
	Worker       string `form:"task_worker"`
	CronTaskId   string `form:"cron_id"`
	ShowRunTask  bool   `form:"show_runtask"`
	RunTaskState string `form:"runtask_state"`
}

type taskCronRequestParam struct {
	DatableRequestParam
	Name   string `form:"task_name"`
	KwArgs string `form:"task_args"`
	Status string `form:"task_status"`
}

type TaskListData struct {
	Id           int    `json:"id"`
	Index        string `json:"index"`
	TaskId       string `json:"task_id"`
	Worker       string `json:"worker"`
	TaskName     string `json:"task_name"`
	State        string `json:"state"`
	Result       string `json:"result"`
	KwArgs       string `json:"kwargs"`
	ReceivedTime string `json:"received"`
	StartedTime  string `json:"started"`
	CreateTime   string `json:"created"`
	UpdateTime   string `json:"updated"`
	Runtime      string `json:"runtime"`
	ResultFile   string `json:"resultfile"`
	TaskType     string `json:"tasktype"`
}

type TaskCronListData struct {
	Id          int    `json:"id"`
	Index       int    `json:"index"`
	TaskId      string `json:"task_id""`
	TaskName    string `json:"task_name"`
	Status      string `json:"status"`
	KwArgs      string `json:"kwargs"`
	CronRule    string `json:"cron_rule"`
	CreateTime  string `json:"create_time"`
	LastRunTime string `json:"lastrun_time"`
	NextRunTime string `json:"nextrun_time"`
	RunCount    int    `json:"run_count"`
	Comment     string `json:"comment"`
}

type TaskInfo struct {
	Id            int
	TaskId        string
	Worker        string
	TaskName      string
	State         string
	Result        string
	KwArgs        string
	ReceivedTime  string
	StartedTime   string
	SucceededTime string
	FailedTime    string
	RetriedTime   string
	RevokedTime   string
	Runtime       string
	CreateTime    string
	UpdateTime    string
	ResultFile    string
	RunTaskInfo   []TaskListData
}

type TaskCronInfo struct {
	Id          int
	TaskId      string
	TaskName    string
	Status      string
	KwArgs      string
	CronRule    string
	LastRunTime string
	CreateTime  string
	UpdateTime  string
	RunCount    int
	Comment     string
}

func (c *TaskController) IndexAction() {
	c.UpdateOnlineUser()
	c.Layout = "base.html"
	c.TplName = "task-list.html"
}

func (c *TaskController) IndexCronAction() {
	c.UpdateOnlineUser()
	c.Layout = "base.html"
	c.TplName = "task-cron-list.html"
}

// ListAction 任务列表的数据
func (c *TaskController) ListAction() {
	c.UpdateOnlineUser()
	defer c.ServeJSON()

	req := taskRequestParam{}
	err := c.ParseForm(&req)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
	}
	c.validateRequestParam(&req)
	resp := c.getTaskListData(req)
	c.Data["json"] = resp
}

// ListCronAction 定时任务列表的数据
func (c *TaskController) ListCronAction() {
	c.UpdateOnlineUser()
	defer c.ServeJSON()

	req := taskCronRequestParam{}
	err := c.ParseForm(&req)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
	}
	c.validateRequestParam2(&req)
	resp := c.getTaskCronListData(req)
	c.Data["json"] = resp
}

// InfoAction 显示一个任务的详情
func (c *TaskController) InfoAction() {
	var taskInfo TaskInfo

	taskId := c.GetString("task_id")
	if taskId != "" {
		taskInfo = c.getTaskInfo(taskId)
	}
	c.Data["task_info"] = taskInfo
	c.Layout = "base.html"
	c.TplName = "task-info.html"
}

// InfoMainAction 显示一个Main任务的详情
func (c *TaskController) InfoMainAction() {
	var taskInfo TaskInfo

	taskId := c.GetString("task_id")
	if taskId != "" {
		taskInfo = c.getTaskMainInfo(taskId)
	}
	c.Data["task_info"] = taskInfo
	c.Layout = "base.html"
	c.TplName = "task-info-main.html"
}

// InfoCronAction 显示一个任务的详情
func (c *TaskController) InfoCronAction() {
	var taskInfo TaskCronInfo

	taskId := c.GetString("task_id")
	if taskId != "" {
		taskInfo = c.getTaskCronInfo(taskId)
	}
	c.Data["task_info"] = taskInfo
	c.Layout = "base.html"
	c.TplName = "task-cron-info.html"
}

// DeleteAction 删除一个记录
func (c *TaskController) DeleteAction() {
	defer c.ServeJSON()

	id, err := c.GetInt("id")
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		c.FailedStatus(err.Error())
	} else {
		task := db.TaskRun{Id: id}
		resultPath := path.Join(conf.GlobalServerConfig().Web.WebFiles, "taskresult")
		if task.Get() {
			filePath := path.Join(resultPath, fmt.Sprintf("%s.json", task.TaskId))
			os.Remove(filePath)
		}
		c.MakeStatusResponse(task.Delete())
	}
}

// DeleteMainAction 删除一个记录
func (c *TaskController) DeleteMainAction() {
	defer c.ServeJSON()

	id, err := c.GetInt("id")
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		c.FailedStatus(err.Error())
	} else {
		task := db.TaskMain{Id: id}
		if task.Get() {
			//同时删除相关的子任务
			deleteRunTaskByMainTaskId(task.TaskId)
		}
		c.MakeStatusResponse(task.Delete())
	}
}

// DeleteBatchAction 批量删除任务
func (c *TaskController) DeleteBatchAction() {
	defer c.ServeJSON()

	taskType := c.GetString("type", "")
	taskTotal := 0
	if taskType == "created" {
		taskTotal += batchDeleteTaskByState(ampq.CREATED)
	} else if taskType == "unfinished" {
		taskTotal += batchDeleteTaskByState(ampq.CREATED)
		taskTotal += batchDeleteTaskByState(ampq.STARTED)
	} else if taskType == "finished" {
		taskTotal += batchDeleteTaskByState(ampq.REVOKED)
		taskTotal += batchDeleteTaskByState(ampq.FAILURE)
		taskTotal += batchDeleteTaskByState(ampq.SUCCESS)
	}
	c.SucceededStatus(fmt.Sprintf("共删除任务:%d", taskTotal))
}

// DeleteCronAction 删除一个记录
func (c *TaskController) DeleteCronAction() {
	defer c.ServeJSON()

	id, err := c.GetInt("id")
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		c.FailedStatus(err.Error())
	} else {
		task := db.TaskCron{Id: id}
		if task.Get() {
			runner.DeleteCronTask(task.TaskId)
			c.MakeStatusResponse(task.Delete())
		} else {
			c.FailedStatus("任务不存在")
		}
	}
}

// StopAction 取消一个未开始执行的任务
func (c *TaskController) StopAction() {
	defer c.ServeJSON()

	taskId := c.GetString("task_id")
	if taskId != "" {
		isRevoked, _ := serverapi.RevokeUnexcusedTask(taskId)
		c.MakeStatusResponse(isRevoked)
		return
	}
	c.MakeStatusResponse(false)
}

// DisableCronTaskAction 禁用一个任务
func (c *TaskController) DisableCronTaskAction() {
	defer c.ServeJSON()

	taskId := c.GetString("task_id")
	if taskId != "" {
		c.MakeStatusResponse(runner.ChangeTaskCronStatus(taskId, "disable"))
		return
	}
	c.MakeStatusResponse(false)
}

// EnableCronTaskAction 启用一个任务
func (c *TaskController) EnableCronTaskAction() {
	defer c.ServeJSON()

	taskId := c.GetString("task_id")
	if taskId != "" {
		c.MakeStatusResponse(runner.ChangeTaskCronStatus(taskId, "enable"))
		return
	}
	c.MakeStatusResponse(false)
}

// RunCronTaskAction 立即执行一个任务
func (c *TaskController) RunCronTaskAction() {
	defer c.ServeJSON()

	taskId := c.GetString("task_id")
	if taskId != "" {
		c.MakeStatusResponse(runner.RunOnceTaskCron(taskId))
		return
	}
	c.MakeStatusResponse(false)
}

// StartPortScanTaskAction 端口扫描任务
func (c *TaskController) StartPortScanTaskAction() {
	defer c.ServeJSON()
	// 解析参数
	var req runner.PortscanRequestParam
	err := c.ParseForm(&req)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		c.FailedStatus(err.Error())
		return
	}
	if req.Target == "" {
		c.FailedStatus("no target")
		return
	}
	if req.Port == "" {
		req.Port = conf.GlobalWorkerConfig().Portscan.Port
	}
	var kwArgs []byte
	var taskId string
	kwArgs, err = json.Marshal(req)
	if err != nil {
		c.FailedStatus(err.Error())
		return
	}
	if req.IsTaskCron {
		taskId = runner.SaveCronTask("portscan", string(kwArgs), req.TaskCronRule, req.TaskCronComment)
		if taskId == "" {
			c.FailedStatus("save to db fail")
			return
		}
		c.SucceededStatus(taskId)
	} else {
		taskId, err = runner.SaveMainTask("portscan", string(kwArgs), "")
		if err != nil {
			c.FailedStatus(err.Error())
			return
		}
		c.SucceededStatus(taskId)
	}
}

// StartBatchScanTaskAction 探测+扫描任务
func (c *TaskController) StartBatchScanTaskAction() {
	defer c.ServeJSON()
	// 解析参数
	var req runner.PortscanRequestParam
	err := c.ParseForm(&req)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		c.FailedStatus(err.Error())
		return
	}
	if req.Target == "" {
		c.FailedStatus("no target")
		return
	}
	var kwArgs []byte
	var taskId string
	kwArgs, err = json.Marshal(req)
	if err != nil {
		c.FailedStatus(err.Error())
		return
	}
	if req.IsTaskCron {
		taskId = runner.SaveCronTask("batchscan", string(kwArgs), req.TaskCronRule, req.TaskCronComment)
		if taskId == "" {
			c.FailedStatus("save to db fail")
			return
		}
		c.SucceededStatus(taskId)
	} else {
		taskId, err = runner.SaveMainTask("batchscan", string(kwArgs), "")
		if err != nil {
			c.FailedStatus(err.Error())
			return
		}
		c.SucceededStatus(taskId)
	}
}

// StartDomainScanTaskAction 域名任务
func (c *TaskController) StartDomainScanTaskAction() {
	defer c.ServeJSON()

	// 解析参数
	var req runner.DomainscanRequestParam
	err := c.ParseForm(&req)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		c.FailedStatus(err.Error())
		return
	}
	if req.Target == "" {
		c.FailedStatus("no target")
		return
	}
	var kwArgs []byte
	var taskId string
	kwArgs, err = json.Marshal(req)
	if err != nil {
		c.FailedStatus(err.Error())
		return
	}
	if req.IsTaskCron {

		taskId = runner.SaveCronTask("domainscan", string(kwArgs), req.TaskCronRule, req.TaskCronComment)
		if taskId == "" {
			c.FailedStatus("save to db fail")
			return
		}
		c.SucceededStatus(taskId)
	} else {
		taskId, err = runner.SaveMainTask("domainscan", string(kwArgs), "")
		if err != nil {
			c.FailedStatus(err.Error())
			return
		}
		c.SucceededStatus(taskId)
	}
}

// StartPocScanTaskAction pocscan任务
func (c *TaskController) StartPocScanTaskAction() {
	defer c.ServeJSON()

	// 解析参数
	var req runner.PocscanRequestParam
	err := c.ParseForm(&req)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		c.FailedStatus(err.Error())
		return
	}
	// 格式化Target
	if req.Target == "" {
		c.FailedStatus("no target")
		return
	}
	var kwArgs []byte
	var taskId string
	kwArgs, err = json.Marshal(req)
	if err != nil {
		c.FailedStatus(err.Error())
		return
	}
	if req.IsTaskCron {
		taskId = runner.SaveCronTask("pocscan", string(kwArgs), req.TaskCronRule, req.TaskCronComment)
		if taskId == "" {
			c.FailedStatus("save to db fail")
			return
		}
		c.SucceededStatus(taskId)
	} else {
		taskId, err = runner.SaveMainTask("pocscan", string(kwArgs), "")
		if err != nil {
			c.FailedStatus(err.Error())
			return
		}
		c.SucceededStatus(taskId)
	}
}

// StartXScanTaskAction 新建Xscan任务
func (c *TaskController) StartXScanTaskAction() {
	c.UpdateOnlineUser()
	defer c.ServeJSON()
	//校验参数
	req := runner.XScanRequestParam{}
	err := c.ParseForm(&req)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		c.FailedStatus(err.Error())
		return
	}
	var taskName string
	if req.XScanType == "xportscan" {
		taskName = "xportscan"
		if req.Target == "" {
			c.FailedStatus("no target")
			return
		}
		if req.Port == "" {
			req.Port = conf.GlobalWorkerConfig().Portscan.Port
		}
	} else if req.XScanType == "xorgipscan" {
		taskName = "xorgscan"
		req.IsOrgIP = true
		if req.OrgId == 0 {
			c.FailedStatus("no org")
			return
		}
	} else if req.XScanType == "xdomainscan" {
		taskName = "xdomainscan"
		if req.Target == "" {
			c.FailedStatus("no target")
			return
		}
	} else if req.XScanType == "xorgdomainscan" {
		taskName = "xorgscan"
		req.IsOrgDomain = true
		if req.OrgId == 0 {
			c.FailedStatus("no org")
			return
		}
	} else if req.XScanType == "xfofa" {
		taskName = "xfofa"
	} else {
		c.FailedStatus("invalide xscan type")
		return
	}
	var kwArgs []byte
	var taskId string
	kwArgs, err = json.Marshal(req)
	if err != nil {
		c.FailedStatus(err.Error())
		return
	}

	var result string
	// 计划任务
	if req.IsTaskCron {
		taskId = runner.SaveCronTask(taskName, string(kwArgs), req.TaskCronRule, req.TaskCronComment)
		if taskId == "" {
			c.FailedStatus("save to db fail")
			return
		}
	} else {
		// 立即执行的任务
		taskId, err = runner.SaveMainTask(taskName, string(kwArgs), "")
		if err != nil {
			c.FailedStatus(err.Error())
			return
		}
	}
	c.SucceededStatus(result)
}

// validateRequestParam 校验请求的参数
func (c *TaskController) validateRequestParam(req *taskRequestParam) {
	if req.Length <= 0 {
		req.Length = 50
	}
	if req.Start < 0 {
		req.Start = 0
	}
}

// validateRequestParam 校验请求的参数
func (c *TaskController) validateRequestParam2(req *taskCronRequestParam) {
	if req.Length <= 0 {
		req.Length = 50
	}
	if req.Start < 0 {
		req.Start = 0
	}
}

// getSearchMap 根据查询参数生成查询条件
func (c *TaskController) getSearchMap(req *taskRequestParam) (searchMap map[string]interface{}) {
	searchMap = make(map[string]interface{})
	if req == nil {
		return
	}
	if req.Name != "" {
		searchMap["task_name"] = req.Name
	}
	if req.State != "" {
		searchMap["state"] = req.State
	}
	if req.KwArgs != "" {
		searchMap["kwargs"] = req.KwArgs
	}
	if req.Worker != "" {
		searchMap["worker"] = req.Worker
	}
	if req.CronTaskId != "" {
		searchMap["cron_id"] = req.CronTaskId
	}
	return
}

// getSearchMap 根据查询参数生成查询条件
func (c *TaskController) getSearchMap2(req taskCronRequestParam) (searchMap map[string]interface{}) {
	searchMap = make(map[string]interface{})

	if req.Name != "" {
		searchMap["task_name"] = req.Name
	}
	if req.KwArgs != "" {
		searchMap["kwargs"] = req.KwArgs
	}
	if req.Status != "" {
		searchMap["status"] = req.Status
	}
	return
}

// getTaskListData 获取列显示的数据
func (c *TaskController) getTaskListData(req taskRequestParam) (resp DataTableResponseData) {
	task := db.TaskMain{}
	searchMap := c.getSearchMap(&req)
	startPage := req.Start/req.Length + 1
	results, total := task.Gets(searchMap, startPage, req.Length)
	for i, taskRow := range results {
		t := TaskListData{}
		t.Id = taskRow.Id
		t.Index = fmt.Sprintf("%d", req.Start+i+1)
		t.TaskId = taskRow.TaskId
		t.TaskName = taskRow.TaskName
		t.Worker = taskRow.ProgressMessage
		t.State = taskRow.State
		t.Result = getResultMsg(taskRow.Result)
		t.KwArgs = runner.ParseTargetFromKwArgs(taskRow.TaskName, taskRow.KwArgs)
		t.ReceivedTime = FormatDateTime(taskRow.ReceivedTime)
		if taskRow.StartedTime != nil {
			t.StartedTime = FormatDateTime(*taskRow.StartedTime)
		}
		if taskRow.SucceededTime != nil && taskRow.StartedTime != nil {
			t.Runtime = taskRow.SucceededTime.Sub(*taskRow.StartedTime).Truncate(time.Second).String()
		}
		t.TaskType = "MainTask"
		resp.Data = append(resp.Data, t)
		if req.ShowRunTask {
			for _, rt := range c.getRunTaskListData(taskRow.TaskId, &req, false, false) {
				resp.Data = append(resp.Data, rt)
			}
		}
	}
	resp.Draw = req.Draw
	resp.RecordsTotal = total
	resp.RecordsFiltered = total
	if resp.Data == nil {
		resp.Data = make([]interface{}, 0)
	}
	return
}

// getRunTaskListData 获取指定maintask的runtask数据
func (c *TaskController) getRunTaskListData(mainTaskId string, req *taskRequestParam, showIndex, showAll bool) (runTaskList []TaskListData) {
	task := db.TaskRun{}
	searchMap := make(map[string]interface{})
	searchMap["main_id"] = mainTaskId
	if req != nil && req.RunTaskState != "" {
		searchMap["state"] = req.RunTaskState
	}
	var results []db.TaskRun
	if showAll {
		results, _ = task.Gets(searchMap, -1, -1)
	} else {
		results, _ = task.Gets(searchMap, 1, 10)
	}
	resultPath := path.Join(conf.GlobalServerConfig().Web.WebFiles, "taskresult")
	for index, taskRow := range results {
		t := TaskListData{}
		if showIndex {
			t.Index = fmt.Sprintf("%d", index+1)
		}
		t.Id = taskRow.Id
		t.TaskId = taskRow.TaskId
		t.TaskName = taskRow.TaskName
		t.Worker = taskRow.Worker
		t.State = taskRow.State
		t.Result = getResultMsg(taskRow.Result)
		t.KwArgs = runner.ParseTargetFromKwArgs(taskRow.TaskName, taskRow.KwArgs)
		t.CreateTime = FormatDateTime(taskRow.CreateDatetime)
		t.UpdateTime = FormatDateTime(taskRow.UpdateDatetime)
		if taskRow.StartedTime != nil {
			t.StartedTime = FormatDateTime(*taskRow.StartedTime)
		}
		if taskRow.ReceivedTime != nil {
			t.ReceivedTime = FormatDateTime(*taskRow.ReceivedTime)
		}
		t.Runtime = formatRuntime(&taskRow)
		if resultPath != "" {
			filePath := path.Join(resultPath, fmt.Sprintf("%s.json", taskRow.TaskId))
			if utils.CheckFileExist(filePath) {
				t.ResultFile = fmt.Sprintf("/webfiles/taskresult/%s.json", taskRow.TaskId)
			}
		}
		t.TaskType = "RunTask"
		runTaskList = append(runTaskList, t)
	}
	return
}

// getTaskListData 获取列显示的数据
func (c *TaskController) getTaskCronListData(req taskCronRequestParam) (resp DataTableResponseData) {
	task := db.TaskCron{}
	searchMap := c.getSearchMap2(req)
	startPage := req.Start/req.Length + 1
	results, total := task.Gets(searchMap, startPage, req.Length)
	for i, taskRow := range results {
		t := TaskCronListData{}
		t.Id = taskRow.Id
		t.Index = req.Start + i + 1
		t.TaskId = taskRow.TaskId
		t.TaskName = taskRow.TaskName
		t.KwArgs = strings.ReplaceAll(runner.ParseTargetFromKwArgs(taskRow.TaskName, taskRow.KwArgs), "\n", ",")
		t.CronRule = taskRow.CronRule
		t.RunCount = taskRow.RunCount
		t.Status = taskRow.Status
		t.Comment = taskRow.Comment
		t.CreateTime = FormatDateTime(taskRow.CreateDatetime)
		if taskRow.LastRunDatetime != taskRow.CreateDatetime {
			t.LastRunTime = FormatDateTime(taskRow.LastRunDatetime)
		}
		if jobExist, dt := runner.GetCronTaskNextRunDatetime(taskRow.TaskId); jobExist {
			t.NextRunTime = FormatDateTime(dt)
		}
		resp.Data = append(resp.Data, t)
	}
	resp.Draw = req.Draw
	resp.RecordsTotal = total
	resp.RecordsFiltered = total
	if resp.Data == nil {
		resp.Data = make([]interface{}, 0)
	}
	return
}

// getTaskInfo 获取一个任务的详情
func (c *TaskController) getTaskInfo(taskId string) (r TaskInfo) {
	task := db.TaskRun{TaskId: taskId}
	if !task.GetByTaskId() {
		return
	}
	r.Id = task.Id
	r.TaskId = task.TaskId
	r.TaskName = task.TaskName
	r.Worker = task.Worker
	r.Result = task.Result
	r.State = task.State
	r.KwArgs = task.KwArgs
	if task.StartedTime != nil {
		r.StartedTime = FormatDateTime(*task.StartedTime)
	}
	if task.ReceivedTime != nil {
		r.ReceivedTime = FormatDateTime(*task.ReceivedTime)
	}
	if task.RetriedTime != nil {
		r.RetriedTime = FormatDateTime(*task.RetriedTime)
	}
	if task.RevokedTime != nil {
		r.RevokedTime = FormatDateTime(*task.RevokedTime)
	}
	if task.FailedTime != nil {
		r.FailedTime = FormatDateTime(*task.FailedTime)
	}
	if task.SucceededTime != nil {
		r.SucceededTime = FormatDateTime(*task.SucceededTime)
	}
	r.Runtime = formatRuntime(&task)
	r.CreateTime = FormatDateTime(task.CreateDatetime)
	r.UpdateTime = FormatDateTime(task.UpdateDatetime)

	resultPath := path.Join(conf.GlobalServerConfig().Web.WebFiles, "taskresult")
	if resultPath != "" {
		filePath := path.Join(resultPath, fmt.Sprintf("%s.json", taskId))
		if utils.CheckFileExist(filePath) {
			r.ResultFile = fmt.Sprintf("/webfiles/taskresult/%s.json", taskId)
		}
	}
	return
}

// getTaskCronInfo 获取一个任务的详情
func (c *TaskController) getTaskCronInfo(taskId string) (r TaskCronInfo) {
	task := db.TaskCron{TaskId: taskId}
	if !task.GetByTaskId() {
		return
	}
	r.Id = task.Id
	r.TaskId = task.TaskId
	r.TaskName = task.TaskName
	r.KwArgs = task.KwArgs
	r.CronRule = task.CronRule
	r.RunCount = task.RunCount
	r.Status = task.Status
	r.Comment = task.Comment
	r.CreateTime = FormatDateTime(task.CreateDatetime)
	r.UpdateTime = FormatDateTime(task.UpdateDatetime)
	if task.LastRunDatetime != task.CreateDatetime {
		r.LastRunTime = FormatDateTime(task.LastRunDatetime)
	}
	return
}

// getTaskMainInfo 获取一个任务的详情
func (c *TaskController) getTaskMainInfo(taskId string) (r TaskInfo) {
	task := db.TaskMain{TaskId: taskId}
	if !task.GetByTaskId() {
		return
	}
	r.Id = task.Id
	r.TaskId = task.TaskId
	r.TaskName = task.TaskName
	r.Result = task.Result
	r.State = task.State
	r.KwArgs = task.KwArgs
	r.ReceivedTime = FormatDateTime(task.ReceivedTime)
	r.Worker = task.ProgressMessage
	if task.StartedTime != nil {
		r.StartedTime = FormatDateTime(*task.StartedTime)
	}
	if task.SucceededTime != nil {
		r.SucceededTime = FormatDateTime(*task.SucceededTime)
	}
	if task.SucceededTime != nil && task.StartedTime != nil {
		r.Runtime = task.SucceededTime.Sub(*task.StartedTime).Truncate(time.Second).String()
	}
	r.CreateTime = FormatDateTime(task.CreateDatetime)
	r.UpdateTime = FormatDateTime(task.UpdateDatetime)
	r.RunTaskInfo = c.getRunTaskListData(taskId, nil, true, true)

	return
}

// formatRuntime 计算任务运行时间
func formatRuntime(t *db.TaskRun) (runtime string) {
	var endTime *time.Time
	if t.SucceededTime != nil {
		endTime = t.SucceededTime
	} else if t.FailedTime != nil {
		endTime = t.FailedTime
	} else {
		return
	}
	var startedTime time.Time
	if t.StartedTime != nil {
		startedTime = *t.StartedTime
	} else if t.ReceivedTime != nil {
		startedTime = *t.ReceivedTime
	} else {
		return
	}
	runtime = endTime.Sub(startedTime).Truncate(time.Second).String()

	return
}

// getResultMsg 从经过JSON反序列化的结果中提取出结果的消息
func getResultMsg(resultJSON string) (msg string) {
	var result ampq.TaskResult
	err := json.Unmarshal([]byte(resultJSON), &result)
	if err != nil {
		return resultJSON
	}
	return result.Msg
}

// batchDeleteTaskByState 批量删除指定状态的任务
func batchDeleteTaskByState(taskState string) (total int) {
	searchMap := make(map[string]interface{})
	searchMap["state"] = taskState
	task := db.TaskMain{}
	results, _ := task.Gets(searchMap, -1, -1)
	for _, taskRow := range results {
		taskDelete := db.TaskMain{Id: taskRow.Id}
		taskDelete.Delete()
		total += deleteRunTaskByMainTaskId(taskRow.TaskId)
	}
	return
}

// deleteRunTaskByMainTaskId 删除maintask下的所有rantask子任务
func deleteRunTaskByMainTaskId(mainTaskId string) (total int) {
	task := db.TaskRun{}
	searchMap := make(map[string]interface{})
	searchMap["main_id"] = mainTaskId
	results, _ := task.Gets(searchMap, -1, -1)
	resultPath := path.Join(conf.GlobalServerConfig().Web.WebFiles, "taskresult")
	for _, taskRow := range results {
		taskDelete := db.TaskRun{Id: taskRow.Id}
		if taskDelete.Delete() && resultPath != "" {
			filePath := path.Join(resultPath, fmt.Sprintf("%s.json", taskRow.TaskId))
			os.Remove(filePath)
			total++
		}
	}
	return
}
