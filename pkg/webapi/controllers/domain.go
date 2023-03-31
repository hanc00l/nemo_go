package controllers

import ctrl "github.com/hanc00l/nemo_go/pkg/web/controllers"

type DomainController struct {
	ctrl.DomainController
}

// @Title List
// @Description 根据指定筛选条件，查询domain资产
// @Param authorization		header string true "token"
// @Param start 			formData int true "查询的资产的起始行数"
// @Param length 			formData int true "返回资产指定的数量"
// @Param org_id 			formData int false "组织机构的ID"
// @Param ip_address 		formData string false "IP地址，单个IP"
// @Param domain_address 	formData string false "域名地址"
// @Param content 			formData string false "域名资产的属性"
// @Param color_tag 		formData string false "颜色标记"
// @Param memo_content 		formData string false "备忘录内容"
// @Param date_delta 		formData int false "更新的日期范围"
// @Param create_date_delta formData int false "创建的日期范围"
// @Param disable_fofa 		formData bool false "禁止显示fofa的来源资产"
// @Param disable_banner 	formData bool false "禁止显示banner"
// @Param select_no_ip 		formData bool false "选择没有解析IP的资产"
// @Param select_order_by_date 	formData bool false "IP按更新日期排序"
// @Param domain_http 			formData string false "http协议中的属性"
// @Success 200 {object} models.DomainDataTableResponseData
// @router /list [post]
func (c *DomainController) List() {
	c.IsServerAPI = true
	c.ListAction()
}

// @Title Info
// @Description 聚合显示一个domain的详细信息
// @Param authorization	header string true "token"
// @Param domain 		formData string true "domain"
// @Param workspace 	formData int true "所在的workspace id"
// @Param disable_fofa 	formData bool true "是否不显示fofa等在线资产的信息"
// @Success 200 {object} models.DomainInfo
// @router /info [post]
func (c *DomainController) Info() {
	c.IsServerAPI = true
	c.InfoAction()
}

// @Title DeleteDomain
// @Description 删除一个domain
// @Param authorization	header string true "token"
// @Param id 			formData int true "id"
// @Success 200 {object} models.StatusResponseData
// @router /delete [post]
func (c *DomainController) DeleteDomain() {
	c.IsServerAPI = true
	c.DeleteDomainAction()
}

// @Title DeleteDomainAttr
// @Description 删除一个域名的属性
// @Param authorization	header string true "token"
// @Param id 			formData int true "id"
// @Success 200 {object} models.StatusResponseData
// @router /domain-attr/delete [post]
func (c *DomainController) DeleteDomainAttr() {
	c.IsServerAPI = true
	c.DeleteDomainAttrAction()
}

// @Title DeleteDomainOnlineAPIAttr
// @Description 删除fofa等属性
// @Param authorization	header string true "token"
// @Param id 			formData int true "id"
// @Success 200 {object} models.StatusResponseData
// @router /domain-api-attr/delete [post]
func (c *DomainController) DeleteDomainOnlineAPIAttr() {
	c.IsServerAPI = true
	c.DeleteDomainOnlineAPIAttrAction()
}

// @Title GetMemo
// @Description 获取指定IP的备忘录信息
// @Param authorization	header string true "token"
// @Param r_id 			formData int true "ip id"
// @Success 200 {object} models.StatusResponseData
// @router /memo [post]
func (c *DomainController) GetMemo() {
	c.IsServerAPI = true
	c.GetMemoAction()
}

// @Title UpdateMemo
// @Description 更新指定IP的备忘录信息
// @Param authorization	header string true "token"
// @Param r_id 			formData int true "domain id"
// @Param memo 			formData string true "memo内容"
// @Success 200 {object} models.StatusResponseData
// @router /memo/update [post]
func (c *DomainController) UpdateMemo() {
	c.IsServerAPI = true
	c.UpdateMemoAction()
}

// @Title MarkColor
// @Description 颜色标记
// @Param authorization	header string true "token"
// @Param r_id 			formData int true "domain id"
// @Param color 		formData string true "标记的颜色值（空或DELETE为清除标记）"
// @Success 200 {object} models.StatusResponseData
// @router /color/mark [post]
func (c *DomainController) MarkColor() {
	c.IsServerAPI = true
	c.MarkColorTagAction()
}

// @Title PinTop
// @Description 置顶/取消在列表中的置顶显示
// @Param authorization	header string true "token"
// @Param id 			formData int true "domain id"
// @Param pin_index 	formData int true "Pin值（1或0），为1/0表示置顶/取消"
// @Success 200 {object} models.StatusResponseData
// @router /pintop [post]
func (c *DomainController) PinTop() {
	c.IsServerAPI = true
	c.PinTopAction()
}

// @Title InfoHttp
// @Description 获取指定的http信息
// @Param authorization	header string true "token"
// @Param r_id 			formData int true "domain id"
// @Success 200 {object} models.StatusResponseData
// @router /http/info [post]
func (c *DomainController) InfoHttp() {
	c.IsServerAPI = true
	c.InfoHttpAction()
}
