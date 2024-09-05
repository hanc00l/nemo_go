package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/hanc00l/nemo_go/v2/pkg/comm"
	"github.com/hanc00l/nemo_go/v2/pkg/logging"
	"github.com/hanc00l/nemo_go/v2/pkg/utils"
	"strings"
	"time"
)

type WorkerController struct {
	BaseController
}

type WorkerStatusData struct {
	Index                    int    `json:"index"`
	WorkName                 string `json:"worker_name"`
	WorkerTopic              string `json:"worker_topic"`
	CreateTime               string `json:"create_time"`
	UpdateTime               string `json:"update_time"`
	TaskExecutedNumber       int    `json:"task_number"`
	TaskStartedNumber        int    `json:"started_number"`
	CPULoad                  string `json:"cpu_load"`
	MemUsage                 string `json:"mem_used"`
	EnableManualReloadFlag   bool   `json:"enable_manual_reload_flag"`
	EnableManualFileSyncFlag bool   `json:"enable_manual_file_sync_flag"`
	HeartColor               string `json:"heart_color"`
	IsDaemonProcess          bool   `json:"daemon_process"`
}

func (c *WorkerController) IndexAction() {
	c.CheckOneAccessRequest(SuperAdmin, true)
	c.Layout = "base.html"
	c.TplName = "worker-list.html"
}

// ListAction 列表的数据
func (c *WorkerController) ListAction() {
	c.CheckOneAccessRequest(SuperAdmin, true)
	defer c.ServeJSON()

	req := DatableRequestParam{}
	err := c.ParseForm(&req)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
	}
	resp := getWorkerAliveList(&req)
	c.Data["json"] = resp
}

// ManualReloadWorkerAction 重启worker
func (c *WorkerController) ManualReloadWorkerAction() {
	defer c.ServeJSON()
	if c.CheckMultiAccessRequest([]RequestRole{SuperAdmin, Admin}, false) == false {
		c.FailedStatus("当前用户权限不允许！")
		return
	}

	worker := c.GetString("worker_name")
	if worker == "" {
		c.FailedStatus("worker name is empty")
		return
	}
	comm.WorkerStatusMutex.Lock()
	defer comm.WorkerStatusMutex.Unlock()

	if _, ok := comm.WorkerStatus[worker]; ok {
		comm.WorkerStatus[worker].ManualReloadFlag = true
		c.SucceededStatus("已设置worker重启标志，等待worker的daemon进程执行！")
	} else {
		c.FailedStatus("无效的worker name")
	}

}

// EditWorkerAction 更改worker
func (c *WorkerController) EditWorkerAction() {
	defer c.ServeJSON()
	if c.CheckMultiAccessRequest([]RequestRole{SuperAdmin, Admin}, false) == false {
		c.FailedStatus("当前用户权限不允许！")
		return
	}

	worker := c.GetString("worker_name")
	if worker == "" {
		c.FailedStatus("worker name is empty")
		return
	}
	comm.WorkerStatusMutex.Lock()
	defer comm.WorkerStatusMutex.Unlock()

	if _, ok := comm.WorkerStatus[worker]; ok {
		var option comm.WorkerOption
		var err error
		err = json.Unmarshal(comm.WorkerStatus[worker].WorkerRunOption, &option)
		if err != nil {
			logging.RuntimeLog.Error(err.Error())
			c.FailedStatus(err.Error())
			return
		}
		c.Data["json"] = option
	} else {
		logging.RuntimeLog.Errorf("无效的worker name:%s", worker)
		c.FailedStatus("无效的worker name")
		return
	}
	return
}

// UpdateWorkerAction 更新worker的启动参数
func (c *WorkerController) UpdateWorkerAction() {
	defer c.ServeJSON()
	if c.CheckMultiAccessRequest([]RequestRole{SuperAdmin, Admin}, false) == false {
		c.FailedStatus("当前用户权限不允许！")
		return
	}
	worker := c.GetString("worker_name")
	if worker == "" {
		c.FailedStatus("worker name is empty")
		return
	}
	req := comm.WorkerOption{}
	if err := c.ParseForm(&req); err != nil {
		c.FailedStatus(err.Error())
		return
	}
	comm.WorkerStatusMutex.Lock()
	defer comm.WorkerStatusMutex.Unlock()

	if _, ok := comm.WorkerStatus[worker]; ok {
		comm.WorkerStatus[worker].ManualUpdateOptionFlag = true
		comm.WorkerStatus[worker].WorkerUpdateOption, _ = json.Marshal(req)
		c.SucceededStatus("已设置Worker更新标志，等待Worker的daemon进程执行！")
	} else {
		logging.RuntimeLog.Errorf("无效的worker name:%s", worker)
		c.FailedStatus("无效的worker name")
		return
	}
	return
}

// getWorkerTopicDescription 获取worker的任务类型的中文描述
func getWorkerTopicDescription(taskMode string) string {
	var modeNameMap = map[string]string{
		"default": "全部任务",
		"active":  "主动扫描",
		"finger":  "指纹识别",
		"passive": "被动收集",
		"pocscan": "漏洞验证",
		"custom":  "自定义任务",
	}
	workerTopicDescription := make(map[string]struct{})
	topicsArray := strings.Split(taskMode, ",")
	if len(topicsArray) >= 4 {
		return modeNameMap["default"]
	}
	for _, v := range topicsArray {
		if modeName, ok := modeNameMap[v]; ok {
			workerTopicDescription[modeName] = struct{}{}
		} else if strings.HasPrefix(v, "custom") {
			workerTopicDescription[modeNameMap["custom"]] = struct{}{}
		} else {
			return "未知模式"
		}
	}
	return utils.SetToString(workerTopicDescription)
}

func getWorkerAliveList(req *DatableRequestParam) (resp DataTableResponseData) {
	comm.WorkerStatusMutex.Lock()
	defer comm.WorkerStatusMutex.Unlock()
	// 对worker按更新时间排序
	taskAliveUpdateList := make(map[string]int)
	for _, v := range comm.WorkerStatus {
		// 只显示最近5分钟有心跳的worker（worker失去同步并清除信息的操作由同步的时候进行检查）
		if time.Now().Sub(v.UpdateTime).Minutes() > 5 {
			continue
		}
		seconds := time.Now().Sub(v.CreateTime).Seconds()
		taskAliveUpdateList[v.WorkerName] = int(seconds)
	}
	index := 1
	sortedWorkers := utils.SortMapByValue(taskAliveUpdateList, false)
	for _, w := range sortedWorkers {
		if v, ok := comm.WorkerStatus[w.Key]; ok {
			wsd := WorkerStatusData{
				Index:              index,
				WorkName:           v.WorkerName,
				WorkerTopic:        getWorkerTopicDescription(v.WorkerTopics),
				CreateTime:         FormatDateTime(v.CreateTime),
				UpdateTime:         fmt.Sprintf("%s前", time.Now().Sub(v.UpdateTime).Truncate(time.Second).String()),
				TaskExecutedNumber: v.TaskExecutedNumber,
				TaskStartedNumber:  v.TaskStartedNumber,
				HeartColor:         "green",
				CPULoad:            v.CPULoad,
				MemUsage:           v.MemUsed,
				IsDaemonProcess:    v.IsDaemonProcess,
			}
			workerHeartDt := time.Now().Sub(v.UpdateTime).Minutes()
			if v.IsDaemonProcess {
				wsd.EnableManualReloadFlag = true
			}
			if workerHeartDt >= 1 && workerHeartDt < 3 {
				wsd.HeartColor = "yellow"
			} else if workerHeartDt >= 3 {
				wsd.HeartColor = "red"
			}
			resp.Data = append(resp.Data, wsd)
			index++
		}
	}
	if len(resp.Data) == 0 {
		resp.Data = make([]interface{}, 0)
	}
	resp.Draw = req.Draw
	resp.RecordsTotal = len(comm.WorkerStatus)
	resp.RecordsFiltered = len(comm.WorkerStatus)

	return
}
