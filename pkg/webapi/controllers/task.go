package controllers

import (
	ctrl "github.com/hanc00l/nemo_go/pkg/web/controllers"
)

type TaskController struct {
	ctrl.TaskController
}

// @Title ListMainTask
// @Description 任务列表的数据
// @Param authorization		header string true "token"
// @Param start 			formData int true "查询起始行数"
// @Param length 			formData int true "返回指定的数量"
// @Param task_state 		formData string false "任务状态"
// @Param task_name 		formData string false "任务名称"
// @Param task_args 		formData string false "任务参数"
// @Param task_worker 		formData string false "任务执行的worker"
// @Param cron_id 			formData string false "计划任务ID"
// @Param show_runtask 		formData bool false "是否显示运行任务"
// @Param runtask_state 	formData string false "运行任务的状态"
// @Success 200 {object} models.TaskDataTableResponseData
// @router /main/list [post]
func (c *TaskController) ListMainTask() {
	c.IsServerAPI = true
	c.ListAction()
}

// @Title ListCronTask
// @Description 计划任务列表的数据
// @Param authorization		header string true "token"
// @Param start 			formData int true "查询起始行数"
// @Param length 			formData int true "返回指定的数量"
// @Param task_name 		formData string false "任务名称"
// @Param task_args 		formData string false "任务参数"
// @Param task_status 		formData string false "任务状态"
// @Success 200 {object} models.TaskDataTableResponseData
// @router /cron/list [post]
func (c *TaskController) ListCronTask() {
	c.IsServerAPI = true
	c.ListCronAction()
}

// @Title InfoMainTask
// @Description 显示一个MainTask任务的详情
// @Param authorization		header string true "token"
// @Param task_id 			formData string true "任务ID"
// @Success 200 {object} models.TaskInfo
// @router /main/info [post]
// InfoMainAction
func (c *TaskController) InfoMainTask() {
	c.IsServerAPI = true
	c.InfoMainAction()
}

// @Title InfoRunTask
// @Description 显示一个RunTask任务的详情
// @Param authorization		header string true "token"
// @Param task_id 			formData string true "任务ID"
// @Success 200 {object} models.TaskInfo
// @router /run/info [post]
// InfoMainAction
func (c *TaskController) InfoRunTask() {
	c.IsServerAPI = true
	c.InfoAction()
}

// @Title InfoCronTask
// @Description 显示一个CronTask任务的详情
// @Param authorization		header string true "token"
// @Param task_id 			formData string true "任务ID"
// @Success 200 {object} models.TaskCronInfo
// @router /cron/info [post]
// InfoMainAction
func (c *TaskController) InfoCronTask() {
	c.IsServerAPI = true
	c.InfoCronAction()
}

// @Title DeleteRunTask
// @Description 删除一个RunTask任务记录
// @Param authorization	header string true "token"
// @Param id 			formData int true "id"
// @Success 200 {object} models.StatusResponseData
// @router /run/delete [post]
func (c *TaskController) DeleteRunTask() {
	c.IsServerAPI = true
	c.DeleteAction()
}

// @Title DeleteMainTask
// @Description 删除一个MainTask任务记录
// @Param authorization	header string true "token"
// @Param id 			formData int true "id"
// @Success 200 {object} models.StatusResponseData
// @router /main/delete [post]
func (c *TaskController) DeleteMainTask() {
	c.IsServerAPI = true
	c.DeleteMainAction()
}

// @Title DeleteCronTask
// @Description 删除一个CronTask任务记录
// @Param authorization	header string true "token"
// @Param id 			formData int true "id"
// @Success 200 {object} models.StatusResponseData
// @router /cron/delete [post]
func (c *TaskController) DeleteCronTask() {
	c.IsServerAPI = true
	c.DeleteCronAction()
}

// @Title DeleteBatchTask
// @Description 批量删除指定类型的任务记录
// @Param authorization	header string true "token"
// @Param type 			formData string true "任务类型"
// @Success 200 {object} models.StatusResponseData
// @router /batch-delete [post]
func (c *TaskController) DeleteBatchTask() {
	c.IsServerAPI = true
	c.DeleteBatchAction()
}

// @Title StopRunTask
// @Description 取消一个未开始执行的任务
// @Param authorization	header string true "token"
// @Param task_id 		formData int true "task id"
// @Success 200 {object} models.StatusResponseData
// @router /run/stop [post]
func (c *TaskController) StopRunTask() {
	c.IsServerAPI = true
	c.StopAction()
}

// @Title DisableCronTask
// @Description 禁用一个计划任务
// @Param authorization	header string true "token"
// @Param task_id 		formData int true "task id"
// @Success 200 {object} models.StatusResponseData
// @router /cron/disable [post]
func (c *TaskController) DisableCronTask() {
	c.IsServerAPI = true
	c.DisableCronTaskAction()
}

// @Title EnableCronTask
// @Description 启用一个禁用的计划任务
// @Param authorization	header string true "token"
// @Param task_id 		formData int true "task id"
// @Success 200 {object} models.StatusResponseData
// @router /cron/enable [post]
func (c *TaskController) EnableCronTask() {
	c.IsServerAPI = true
	c.EnableCronTaskAction()
}

// @Title RunCronTask
// @Description 立即执行一个计划任务
// @Param authorization	header string true "token"
// @Param task_id 		formData int true "task id"
// @Success 200 {object} models.StatusResponseData
// @router /cron/run [post]
func (c *TaskController) RunCronTask() {
	c.IsServerAPI = true
	c.RunCronTaskAction()
}

/*
type XScanRequestParam struct {
	XScanType       string `form:"xscan_type"`
	Target          string `form:"target"`
	Port            string `form:"port"`
	OrgId           int    `form:"org_id"`
	IsOrgIP         bool
	IsOrgDomain     bool
	IsOnlineAPI    bool   `form:"fofa"`
	IsFingerprint   bool   `form:"fingerprint"`
	IsXrayPocscan   bool   `form:"xraypoc"`
	XrayPocFile     string `form:"xraypocfile"`
	IsTaskCron      bool   `form:"taskcron" json:"-"`
	TaskCronRule    string `form:"cronrule" json:"-"`
	TaskCronComment string `form:"croncomment" json:"-"`
	IsCn            bool   `form:"is_CN"`
}
*/
// @Title StartXScanTask
// @Description 执行一个XScan任务
// @Param authorization	header string true "token"
// @Param target 		formData string true "任务目标(ip、ip/掩码或域名），多个任务以,分开"
// @Param port 			formData string false "ip目标扫描的端口"
// @Param org_id 		formData int false "关联的组机构"
// @Param onlineapi 	formData bool false "是否要执行fofa、quake、hunter等任务"
// @Param xraypoc 		formData bool false "是否要执行xraypoc扫描"
// @Param xraypocfile 	formData string false "xraypoc使用的pocfile，格式为\"poc类型|poc文件名\"；poc类型为default或custom，poc文件名可为空（全部poc）或xray支持的模糊匹配方式"
// @Param nucleipoc 	formData bool false "是否要执行nuclei扫描"
// @Param nucleipocfile formData string false "nucleipoc使用的pocfile"
// @Param taskcron 		formData bool false "是否为计划任务"
// @Param cronrule 		formData string false "计划任务的规则"
// @Param croncomment 	formData string false "计划任务的名称"
// @Success 200 {object} models.StatusResponseData
// @router /xscan [post]
func (c *TaskController) StartXScanTask() {
	c.IsServerAPI = true
	c.StartXScanTaskAction()
}
