package controllers

import (
	"github.com/hanc00l/nemo_go/v3/pkg/db"
	"github.com/hanc00l/nemo_go/v3/pkg/logging"
	"go.mongodb.org/mongo-driver/v2/bson"
	"time"
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
	Id         string `json:"id"`
	Index      int    `json:"index"`
	Source     string `json:"source"`
	File       string `json:"file"`
	Func       string `json:"func"`
	Level      string `json:"level"`
	Message    string `json:"message"`
	CreateTime string `json:"create_datetime"`
	UpdateTime string `json:"update_datetime"`
}

func (c *RuntimeLogController) IndexAction() {
	c.CheckOneAccessRequest(SuperAdmin, true)
	c.Layout = "base.html"
	c.TplName = "runtimelog-list.html"
}

// ListAction 列表的数据
func (c *RuntimeLogController) ListAction() {
	defer func(c *RuntimeLogController, encoding ...bool) {
		_ = c.ServeJSON(encoding...)
	}(c)
	if !c.CheckOneAccessRequest(SuperAdmin, false) {
		c.FailedStatus("没有权限")
		return
	}

	req := runtimeLogRequestParam{}
	err := c.ParseForm(&req)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
	}
	c.validateRequestParam(&req)
	resp := c.getRuntimeLogListData(req)
	c.Data["json"] = resp
}

// DeleteAction 删除一个记录
func (c *RuntimeLogController) DeleteAction() {
	defer func(c *RuntimeLogController, encoding ...bool) {
		_ = c.ServeJSON(encoding...)
	}(c)
	if !c.CheckOneAccessRequest(SuperAdmin, false) {
		c.FailedStatus("没有权限")
		return
	}
	id := c.GetString("id")
	if len(id) == 0 {
		c.FailedStatus("empty id")
		return
	}
	mongoClient, err := db.GetClient()
	if err != nil {
		logging.RuntimeLog.Error(err)
		c.FailedStatus(err.Error())
		return
	}
	defer db.CloseClient(mongoClient)

	c.CheckErrorAndStatus(db.NewRuntimeLog(mongoClient).Delete(id))
}

// BatchDeleteAction 批量清除记录
func (c *RuntimeLogController) BatchDeleteAction() {
	defer func(c *RuntimeLogController, encoding ...bool) {
		_ = c.ServeJSON(encoding...)
	}(c)
	if !c.CheckOneAccessRequest(SuperAdmin, false) {
		c.FailedStatus("没有权限")
		return
	}
	req := runtimeLogRequestParam{}
	err := c.ParseForm(&req)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
	}

	mongoClient, err := db.GetClient()
	if err != nil {
		logging.RuntimeLog.Error(err)
		c.FailedStatus(err.Error())
		return
	}
	defer db.CloseClient(mongoClient)

	c.CheckErrorAndStatus(db.NewRuntimeLog(mongoClient).DeleteMany(getSearchFilter(req)))
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

// getRuntimeLogListData 获取列显示的数据
func (c *RuntimeLogController) getRuntimeLogListData(req runtimeLogRequestParam) (resp DataTableResponseData) {
	defer func() {
		resp.Draw = req.Draw
		if len(resp.Data) == 0 {
			resp.Data = make([]interface{}, 0)
		}
	}()
	mongoClient, err := db.GetClient()
	if err != nil {
		logging.RuntimeLog.Error(err)
		c.FailedStatus(err.Error())
		return
	}
	defer db.CloseClient(mongoClient)

	rlog := db.NewRuntimeLog(mongoClient)
	startPage := req.Start/req.Length + 1
	filter := getSearchFilter(req)
	results, err := rlog.Find(filter, startPage, req.Length)
	if err != nil {
		logging.RuntimeLog.Error(err)
		c.FailedStatus(err.Error())
		return
	}
	for i, row := range results {
		u := RuntimeLogData{
			Id:         row.Id.Hex(),
			Index:      req.Start + i + 1,
			Source:     row.Source,
			File:       row.File,
			Func:       row.Func,
			Level:      row.Level,
			Message:    row.Message,
			CreateTime: FormatDateTime(row.CreateTime),
			UpdateTime: FormatDateTime(row.UpdateTime),
		}
		resp.Data = append(resp.Data, u)
	}
	total, _ := rlog.Count(filter)
	resp.RecordsTotal = total
	resp.RecordsFiltered = total

	return resp
}

func getSearchFilter(req runtimeLogRequestParam) bson.M {
	filter := bson.M{}
	if len(req.Source) > 0 {
		filter["source"] = bson.M{"$regex": req.Source, "$options": "i"}
	}
	if len(req.Func) > 0 {
		filter["func"] = bson.M{"$regex": req.Func, "$options": "i"}
	}
	if req.Level > 0 {
		filter["level_int"] = req.Level
	}
	if len(req.Message) > 0 {
		filter["message"] = bson.M{"$regex": req.Message, "$options": "i"}
	}
	// 日期筛选，注意日志是查看最远期的，所以是小于等于；比如一天前、三天前、一周前等；这个和资产的日是期是相反的
	if req.DateDelta > 0 {
		filter["create_time"] = bson.M{"$lte": GetDateBeforeNow(req.DateDelta)}
	}
	return filter
}

func GetDateBeforeNow(delta int) time.Time {
	now := time.Now()
	return now.AddDate(0, 0, -delta)
}
