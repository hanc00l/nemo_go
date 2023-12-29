package controllers

import (
	"fmt"
	"github.com/hanc00l/nemo_go/pkg/comm"
	"github.com/hanc00l/nemo_go/pkg/db"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/ampq"
	"github.com/hanc00l/nemo_go/pkg/task/custom"
	"github.com/hanc00l/nemo_go/pkg/task/runner"
	"github.com/hanc00l/nemo_go/pkg/utils"
	"time"
)

type DashboardController struct {
	BaseController
}

type DashboardStatisticData struct {
	IP            int `json:"ip_count"`
	Domain        int `json:"domain_count"`
	Vulnerability int `json:"vulnerability_count"`
	ActiveTask    int `json:"task_active"`
	Worker        int `json:"worker_count"`
}

type TaskInfoData struct {
	TaskInfo string `json:"task_info"`
}

type StartedTaskInfo struct {
	TaskName     string `json:"task_name"`
	TaskArgs     string `json:"task_args"`
	TaskStarting string `json:"task_starting"`
}

type OnlineUserInfoData struct {
	Index        int    `json:"index"`
	IP           string `json:"ip"`
	LoginTime    string `json:"login_time"`
	UpdateTime   string `json:"update_time"`
	UpdateNumber int64  `json:"update_number"`
}

// IndexAction dashboard首页
func (c *DashboardController) IndexAction() {
	c.Layout = "base.html"
	c.TplName = "dashboard.html"
}

// GetStatisticDataAction 获取统计信息
func (c *DashboardController) GetStatisticDataAction() {
	defer c.ServeJSON()

	searchMap := make(map[string]interface{})
	workspaceId := c.GetCurrentWorkspace()
	if workspaceId > 0 {
		searchMap["workspace_id"] = workspaceId
	}
	ip := &db.Ip{}
	domain := &db.Domain{}
	vul := &db.Vulnerability{}
	task := &db.TaskRun{}
	data := DashboardStatisticData{
		IP:            ip.Count(searchMap),
		Domain:        domain.Count(searchMap),
		Vulnerability: vul.Count(searchMap),
	}
	searchMapTask := make(map[string]interface{})
	searchMapTask["state"] = ampq.STARTED
	if workspaceId > 0 {
		searchMapTask["workspace_id"] = workspaceId
	}
	data.ActiveTask = task.Count(searchMapTask)
	//
	comm.WorkerStatusMutex.Lock()
	defer comm.WorkerStatusMutex.Unlock()
	for _, v := range comm.WorkerStatus {
		if time.Now().Sub(v.UpdateTime).Minutes() > 5 {
			delete(comm.WorkerStatus, v.WorkerName)
			continue
		}
	}
	data.Worker = len(comm.WorkerStatus)
	c.Data["json"] = data
}

// GetTaskInfoAction 获取任务数据
func (c *DashboardController) GetTaskInfoAction() {
	defer c.ServeJSON()

	searchMapActivated := make(map[string]interface{})
	searchMapActivated["state"] = ampq.STARTED
	searchMapCreated := make(map[string]interface{})
	searchMapCreated["state"] = ampq.CREATED
	searchMapALL := make(map[string]interface{})
	searchMapALL["date_delta"] = 7

	workspaceId := c.GetCurrentWorkspace()
	if workspaceId > 0 {
		searchMapActivated["workspace_id"] = workspaceId
		searchMapCreated["workspace_id"] = workspaceId
		searchMapALL["workspace_id"] = workspaceId
	}

	task := &db.TaskRun{}
	data := TaskInfoData{}
	data.TaskInfo = fmt.Sprintf("%d/%d/%d", task.Count(searchMapActivated), task.Count(searchMapCreated), task.Count(searchMapALL))
	c.Data["json"] = data
}

// GetStartedTaskInfoAction 获取正在执行的任务数据
func (c *DashboardController) GetStartedTaskInfoAction() {
	defer c.ServeJSON()

	searchMapActivated := make(map[string]interface{})
	workspaceId := c.GetCurrentWorkspace()
	if workspaceId > 0 {
		searchMapActivated["workspace_id"] = workspaceId
	}
	searchMapActivated["state"] = ampq.STARTED
	task := &db.TaskRun{}
	var tis []StartedTaskInfo
	rows, _ := task.Gets(searchMapActivated, 1, 10)
	for _, row := range rows {
		ti := StartedTaskInfo{
			TaskName:     row.TaskName,
			TaskArgs:     runner.ParseTargetFromKwArgs(row.TaskName, row.KwArgs),
			TaskStarting: fmt.Sprintf("%s前", time.Now().Sub(*row.StartedTime).Truncate(time.Second).String()),
		}
		tis = append(tis, ti)
	}
	c.Data["json"] = tis
}

// WorkerAliveListAction 获取worker数据，用于dashboard列表显示
func (c *DashboardController) WorkerAliveListAction() {
	defer c.ServeJSON()

	req := DatableRequestParam{}
	err := c.ParseForm(&req)
	if err != nil {
		logging.RuntimeLog.Error(err)
		logging.CLILog.Error(err)
	}
	resp := getWorkerAliveList(&req)
	c.Data["json"] = resp
}

// OnlineUserListAction 获取在线用户数据，用于Dashboard表表显示
func (c *DashboardController) OnlineUserListAction() {
	defer c.ServeJSON()

	req := DatableRequestParam{}
	err := c.ParseForm(&req)
	if err != nil {
		logging.RuntimeLog.Error(err)
		logging.CLILog.Error(err)
	}
	resp := DataTableResponseData{}
	OnlineUserMutex.Lock()
	defer OnlineUserMutex.Unlock()

	ipLocation := custom.NewIPv4Location()
	onlineUserAliveUpdateList := make(map[string]int)
	for _, v := range OnlineUser {
		if time.Now().Sub(v.UpdateTime).Hours() > 24 {
			delete(OnlineUser, v.IP)
			continue
		}
		seconds := time.Now().Sub(v.UpdateTime).Seconds()
		onlineUserAliveUpdateList[v.IP] = int(seconds)
	}
	sortedOnlineUsers := utils.SortMapByValue(onlineUserAliveUpdateList, false)
	index := 1
	for _, w := range sortedOnlineUsers {
		if v, ok := OnlineUser[w.Key]; ok {
			ipl := ipLocation.FindCustomIP(v.IP)
			if ipl == "" {
				ipl = ipLocation.FindPublicIP(v.IP)
			}
			resp.Data = append(resp.Data, OnlineUserInfoData{
				Index:        index,
				IP:           fmt.Sprintf("%s (%s)", v.IP, ipl),
				LoginTime:    FormatDateTime(v.LoginTime),
				UpdateTime:   fmt.Sprintf("%s前", time.Now().Sub(v.UpdateTime).Truncate(time.Second).String()),
				UpdateNumber: v.UpdateNumber,
			})
			index++
		}
	}

	resp.Draw = req.Draw
	resp.RecordsTotal = len(OnlineUser)
	resp.RecordsFiltered = len(OnlineUser)
	c.Data["json"] = resp
}
