package controllers

import ctrl "github.com/hanc00l/nemo_go/v2/pkg/web/controllers"

type OrganizationController struct {
	ctrl.OrganizationController
}

// @Title List
// @Description 根据指定筛选条件，获取列表显示的数据
// @Param authorization		header string true "token"
// @Param start 			formData int true "查询的资产的起始行数"
// @Param length 			formData int true "返回资产指定的数量"
// @Success 200 {object} models.OrgDataTableResponseData
// @router /list [post]
func (c *OrganizationController) List() {
	c.IsServerAPI = true
	c.ListAction()
}

// @Title Info
// @Description 显示一个资产的详情
// @Param authorization		header string true "token"
// @Param id 				formData int true "id"
// @Success 200 {object} models.OrganizationData
// @router /info [post]
func (c *OrganizationController) Info() {
	c.IsServerAPI = true
	c.GetAction()
}

// @Title GetAll
// @Description 显示一个资产的详情
// @Param authorization		header string true "token"
// @Param id 				formData int true "id"
// @Success 200 {object} models.OrganizationAllData
// @router /getall [post]
func (c *OrganizationController) GetAll() {
	c.IsServerAPI = true
	c.GetAllAction()
}

// @Title DeleteOrg
// @Description 删除一个组织
// @Param authorization	header string true "token"
// @Param id 			formData int true "id"
// @Success 200 {object} models.StatusResponseData
// @router /delete [post]
func (c *OrganizationController) DeleteOrg() {
	c.IsServerAPI = true
	c.DeleteAction()
}

// @Title SaveOrg
// @Description 保存一个新增的记录
// @Param authorization		header string true "token"
// @Param org_name 			formData string true "组织名称"
// @Param status 			formData string true "组织状态（enable/disable）"
// @Param sort_order 		formData int true "排序号（默认100）"
// @Success 200 {object} models.StatusResponseData
// @router /save [post]
func (c *OrganizationController) SaveOrg() {
	c.IsServerAPI = true
	c.AddSaveAction()
}

// @Title UpdateOrg
// @Description 更新一个已有的记录
// @Param authorization		header string true "token"
// @Param id		 		formData int true "id"
// @Param org_name 			formData string true "组织名称"
// @Param status 			formData string true "组织状态（enable/disable）"
// @Param sort_order 		formData int true "排序号（默认100）"
// @Success 200 {object} models.StatusResponseData
// @router /update [post]
func (c *OrganizationController) UpdateOrg() {
	c.IsServerAPI = true
	c.UpdateAction()
}
