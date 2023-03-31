package controllers

import ctrl "github.com/hanc00l/nemo_go/pkg/web/controllers"

type WorkspaceController struct {
	ctrl.WorkspaceController
}

// @Title List
// @Description 根据指定筛选条件，获取列表显示的数据
// @Param authorization		header string true "token"
// @Param start 			formData int true "查询的资产的起始行数"
// @Param length 			formData int true "返回资产指定的数量"
// @Success 200 {object} models.WorkspaceDataTableResponseData
// @router /list [post]
func (c *WorkspaceController) List() {
	c.IsServerAPI = true
	c.ListAction()
}

// @Title Info
// @Description 显示一个workspace的详情
// @Param authorization		header string true "token"
// @Param id 				formData int true "id"
// @Success 200 {object} models.WorkspaceData
// @router /info [post]
func (c *WorkspaceController) Info() {
	c.IsServerAPI = true
	c.GetAction()
}

// @Title UserWorkspace
// @Description 显示一个用户具有的workspace列表
// @Param authorization		header string true "token"
// @Success 200 {object} models.WorkspaceInfo
// @router /user-ownerd [post]
func (c *WorkspaceController) UserWorkspace() {
}

// @Title DeleteWorkspace
// @Description 删除一个workspace
// @Param authorization	header string true "token"
// @Param id 			formData int true "id"
// @Success 200 {object} models.StatusResponseData
// @router /delete [post]
func (c *WorkspaceController) DeleteWorkspace() {
	c.IsServerAPI = true
	c.DeleteAction()
}

// @Title SaveWorkspace
// @Description 保存一个新增的记录
// @Param authorization				header string true "token"
// @Param workspace_name 			formData string true "名称"
// @Param workspace_guid 			formData string true "guid
// @Param workspace_description 	formData string true "描述"
// @Param state 					formData string true "状态（enable/disable）"
// @Param sort_order 				formData int true "排序号（默认100）"
// @Success 200 {object} models.StatusResponseData
// @router /save [post]
func (c *WorkspaceController) SaveWorkspace() {
	c.IsServerAPI = true
	c.AddSaveAction()
}

// @Title UpdateWorkspace
// @Description 更新一个已有的记录
// @Param authorization		header string true "token"
// @Param id		 		formData int true "id"
// @Param workspace_name 			formData string true "名称"
// @Param workspace_description 	formData string true "描述"
// @Param state 					formData string true "状态（enable/disable）"
// @Param sort_order 				formData int true "排序号（默认100）"
// @Success 200 {object} models.StatusResponseData
// @router /update [post]
func (c *WorkspaceController) UpdateWorkspace() {
	c.IsServerAPI = true
	c.UpdateAction()
}

// @Title ChangeWorkspaceSelect
// @Description  切换到指定的workspace、更新JWT
// @Param authorization		header string true "token"
// @Param workspace			formData int true "workspace id"
// @Success 200 {object} models.StatusResponseData
// @router /change [post]
func (c *WorkspaceController) ChangeWorkspaceSelect() {
	c.IsServerAPI = true
	c.ChangeWorkspaceSelectAction()
}
