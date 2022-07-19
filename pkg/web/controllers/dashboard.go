package controllers

import (
	"fmt"
	"github.com/hanc00l/nemo_go/pkg/comm"
	"github.com/hanc00l/nemo_go/pkg/db"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/ampq"
	"github.com/hanc00l/nemo_go/pkg/task/custom"
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
}

type WorkerStatusData struct {
	Index              int    `json:"index"`
	WorkName           string `json:"worker_name"`
	CreateTime         string `json:"create_time"`
	UpdateTime         string `json:"update_time"`
	TaskExecutedNumber int    `json:"task_number"`
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
	c.UpdateOnlineUser()
	c.Layout = "base.html"
	c.TplName = "dashboard.html"
}

// GetStatisticDataAction 获取统计信息
func (c *DashboardController) GetStatisticDataAction() {
	defer c.ServeJSON()

	searchMap := make(map[string]interface{})
	ip := &db.Ip{}
	domain := &db.Domain{}
	vul := &db.Vulnerability{}
	task := &db.Task{}
	data := DashboardStatisticData{
		IP:            ip.Count(searchMap),
		Domain:        domain.Count(searchMap),
		Vulnerability: vul.Count(searchMap),
	}
	searchMapTask := make(map[string]interface{})
	searchMapTask["state"] = ampq.STARTED
	data.ActiveTask = task.Count(searchMapTask)
	c.Data["json"] = data
}

// GetTaskInfoAction 获取任务数据
func (c *DashboardController) GetTaskInfoAction() {
	c.UpdateOnlineUser()
	defer c.ServeJSON()

	searchMapActivated := make(map[string]interface{})
	searchMapActivated["state"] = ampq.STARTED
	searchMapCreated := make(map[string]interface{})
	searchMapCreated["state"] = ampq.CREATED
	searchMapALL := make(map[string]interface{})
	searchMapALL["date_delta"] = 7
	task := &db.Task{}
	data := TaskInfoData{}
	data.TaskInfo = fmt.Sprintf("%d/%d/%d", task.Count(searchMapActivated), task.Count(searchMapCreated), task.Count(searchMapALL))
	c.Data["json"] = data
}

// GetStartedTaskInfoAction 获取正在执行的任务数据
func (c *DashboardController) GetStartedTaskInfoAction() {
	c.UpdateOnlineUser()
	defer c.ServeJSON()

	searchMapActivated := make(map[string]interface{})
	searchMapActivated["state"] = ampq.STARTED
	task := &db.Task{}
	var tis []StartedTaskInfo
	rows, _ := task.Gets(searchMapActivated, 1, 10)
	for _, row := range rows {
		ti := StartedTaskInfo{
			TaskName:     row.TaskName,
			TaskArgs:     ParseTargetFromKwArgs(row.KwArgs),
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
		logging.RuntimeLog.Error(err.Error())
	}
	index := 1
	resp := DataTableResponseData{}

	comm.WorkerStatusMutex.Lock()
	for _, v := range comm.WorkerStatus {
		if time.Now().Sub(v.UpdateTime).Minutes() > 5 {
			delete(comm.WorkerStatus, v.WorkerName)
			continue
		}
		resp.Data = append(resp.Data, WorkerStatusData{
			Index:              index,
			WorkName:           v.WorkerName,
			CreateTime:         FormatDateTime(v.CreateTime),
			UpdateTime:         fmt.Sprintf("%s前", time.Now().Sub(v.UpdateTime).Truncate(time.Second).String()),
			TaskExecutedNumber: v.TaskExecutedNumber,
		})
		index++
	}
	comm.WorkerStatusMutex.Unlock()

	resp.Draw = req.Draw
	resp.RecordsTotal = len(comm.WorkerStatus)
	resp.RecordsFiltered = len(comm.WorkerStatus)
	c.Data["json"] = resp
}

// OnlineUserListAction 获取在线用户数据，用于Dashboard表表显示
func (c *DashboardController) OnlineUserListAction() {
	defer c.ServeJSON()

	req := DatableRequestParam{}
	err := c.ParseForm(&req)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
	}
	index := 1
	resp := DataTableResponseData{}
	OnlineUserMutex.Lock()
	defer OnlineUserMutex.Unlock()
	ipLocation := custom.NewIPLocation()

	for _, v := range OnlineUser {
		if time.Now().Sub(v.UpdateTime).Hours() > 24 {
			delete(OnlineUser, v.IP)
			continue
		}
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

	resp.Draw = req.Draw
	resp.RecordsTotal = len(OnlineUser)
	resp.RecordsFiltered = len(OnlineUser)
	c.Data["json"] = resp
}
