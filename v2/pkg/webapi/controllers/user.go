package controllers

import ctrl "github.com/hanc00l/nemo_go/v2/pkg/web/controllers"

type UserController struct {
	ctrl.UserController
}

// @Title List
// @Description 根据指定筛选条件，获取列表显示的数据
// @Param authorization		header string true "token"
// @Param start 			formData int true "查询的资产的起始行数"
// @Param length 			formData int true "返回资产指定的数量"
// @Success 200 {object} models.UserDataTableResponseData
// @router /list [post]
func (c *UserController) List() {
	c.IsServerAPI = true
	c.ListAction()
}

// @Title Info
// @Description 显示一个资产的详情
// @Param authorization		header string true "token"
// @Param id 				formData int true "id"
// @Success 200 {object} models.UserData
// @router /info [post]
func (c *UserController) Info() {
	c.IsServerAPI = true
	c.GetAction()
}

// @Title DeleteUser
// @Description 删除一个用户
// @Param authorization	header string true "token"
// @Param id 			formData int true "id"
// @Success 200 {object} models.StatusResponseData
// @router /delete [post]
func (c *UserController) DeleteUser() {
	c.IsServerAPI = true
	c.DeleteAction()
}

// @Title SaveUser
// @Description 保存一个新增的记录
// @Param authorization		header string true "token"
// @Param user_name 		formData string true "用户名称"
// @Param user_password 	formData string true "用户密码
// @Param user_role 		formData string true "用户角色"
// @Param user_description 	formData string true "用户描述"
// @Param state 			formData string true "用户状态（enable/disable）"
// @Param sort_order 		formData int true "排序号（默认100）"
// @Success 200 {object} models.StatusResponseData
// @router /save [post]
func (c *UserController) SaveUser() {
	c.IsServerAPI = true
	c.AddSaveAction()
}

// @Title UpdateUser
// @Description 更新一个已有的记录
// @Param authorization		header string true "token"
// @Param id		 		formData int true "id"
// @Param user_name 		formData string true "用户名称"
// @Param user_role 		formData string true "用户角色"
// @Param user_description 	formData string true "用户描述"
// @Param state 			formData string true "用户状态（enable/disable）"
// @Param sort_order 		formData int true "排序号（默认100）"
// @Success 200 {object} models.StatusResponseData
// @router /update [post]
func (c *UserController) UpdateUser() {
	c.IsServerAPI = true
	c.UpdateAction()
}

// @Title ResetPassword
// @Description 重置用户密码
// @Param authorization		header string true "token"
// @Param id		 		formData int true "用户的id"
// @Param user_password 	formData string true "用户密码"
// @Success 200 {object} models.StatusResponseData
// @router /reset [post]
func (c *UserController) ResetPassword() {
	c.IsServerAPI = true
	c.ResetPasswordAction()
}

// @Title ListUserWorkspace
// @Description 获取用户相关的工作空间权限设置情况
// @Param authorization		header string true "token"
// @Param user_id		 	formData int true "用户的id"
// @Success 200 {object} models.UserWorkspaceData
// @router /workspace/list [post]
func (c *UserController) ListUserWorkspace() {
	c.IsServerAPI = true
	c.ListUserWorkspaceAction()
}

// @Title UpdateUserWorkspace
// @Description 获取用户相关的工作空间权限设置情况
// @Param authorization		header string true "token"
// @Param user_id		 	formData int true "用户的id"
// @Param workspace_id		formData int true "workspace id，多个id用\",\"分隔"
// @Success 200 {object} models.StatusResponseData
// @router /workspace/update [post]
func (c *UserController) UpdateUserWorkspace() {
	c.IsServerAPI = true
	c.UpdateUserWorkspaceAction()
}
