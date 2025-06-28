package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/hanc00l/nemo_go/v3/pkg/core"
	"github.com/hanc00l/nemo_go/v3/pkg/db"
	"github.com/hanc00l/nemo_go/v3/pkg/logging"
	"github.com/hanc00l/nemo_go/v3/pkg/task/execute"
	"go.mongodb.org/mongo-driver/v2/bson"
	"time"
)

type MainTaskController struct {
	BaseController
}

type mainTaskRequestParam struct {
	DatableRequestParam
	TaskType string `json:"task_type" form:"task_type"`
	Status   string `json:"status" form:"status"`
	Name     string `json:"name" form:"name"`
	Target   string `json:"target" form:"target"`
}

type TaskData struct {
	Id             string `json:"id"`
	Index          int    `json:"index"`
	TaskId         string `json:"task_id"`
	Name           string `json:"name"`
	Target         string `json:"target"`
	ProfileName    string `json:"profile"`
	Status         string `json:"status"`
	Result         string `json:"result"`
	ProgressRate   string `json:"progress_rate"`
	Progress       string `json:"progress"`
	CreateDatetime string `json:"create_time"`
	UpdateDatetime string `json:"update_time"`
	StartDatetime  string `json:"start_time"`
	Runtime        string `json:"runtime"`
	IsCron         bool   `json:"cron"`
}

type MainTaskInfoData struct {
	Id             string `json:"id" form:"id"`
	Name           string `json:"name" form:"name"`
	ProfileName    string `json:"profile_name" form:"profile_name"`
	ProfileId      string `json:"profile_id" form:"profile_id"`
	Description    string `json:"description" form:"description"`
	Target         string `json:"target" form:"target"`
	ExcludeTarget  string `json:"exclude_target" form:"exclude_target"`
	TargetSplit    int    `json:"target_split" form:"target_split"`
	TargetSplitNum int    `json:"target_split_num" form:"target_split_num"`
	OrgId          string `json:"org_id" form:"org_id"`
	IsCronTask     bool   `json:"is_cron_task" form:"is_cron_task"`
	CronExpr       string `json:"cron_expr" form:"cron_expr"`
	IsProxy        bool   `json:"is_proxy" form:"is_proxy"`
}

type MainTaskDetailInfoData struct {
	TaskId        string `json:"task_id"`
	Name          string `json:"name"`
	ProfileName   string `json:"profile_name"`
	Description   string `json:"description"`
	Target        string `json:"target"`
	ExcludeTarget string `json:"exclude_target"`
	TargetSplit   string `json:"target_split"`
	OrgName       string `json:"org_name"`
	IsCronTask    bool   `json:"is_cron_task" form:"is_cron_task"`
	CronTask      string `json:"cron_task"`
	Proxy         string `json:"proxy"`
	WorkspaceName string `json:"workspace"`
	Args          string `json:"args"`

	Status         string `json:"status"`
	Result         string `json:"result"`
	Progress       string `json:"progress"`
	ProgressRate   string `json:"progress_rate"`
	CreateDatetime string `json:"create_time"`
	StartDatetime  string `json:"start_time"`
	Runtime        string `json:"runtime"`
	EndDatetime    string `json:"end_time"`
}

type ExecutorTaskInfoData struct {
	Id             string `json:"id"`
	Executor       string `json:"executor"`
	Target         string `json:"target"`
	Args           string `json:"args"`
	Status         string `json:"status"`
	Result         string `json:"result"`
	Worker         string `json:"worker"`
	ProgressRate   string `json:"progress_rate"`
	CreateDatetime string `json:"create_time"`
	UpdateDatetime string `json:"update_time"`
	StartDatetime  string `json:"start_time"`
	Runtime        string `json:"runtime"`
	EndDatetime    string `json:"end_time"`
}

// IndexAction 显示列表页面
func (c *MainTaskController) IndexAction() {
	c.CheckOneAccessRequest(SuperAdmin, true)

	c.Layout = "base.html"
	c.TplName = "maintask-list.html"
}

// ListAction 获取列表显示的数据
func (c *MainTaskController) ListAction() {
	defer func(c *MainTaskController, encoding ...bool) {
		_ = c.ServeJSON(encoding...)
	}(c)
	if !c.CheckMultiAccessRequest([]RequestRole{SuperAdmin, Admin}, false) {
		c.FailedStatus("无权访问！")
		return
	}

	req := mainTaskRequestParam{}
	err := c.ParseForm(&req)
	if err != nil {
		logging.RuntimeLog.Error(err)
	}
	c.validateRequestParam(&req)
	resp := c.getListData(req)
	c.Data["json"] = resp
}

// validateRequestParam 校验请求的参数
func (c *MainTaskController) validateRequestParam(req *mainTaskRequestParam) {
	if req.Length <= 0 {
		req.Length = 50
	}
	if req.Start < 0 {
		req.Start = 0
	}
}

// getListData 获取列表数据
func (c *MainTaskController) getListData(req mainTaskRequestParam) (resp DataTableResponseData) {
	defer func() {
		resp.Draw = req.Draw
		if len(resp.Data) == 0 {
			resp.Data = make([]interface{}, 0)
		}
	}()
	workspaceId := c.GetWorkspace()
	if len(workspaceId) == 0 {
		c.FailedStatus("未选择当前的工作空间！")
		return
	}
	mongoClient, err := db.GetClient()
	if err != nil {
		logging.RuntimeLog.Error(err)
		c.FailedStatus(err.Error())
		return
	}
	defer db.CloseClient(mongoClient)

	filter := bson.M{"workspaceId": workspaceId}
	if req.Name != "" {
		filter["name"] = bson.M{"$regex": req.Name, "$options": "i"}
	}
	if req.Status != "" {
		filter["status"] = req.Status
	}
	if req.Target != "" {
		filter["target"] = bson.M{"$regex": req.Target, "$options": "i"}
	}
	if req.TaskType == "cron" {
		filter["cron"] = true
	} else {
		filter["cron"] = false
	}
	mainTask := db.NewMainTask(mongoClient)
	startPage := req.Start/req.Length + 1
	results, err := mainTask.Find(filter, startPage, req.Length)
	if err != nil {
		logging.RuntimeLog.Error(err)
		c.FailedStatus(err.Error())
		return
	}

	for i, row := range results {
		task := TaskData{
			Id:             row.Id.Hex(),
			Index:          req.Start + i + 1,
			TaskId:         row.TaskId,
			Name:           row.TaskName,
			ProfileName:    row.ProfileName,
			Status:         row.Status,
			Result:         row.Result,
			IsCron:         row.IsCron,
			ProgressRate:   fmt.Sprintf("%d%%", int(row.ProgressRate*100.0)),
			Progress:       row.Progress,
			CreateDatetime: FormatDateTime(row.CreateTime),
			UpdateDatetime: FormatDateTime(row.UpdateTime),
		}
		if len(row.Target) > 100 {
			task.Target = row.Target[:100] + "..."
		} else {
			task.Target = row.Target
		}
		if row.StartTime != nil {
			task.StartDatetime = FormatDateTime(row.UpdateTime)
		}
		if row.Status == core.SUCCESS || row.Status == core.FAILURE {
			if row.EndTime != nil && row.StartTime != nil {
				task.Runtime = row.EndTime.Sub(*row.StartTime).Truncate(time.Second).String()
			}
		}
		//如果是定时任务，则显示cron表达式
		if row.IsCron {
			if row.CronTaskInfo.CronEnabled {
				task.Status = "enabled"
			} else {
				task.Status = "disabled"
			}
			task.ProgressRate = fmt.Sprintf("已执行: %d次", row.CronTaskInfo.CronRunCount)
			task.Result = "启动规则: " + row.CronTaskInfo.CronExpr
			if row.CronTaskInfo.CronLastRun != nil {
				task.StartDatetime = FormatDateTime(*row.CronTaskInfo.CronLastRun)
			}
		}

		resp.Data = append(resp.Data, task)
	}
	total, _ := mainTask.Count(filter)
	resp.RecordsTotal = total
	resp.RecordsFiltered = total

	return
}

func (c *MainTaskController) AddIndexAction() {
	c.CheckMultiAccessRequest([]RequestRole{SuperAdmin, Admin}, true)
	c.Layout = "base.html"
	c.TplName = "maintask-add.html"

	return
}

func (c *MainTaskController) AddSaveAction() {
	defer func(c *MainTaskController, encoding ...bool) {
		_ = c.ServeJSON(encoding...)
	}(c)
	if c.CheckMultiAccessRequest([]RequestRole{SuperAdmin, Admin}, false) == false {
		c.FailedStatus("当前用户权限不允许！")
		return
	}
	workspaceId := c.GetWorkspace()
	if len(workspaceId) == 0 {
		c.FailedStatus("未选择当前的工作空间！")
		return
	}
	// 前端 form 表单提交的数据
	var mainTaskInfoData MainTaskInfoData
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &mainTaskInfoData)
	if err != nil {
		logging.RuntimeLog.Error(err)
		c.FailedStatus(err.Error())
		return
	}
	// 处理数据
	doc, err := processMainTaskInfoData(mainTaskInfoData)
	if err != nil {
		c.FailedStatus(err.Error())
		return
	}
	mongoClient, err := db.GetClient()
	if err != nil {
		logging.RuntimeLog.Error(err)
		c.FailedStatus(err.Error())
		return
	}
	defer db.CloseClient(mongoClient)
	// 从profile中读取args
	profile := db.NewProfile(workspaceId, mongoClient)
	profileDoc, err := profile.Get(mainTaskInfoData.ProfileId)
	if err != nil {
		c.FailedStatus(err.Error())
		return
	}
	if profileDoc.Id.Hex() != mainTaskInfoData.ProfileId {
		c.FailedStatus("profile id 错误！")
		return
	}
	doc.WorkspaceId = workspaceId
	doc.Args = profileDoc.Args
	doc.TaskId = uuid.New().String()
	doc.Status = core.CREATED
	// 保存到数据库
	if c.CheckErrorAndStatus(db.NewMainTask(mongoClient).Insert(doc)) {
		if doc.IsCron {
			_ = core.SetCronTaskUpdateFlag("true")
		}
	}

	return
}

func processMainTaskInfoData(data MainTaskInfoData) (mainTask db.MainTaskDocument, err error) {
	mainTask.TaskName = data.Name
	mainTask.ProfileName = data.ProfileName
	mainTask.Description = data.Description
	mainTask.Target = data.Target
	mainTask.ExcludeTarget = data.ExcludeTarget
	mainTask.TargetSliceType = data.TargetSplit
	mainTask.TargetSliceNum = data.TargetSplitNum
	mainTask.OrgId = data.OrgId
	mainTask.IsCron = data.IsCronTask
	if mainTask.IsCron {
		mainTask.CronTaskInfo = &db.CronTask{}
		mainTask.CronTaskInfo.CronExpr = data.CronExpr
		mainTask.CronTaskInfo.CronEnabled = true
	}
	mainTask.IsProxy = data.IsProxy
	return
}

func (c *MainTaskController) GetTaskProfileListAction() {
	defer func(c *MainTaskController, encoding ...bool) {
		_ = c.ServeJSON(encoding...)
	}(c)

	workspaceId := c.GetWorkspace()
	if len(workspaceId) == 0 {
		c.FailedStatus("未选择当前的工作空间！")
		return
	}
	mongoClient, err := db.GetClient()
	if err != nil {
		logging.RuntimeLog.Error(err)
		c.FailedStatus(err.Error())
		return
	}
	defer db.CloseClient(mongoClient)

	profile := db.NewProfile(workspaceId, mongoClient)
	filter := bson.M{"status": "enable"}
	results, err := profile.Find(filter, 0, 0)
	if err != nil {
		logging.RuntimeLog.Error(err)
		c.FailedStatus(err.Error())
		return
	}
	type profileItem struct {
		Id   string `json:"id"`
		Name string `json:"name"`
	}
	var items []profileItem
	for _, row := range results {
		item := profileItem{
			Id:   row.Id.Hex(),
			Name: row.ProfileName,
		}
		items = append(items, item)
	}
	c.Data["json"] = items
	return
}

func (c *MainTaskController) GetProfileInfoAction() {
	defer func(c *MainTaskController, encoding ...bool) {
		_ = c.ServeJSON(encoding...)
	}(c)
	if c.CheckMultiAccessRequest([]RequestRole{SuperAdmin, Admin}, false) == false {
		c.FailedStatus("当前用户权限不允许！")
		return
	}
	id := c.GetString("id")
	if len(id) == 0 {
		logging.RuntimeLog.Error("empty id")
		c.FailedStatus("empty id")
		return
	}
	workspaceId := c.GetWorkspace()
	if len(workspaceId) == 0 {
		c.FailedStatus("未选择当前的工作空间！")
		return
	}

	profileInfo, err := getProfileInfoData(id, workspaceId)
	if err != nil {
		c.FailedStatus(err.Error())
		return
	}
	c.Data["json"] = profileInfo
	return
}

// DeleteAction 删除一条记录
func (c *MainTaskController) DeleteAction() {
	defer func(c *MainTaskController, encoding ...bool) {
		_ = c.ServeJSON(encoding...)
	}(c)
	if c.CheckMultiAccessRequest([]RequestRole{SuperAdmin, Admin}, false) == false {
		c.FailedStatus("当前用户权限不允许！")
		return
	}
	workspaceId := c.GetWorkspace()
	if len(workspaceId) == 0 {
		c.FailedStatus("未选择当前的工作空间！")
		return
	}
	id := c.GetString("id")
	if len(id) == 0 {
		logging.RuntimeLog.Error("empty id")
		c.FailedStatus("empty id")
		return
	}
	mongoClient, err := db.GetClient()
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		c.FailedStatus(err.Error())
		return
	}
	defer db.CloseClient(mongoClient)
	// 删除记录，只能删除当前指定的工作空间的记录以防止越权删除
	mainTask := db.NewMainTask(mongoClient)
	doc, err := mainTask.GetById(id)
	if err != nil {
		c.FailedStatus(err.Error())
		return
	}
	if doc.WorkspaceId != workspaceId {
		c.FailedStatus("该任务不属于当前工作空间！")
		return
	}
	if c.CheckErrorAndStatus(mainTask.Delete(bson.M{"_id": doc.Id})) {
		// 更新定时任务
		if doc.IsCron {
			_ = core.SetCronTaskUpdateFlag("true")
		}
		// 删除executor任务
		if _, err = db.NewExecutorTask(mongoClient).DeleteByMainTaskId(doc.TaskId); err != nil {
			logging.RuntimeLog.Error(err)
		}
		// 删除相关的任务资产结果
		if _, err = db.NewAsset(workspaceId, db.TaskAsset, doc.TaskId, mongoClient).DeleteByTaskId(doc.TaskId); err != nil {
			logging.RuntimeLog.Error(err)
		}
	}

	return
}

func (c *MainTaskController) InfoIndexAction() {
	c.CheckMultiAccessRequest([]RequestRole{SuperAdmin, Admin}, true)

	c.Layout = "base.html"
	c.TplName = "maintask-info.html"

	return
}

func (c *MainTaskController) InfoAction() {
	defer func(c *MainTaskController, encoding ...bool) {
		_ = c.ServeJSON(encoding...)
	}(c)
	if c.CheckMultiAccessRequest([]RequestRole{SuperAdmin, Admin}, false) == false {
		c.FailedStatus("当前用户权限不允许！")
		return
	}
	id := c.GetString("id")
	if len(id) == 0 {
		logging.RuntimeLog.Error("empty id")
		c.FailedStatus("empty id")
		return
	}
	workspaceId := c.GetWorkspace()
	if len(workspaceId) == 0 {
		c.FailedStatus("未选择当前的工作空间！")
		return
	}
	mongoClient, err := db.GetClient()
	if err != nil {
		logging.RuntimeLog.Error(err)
		c.FailedStatus(err.Error())
		return
	}
	defer db.CloseClient(mongoClient)
	mainTask := db.NewMainTask(mongoClient)
	doc, err := mainTask.GetById(id)
	if err != nil {
		logging.RuntimeLog.Error(err)
		c.FailedStatus(err.Error())
		return
	}
	if doc.WorkspaceId != workspaceId {
		c.FailedStatus("该任务不属于当前工作空间！")
		return
	}

	var infoData MainTaskDetailInfoData
	infoData.TaskId = doc.TaskId
	infoData.Name = doc.TaskName
	infoData.ProfileName = doc.ProfileName
	infoData.Description = doc.Description
	infoData.Target = doc.Target
	infoData.ExcludeTarget = doc.ExcludeTarget
	infoData.IsCronTask = doc.IsCron
	if doc.TargetSliceType == 0 {
		infoData.TargetSplit = "不拆分"
	} else if doc.TargetSliceType == 1 {
		infoData.TargetSplit = fmt.Sprintf("按行每%d个拆分", doc.TargetSliceNum)
	} else if doc.TargetSliceType == 2 {
		infoData.TargetSplit = fmt.Sprintf("按IP每%d个拆分", doc.TargetSliceNum)
	} else {
		infoData.TargetSplit = "错误的拆分类型"
	}
	if doc.IsCron && doc.CronTaskInfo != nil {
		infoData.CronTask = fmt.Sprintf("是：%s", doc.CronTaskInfo.CronExpr)
	} else {
		infoData.CronTask = "否"
	}
	if doc.IsProxy {
		infoData.Proxy = "是"
	} else {
		infoData.Proxy = "否"
	}
	infoData.Args = doc.Args
	infoData.Status = doc.Status
	infoData.Result = doc.Result
	infoData.Progress = doc.Progress
	infoData.ProgressRate = fmt.Sprintf("%d", int(doc.ProgressRate*100.0)) + "%"
	infoData.CreateDatetime = FormatDateTime(doc.CreateTime)
	if doc.StartTime != nil {
		infoData.StartDatetime = FormatDateTime(*doc.StartTime)
	}
	if doc.EndTime != nil {
		infoData.EndDatetime = FormatDateTime(*doc.EndTime)
	}
	if doc.Status == core.SUCCESS || doc.Status == core.FAILURE {
		if doc.StartTime != nil && doc.EndTime != nil {
			infoData.Runtime = doc.EndTime.Sub(*doc.StartTime).Truncate(time.Second).String()
		}
	}
	infoData.CreateDatetime = FormatDateTime(doc.CreateTime)

	if doc.OrgId != "" {
		org := db.NewOrg(workspaceId, mongoClient)
		orgDoc, err := org.Get(doc.OrgId)
		if err != nil {
			logging.RuntimeLog.Error(err)
		} else {
			infoData.OrgName = orgDoc.Name
		}
	}
	if doc.WorkspaceId != "" {
		workspace := db.NewWorkspace(mongoClient)
		workspaceDoc, err := workspace.GetByWorkspaceId(doc.WorkspaceId)
		if err != nil {
			logging.RuntimeLog.Error(err)
		} else {
			infoData.WorkspaceName = workspaceDoc.Name
		}
	}
	c.Data["json"] = infoData
	return
}

func (c *MainTaskController) InfoExecutorTaskAction() {
	defer func(c *MainTaskController, encoding ...bool) {
		_ = c.ServeJSON(encoding...)
	}(c)
	if c.CheckMultiAccessRequest([]RequestRole{SuperAdmin, Admin}, false) == false {
		c.FailedStatus("当前用户权限不允许！")
		return
	}
	maintaskId := c.GetString("maintaskId")
	if len(maintaskId) == 0 {
		logging.RuntimeLog.Error("empty id")
		c.FailedStatus("empty id")
		return
	}
	taskStatus := c.GetString("taskStatus")
	workspaceId := c.GetWorkspace()
	if len(workspaceId) == 0 {
		c.FailedStatus("未选择当前的工作空间！")
		return
	}
	mongoClient, err := db.GetClient()
	if err != nil {
		logging.RuntimeLog.Error(err)
		c.FailedStatus(err.Error())
		return
	}
	defer db.CloseClient(mongoClient)

	executorTask := db.NewExecutorTask(mongoClient)
	filter := bson.M{"mainTaskId": maintaskId, "workspaceId": workspaceId}
	if len(taskStatus) > 0 {
		filter["status"] = taskStatus
	}
	results, err := executorTask.Find(filter, 0, 0)
	if err != nil {
		logging.RuntimeLog.Error(err)
		c.FailedStatus(err.Error())
		return
	}
	var infoData []ExecutorTaskInfoData
	for _, row := range results {
		item := ExecutorTaskInfoData{
			Id:             row.Id.Hex(),
			Executor:       row.Executor,
			Target:         row.Target,
			Worker:         row.Worker,
			Status:         row.Status,
			Result:         row.Result,
			Args:           row.Args,
			CreateDatetime: FormatDateTime(row.CreateTime),
			UpdateDatetime: FormatDateTime(row.UpdateTime),
		}
		if row.StartTime != nil {
			item.StartDatetime = FormatDateTime(*row.StartTime)
		}
		if row.EndTime != nil {
			item.EndDatetime = FormatDateTime(*row.EndTime)
		}
		if row.Status == core.SUCCESS || row.Status == core.FAILURE {
			if row.StartTime != nil && row.EndTime != nil {
				item.Runtime = row.EndTime.Sub(*row.StartTime).Truncate(time.Second).String()
			}
		}
		infoData = append(infoData, item)
	}
	c.Data["json"] = infoData
	return
}

// DeleteExecutorTaskAction 删除一条记录
func (c *MainTaskController) DeleteExecutorTaskAction() {
	defer func(c *MainTaskController, encoding ...bool) {
		_ = c.ServeJSON(encoding...)
	}(c)
	if c.CheckMultiAccessRequest([]RequestRole{SuperAdmin, Admin}, false) == false {
		c.FailedStatus("当前用户权限不允许！")
		return
	}

	workspaceId := c.GetWorkspace()
	if len(workspaceId) == 0 {
		c.FailedStatus("未选择当前的工作空间！")
		return
	}
	id := c.GetString("id")
	if len(id) == 0 {
		logging.RuntimeLog.Error("empty id")
		c.FailedStatus("empty id")
		return
	}
	mongoClient, err := db.GetClient()
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		c.FailedStatus(err.Error())
		return
	}
	defer db.CloseClient(mongoClient)
	// 删除记录，只能删除当前指定的工作空间的记录以防止越权删除
	executorTask := db.NewExecutorTask(mongoClient)
	doc, err := executorTask.GetById(id)
	if err != nil {
		c.FailedStatus(err.Error())
		return
	}
	if doc.WorkspaceId != workspaceId {
		c.FailedStatus("该任务不属于当前工作空间！")
		return
	}
	_ = c.CheckErrorAndStatus(executorTask.Delete(id))

	return
}

// RedoMaintaskAction 重新执行一次任务
func (c *MainTaskController) RedoMaintaskAction() {
	defer func(c *MainTaskController, encoding ...bool) {
		_ = c.ServeJSON(encoding...)
	}(c)
	if c.CheckMultiAccessRequest([]RequestRole{SuperAdmin, Admin}, false) == false {
		c.FailedStatus("当前用户权限不允许！")
		return
	}

	workspaceId := c.GetWorkspace()
	if len(workspaceId) == 0 {
		c.FailedStatus("未选择当前的工作空间！")
		return
	}
	id := c.GetString("id")
	if len(id) == 0 {
		logging.RuntimeLog.Error("empty id")
		c.FailedStatus("empty id")
		return
	}
	mongoClient, err := db.GetClient()
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		c.FailedStatus(err.Error())
		return
	}
	defer db.CloseClient(mongoClient)

	mainTask := db.NewMainTask(mongoClient)
	doc, err := mainTask.GetById(id)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		c.FailedStatus(err.Error())
		return
	}
	if doc.Id.Hex() != id {
		c.FailedStatus("该任务不存在！")
		return
	}
	docNew := db.MainTaskDocument{
		WorkspaceId:     doc.WorkspaceId,
		TaskId:          uuid.New().String(),
		TaskName:        doc.TaskName,
		Description:     "(来自重做任务：" + doc.TaskName + ")",
		Target:          doc.Target,
		ExcludeTarget:   doc.ExcludeTarget,
		ProfileName:     doc.ProfileName,
		OrgId:           doc.OrgId,
		IsProxy:         doc.IsProxy,
		Args:            doc.Args,
		Status:          core.CREATED,
		TargetSliceType: doc.TargetSliceType,
		TargetSliceNum:  doc.TargetSliceNum,
		IsCron:          doc.IsCron,
		CronTaskInfo:    doc.CronTaskInfo,
	}
	c.CheckErrorAndStatus(mainTask.Insert(docNew))

	return
}

// RedoExecutorTaskAction 重新执行一Executor任务
func (c *MainTaskController) RedoExecutorTaskAction() {
	defer func(c *MainTaskController, encoding ...bool) {
		_ = c.ServeJSON(encoding...)
	}(c)
	if c.CheckMultiAccessRequest([]RequestRole{SuperAdmin, Admin}, false) == false {
		c.FailedStatus("当前用户权限不允许！")
		return
	}

	workspaceId := c.GetWorkspace()
	if len(workspaceId) == 0 {
		c.FailedStatus("未选择当前的工作空间！")
		return
	}
	id := c.GetString("id")
	if len(id) == 0 {
		logging.RuntimeLog.Error("empty id")
		c.FailedStatus("empty id")
		return
	}
	mongoClient, err := db.GetClient()
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		c.FailedStatus(err.Error())
		return
	}
	defer db.CloseClient(mongoClient)

	executorTask := db.NewExecutorTask(mongoClient)
	executorDoc, err := executorTask.GetById(id)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		c.FailedStatus(err.Error())
		return
	}
	if executorDoc.Id.Hex() != id {
		c.FailedStatus("该任务不存在！")
		return
	}
	mainTask := db.NewMainTask(mongoClient)
	mainTaskDoc, err := mainTask.GetByTaskId(executorDoc.MainTaskId)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		c.FailedStatus(err.Error())
		return
	}
	if mainTaskDoc.WorkspaceId != workspaceId {
		c.FailedStatus("Maintask任务不存在或不属于当前工作空间！")
		return
	}
	if mainTaskDoc.Status == core.SUCCESS || mainTaskDoc.Status == core.FAILURE {
		c.FailedStatus("Maintask任务已执行完毕，无法重新执行！")
		return
	}
	var executorConfig execute.ExecutorConfig
	err = json.Unmarshal([]byte(mainTaskDoc.Args), &executorConfig)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		c.FailedStatus("参数解析失败！")
		return
	}
	executorTaskInfoNew := execute.ExecutorTaskInfo{
		Executor:  executorDoc.Executor,
		TaskId:    uuid.New().String(),
		PreTaskId: executorDoc.PreTaskId,
		MainTaskInfo: execute.MainTaskInfo{
			Target:          executorDoc.Target,
			ExcludeTarget:   mainTaskDoc.ExcludeTarget,
			ExecutorConfig:  executorConfig,
			OrgId:           mainTaskDoc.OrgId,
			WorkspaceId:     mainTaskDoc.WorkspaceId,
			MainTaskId:      mainTaskDoc.TaskId,
			IsProxy:         mainTaskDoc.IsProxy,
			TargetSliceType: mainTaskDoc.TargetSliceType,
			TargetSliceNum:  mainTaskDoc.TargetSliceNum,
		},
	}
	err = core.NewExecutorTask(executorTaskInfoNew)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		c.FailedStatus("重新执行任务失败！")
		return
	}
	c.SucceededStatus("重新执行任务成功！")
	return
}

// ChangeCronTaskStatusAction 更改定时任务状态
func (c *MainTaskController) ChangeCronTaskStatusAction() {
	defer func(c *MainTaskController, encoding ...bool) {
		_ = c.ServeJSON(encoding...)
	}(c)
	if c.CheckMultiAccessRequest([]RequestRole{SuperAdmin, Admin}, false) == false {
		c.FailedStatus("当前用户权限不允许！")
		return
	}

	workspaceId := c.GetWorkspace()
	if len(workspaceId) == 0 {
		c.FailedStatus("未选择当前的工作空间！")
		return
	}
	id := c.GetString("id")
	if len(id) == 0 {
		logging.RuntimeLog.Error("empty id")
		c.FailedStatus("empty id")
		return
	}
	mongoClient, err := db.GetClient()
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		c.FailedStatus(err.Error())
		return
	}
	defer db.CloseClient(mongoClient)

	mainTask := db.NewMainTask(mongoClient)
	mainTaskDoc, err := mainTask.GetById(id)
	if err != nil {
		c.FailedStatus(err.Error())
		return
	}
	if mainTaskDoc.WorkspaceId != workspaceId {
		c.FailedStatus("该任务不属于当前工作空间！")
		return
	}
	if mainTaskDoc.IsCron == false {
		c.FailedStatus("该任务不是定时任务！")
		return
	}
	if mainTaskDoc.CronTaskInfo == nil {
		c.FailedStatus("该任务的定时任务信息不存在！")
		return
	}
	if mainTaskDoc.CronTaskInfo.CronEnabled == false {
		mainTaskDoc.CronTaskInfo.CronEnabled = true
	} else {
		mainTaskDoc.CronTaskInfo.CronEnabled = false
	}
	updateDoc := bson.M{"cronTaskInfo.enabled": mainTaskDoc.CronTaskInfo.CronEnabled}
	if c.CheckErrorAndStatus(mainTask.Update(id, updateDoc)) {
		_ = core.SetCronTaskUpdateFlag("true")
	}

	return
}

func (c *MainTaskController) MaintaskTreeAction() {
	c.CheckOneAccessRequest(SuperAdmin, true)

	c.Layout = "base.html"
	c.TplName = "maintask-tree.html"
}

func (c *MainTaskController) MaintaskTreeDataAction() {
	defer func(c *MainTaskController, encoding ...bool) {
		_ = c.ServeJSON(encoding...)
	}(c)
	if c.CheckOneAccessRequest(SuperAdmin, false) == false {
		c.FailedStatus("当前用户权限不允许！")
		return
	}

	workspaceId := c.GetWorkspace()
	if len(workspaceId) == 0 {
		c.FailedStatus("未选择当前的工作空间！")
		return
	}
	maintaskId := c.GetString("maintaskId")
	if len(maintaskId) == 0 {
		logging.RuntimeLog.Error("empty id")
		c.FailedStatus("empty id")
		return
	}

	mongoClient, err := db.GetClient()
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		c.FailedStatus(err.Error())
		return
	}
	defer db.CloseClient(mongoClient)
	executorTask := db.NewExecutorTask(mongoClient)
	treeData, err := executorTask.GetTaskChainByMainTaskID(maintaskId)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		c.FailedStatus(err.Error())
		return
	}
	if len(treeData.ApexTree) > 0 {
		c.Data["json"] = treeData.ApexTree[0]
	} else {
		c.FailedStatus("无子任务流程节点")
	}
	return
}
