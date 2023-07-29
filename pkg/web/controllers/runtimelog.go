package controllers

import (
	"fmt"
	"github.com/hanc00l/nemo_go/pkg/db"
	"github.com/hanc00l/nemo_go/pkg/logging"
)

type RuntimeLogController struct {
	BaseController
}

// runtimeLogRequestParam 请求参数
type runtimeLogRequestParam struct {
	DatableRequestParam
	Source    string `form:"log_source"`
	Func      string `form:"log_func"`
	Level     int    `form:"log_level"`
	Message   string `form:"log_message"`
	DateDelta int    `form:"date_delta"`
}

type RuntimeLogData struct {
	Id         int    `json:"id"`
	Index      int    `json:"index"`
	Source     string `json:"source"`
	File       string `json:"file"`
	Func       string `json:"func"`
	Level      string `json:"level"`
	Message    string `json:"message"`
	CreateTime string `json:"create_datetime"`
	UpdateTime string `json:"update_datetime"`
}

type RuntimeLogInfo struct {
	Id         int
	Source     string
	File       string
	Func       string
	Level      string
	Message    string
	CreateTime string
	UpdateTime string
}

func (c *RuntimeLogController) IndexAction() {
	c.CheckOneAccessRequest(SuperAdmin, true)
	c.Layout = "base.html"
	c.TplName = "runtimelog-list.html"
}

// ListAction 列表的数据
func (c *RuntimeLogController) ListAction() {
	c.CheckOneAccessRequest(SuperAdmin, true)
	defer c.ServeJSON()

	req := runtimeLogRequestParam{}
	err := c.ParseForm(&req)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
	}
	c.validateRequestParam(&req)
	resp := c.getRuntimeLogListData(req)
	c.Data["json"] = resp
}

// InfoAction 显示一个详情
func (c *RuntimeLogController) InfoAction() {
	c.CheckOneAccessRequest(SuperAdmin, true)

	var runtimeLogInfo RuntimeLogInfo
	logId, err := c.GetInt("id")
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
	} else {
		runtimeLogInfo = getRuntimeLogInfo(logId)
	}
	if c.IsServerAPI {
		c.Data["json"] = runtimeLogInfo
		c.ServeJSON()
	} else {
		c.Data["runtimelog_info"] = runtimeLogInfo
		c.Layout = "base.html"
		c.TplName = "runtimelog-info.html"
	}
}

// DeleteAction 删除一个记录
func (c *RuntimeLogController) DeleteAction() {
	defer c.ServeJSON()
	if c.CheckOneAccessRequest(SuperAdmin, false) == false {
		c.FailedStatus("当前用户权限不允许！")
		return
	}

	id, err := c.GetInt("id")
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		c.FailedStatus(err.Error())
		return
	}
	rtlog := db.RuntimeLog{Id: id}
	c.MakeStatusResponse(rtlog.Delete())
}

// BatchDeleteAction 批量清除记录
func (c *RuntimeLogController) BatchDeleteAction() {
	defer c.ServeJSON()
	if c.CheckOneAccessRequest(SuperAdmin, false) == false {
		c.FailedStatus("当前用户权限不允许！")
		return
	}
	req := runtimeLogRequestParam{}
	err := c.ParseForm(&req)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
	}
	searchMap := c.getSearchMap(req)
	fmt.Println(searchMap)
	rtlog := db.RuntimeLog{}
	c.MakeStatusResponse(rtlog.DeleteLogs(searchMap))
}

// validateRequestParam 校验请求的参数
func (c *RuntimeLogController) validateRequestParam(req *runtimeLogRequestParam) {
	if req.Length <= 0 {
		req.Length = 50
	}
	if req.Start < 0 {
		req.Start = 0
	}
}

// getSearchMap 根据查询参数生成查询条件
func (c *RuntimeLogController) getSearchMap(req runtimeLogRequestParam) (searchMap map[string]interface{}) {
	searchMap = make(map[string]interface{})

	if req.Source != "" {
		searchMap["source"] = req.Source
	}
	if req.Func != "" {
		searchMap["func"] = req.Func
	}
	if req.Message != "" {
		searchMap["message"] = req.Message
	}
	if req.Level > 0 {
		searchMap["level_int"] = req.Level
	}
	if req.DateDelta > 0 {
		searchMap["date_delta"] = req.DateDelta
	}
	return
}

// getRuntimeLogListData 获取列显示的数据
func (c *RuntimeLogController) getRuntimeLogListData(req runtimeLogRequestParam) (resp DataTableResponseData) {
	rtlog := db.RuntimeLog{}
	searchMap := c.getSearchMap(req)
	startPage := req.Start/req.Length + 1
	results, total := rtlog.Gets(searchMap, startPage, req.Length)
	for i, logRow := range results {
		v := RuntimeLogData{}
		v.Id = logRow.Id
		v.Index = req.Start + i + 1
		v.Source = logRow.Source
		v.File = logRow.File
		v.Func = logRow.Func
		v.Level = logRow.Level
		v.Message = logRow.Message
		v.CreateTime = FormatDateTime(logRow.CreateDatetime)
		v.UpdateTime = FormatDateTime(logRow.UpdateDatetime)
		resp.Data = append(resp.Data, v)
	}
	resp.Draw = req.Draw
	resp.RecordsTotal = total
	resp.RecordsFiltered = total
	if resp.Data == nil {
		resp.Data = make([]interface{}, 0)
	}
	return
}

// getRuntimeLogInfo 获取一个详情
func getRuntimeLogInfo(id int) (r RuntimeLogInfo) {
	rtlog := db.RuntimeLog{Id: id}
	if !rtlog.Get() {
		return r
	}
	r.Id = id
	r.Source = rtlog.Source
	r.File = rtlog.File
	r.Func = rtlog.Func
	r.Level = rtlog.Level
	r.Message = rtlog.Message
	r.CreateTime = FormatDateTime(rtlog.CreateDatetime)
	r.UpdateTime = FormatDateTime(rtlog.UpdateDatetime)

	return
}
