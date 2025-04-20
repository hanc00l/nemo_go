package controllers

import (
	"github.com/hanc00l/nemo_go/v3/pkg/db"
	"github.com/hanc00l/nemo_go/v3/pkg/logging"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type OrganizationController struct {
	BaseController
}

type orgRequestParam struct {
	DatableRequestParam
}

type OrganizationData struct {
	Id             string `json:"id" form:"id"`
	Index          int    `json:"index" form:"-"`
	OrgName        string `json:"org_name" form:"org_name"`
	Description    string `json:"description" form:"description"`
	Status         string `json:"status" form:"status"`
	SortNumber     int    `json:"sort_number" form:"sort_number"`
	CreateDatetime string `json:"create_time" form:"-"`
	UpdateDatetime string `json:"update_time" form:"-"`
}

type OrganizationSelectData struct {
	Id      string `json:"id"`
	OrgName string `json:"name"`
}

// GetAllAction 获取所有的记录
func (c *OrganizationController) GetAllAction() {
	defer func(c *OrganizationController, encoding ...bool) {
		_ = c.ServeJSON(encoding...)
	}(c)

	workspaceId := c.GetWorkspace()
	if len(workspaceId) == 0 {
		c.FailedStatus("未选择当前的工作空间！")
		return
	}
	mongoClient, err := db.GetClient()
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		c.FailedStatus(err.Error())
		return
	}
	defer db.CloseClient(mongoClient)

	org := db.NewOrg(workspaceId, mongoClient)
	orgData, err := org.Find(bson.M{"status": "enable"}, 0, 0)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		c.FailedStatus(err.Error())
		return
	}
	var results []OrganizationSelectData
	for _, r := range orgData {
		results = append(results, OrganizationSelectData{Id: r.Id.Hex(), OrgName: r.Name})
	}
	if len(results) == 0 {
		results = make([]OrganizationSelectData, 0)
	}
	c.Data["json"] = results
}

// IndexAction 显示列表页面
func (c *OrganizationController) IndexAction() {
	c.CheckMultiAccessRequest([]RequestRole{SuperAdmin, Admin}, true)

	c.Layout = "base.html"
	c.TplName = "org-list.html"
}

// ListAction 获取列表显示的数据
func (c *OrganizationController) ListAction() {
	defer func(c *OrganizationController, encoding ...bool) {
		_ = c.ServeJSON(encoding...)
	}(c)
	if c.CheckMultiAccessRequest([]RequestRole{SuperAdmin, Admin}, false) == false {
		c.FailedStatus("当前用户权限不允许！")
		return
	}
	req := orgRequestParam{}
	err := c.ParseForm(&req)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
	}
	c.validateRequestParam(&req)
	resp := c.getListData(req)
	c.Data["json"] = resp
}

// DeleteAction 删除一条记录
func (c *OrganizationController) DeleteAction() {
	defer func(c *OrganizationController, encoding ...bool) {
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

	_ = c.CheckErrorAndStatus(db.NewOrg(workspaceId, mongoClient).Delete(id))

	return
}

// AddIndexAction 新增页面显示
func (c *OrganizationController) AddIndexAction() {
	c.CheckMultiAccessRequest([]RequestRole{SuperAdmin, Admin}, true)

	c.Layout = "base.html"
	c.TplName = "org-add.html"
}

// AddSaveAction 保存新增的记录
func (c *OrganizationController) AddSaveAction() {
	defer func(c *OrganizationController, encoding ...bool) {
		_ = c.ServeJSON(encoding...)
	}(c)
	if c.CheckMultiAccessRequest([]RequestRole{SuperAdmin, Admin}, false) == false {
		c.FailedStatus("当前用户权限不允许！")
		return
	}
	orgData := OrganizationData{}
	err := c.ParseForm(&orgData)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		c.FailedStatus(err.Error())
		return
	}
	workspaceId := c.GetWorkspace()
	if len(workspaceId) == 0 {
		c.FailedStatus("未选择当前的工作空间！")
		return
	}
	mongoClient, err := db.GetClient()
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		c.FailedStatus(err.Error())
		return
	}
	defer db.CloseClient(mongoClient)

	doc := db.OrgDocument{
		Name:        orgData.OrgName,
		Description: orgData.Description,
		SortNumber:  orgData.SortNumber,
		Status:      orgData.Status,
	}
	_ = c.CheckErrorAndStatus(db.NewOrg(workspaceId, mongoClient).Insert(doc))
}

// GetAction 根据ID获取一个记录
func (c *OrganizationController) GetAction() {
	defer func(c *OrganizationController, encoding ...bool) {
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
		logging.RuntimeLog.Error(err.Error())
		c.FailedStatus(err.Error())
		return
	}
	defer db.CloseClient(mongoClient)

	doc, err := db.NewOrg(workspaceId, mongoClient).Get(id)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		c.FailedStatus(err.Error())
		return
	}
	if doc.Id.Hex() != id {
		c.FailedStatus("记录不存在！")
		return
	}
	r := OrganizationData{}
	r.Id = doc.Id.Hex()
	r.OrgName = doc.Name
	r.Description = doc.Description
	r.Status = doc.Status
	r.SortNumber = doc.SortNumber
	r.CreateDatetime = FormatDateTime(doc.CreateTime)
	r.UpdateDatetime = FormatDateTime(doc.UpdateTime)
	c.Data["json"] = r
}

// UpdateAction 更新一个记录
func (c *OrganizationController) UpdateAction() {
	defer func(c *OrganizationController, encoding ...bool) {
		_ = c.ServeJSON(encoding...)
	}(c)
	if c.CheckMultiAccessRequest([]RequestRole{SuperAdmin, Admin}, false) == false {
		c.FailedStatus("当前用户权限不允许！")
		return
	}
	orgData := OrganizationData{}
	err := c.ParseForm(&orgData)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		c.FailedStatus(err.Error())
		return
	}
	if len(orgData.Id) == 0 {
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
		logging.RuntimeLog.Error(err.Error())
		c.FailedStatus(err.Error())
		return
	}
	defer db.CloseClient(mongoClient)

	doc, err := db.NewOrg(workspaceId, mongoClient).Get(orgData.Id)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		c.FailedStatus(err.Error())
		return
	}
	if doc.Id.Hex() != orgData.Id {
		c.FailedStatus("记录不存在！")
		return
	}
	doc.Name = orgData.OrgName
	doc.Description = orgData.Description
	doc.SortNumber = orgData.SortNumber
	doc.Status = orgData.Status
	_ = c.CheckErrorAndStatus(db.NewOrg(workspaceId, mongoClient).Update(orgData.Id, doc))

	return
}

// validateRequestParam 校验请求的参数
func (c *OrganizationController) validateRequestParam(req *orgRequestParam) {
	if req.Length <= 0 {
		req.Length = 50
	}
	if req.Start < 0 {
		req.Start = 0
	}
}

// getListData 获取列表数据
func (c *OrganizationController) getListData(req orgRequestParam) (resp DataTableResponseData) {
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
		logging.RuntimeLog.Error(err.Error())
		c.FailedStatus(err.Error())
		return
	}
	defer db.CloseClient(mongoClient)

	org := db.NewOrg(workspaceId, mongoClient)
	startPage := req.Start/req.Length + 1
	results, err := org.Find(bson.M{}, startPage, req.Length)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		c.FailedStatus(err.Error())
		return
	}
	for i, row := range results {
		r := OrganizationData{}
		r.Id = row.Id.Hex()
		r.Status = row.Status
		r.Index = req.Start + i + 1
		r.OrgName = row.Name
		r.SortNumber = row.SortNumber
		r.UpdateDatetime = FormatDateTime(row.UpdateTime)
		r.CreateDatetime = FormatDateTime(row.CreateTime)

		resp.Data = append(resp.Data, r)
	}
	total, _ := org.Count(bson.M{})
	resp.RecordsTotal = total
	resp.RecordsFiltered = total

	return
}
