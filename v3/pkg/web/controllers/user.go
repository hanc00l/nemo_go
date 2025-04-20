package controllers

import (
	"fmt"
	"github.com/hanc00l/nemo_go/v3/pkg/db"
	"github.com/hanc00l/nemo_go/v3/pkg/logging"
	"github.com/hanc00l/nemo_go/v3/pkg/utils"
	"go.mongodb.org/mongo-driver/v2/bson"
	"strings"
	"time"
)

type UserController struct {
	BaseController
}

type userRequestParam struct {
	DatableRequestParam
}

type UserData struct {
	Id              string `json:"id" form:"id"`
	Index           int    `json:"index" form:"-"`
	UserName        string `json:"user_name" form:"user_name"`
	UserPassword    string `json:"user_password" form:"user_password"`
	UserDescription string `json:"user_description" form:"user_description"`
	UserRole        string `json:"user_role" form:"user_role"`
	Status          string `json:"status" form:"status"`
	SortNumber      int    `json:"sort_number" form:"sort_number"`
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
	defer func(c *UserController, encoding ...bool) {
		_ = c.ServeJSON(encoding...)
	}(c)
	if c.CheckOneAccessRequest(SuperAdmin, false) == false {
		c.FailedStatus("当前用户权限不允许！")
		return
	}

	req := userRequestParam{}
	err := c.ParseForm(&req)
	if err != nil {
		logging.RuntimeLog.Error(err)
	}
	c.validateRequestParam(&req)
	resp := c.getListData(req)
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

// getListData 获取列表数据
func (c *UserController) getListData(req userRequestParam) (resp DataTableResponseData) {
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

	user := db.NewUser(mongoClient)
	startPage := req.Start/req.Length + 1
	results, err := user.Find(bson.M{}, startPage, req.Length)
	if err != nil {
		logging.RuntimeLog.Error(err)
		c.FailedStatus(err.Error())
		return
	}
	for i, row := range results {
		u := UserData{}
		u.Id = row.Id.Hex()
		u.Status = row.Status
		u.Index = req.Start + i + 1
		u.UserName = row.Username
		switch row.Role {
		case SuperAdmin:
			u.UserRole = "超级管理员"
		case Admin:
			u.UserRole = "管理员"
		case Guest:
			u.UserRole = "普通用户"
		}
		u.UserDescription = row.Description
		u.SortNumber = row.SortNumber
		u.UpdateDatetime = FormatDateTime(row.UpdateTime)
		u.CreateDatetime = FormatDateTime(row.CreateTime)
		resp.Data = append(resp.Data, u)
	}
	total, _ := user.Count(bson.M{})
	resp.RecordsTotal = total
	resp.RecordsFiltered = total

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
	defer func(c *UserController, encoding ...bool) {
		_ = c.ServeJSON(encoding...)
	}(c)
	if c.CheckOneAccessRequest(SuperAdmin, false) == false {
		c.FailedStatus("当前用户权限不允许！")
		return
	}

	userData := UserData{}
	err := c.ParseForm(&userData)
	if err != nil {
		logging.RuntimeLog.Error(err)
		logging.CLILog.Error(err)
		c.FailedStatus(err.Error())
		return
	}
	doc := db.UserDocument{
		Username:    userData.UserName,
		Password:    utils.SHA512(fmt.Sprintf("%s%s", SecuritySalt, userData.UserPassword)),
		Description: userData.UserDescription,
		Role:        userData.UserRole,
		SortNumber:  userData.SortNumber,
		Status:      userData.Status,
	}
	mongoClient, err := db.GetClient()
	if err != nil {
		logging.RuntimeLog.Error(err)
	}
	defer db.CloseClient(mongoClient)

	user := db.NewUser(mongoClient)
	if c.CheckErrorAndStatus(user.Insert(doc)) {
		logging.RuntimeLog.Infof("add new user:%s,type:%s", userData.UserName, userData.UserRole)
	}
}

// GetAction 根据ID获取一个记录
func (c *UserController) GetAction() {
	defer func(c *UserController, encoding ...bool) {
		_ = c.ServeJSON(encoding...)
	}(c)
	if c.CheckOneAccessRequest(SuperAdmin, false) == false {
		c.FailedStatus("当前用户权限不允许！")
		return
	}

	id := c.GetString("id", "")
	if len(id) == 0 {
		c.FailedStatus("empty id")
		return
	}
	userData := UserData{}
	mongoClient, err := db.GetClient()
	if err != nil {
		logging.RuntimeLog.Error(err)
		return
	}
	defer db.CloseClient(mongoClient)

	user := db.NewUser(mongoClient)
	doc, err := user.Get(id)
	if err != nil {
		logging.RuntimeLog.Error(err)
		c.FailedStatus(err.Error())
		return
	}
	userData.Id = doc.Id.Hex()
	userData.UserName = doc.Username
	userData.UserDescription = doc.Description
	userData.UserRole = doc.Role
	userData.Status = doc.Status
	userData.SortNumber = doc.SortNumber
	userData.UpdateDatetime = FormatDateTime(doc.UpdateTime)
	userData.CreateDatetime = FormatDateTime(doc.CreateTime)
	c.Data["json"] = userData

	return
}

// UpdateAction 更新一个记录
func (c *UserController) UpdateAction() {
	defer func(c *UserController, encoding ...bool) {
		_ = c.ServeJSON(encoding...)
	}(c)
	if c.CheckOneAccessRequest(SuperAdmin, false) == false {
		c.FailedStatus("当前用户权限不允许！")
		return
	}

	id := c.GetString("id", "")
	if len(id) == 0 {
		c.FailedStatus("empty id")
		return
	}
	userData := UserData{}
	err := c.ParseForm(&userData)
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
	user := db.NewUser(mongoClient)
	doc, err := user.Get(id)
	if err != nil {
		logging.RuntimeLog.Error(err)
		c.FailedStatus(err.Error())
		return
	}
	if doc.Id.Hex() != id {
		c.FailedStatus("user not exist")
		return
	}
	doc.Username = userData.UserName
	doc.Password = PasswordEncrypt(userData.UserPassword)
	doc.Description = userData.UserDescription
	doc.Role = userData.UserRole
	doc.SortNumber = userData.SortNumber
	doc.Status = userData.Status
	doc.UpdateTime = time.Now()
	if c.CheckErrorAndStatus(user.Update(doc.Id.Hex(), doc)) {
		logging.RuntimeLog.Infof("update user:%s,type:%s", userData.UserName, userData.UserRole)
	}
}

// DeleteAction 删除一条记录
func (c *UserController) DeleteAction() {
	defer func(c *UserController, encoding ...bool) {
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

	user := db.NewUser(mongoClient)
	doc, err := user.Get(id)
	if err != nil {
		logging.RuntimeLog.Error(err)
		c.FailedStatus(err.Error())
		return
	}
	if doc.Id.Hex() != id {
		c.FailedStatus("user not exist")
		return
	}
	// 删除用户相关的工作空间权限
	if c.CheckErrorAndStatus(user.Delete(doc.Id.Hex())) {
		logging.RuntimeLog.Infof("delete user:%s,type:%s", doc.Username, doc.Role)
	}
}

// ResetPasswordAction 重置指定用户的密码
func (c *UserController) ResetPasswordAction() {
	defer func(c *UserController, encoding ...bool) {
		_ = c.ServeJSON(encoding...)
	}(c)
	if c.CheckOneAccessRequest(SuperAdmin, false) == false {
		c.FailedStatus("当前用户权限不允许！")
		return
	}

	userData := UserData{}
	err := c.ParseForm(&userData)
	if err != nil {
		logging.RuntimeLog.Error(err)
		c.FailedStatus(err.Error())
		return
	}
	if len(userData.Id) == 0 || len(userData.UserPassword) == 0 {
		c.FailedStatus("用户id或密码为空")
		return
	}
	mongoClient, err := db.GetClient()
	if err != nil {
		logging.RuntimeLog.Error(err)
		c.FailedStatus(err.Error())
		return
	}
	defer db.CloseClient(mongoClient)
	user := db.NewUser(mongoClient)
	doc, err := user.Get(userData.Id)
	if err != nil {
		logging.RuntimeLog.Error(err)
		c.FailedStatus(err.Error())
		return
	}
	if doc.Id.Hex() != userData.Id {
		c.FailedStatus("user not exist")
		return
	}
	doc.Password = PasswordEncrypt(userData.UserPassword)
	if c.CheckErrorAndStatus(user.Update(doc.Id.Hex(), doc)) {
		logging.RuntimeLog.Infof("reset user:%s,type:%s", doc.Username, doc.Role)
		c.SucceededStatus("重置密码成功！")
	} else {
		c.FailedStatus("重置密码失败！")
	}
}

// ListUserWorkspaceAction 获取用户相关的工作空间权限设置情况
func (c *UserController) ListUserWorkspaceAction() {
	defer func(c *UserController, encoding ...bool) {
		_ = c.ServeJSON(encoding...)
	}(c)
	if c.CheckOneAccessRequest(SuperAdmin, false) == false {
		c.FailedStatus("当前用户权限不允许！")
		return
	}

	var workspaceInfoList []WorkspaceInfoData
	userId := c.GetString("user_id")
	if len(userId) == 0 {
		c.Data["json"] = workspaceInfoList
		return
	}
	mongoClient, err := db.GetClient()
	if err != nil {
		logging.RuntimeLog.Error(err)
		c.Data["json"] = workspaceInfoList
		return
	}
	defer db.CloseClient(mongoClient)
	user := db.NewUser(mongoClient)
	userDoc, err := user.Get(userId)
	if err != nil {
		logging.RuntimeLog.Error(err)
		c.Data["json"] = workspaceInfoList
		return
	}
	if userDoc.Id.Hex() != userId {
		c.Data["json"] = workspaceInfoList
		return
	}
	workspace := db.NewWorkspace(mongoClient)
	// 获取用户已有的工作空间权限
	workspaceListDocs, err := workspace.Find(bson.M{}, 0, 0)
	if err != nil {
		logging.RuntimeLog.Error(err)
		c.Data["json"] = workspaceInfoList
		return
	}
	for _, doc := range workspaceListDocs {
		workspaceInfo := WorkspaceInfoData{}
		workspaceInfo.WorkspaceId = doc.WorkspaceId
		workspaceInfo.WorkspaceName = doc.Name
		for _, uw := range userDoc.WorkspaceId {
			if uw == doc.WorkspaceId {
				workspaceInfo.Enable = true
				break
			}
		}
		workspaceInfoList = append(workspaceInfoList, workspaceInfo)
	}
	c.Data["json"] = workspaceInfoList
}

// UpdateUserWorkspaceAction 更新用户的工作空间访问权限
func (c *UserController) UpdateUserWorkspaceAction() {
	defer func(c *UserController, encoding ...bool) {
		_ = c.ServeJSON(encoding...)
	}(c)
	if c.CheckOneAccessRequest(SuperAdmin, false) == false {
		c.FailedStatus("当前用户权限不允许！")
		return
	}

	userId := c.GetString("user_id")
	if len(userId) == 0 {
		c.FailedStatus("user id is empty")
		return
	}
	mongoClient, err := db.GetClient()
	if err != nil {
		logging.RuntimeLog.Error(err)
		c.FailedStatus(err.Error())
		return
	}
	defer db.CloseClient(mongoClient)
	user := db.NewUser(mongoClient)
	userDoc, err := user.Get(userId)
	if err != nil {
		logging.RuntimeLog.Error(err)
		c.FailedStatus(err.Error())
		return
	}
	if userDoc.Id.Hex() != userId {
		c.FailedStatus("user not exist")
		return
	}
	userDoc.WorkspaceId = []string{}
	// 将用户与工作空间的关联写入到数据库
	for _, workspaceId := range strings.Split(c.GetString("workspace_id", ""), ",") {
		if len(workspaceId) == 0 {
			continue
		}
		userDoc.WorkspaceId = append(userDoc.WorkspaceId, workspaceId)
	}
	if c.CheckErrorAndStatus(user.Update(userDoc.Id.Hex(), userDoc)) {
		logging.RuntimeLog.Infof("update user workspace:%s,type:%s", userDoc.Username, userDoc.Role)
	}
	return
}
