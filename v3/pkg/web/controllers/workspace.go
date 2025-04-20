package controllers

import (
	"github.com/hanc00l/nemo_go/v3/pkg/db"
	"github.com/hanc00l/nemo_go/v3/pkg/logging"
	"go.mongodb.org/mongo-driver/v2/bson"
	"strings"
)

type WorkspaceController struct {
	BaseController
}

type workspaceRequestParam struct {
	DatableRequestParam
}

type WorkspaceData struct {
	Id                   string `json:"id" form:"id"`
	Index                int    `json:"index" form:"-"`
	WorkspaceName        string `json:"workspace_name" form:"workspace_name"`
	WorkspaceId          string `json:"workspace_id" form:"workspace_id"`
	WorkspaceDescription string `json:"workspace_description" form:"workspace_description"`
	Status               string `json:"status" form:"status"`
	SortNumber           int    `json:"sort_number" form:"sort_number"`
	Notify               string `json:"notify" form:"notify"`

	CreateDatetime string `json:"create_time" form:"-"`
	UpdateDatetime string `json:"update_time" form:"-"`
}

type WorkspaceInfoData struct {
	WorkspaceId   string `json:"workspace_id"`
	WorkspaceName string `json:"workspace_name"`
	Enable        bool   `json:"enable"`
}

type WorkspaceInfo struct {
	WorkspaceInfoList []WorkspaceInfoData
	CurrentWorkspace  string
}

// IndexAction 显示列表页面
func (c *WorkspaceController) IndexAction() {
	c.CheckOneAccessRequest(SuperAdmin, true)

	c.Layout = "base.html"
	c.TplName = "workspace-list.html"
}

// ListAction 获取列表显示的数据
func (c *WorkspaceController) ListAction() {
	defer func(c *WorkspaceController, encoding ...bool) {
		_ = c.ServeJSON(encoding...)
	}(c)
	if c.CheckOneAccessRequest(SuperAdmin, false) == false {
		c.FailedStatus("当前用户权限不允许！")
		return
	}

	req := workspaceRequestParam{}
	err := c.ParseForm(&req)
	if err != nil {
		logging.RuntimeLog.Error(err)
		logging.CLILog.Error(err)
	}
	c.validateRequestParam(&req)
	resp := c.getListData(req)
	c.Data["json"] = resp
}

// validateRequestParam 校验请求的参数
func (c *WorkspaceController) validateRequestParam(req *workspaceRequestParam) {
	if req.Length <= 0 {
		req.Length = 50
	}
	if req.Start < 0 {
		req.Start = 0
	}
}

// getListData 获取列表数据
func (c *WorkspaceController) getListData(req workspaceRequestParam) (resp DataTableResponseData) {
	defer func() {
		resp.Draw = req.Draw
		if len(resp.Data) == 0 {
			resp.Data = make([]interface{}, 0)
		}
	}()
	mongoClient, err := db.GetClient()
	if err != nil {
		logging.RuntimeLog.Error(err)
		return
	}
	defer db.CloseClient(mongoClient)

	workspace := db.NewWorkspace(mongoClient)
	startPage := req.Start/req.Length + 1
	results, err := workspace.Find(bson.M{}, startPage, req.Length)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return
	}
	notify := db.NewNotify(mongoClient)
	for i, row := range results {
		wData := WorkspaceData{}
		wData.Id = row.Id.Hex()
		wData.Status = row.Status
		wData.Index = req.Start + i + 1
		wData.WorkspaceId = row.WorkspaceId
		wData.WorkspaceName = row.Name
		wData.SortNumber = row.SortNumber
		if len(row.NotifyId) > 0 {
			var notifyList []string
			for _, notifyId := range row.NotifyId {
				notifyDoc, err := notify.Get(notifyId)
				if err != nil {
					logging.RuntimeLog.Error(err)
					continue
				}
				notifyList = append(notifyList, notifyDoc.Name)
			}
			wData.Notify = strings.Join(notifyList, ",")
		}
		wData.UpdateDatetime = FormatDateTime(row.UpdateTime)
		wData.CreateDatetime = FormatDateTime(row.CreateTime)

		resp.Data = append(resp.Data, wData)
	}
	total, _ := workspace.Count(bson.M{})
	resp.RecordsTotal = total
	resp.RecordsFiltered = total

	return resp
}

// AddIndexAction 新增页面显示
func (c *WorkspaceController) AddIndexAction() {
	c.CheckOneAccessRequest(SuperAdmin, true)

	c.Data["workspaceId"] = ""
	c.Layout = "base.html"
	c.TplName = "workspace-edit.html"
}

func (c *WorkspaceController) EditIndexAction() {
	c.CheckOneAccessRequest(SuperAdmin, true)

	id := c.GetString("id")
	if len(id) == 0 {
		logging.RuntimeLog.Error("empty id")
	}
	c.Data["workspaceId"] = id
	c.Layout = "base.html"
	c.TplName = "workspace-edit.html"
}

// GetAction 根据ID获取一个记录
func (c *WorkspaceController) GetAction() {
	defer func(c *WorkspaceController, encoding ...bool) {
		_ = c.ServeJSON(encoding...)
	}(c)
	if c.CheckOneAccessRequest(SuperAdmin, false) == false {
		c.FailedStatus("当前用户权限不允许！")
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
	doc, err := db.NewWorkspace(mongoClient).Get(id)
	if err != nil {
		logging.RuntimeLog.Error(err)
		c.FailedStatus(err.Error())
		return
	}
	if doc.Id.Hex() != id {
		c.FailedStatus("not found")
		return
	}
	wData := WorkspaceData{
		Id:                   doc.Id.Hex(),
		WorkspaceName:        doc.Name,
		WorkspaceId:          doc.WorkspaceId,
		WorkspaceDescription: doc.Description,
		Status:               doc.Status,
		SortNumber:           doc.SortNumber,
		Notify:               strings.Join(doc.NotifyId, ","),
		CreateDatetime:       FormatDateTime(doc.CreateTime),
		UpdateDatetime:       FormatDateTime(doc.UpdateTime),
	}
	c.Data["json"] = wData
}

// SaveAction 更新一个记录
func (c *WorkspaceController) SaveAction() {
	defer func(c *WorkspaceController, encoding ...bool) {
		_ = c.ServeJSON(encoding...)
	}(c)
	if c.CheckOneAccessRequest(SuperAdmin, false) == false {
		c.FailedStatus("当前用户权限不允许！")
		return
	}

	wData := WorkspaceData{}
	err := c.ParseForm(&wData)
	if err != nil {
		logging.RuntimeLog.Error(err)
		c.FailedStatus(err.Error())
		return
	}
	mongoClient, err := db.GetClient()
	if err != nil {
		logging.RuntimeLog.Error(err)
		c.FailedStatus(err.Error())
		return
	}
	defer db.CloseClient(mongoClient)
	workspace := db.NewWorkspace(mongoClient)
	// 新增
	if len(wData.Id) == 0 {
		doc := db.WorkspaceDocument{
			Name:        wData.WorkspaceName,
			Description: wData.WorkspaceDescription,
			Status:      wData.Status,
			SortNumber:  wData.SortNumber,
			NotifyId:    strings.Split(wData.Notify, ","),
		}
		doc.Id = bson.NewObjectID()
		if c.CheckErrorAndStatus(workspace.Insert(doc)) {
			logging.RuntimeLog.Infof("新增工作空间:%s", wData.WorkspaceName)
			c.SucceededStatus(doc.Id.Hex())
		}
	} else {
		doc, err := workspace.Get(wData.Id)
		if err != nil {
			logging.RuntimeLog.Error(err)
			c.FailedStatus(err.Error())
			return
		}
		if doc.Id.Hex() != wData.Id {
			c.FailedStatus("not found")
			return
		}
		doc.Name = wData.WorkspaceName
		doc.Description = wData.WorkspaceDescription
		doc.Status = wData.Status
		doc.SortNumber = wData.SortNumber
		doc.NotifyId = strings.Split(wData.Notify, ",")
		if c.CheckErrorAndStatus(workspace.Update(doc.Id, doc)) {
			logging.RuntimeLog.Infof("更新工作空间:%s", wData.WorkspaceName)
			c.SucceededStatus(doc.Id.Hex())
		}
	}
	return
}

// DeleteAction 删除一条记录
func (c *WorkspaceController) DeleteAction() {
	defer func(c *WorkspaceController, encoding ...bool) {
		_ = c.ServeJSON(encoding...)
	}(c)
	if c.CheckOneAccessRequest(SuperAdmin, false) == false {
		c.FailedStatus("当前用户权限不允许！")
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
	workspace := db.NewWorkspace(mongoClient)
	if c.CheckErrorAndStatus(workspace.Delete(id)) {
		logging.RuntimeLog.Infof("delete workspace:%s", id)
	}

	return
}
