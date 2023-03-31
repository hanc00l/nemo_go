package controllers

import (
	ctrl "github.com/hanc00l/nemo_go/pkg/web/controllers"
)

type IPController struct {
	ctrl.IPController
}

// @Title List
// @Description 根据指定筛选条件，查询IP资产
// @Param authorization		header string true "token"
// @Param start 			formData int true "查询的资产的起始行数"
// @Param length 			formData int true "返回资产指定的数量"
// @Param org_id 			formData int false "组织机构的ID"
// @Param ip_address 		formData string false "IP地址，单个IP或者掩码"
// @Param domain_address 	formData string false "域名地址"
// @Param port 				formData string false "端口，单个或多个"
// @Param content 			formData string false "IP端口的属性"
// @Param iplocation 		formData string false "IP归属地"
// @Param port_status 		formData string false "端口状态"
// @Param color_tag 		formData string false "颜色标记"
// @Param memo_content 		formData string false "备忘录内容"
// @Param date_delta 		formData int false "更新的日期范围"
// @Param create_date_delta formData int false "创建的日期范围"
// @Param disable_fofa 		formData bool false "禁止显示fofa的来源资产"
// @Param disable_banner 	formData bool false "禁止显示banner"
// @Param disable_outof_china 	formData bool false "禁止显示中国大陆以外的IP"
// @Param select_outof_china 	formData bool false "选择中国大陆以外的IP"
// @Param select_no_openedport 	formData bool false "选择没有开放端口的IP"
// @Param select_order_by_date 	formData bool false "IP按更新日期排序"
// @Param ip_http 			formData string false "http协议中的属性"
// @Success 200 {object} models.IPDataTableResponseData
// @router /list [post]
func (c *IPController) List() {
	c.IsServerAPI = true
	c.ListAction()
}

// @Title Info
// @Description 聚合显示一个IP的详细信息
// @Param authorization	header string true "token"
// @Param ip 			formData string true "ip"
// @Param workspace 	formData int true "所在的workspace id"
// @Param disable_fofa 	formData bool true "是否不显示fofa等在线资产的信息"
// @Success 200 {object} models.IPInfo
// @router /info [post]
func (c *IPController) Info() {
	c.IsServerAPI = true
	c.InfoAction()
}

// @Title DeleteIP
// @Description 删除一个IP
// @Param authorization	header string true "token"
// @Param id 			formData int true "id"
// @Success 200 {object} models.StatusResponseData
// @router /delete [post]
func (c *IPController) DeleteIP() {
	c.IsServerAPI = true
	c.DeleteIPAction()
}

// @Title DeletePortAttr
// @Description 删除一个IP的端口的属性
// @Param authorization	header string true "token"
// @Param id 			formData int true "id"
// @Success 200 {object} models.StatusResponseData
// @router /port-attr/delete [post]
func (c *IPController) DeletePortAttr() {
	c.IsServerAPI = true
	c.DeletePortAttrAction()
}

// @Title GetMemo
// @Description 获取指定IP的备忘录信息
// @Param authorization	header string true "token"
// @Param r_id 			formData int true "ip id"
// @Success 200 {object} models.StatusResponseData
// @router /memo [post]
func (c *IPController) GetMemo() {
	c.IsServerAPI = true
	c.GetMemoAction()
}

// @Title UpdateMemo
// @Description 更新指定IP的备忘录信息
// @Param authorization	header string true "token"
// @Param r_id 			formData int true "ip id"
// @Param memo 			formData string true "memo内容"
// @Success 200 {object} models.StatusResponseData
// @router /memo/update [post]
func (c *IPController) UpdateMemo() {
	c.IsServerAPI = true
	c.UpdateMemoAction()
}

// @Title MarkColor
// @Description 颜色标记
// @Param authorization	header string true "token"
// @Param r_id 			formData int true "ip id"
// @Param color 		formData string true "标记的颜色值（空或DELETE为清除标记）"
// @Success 200 {object} models.StatusResponseData
// @router /color/mark [post]
func (c *IPController) MarkColor() {
	c.IsServerAPI = true
	c.MarkColorTagAction()
}

// @Title PinTop
// @Description 置顶/取消在列表中的置顶显示
// @Param authorization	header string true "token"
// @Param id 			formData int true "ip id"
// @Param pin_index 	formData int true "Pin值（1或0），为1/0表示置顶/取消"
// @Success 200 {object} models.StatusResponseData
// @router /pintop [post]
func (c *IPController) PinTop() {
	c.IsServerAPI = true
	c.PinTopAction()
}

// @Title InfoHttp
// @Description 获取指定的http信息
// @Param authorization	header string true "token"
// @Param r_id 			formData int true "ip id"
// @Success 200 {object} models.StatusResponseData
// @router /http/info [post]
func (c *IPController) InfoHttp() {
	c.IsServerAPI = true
	c.InfoHttpAction()
}

// @Title ImportPortscanResult
// @Description 导入portscan扫描结果
// @Param authorization	header string true "token"
// @Param bin 			formData string true "扫描结果类型"
// @Param org_id 		formData int true "关联的组织id"
// @Param file 			formData string true "文件内容"
// @Success 200 {object} models.StatusResponseData
// @router /result/import [post]
func (c *IPController) ImportPortscanResult() {
	c.IsServerAPI = true
	c.ImportPortscanResultAction()
}
