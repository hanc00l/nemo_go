package controllers

import (
	"github.com/hanc00l/nemo_go/v3/pkg/db"
	"github.com/hanc00l/nemo_go/v3/pkg/logging"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type NotifyController struct {
	BaseController
}

type notifyRequestParam struct {
	DatableRequestParam
}

type NotifyListData struct {
	Id             string `json:"id" form:"id"`
	Index          int    `json:"index" form:"-"`
	Name           string `json:"name" form:"name"`
	Category       string `json:"category" form:"category"`
	Status         string `json:"status" form:"status"`
	SortNumber     int    `json:"sort_number" form:"sort_number"`
	CreateDatetime string `json:"create_time" form:"-"`
	UpdateDatetime string `json:"update_time" form:"-"`
}

type NotifyInfoData struct {
	Id          string `json:"id" form:"id"`
	Name        string `json:"name" form:"name"`
	Description string `json:"description" form:"description"`
	Category    string `json:"category" form:"category"`
	Token       string `json:"token" form:"token"`
	Template    string `json:"template" form:"template"`
	Status      string `json:"status" form:"status"`
	SortNumber  int    `json:"sort_number" form:"sort_number"`
}
type NotifySelectData struct {
	Id   string `json:"id" form:"id"`
	Name string `json:"name" form:"name"`
}

// IndexAction 显示列表页面
func (c *NotifyController) IndexAction() {
	c.CheckOneAccessRequest(SuperAdmin, true)

	c.Layout = "base.html"
	c.TplName = "notify-list.html"
}

// ListAction 获取列表显示的数据
func (c *NotifyController) ListAction() {
	defer func(c *NotifyController, encoding ...bool) {
		_ = c.ServeJSON(encoding...)
	}(c)
	if !c.CheckOneAccessRequest(SuperAdmin, false) {
		c.FailedStatus("没有权限")
		return
	}

	req := notifyRequestParam{}
	err := c.ParseForm(&req)
	if err != nil {
		logging.RuntimeLog.Error(err)
	}
	c.validateRequestParam(&req)
	resp := c.getListData(req)
	c.Data["json"] = resp
}

// validateRequestParam 校验请求的参数
func (c *NotifyController) validateRequestParam(req *notifyRequestParam) {
	if req.Length <= 0 {
		req.Length = 50
	}
	if req.Start < 0 {
		req.Start = 0
	}
}

// getListData 获取列表数据
func (c *NotifyController) getListData(req notifyRequestParam) (resp DataTableResponseData) {
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

	notify := db.NewNotify(mongoClient)
	startPage := req.Start/req.Length + 1
	results, err := notify.Find(bson.M{}, startPage, req.Length)
	if err != nil {
		logging.RuntimeLog.Error(err)
		c.FailedStatus(err.Error())
		return
	}
	for i, row := range results {
		u := NotifyListData{
			Id:             row.Id.Hex(),
			Index:          req.Start + i + 1,
			Name:           row.Name,
			Category:       row.Category,
			Status:         row.Status,
			SortNumber:     row.SortNumber,
			CreateDatetime: FormatDateTime(row.CreateTime),
			UpdateDatetime: FormatDateTime(row.UpdateTime),
		}
		resp.Data = append(resp.Data, u)
	}
	total, _ := notify.Count(bson.M{})
	resp.RecordsTotal = total
	resp.RecordsFiltered = total

	return resp
}

// AddIndexAction 新增页面显示
func (c *NotifyController) AddIndexAction() {
	c.CheckOneAccessRequest(SuperAdmin, true)

	c.Layout = "base.html"
	c.TplName = "notify-edit.html"
}

// AddSaveAction 保存新增的记录
func (c *NotifyController) AddSaveAction() {
	defer func(c *NotifyController, encoding ...bool) {
		_ = c.ServeJSON(encoding...)
	}(c)
	if !c.CheckOneAccessRequest(SuperAdmin, false) {
		c.FailedStatus("没有权限")
		return
	}
	data := NotifyInfoData{}
	err := c.ParseForm(&data)
	if err != nil {
		logging.RuntimeLog.Error(err)
		c.FailedStatus(err.Error())
		return
	}
	if len(data.Name) == 0 {
		c.FailedStatus("名称不能为空")
		return
	}
	if len(data.Category) == 0 {
		c.FailedStatus("分类不能为空")
		return
	}
	if len(data.Template) == 0 {
		c.FailedStatus("模板不能为空")
		return
	}
	if data.Status != "enable" && data.Status != "disable" {
		c.FailedStatus("状态不正确")
		return
	}
	if data.SortNumber < 0 {
		data.SortNumber = 100
	}
	doc := db.NotifyDocument{
		Name:        data.Name,
		Description: data.Description,
		Category:    data.Category,
		Template:    data.Template,
		Token:       data.Token,
		Status:      data.Status,
		SortNumber:  data.SortNumber,
	}
	mongoClient, err := db.GetClient()
	if err != nil {
		logging.RuntimeLog.Error(err)
	}
	defer db.CloseClient(mongoClient)
	c.CheckErrorAndStatus(db.NewNotify(mongoClient).Insert(doc))
}

// GetAction 根据ID获取一个记录
func (c *NotifyController) GetAction() {
	defer func(c *NotifyController, encoding ...bool) {
		_ = c.ServeJSON(encoding...)
	}(c)
	if !c.CheckOneAccessRequest(SuperAdmin, false) {
		c.FailedStatus("没有权限")
		return
	}

	id := c.GetString("id", "")
	if len(id) == 0 {
		c.FailedStatus("empty id")
		return
	}

	mongoClient, err := db.GetClient()
	if err != nil {
		logging.RuntimeLog.Error(err)
		return
	}
	defer db.CloseClient(mongoClient)

	doc, err := db.NewNotify(mongoClient).Get(id)
	if err != nil {
		logging.RuntimeLog.Error(err)
		c.FailedStatus(err.Error())
		return
	}
	data := NotifyInfoData{
		Id:          doc.Id.Hex(),
		Name:        doc.Name,
		Description: doc.Description,
		Category:    doc.Category,
		Template:    doc.Template,
		Status:      doc.Status,
		SortNumber:  doc.SortNumber,
		Token:       doc.Token,
	}
	c.Data["json"] = data

	return
}
func (c *NotifyController) UpdateIndexAction() {
	c.CheckOneAccessRequest(SuperAdmin, true)

	c.Layout = "base.html"
	c.TplName = "notify-edit.html"
}

// UpdateAction 更新一个记录
func (c *NotifyController) UpdateAction() {
	defer func(c *NotifyController, encoding ...bool) {
		_ = c.ServeJSON(encoding...)
	}(c)
	if !c.CheckOneAccessRequest(SuperAdmin, false) {
		c.FailedStatus("没有权限")
		return
	}

	id := c.GetString("id", "")
	if len(id) == 0 {
		c.FailedStatus("empty id")
		return
	}
	data := NotifyInfoData{}
	err := c.ParseForm(&data)
	if err != nil {
		logging.RuntimeLog.Error(err)
		c.FailedStatus(err.Error())
		return
	}
	mongoClient, err := db.GetClient()
	if err != nil {
		logging.RuntimeLog.Error(err)
		return
	}
	defer db.CloseClient(mongoClient)
	notify := db.NewNotify(mongoClient)
	doc, err := notify.Get(id)
	if err != nil {
		logging.RuntimeLog.Error(err)
		c.FailedStatus(err.Error())
		return
	}
	if doc.Id.Hex() != id {
		c.FailedStatus("id错误")
		return
	}
	doc.Name = data.Name
	doc.Description = data.Description
	doc.Category = data.Category
	doc.Token = data.Token
	doc.Template = data.Template
	doc.Status = data.Status
	doc.SortNumber = data.SortNumber
	c.CheckErrorAndStatus(notify.Update(id, doc))

	return
}

// DeleteAction 删除一条记录
func (c *NotifyController) DeleteAction() {
	defer func(c *NotifyController, encoding ...bool) {
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
	c.CheckErrorAndStatus(db.NewNotify(mongoClient).Delete(id))
}

// ListSelectNotifyAction 获取可用的通知列表
func (c *NotifyController) ListSelectNotifyAction() {
	defer func(c *NotifyController, encoding ...bool) {
		_ = c.ServeJSON(encoding...)
	}(c)
	if c.CheckMultiAccessRequest([]RequestRole{SuperAdmin, Admin}, false) == false {
		c.FailedStatus("当前用户权限不允许！")
		return
	}
	mongoClient, err := db.GetClient()
	if err != nil {
		logging.RuntimeLog.Error(err)
		c.Data["json"] = []NotifySelectData{}
		return
	}
	defer db.CloseClient(mongoClient)

	results, err := db.NewNotify(mongoClient).Find(bson.M{"status": "enable"}, 0, 0)
	if err != nil {
		logging.RuntimeLog.Error(err)
		c.FailedStatus(err.Error())
		return
	}
	var notifyList []NotifySelectData
	for _, row := range results {
		notifyList = append(notifyList, NotifySelectData{
			Id:   row.Id.Hex(),
			Name: row.Name,
		})
	}

	c.Data["json"] = notifyList
}
