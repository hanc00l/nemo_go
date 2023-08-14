package controllers

import (
	"fmt"
	"github.com/hanc00l/nemo_go/pkg/db"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"strconv"
	"strings"
)

type UserController struct {
	BaseController
}

type userRequestParam struct {
	DatableRequestParam
}

type UserData struct {
	Id              int    `json:"id" form:"id"`
	Index           int    `json:"index" form:"-"`
	UserName        string `json:"user_name" form:"user_name"`
	UserPassword    string `json:"user_password" form:"user_password"`
	UserDescription string `json:"user_description" form:"user_description"`
	UserRole        string `json:"user_role" form:"user_role"`
	State           string `json:"state" form:"state"`
	SortOrder       int    `json:"sort_order" form:"sort_order"`
	CreateDatetime  string `json:"create_time" form:"-"`
	UpdateDatetime  string `json:"update_time" form:"-"`
}

// IndexAction 显示列表页面
func (c *UserController) IndexAction() {
	c.CheckOneAccessRequest(SuperAdmin, true)

	c.Layout = "base.html"
	c.TplName = "user-list.html"
}

// ListAction 获取列表显示的数据
func (c *UserController) ListAction() {
	c.CheckOneAccessRequest(SuperAdmin, true)
	defer c.ServeJSON()

	req := userRequestParam{}
	err := c.ParseForm(&req)
	if err != nil {
		logging.RuntimeLog.Error(err)
		logging.CLILog.Error(err)
	}
	c.validateRequestParam(&req)
	resp := c.getUserListData(req)
	c.Data["json"] = resp
}

// validateRequestParam 校验请求的参数
func (c *UserController) validateRequestParam(req *userRequestParam) {
	if req.Length <= 0 {
		req.Length = 50
	}
	if req.Start < 0 {
		req.Start = 0
	}
}

// getUserListData 获取列表数据
func (c *UserController) getUserListData(req userRequestParam) (resp DataTableResponseData) {
	user := db.User{}
	searchMap := make(map[string]interface{})
	startPage := req.Start/req.Length + 1
	results, _ := user.Gets(searchMap, startPage, req.Length)
	for i, userRow := range results {
		u := UserData{}
		u.Id = userRow.Id
		u.State = userRow.State
		u.Index = req.Start + i + 1
		u.UserName = userRow.UserName
		switch userRow.UserRole {
		case SuperAdmin:
			u.UserRole = "超级管理员"
		case Admin:
			u.UserRole = "管理员"
		case Guest:
			u.UserRole = "普通用户"
		}
		u.UserDescription = userRow.UserDescription
		u.SortOrder = userRow.SortOrder
		u.UpdateDatetime = FormatDateTime(userRow.UpdateDatetime)
		u.CreateDatetime = FormatDateTime(userRow.CreateDatetime)
		resp.Data = append(resp.Data, u)
	}
	resp.Draw = req.Draw
	total := user.Count(searchMap)
	resp.RecordsTotal = total
	resp.RecordsFiltered = total
	if resp.Data == nil {
		resp.Data = make([]interface{}, 0)
	}

	return resp
}

// AddIndexAction 新增页面显示
func (c *UserController) AddIndexAction() {
	c.CheckOneAccessRequest(SuperAdmin, true)

	c.Layout = "base.html"
	c.TplName = "user-add.html"
}

// AddSaveAction 保存新增的记录
func (c *UserController) AddSaveAction() {
	c.CheckOneAccessRequest(SuperAdmin, true)

	defer c.ServeJSON()
	userData := UserData{}
	err := c.ParseForm(&userData)
	if err != nil {
		logging.RuntimeLog.Error(err)
		logging.CLILog.Error(err)
		c.FailedStatus(err.Error())
		return
	}
	user := db.User{}
	user.UserName = userData.UserName
	user.UserPassword = ProcessPasswordHash(userData.UserPassword)
	user.UserRole = userData.UserRole
	user.UserDescription = userData.UserDescription
	user.State = userData.State
	user.SortOrder = userData.SortOrder
	c.MakeStatusResponse(user.Add())

	logging.RuntimeLog.Infof("add new user:%s,type:%s", userData.UserName, userData.UserRole)
}

// GetAction 根据ID获取一个记录
func (c *UserController) GetAction() {
	c.CheckOneAccessRequest(SuperAdmin, true)
	defer c.ServeJSON()

	id, err := c.GetInt("id")
	if err != nil {
		logging.RuntimeLog.Error(err)
		logging.CLILog.Error(err)
		c.FailedStatus(err.Error())
		return
	}
	userData := UserData{}
	user := db.User{Id: id}
	if user.Get() {
		userData.Id = user.Id
		userData.UserName = user.UserName
		userData.SortOrder = user.SortOrder
		userData.UserRole = user.UserRole
		userData.State = user.State
		userData.UserDescription = user.UserDescription
		userData.UpdateDatetime = FormatDateTime(user.UpdateDatetime)
		userData.CreateDatetime = FormatDateTime(user.CreateDatetime)
	}
	c.Data["json"] = userData
}

// UpdateAction 更新一个记录
func (c *UserController) UpdateAction() {
	c.CheckOneAccessRequest(SuperAdmin, true)
	defer c.ServeJSON()

	id, err := c.GetInt("id")
	if err != nil {
		logging.RuntimeLog.Error(err)
		logging.CLILog.Error(err)
		c.FailedStatus(err.Error())
		return
	}
	userData := UserData{}
	err = c.ParseForm(&userData)
	if err != nil {
		logging.RuntimeLog.Error(err)
		logging.CLILog.Error(err)
		c.FailedStatus(err.Error())
		return
	}
	user := db.User{Id: id}
	updateMap := make(map[string]interface{})
	updateMap["user_name"] = userData.UserName
	updateMap["sort_order"] = userData.SortOrder
	updateMap["state"] = userData.State
	updateMap["user_description"] = userData.UserDescription
	updateMap["user_role"] = userData.UserRole
	c.MakeStatusResponse(user.Update(updateMap))

	logging.RuntimeLog.Infof("update user:%s,type:%s", userData.UserName, userData.UserRole)

}

// DeleteAction 删除一条记录
func (c *UserController) DeleteAction() {
	c.CheckOneAccessRequest(SuperAdmin, true)
	defer c.ServeJSON()

	id, err := c.GetInt("id")
	if err != nil {
		logging.RuntimeLog.Error(err)
		logging.CLILog.Error(err)
		c.FailedStatus(err.Error())
		return
	}
	user := db.User{Id: id}
	if user.Get() {
		logging.RuntimeLog.Infof("delete user:%s,type:%s", user.UserName, user.UserRole)
		c.MakeStatusResponse(user.Delete())
	} else {
		c.FailedStatus("delete user error: user not exist")
	}
}

// ResetPasswordAction 重置指定用户的密码
func (c *UserController) ResetPasswordAction() {
	c.CheckOneAccessRequest(SuperAdmin, true)
	defer c.ServeJSON()

	userData := UserData{}
	err := c.ParseForm(&userData)
	if err != nil {
		logging.RuntimeLog.Error(err)
		logging.CLILog.Error(err)
		c.FailedStatus(err.Error())
		return
	}
	if userData.Id <= 0 || len(userData.UserPassword) == 0 {
		logging.CLILog.Error("用户id或密码为空")
		c.FailedStatus("用户id或密码为空")
		return
	}
	user := db.User{Id: userData.Id}
	if user.Get() {
		updateMap := make(map[string]interface{})
		updateMap["user_password"] = ProcessPasswordHash(userData.UserPassword)
		if user.Update(updateMap) {
			c.SucceededStatus("重置密码成功！")
			logging.RuntimeLog.Infof("reset user:%s,type:%s", user.UserName, user.UserRole)

			return
		}
	}
	c.FailedStatus("重置密码失败！")
}

// ListUserWorkspaceAction 获取用户相关的工作空间权限设置情况
func (c *UserController) ListUserWorkspaceAction() {
	c.CheckOneAccessRequest(SuperAdmin, true)
	defer c.ServeJSON()

	var workspaceInfoList []WorkspaceInfoData
	userId, err := c.GetInt("user_id")
	if err != nil || userId <= 0 {
		c.Data["json"] = workspaceInfoList
		return
	}
	// 获取已具有的访问权限
	userWorkspaceMap := make(map[int]interface{})
	userWorkspace := db.UserWorkspace{}
	userWorkspaceList := userWorkspace.GetsByUserId(userId)
	for _, uw := range userWorkspaceList {
		userWorkspaceMap[uw.WorkspaceId] = struct{}{}
	}
	// 获取所有的工作空间并生成id、name和用户是否有权限的列表数据
	workspace := db.Workspace{}
	results, _ := workspace.Gets(make(map[string]interface{}), -1, -1)
	for _, wRow := range results {
		wInfoData := WorkspaceInfoData{
			WorkspaceId:   fmt.Sprintf("%d", wRow.Id),
			WorkspaceName: wRow.WorkspaceName,
			Enable:        false,
		}
		if _, ok := userWorkspaceMap[wRow.Id]; ok {
			wInfoData.Enable = true
		}
		workspaceInfoList = append(workspaceInfoList, wInfoData)
	}
	c.Data["json"] = workspaceInfoList
}

// UpdateUserWorkspaceAction 更新用户的工作空间访问权限
func (c *UserController) UpdateUserWorkspaceAction() {
	c.CheckOneAccessRequest(SuperAdmin, true)
	defer c.ServeJSON()

	userId, err := c.GetInt("user_id")
	if err != nil || userId <= 0 {
		c.FailedStatus("user id is empty")
		return
	}
	// 先删除用户已有的workspace：
	userWorkspace := db.UserWorkspace{}
	userWorkspace.RemoveUserWorkspace(userId)
	// 将用户与工作空间的关联写入到数据库
	for _, workspaceId := range strings.Split(c.GetString("workspace_id", ""), ",") {
		if len(workspaceId) == 0 {
			continue
		}
		uw := db.UserWorkspace{}
		wid, errInt := strconv.ParseInt(workspaceId, 10, 64)
		if errInt == nil && wid > 0 {
			uw.WorkspaceId = int(wid)
			uw.UserId = userId
			if uw.Add() == false {
				c.FailedStatus("update fail")
				return
			}
		}
	}

	c.SucceededStatus("update success")
}
