package controllers

import (
	"github.com/hanc00l/nemo_go/pkg/db"
	"github.com/hanc00l/nemo_go/pkg/logging"
)

type OrganizationController struct {
	BaseController
}

type orgRequestParam struct {
	DatableRequestParam
}

type OrganizationData struct {
	Id             int    `json:"id" form:"id"`
	Index          int    `json:"index" form:"-"`
	OrgName        string `json:"org_name" form:"org_name"`
	Status         string `json:"status" form:"status"`
	SortOrder      int    `json:"sort_order" form:"sort_order"`
	CreateDatetime string `json:"create_time" form:"-"`
	UpdateDatetime string `json:"update_time" form:"-"`
}

type OrganizationSelectData struct {
	Id      int    `json:"id"`
	OrgName string `json:"name"`
}

// GetAllAction 获取所有的记录
func (c *OrganizationController) GetAllAction() {
	defer c.ServeJSON()

	org := &db.Organization{}
	orgData := org.Gets(make(map[string]interface{}), -1, -1)
	var results []OrganizationSelectData
	for _, r := range orgData {
		results = append(results, OrganizationSelectData{Id: r.Id, OrgName: r.OrgName})
	}
	if len(results) == 0 {
		results = make([]OrganizationSelectData, 0)
	}
	c.Data["json"] = results
}

// IndexAction 显示列表页面
func (c *OrganizationController) IndexAction() {
	c.UpdateOnlineUser()
	c.Layout = "base.html"
	c.TplName = "org-list.html"
}

// ListAction 获取列表显示的数据
func (c *OrganizationController) ListAction() {
	c.UpdateOnlineUser()
	defer c.ServeJSON()

	req := orgRequestParam{}
	err := c.ParseForm(&req)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
	}
	c.validateRequestParam(&req)
	resp := c.getOrganizationListData(req)
	c.Data["json"] = resp
}

// DeleteAction 删除一条记录
func (c *OrganizationController) DeleteAction() {
	defer c.ServeJSON()

	id, err := c.GetInt("id")
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		c.FailedStatus(err.Error())
		return
	}
	org := db.Organization{Id: id}
	c.MakeStatusResponse(org.Delete())
}

// AddIndexAction 新增页面显示
func (c *OrganizationController) AddIndexAction() {
	c.Layout = "base.html"
	c.TplName = "org-add.html"
}

// AddSaveAction 保存新增的记录
func (c *OrganizationController) AddSaveAction() {
	defer c.ServeJSON()

	orgData := OrganizationData{}
	err := c.ParseForm(&orgData)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		c.FailedStatus(err.Error())
		return
	}
	org := db.Organization{}
	org.OrgName = orgData.OrgName
	org.Status = orgData.Status
	org.SortOrder = orgData.SortOrder
	c.MakeStatusResponse(org.Add())
}

// GetAction 根据ID获取一个记录
func (c *OrganizationController) GetAction() {
	defer c.ServeJSON()

	id, err := c.GetInt("id")
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		c.FailedStatus(err.Error())
		return
	}
	r := OrganizationData{}
	org := db.Organization{Id: id}
	if org.Get() {
		r.Id = org.Id
		r.OrgName = org.OrgName
		r.Status = org.Status
		r.SortOrder = org.SortOrder
		r.UpdateDatetime = FormatDateTime(org.UpdateDatetime)
		r.CreateDatetime = FormatDateTime(org.CreateDatetime)
	}
	c.Data["json"] = r
}

// UpdateAction 更新一个记录
func (c *OrganizationController) UpdateAction() {
	defer c.ServeJSON()

	id, err := c.GetInt("id")
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		c.FailedStatus(err.Error())
		return
	}
	orgData := OrganizationData{}
	err = c.ParseForm(&orgData)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		c.FailedStatus(err.Error())
		return
	}
	org := db.Organization{Id: id}
	updateMap := make(map[string]interface{})
	updateMap["org_name"] = orgData.OrgName
	updateMap["sort_order"] = orgData.SortOrder
	updateMap["status"] = orgData.Status
	c.MakeStatusResponse(org.Update(updateMap))
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

// getOrganizationListData 获取列表数据
func (c *OrganizationController) getOrganizationListData(req orgRequestParam) (resp DataTableResponseData) {
	org := db.Organization{}
	searchMap := make(map[string]interface{})
	startPage := req.Start/req.Length + 1
	results := org.Gets(searchMap, startPage, req.Length)
	for i, orgRow := range results {
		r := OrganizationData{}
		r.Id = orgRow.Id
		r.Status = orgRow.Status
		r.Index = req.Start + i + 1
		r.OrgName = orgRow.OrgName
		r.SortOrder = orgRow.SortOrder
		r.UpdateDatetime = FormatDateTime(orgRow.UpdateDatetime)
		r.CreateDatetime = FormatDateTime(orgRow.CreateDatetime)

		resp.Data = append(resp.Data, r)
	}
	resp.Draw = req.Draw
	total := org.Count(searchMap)
	resp.RecordsTotal = total
	resp.RecordsFiltered = total
	if resp.Data == nil {
		resp.Data = make([]interface{}, 0)
	}

	return resp
}
