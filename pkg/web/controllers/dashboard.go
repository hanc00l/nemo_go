package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/hanc00l/nemo_go/pkg/comm"
	"github.com/hanc00l/nemo_go/pkg/db"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/asynctask"
	"sync"
	"time"
)

type DashboardController struct {
	BaseController
	workerStatusMutex sync.Mutex
	WorkerStatus map[string]*asynctask.WorkerStatus
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

// IndexAction dashboard首页
func (c *DashboardController) IndexAction() {
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
	searchMapTask["state"] = asynctask.STARTED
	data.ActiveTask = task.Count(searchMapTask)
	c.Data["json"] = data
}

// GetTaskInfoAction 获取任务数据
func (c *DashboardController) GetTaskInfoAction() {
	defer c.ServeJSON()

	searchMapActivated := make(map[string]interface{})
	searchMapActivated["state"] = asynctask.STARTED
	searchMapALL := make(map[string]interface{})
	searchMapALL["date_delta"] = 7
	task := &db.Task{}
	data := TaskInfoData{}
	data.TaskInfo = fmt.Sprintf("%d/%d", task.Count(searchMapActivated), task.Count(searchMapALL))
	c.Data["json"] = data
}

// WorkerAliveAction worker keep alive
func (c *DashboardController) WorkerAliveAction() {
	var rData *StatusResponseData
	defer func() {
		r, _ := json.Marshal(rData)
		c.writeByteContent(comm.EncryptData(r))
	}()

	requestData := comm.DecryptData(c.Ctx.Input.RequestBody)
	var kai comm.KeepAliveInfo
	err := json.Unmarshal(requestData, &kai)
	if err != nil {
		logging.RuntimeLog.Error(err)
		rData = &StatusResponseData{Status: Fail, Msg: err.Error()}
		return
	}
	if kai.WorkerStatus.WorkerName == "" {
		rData = &StatusResponseData{Status: Fail, Msg: "no worker name"}
		return
	}
	c.workerStatusMutex.Lock()
	c.WorkerStatus[kai.WorkerStatus.WorkerName] = &kai.WorkerStatus
	c.WorkerStatus[kai.WorkerStatus.WorkerName].UpdateTime = time.Now()
	c.workerStatusMutex.Unlock()

	responseData := comm.NewKeepAliveResponseInfo(kai.CustomFiles)
	rData = &StatusResponseData{Status: Success, Msg: string(responseData)}
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

	c.workerStatusMutex.Lock()
	for _, v := range c.WorkerStatus {
		if time.Now().Sub(v.UpdateTime).Minutes() > 5 {
			delete(c.WorkerStatus, v.WorkerName)
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
	c.workerStatusMutex.Unlock()

	resp.Draw = req.Draw
	resp.RecordsTotal = len(c.WorkerStatus)
	resp.RecordsFiltered = len(c.WorkerStatus)
	c.Data["json"] = resp
}
