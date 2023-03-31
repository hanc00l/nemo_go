package controllers

import (
	"fmt"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/db"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"os"
	"path/filepath"
)

type WorkspaceController struct {
	BaseController
}

type workspaceRequestParam struct {
	DatableRequestParam
}

type WorkspaceData struct {
	Id                   int    `json:"id" form:"id"`
	Index                int    `json:"index" form:"-"`
	WorkspaceName        string `json:"workspace_name" form:"workspace_name"`
	WorkspaceGUID        string `json:"workspace_guid" form:"workspace_guid"`
	WorkspaceDescription string `json:"workspace_description" form:"workspace_description"`
	State                string `json:"state" form:"state"`
	SortOrder            int    `json:"sort_order" form:"sort_order"`
	CreateDatetime       string `json:"create_time" form:"-"`
	UpdateDatetime       string `json:"update_time" form:"-"`
}

type WorkspaceInfoData struct {
	WorkspaceId   string `json:"workspaceId"`
	WorkspaceName string `json:"workspaceName"`
	Enable        bool   `json:"enable"`
}

type WorkspaceInfo struct {
	WorkspaceInfoList []WorkspaceInfoData
	CurrentWorkspace  string
}

// UserWorkspaceAction 获取用户的workspace， 用于用户进行用户工作的切换
func (c *WorkspaceController) UserWorkspaceAction() {
	defer c.ServeJSON()

	username := c.GetCurrentUser()
	workspaceId := c.GetCurrentWorkspace()

	var workspaceInfo WorkspaceInfo
	if len(username) == 0 {
		c.Data["json"] = workspaceInfo
		return
	}
	user := db.User{UserName: username}
	if user.GetByUsername() == false {
		c.Data["json"] = workspaceInfo
		return
	}
	if user.UserRole == SuperAdmin {
		workspaceInfo.WorkspaceInfoList = append(workspaceInfo.WorkspaceInfoList, WorkspaceInfoData{
			WorkspaceId:   "0",
			WorkspaceName: "--工作空间--",
			Enable:        true,
		})
		workspace := db.Workspace{}
		searchMap := make(map[string]interface{})
		searchMap["state"] = "enable"
		workspaceList, _ := workspace.Gets(searchMap, -1, -1)
		for _, w := range workspaceList {
			workspaceInfo.WorkspaceInfoList = append(workspaceInfo.WorkspaceInfoList, WorkspaceInfoData{
				WorkspaceId:   fmt.Sprintf("%d", w.Id),
				WorkspaceName: w.WorkspaceName,
				Enable:        true,
			})
		}
	} else {
		userWorkspace := db.UserWorkspace{}
		userWorkspaceList := userWorkspace.GetsByUserId(user.Id)
		for _, uw := range userWorkspaceList {
			workspace := db.Workspace{Id: uw.WorkspaceId}
			if workspace.Get() == false {
				continue
			}
			if workspace.State == "enable" {
				workspaceInfo.WorkspaceInfoList = append(workspaceInfo.WorkspaceInfoList, WorkspaceInfoData{
					WorkspaceId:   fmt.Sprintf("%d", workspace.Id),
					WorkspaceName: workspace.WorkspaceName,
					Enable:        true,
				})
			}
		}
	}
	workspaceInfo.CurrentWorkspace = fmt.Sprintf("%d", workspaceId)
	c.Data["json"] = workspaceInfo
}

// ChangeWorkspaceSelectAction 切换到指定的workspace、更新JWTData或session
func (c *WorkspaceController) ChangeWorkspaceSelectAction() {
	defer c.ServeJSON()

	userName := c.GetCurrentUser()
	newWorkspaceId, err := c.GetInt("workspace")
	if err != nil {
		c.FailedStatus(err.Error())
		return
	}
	user := db.User{UserName: userName}
	if user.GetByUsername() == false {
		c.FailedStatus("user not exist!")
		return
	}
	if user.UserRole == Guest || user.UserRole == Admin {
		if newWorkspaceId == 0 {
			c.FailedStatus("未选择当前的工作空间！")
			return
		}
		userWorkspace := db.UserWorkspace{UserId: user.Id, WorkspaceId: newWorkspaceId}
		if userWorkspace.GetByUserAndWorkspaceId() == false {
			c.FailedStatus("user or workspace not permit!")
			return
		}
		workspace := db.Workspace{Id: newWorkspaceId}
		if !workspace.Get() || workspace.State != "enable" {
			c.FailedStatus("workspace not exist or disabled!")
			return
		}
	}
	// 生成新的token
	if c.IsServerAPI {
		tokenString, err := GenerateToken(TokenData{
			User:      user.UserName,
			UserRole:  user.UserRole,
			Workspace: newWorkspaceId,
		})
		if err != nil || tokenString == "" {
			c.FailedStatus("生成新的token失败")
		}
		c.SucceededStatus(tokenString)
	} else {
		// 设置新的workspaceId
		c.SetSession("Workspace", newWorkspaceId)
		c.SucceededStatus("")
	}
}

// IndexAction 显示列表页面
func (c *WorkspaceController) IndexAction() {
	c.CheckOneAccessRequest(SuperAdmin, true)

	c.Layout = "base.html"
	c.TplName = "workspace-list.html"
}

// ListAction 获取列表显示的数据
func (c *WorkspaceController) ListAction() {
	c.CheckOneAccessRequest(SuperAdmin, true)
	defer c.ServeJSON()

	req := workspaceRequestParam{}
	err := c.ParseForm(&req)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
	}
	c.validateRequestParam(&req)
	resp := c.getWorkspaceListData(req)
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

// getWorkspaceListData 获取列表数据
func (c *WorkspaceController) getWorkspaceListData(req workspaceRequestParam) (resp DataTableResponseData) {
	workspace := db.Workspace{}
	searchMap := make(map[string]interface{})
	startPage := req.Start/req.Length + 1
	results, _ := workspace.Gets(searchMap, startPage, req.Length)
	for i, workspaceRow := range results {
		wData := WorkspaceData{}
		wData.Id = workspaceRow.Id
		wData.State = workspaceRow.State
		wData.Index = req.Start + i + 1
		wData.WorkspaceGUID = workspaceRow.WorkspaceGUID
		wData.WorkspaceName = workspaceRow.WorkspaceName
		wData.WorkspaceDescription = workspaceRow.WorkspaceDescription
		wData.SortOrder = workspaceRow.SortOrder
		wData.UpdateDatetime = FormatDateTime(workspaceRow.UpdateDatetime)
		wData.CreateDatetime = FormatDateTime(workspaceRow.CreateDatetime)
		resp.Data = append(resp.Data, wData)
	}
	resp.Draw = req.Draw
	total := workspace.Count(searchMap)
	resp.RecordsTotal = total
	resp.RecordsFiltered = total
	if resp.Data == nil {
		resp.Data = make([]interface{}, 0)
	}

	return resp
}

// AddIndexAction 新增页面显示
func (c *WorkspaceController) AddIndexAction() {
	c.CheckOneAccessRequest(SuperAdmin, true)

	c.Layout = "base.html"
	c.TplName = "workspace-add.html"
}

// AddSaveAction 保存新增的记录
func (c *WorkspaceController) AddSaveAction() {
	c.CheckOneAccessRequest(SuperAdmin, true)
	defer c.ServeJSON()

	wData := WorkspaceData{}
	err := c.ParseForm(&wData)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		c.FailedStatus(err.Error())
		return
	}
	workspace := db.Workspace{}
	workspace.WorkspaceName = wData.WorkspaceName
	workspace.State = wData.State
	workspace.SortOrder = wData.SortOrder
	workspace.WorkspaceDescription = wData.WorkspaceDescription
	c.MakeStatusResponse(workspace.Add())
}

// GetAction 根据ID获取一个记录
func (c *WorkspaceController) GetAction() {
	c.CheckOneAccessRequest(SuperAdmin, true)
	defer c.ServeJSON()

	id, err := c.GetInt("id")
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		c.FailedStatus(err.Error())
		return
	}
	wData := WorkspaceData{}
	workspace := db.Workspace{Id: id}
	if workspace.Get() {
		wData.Id = workspace.Id
		wData.WorkspaceName = workspace.WorkspaceName
		wData.State = workspace.State
		wData.SortOrder = workspace.SortOrder
		wData.WorkspaceDescription = workspace.WorkspaceDescription
		wData.WorkspaceGUID = workspace.WorkspaceGUID
		wData.UpdateDatetime = FormatDateTime(workspace.UpdateDatetime)
		wData.CreateDatetime = FormatDateTime(workspace.CreateDatetime)
	}
	c.Data["json"] = wData
}

// UpdateAction 更新一个记录
func (c *WorkspaceController) UpdateAction() {
	c.CheckOneAccessRequest(SuperAdmin, true)
	defer c.ServeJSON()

	id, err := c.GetInt("id")
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		c.FailedStatus(err.Error())
		return
	}
	wData := WorkspaceData{}
	err = c.ParseForm(&wData)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		c.FailedStatus(err.Error())
		return
	}
	workspace := db.Workspace{Id: id}
	updateMap := make(map[string]interface{})
	updateMap["workspace_name"] = wData.WorkspaceName
	updateMap["sort_order"] = wData.SortOrder
	updateMap["state"] = wData.State
	updateMap["workspace_description"] = wData.WorkspaceDescription
	c.MakeStatusResponse(workspace.Update(updateMap))
}

// DeleteAction 删除一条记录
func (c *WorkspaceController) DeleteAction() {
	c.CheckOneAccessRequest(SuperAdmin, true)
	defer c.ServeJSON()

	id, err := c.GetInt("id")
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		c.FailedStatus(err.Error())
		return
	}
	workspace := db.Workspace{Id: id}
	if workspace.Get() {
		domainPath := filepath.Join(conf.GlobalServerConfig().Web.WebFiles, workspace.WorkspaceGUID)
		os.RemoveAll(domainPath)
		c.MakeStatusResponse(workspace.Delete())
	}
	c.MakeStatusResponse(false)
}
