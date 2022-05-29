package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/db"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/ampq"
	"github.com/hanc00l/nemo_go/pkg/task/execute"
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
	State      string `form:"task_state"`
	Name       string `form:"task_name"`
	KwArgs     string `form:"task_args"`
	Worker     string `form:"task_worker"`
	CronTaskId string `form:"cron_id"`
}

type taskCronRequestParam struct {
	DatableRequestParam
	Name   string `form:"task_name"`
	KwArgs string `form:"task_args"`
	Status string `form:"task_status"`
}

type TaskListData struct {
	Id           int    `json:"id"`
	Index        int    `json:"index"`
	TaskId       string `json:"task_id""`
	Worker       string `json:"worker"`
	TaskName     string `json:"task_name"`
	State        string `json:"state"`
	Result       string `json:"result"`
	KwArgs       string `json:"kwargs"`
	ReceivedTime string `json:"received"`
	StartedTime  string `json:"started"`
	Runtime      string `json:"runtime"`
	ResultFile   string `json:"resultfile"`
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
		taskInfo = getTaskInfo(taskId)
	}
	c.Data["task_info"] = taskInfo
	c.Layout = "base.html"
	c.TplName = "task-info.html"
}

// InfoCronAction 显示一个任务的详情
func (c *TaskController) InfoCronAction() {
	var taskInfo TaskCronInfo

	taskId := c.GetString("task_id")
	if taskId != "" {
		taskInfo = getTaskCronInfo(taskId)
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
		task := db.Task{Id: id}
		resultPath := conf.GlobalServerConfig().Web.TaskResultPath
		if resultPath != "" && task.Get() {
			filePath := path.Join(resultPath, fmt.Sprintf("%s.json", task.TaskId))
			os.Remove(filePath)
		}
		c.MakeStatusResponse(task.Delete())
	}
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
			execute.DeleteCronTask(task.TaskId)
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
		c.MakeStatusResponse(execute.ChangeTaskCronStatus(taskId, "disable"))
		return
	}
	c.MakeStatusResponse(false)
}

// EnableCronTaskAction 启用一个任务
func (c *TaskController) EnableCronTaskAction() {
	defer c.ServeJSON()

	taskId := c.GetString("task_id")
	if taskId != "" {
		c.MakeStatusResponse(execute.ChangeTaskCronStatus(taskId, "enable"))
		return
	}
	c.MakeStatusResponse(false)
}

// StartPortScanTaskAction 端口扫描任务
func (c *TaskController) StartPortScanTaskAction() {
	defer c.ServeJSON()
	// 解析参数
	var req execute.PortscanRequestParam
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
	if req.IsTaskCron {
		kwargs, err := json.Marshal(req)
		if err != nil {
			c.FailedStatus(err.Error())
			return
		}
		taskId := execute.SaveCronTask("portscan", string(kwargs), req.TaskCronRule)
		if taskId == "" {
			c.FailedStatus("save to db fail")
			return
		}
		c.SucceededStatus(taskId)
	} else {
		taskId, err := execute.StartPortScanTask(req, "")
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
	var req execute.PortscanRequestParam
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
	if req.IsTaskCron {
		kwargs, err := json.Marshal(req)
		if err != nil {
			c.FailedStatus(err.Error())
			return
		}
		taskId := execute.SaveCronTask("batchscan", string(kwargs), req.TaskCronRule)
		if taskId == "" {
			c.FailedStatus("save to db fail")
			return
		}
		c.SucceededStatus(taskId)
	} else {
		taskId, err := execute.StartBatchScanTask(req, "")
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
	var req execute.DomainscanRequestParam
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
	if req.IsTaskCron {
		kwargs, err := json.Marshal(req)
		if err != nil {
			c.FailedStatus(err.Error())
			return
		}
		taskId := execute.SaveCronTask("domainscan", string(kwargs), req.TaskCronRule)
		if taskId == "" {
			c.FailedStatus("save to db fail")
			return
		}
		c.SucceededStatus(taskId)
	} else {
		taskId, err := execute.StartDomainScanTask(req, "")
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
	var req execute.PocscanRequestParam
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
	if req.IsTaskCron {
		kwargs, err := json.Marshal(req)
		if err != nil {
			c.FailedStatus(err.Error())
			return
		}
		taskId := execute.SaveCronTask("pocscan", string(kwargs), req.TaskCronRule)
		if taskId == "" {
			c.FailedStatus("save to db fail")
			return
		}
		c.SucceededStatus(taskId)
	} else {
		taskId, err := execute.StartPocScanTask(req, "")
		if err != nil {
			c.FailedStatus(err.Error())
			return
		}
		c.SucceededStatus(taskId)
	}
}

//validateRequestParam 校验请求的参数
func (c *TaskController) validateRequestParam(req *taskRequestParam) {
	if req.Length <= 0 {
		req.Length = 50
	}
	if req.Start < 0 {
		req.Start = 0
	}
}

//validateRequestParam 校验请求的参数
func (c *TaskController) validateRequestParam2(req *taskCronRequestParam) {
	if req.Length <= 0 {
		req.Length = 50
	}
	if req.Start < 0 {
		req.Start = 0
	}
}

// getSearchMap 根据查询参数生成查询条件
func (c *TaskController) getSearchMap(req taskRequestParam) (searchMap map[string]interface{}) {
	searchMap = make(map[string]interface{})

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
	task := db.Task{}
	searchMap := c.getSearchMap(req)
	startPage := req.Start/req.Length + 1
	results, total := task.Gets(searchMap, startPage, req.Length)
	resultPath := conf.GlobalServerConfig().Web.TaskResultPath
	for i, taskRow := range results {
		t := TaskListData{}
		t.Id = taskRow.Id
		t.Index = req.Start + i + 1
		t.TaskId = taskRow.TaskId
		t.TaskName = taskRow.TaskName
		t.Worker = taskRow.Worker
		t.State = taskRow.State
		t.Result = getResultMsg(taskRow.Result)
		t.KwArgs = ParseTargetFromKwArgs(taskRow.KwArgs)
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
				t.ResultFile = fmt.Sprintf("/taskresult/%s.json", taskRow.TaskId)
			}
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
		t.KwArgs = strings.ReplaceAll(ParseTargetFromKwArgs(taskRow.KwArgs), "\n", ",")
		t.CronRule = taskRow.CronRule
		t.RunCount = taskRow.RunCount
		t.Status = taskRow.Status
		t.CreateTime = FormatDateTime(taskRow.CreateDatetime)
		if taskRow.LastRunDatetime != taskRow.CreateDatetime {
			t.LastRunTime = FormatDateTime(taskRow.LastRunDatetime)
		}
		if jobExist, dt := execute.GetCronTaskNextRunDatetime(taskRow.TaskId); jobExist {
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
func getTaskInfo(taskId string) (r TaskInfo) {
	task := db.Task{TaskId: taskId}
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

	resultPath := conf.GlobalServerConfig().Web.TaskResultPath
	if resultPath != "" {
		filePath := path.Join(resultPath, fmt.Sprintf("%s.json", taskId))
		if utils.CheckFileExist(filePath) {
			r.ResultFile = fmt.Sprintf("/taskresult/%s.json", taskId)
		}
	}
	return
}

// getTaskCronInfo 获取一个任务的详情
func getTaskCronInfo(taskId string) (r TaskCronInfo) {
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
	r.CreateTime = FormatDateTime(task.CreateDatetime)
	r.UpdateTime = FormatDateTime(task.UpdateDatetime)
	if task.LastRunDatetime != task.CreateDatetime {
		r.LastRunTime = FormatDateTime(task.LastRunDatetime)
	}
	return
}

// formatRuntime 计算任务运行时间
func formatRuntime(t *db.Task) (runtime string) {
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

// ParseTargetFromKwArgs 从经过JSON序列化的参数中单独提取出target
func ParseTargetFromKwArgs(args string) (target string) {
	const displayedLength = 100
	type TargetStrut struct {
		Target string `json:"target"`
	}
	var t TargetStrut
	err := json.Unmarshal([]byte(args), &t)
	if err != nil {
		target = args
	} else {
		target = t.Target
	}
	if len(target) > displayedLength {
		return fmt.Sprintf("%s...", target[:displayedLength])
	}
	return
}
