package controllers

import ctrl "github.com/hanc00l/nemo_go/pkg/web/controllers"

type DashboardController struct {
	ctrl.DashboardController
}

// @Title GetStatisticData
// @Description 获取统计信息
// @Param authorization		header string true "token"
// @Success 200 {object} models.DashboardStatisticData
// @router /statistic [post]
func (c *DashboardController) GetStatisticData() {
	c.IsServerAPI = true
	c.GetStatisticDataAction()
}

// @Title GetTaskInfo
// @Description 获取任务数据
// @Param authorization		header string true "token"
// @Success 200 {object} models.TaskInfoData
// @router /task [post]
func (c *DashboardController) GetTaskInfo() {
	c.IsServerAPI = true
	c.GetTaskInfoAction()
}

// @Title WorkerAliveList
// @Description 获取worker数据，用于dashboard列表显示
// @Param authorization		header string true "token"
// @Success 200 {object} models.WorkerStatusData
// @router /worker/list [post]
func (c *DashboardController) WorkerAliveList() {
	c.IsServerAPI = true
	c.WorkerAliveListAction()
}

// @Title ManualReloadWorker
// @Description ManualReloadWorkerAction 重启worker
// @Param authorization		header string true "token"
// @Param worker_name		formData string true "worker name"
// @Success 200 {object} models.WorkspaceDataTableResponseData
// @router /worker/reload [post]
func (c *DashboardController) ManualReloadWorker() {
	c.IsServerAPI = true
	c.ManualReloadWorkerAction()
}

// @Title ManualWorkerFileSync
// @Description 同步worker文件
// @Param authorization		header string true "token"
// @Param worker_name		formData string true "worker name"
// @Success 200 {object} models.WorkspaceDataTableResponseData
// @router /worker/filesync [post]
func (c *DashboardController) ManualWorkerFileSync() {
	c.IsServerAPI = true
	c.ManualWorkerFileSyncAction()
}

// @Title OnlineUserList
// @Description 获取在线用户数据，用于Dashboard表表显示
// @Param authorization		header string true "token"
// @Success 200 {object} models.OnlineUserDataTableResponseData
// @router /user [post]
func (c *DashboardController) OnlineUserList() {
	c.IsServerAPI = true
	c.OnlineUserListAction()
}
