package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hanc00l/nemo_go/v3/pkg/core"
	"github.com/hanc00l/nemo_go/v3/pkg/logging"
	"github.com/hanc00l/nemo_go/v3/pkg/utils"
	"github.com/redis/go-redis/v9"
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
	IsIPv6Support            bool   `json:"ipv6"`
}

func (c *WorkerController) IndexAction() {
	c.CheckMultiAccessRequest([]RequestRole{SuperAdmin, Admin}, true)
	c.Layout = "base.html"
	c.TplName = "worker-list.html"
}

// ListAction 列表的数据
func (c *WorkerController) ListAction() {
	defer func(c *WorkerController, encoding ...bool) {
		_ = c.ServeJSON(encoding...)
	}(c)
	if c.CheckMultiAccessRequest([]RequestRole{SuperAdmin, Admin}, false) == false {
		c.FailedStatus("当前用户权限不允许！")
		return
	}
	req := DatableRequestParam{}
	err := c.ParseForm(&req)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
	}
	resp := getWorkerAliveList(&req)
	c.Data["json"] = resp
}

// EditWorkerAction 更改worker
func (c *WorkerController) EditWorkerAction() {
	defer func(c *WorkerController, encoding ...bool) {
		_ = c.ServeJSON(encoding...)
	}(c)
	if c.CheckMultiAccessRequest([]RequestRole{SuperAdmin, Admin}, false) == false {
		c.FailedStatus("当前用户权限不允许！")
		return
	}

	worker := c.GetString("worker_name")
	if worker == "" {
		c.FailedStatus("worker name is empty")
		return
	}

	rdb, err := core.GetRedisClient()
	if err != nil {
		logging.RuntimeLog.Errorf("get redis client fail:%v", err)
		return
	}
	defer rdb.Close()

	workerAliveStatus, err := core.GetWorkerStatusFromRedis(rdb, worker)
	if errors.Is(err, redis.Nil) {
		c.FailedStatus("worker不存在！")
		return
	} else if err != nil {
		logging.RuntimeLog.Errorf(err.Error())
		return
	}
	var option core.WorkerOption
	err = json.Unmarshal(workerAliveStatus.WorkerRunOption, &option)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		c.FailedStatus(err.Error())
		return
	}
	c.Data["json"] = option

	return
}

func (c *WorkerController) ManualUpdateWorkerAction() {
	c.updateWorkAliveOptions("update")
}

// ManualReloadWorkerAction 重启worker
func (c *WorkerController) ManualReloadWorkerAction() {
	c.updateWorkAliveOptions("reload")
}

func (c *WorkerController) ManualInitWorkerAction() {
	c.updateWorkAliveOptions("init")
}
func (c *WorkerController) ManualSyncWorkerAction() {
	c.updateWorkAliveOptions("sync")
}

func (c *WorkerController) updateWorkAliveOptions(updateType string) {
	defer func(c *WorkerController, encoding ...bool) {
		_ = c.ServeJSON(encoding...)
	}(c)
	if c.CheckMultiAccessRequest([]RequestRole{SuperAdmin, Admin}, false) == false {
		c.FailedStatus("当前用户权限不允许！")
		return
	}
	worker := c.GetString("worker_name")
	if worker == "" {
		c.FailedStatus("worker name is empty")
		return
	}
	req := core.WorkerOption{}
	if err := c.ParseForm(&req); err != nil {
		c.FailedStatus(err.Error())
		return
	}
	rdb, err := core.GetRedisClient()
	if err != nil {
		logging.RuntimeLog.Errorf("get redis client fail:%v", err)
		return
	}
	defer rdb.Close()

	workerAliveStatus, err := core.GetWorkerStatusFromRedis(rdb, worker)
	if errors.Is(err, redis.Nil) {
		c.FailedStatus("worker不存在！")
		return
	} else if err != nil {
		logging.RuntimeLog.Errorf(err.Error())
		return
	}
	if workerAliveStatus.IsDaemonProcess == false {
		c.FailedStatus("worker不是daemon进程！")
		return
	}
	if updateType == "reload" {
		workerAliveStatus.ManualReloadFlag = true
	} else if updateType == "sync" {
		workerAliveStatus.ManualConfigAndPocSyncFlag = true
	} else if updateType == "init" {
		workerAliveStatus.ManualInitEnvFlag = true
	} else if updateType == "update" {
		workerAliveStatus.ManualUpdateOptionFlag = true
	}
	workerAliveStatus.WorkerUpdateOption, err = json.Marshal(req)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		c.FailedStatus(err.Error())
		return
	}
	err = core.SetWorkerStatusToRedis(rdb, worker, workerAliveStatus)
	if err != nil {
		logging.RuntimeLog.Errorf(err.Error())
		c.FailedStatus("设置参数失败！")
		return
	}
	c.SucceededStatus("已设置，等待worker的daemon进程执行！")
	return
}

// getWorkerTopicDescription 获取worker的任务类型的中文描述
func getWorkerTopicDescription(taskMode string) string {
	var modeNameMap = map[string]string{
		"default":    "分布式任务",
		"active":     "主动扫描",
		"finger":     "指纹识别",
		"passive":    "被动收集",
		"pocscan":    "漏洞验证",
		"custom":     "自定义任务",
		"standalone": "独立任务",
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
	defer func() {
		resp.Draw = req.Draw
		if len(resp.Data) == 0 {
			resp.Data = make([]interface{}, 0)
		}
	}()
	rdb, err := core.GetRedisClient()
	if err != nil {
		logging.RuntimeLog.Errorf("get redis client fail:%v", err)
		return
	}
	defer rdb.Close()

	workerAliveStatus, err := core.LoadWorkerStatusFromRedis(rdb)
	if errors.Is(err, redis.Nil) {
		return
	} else if err != nil {
		logging.RuntimeLog.Errorf(err.Error())
		return
	}
	// 对worker按更新时间排序
	taskAliveUpdateList := make(map[string]int)
	for _, v := range workerAliveStatus {
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
		if v, ok := workerAliveStatus[w.Key]; ok {
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
			var workerOption core.WorkerOption
			err = json.Unmarshal(v.WorkerRunOption, &workerOption)
			if err != nil {
				logging.RuntimeLog.Error(err.Error())
			} else {
				wsd.IsIPv6Support = workerOption.IpV6Support
			}
			resp.Data = append(resp.Data, wsd)
			index++
		}
	}
	resp.RecordsTotal = len(resp.Data)
	resp.RecordsFiltered = len(resp.Data)

	return
}
