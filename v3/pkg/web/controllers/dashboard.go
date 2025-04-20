package controllers

import (
	"github.com/hanc00l/nemo_go/v3/pkg/db"
	"github.com/hanc00l/nemo_go/v3/pkg/logging"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type DashboardController struct {
	BaseController
}

type WorkspaceList struct {
	Name        string `json:"name"`
	WorkspaceId string `json:"id"`
}

type AssetStatistics struct {
	TotalAssetCount  int `json:"total"`
	IpAssetCount     int `json:"ip"`
	DomainAssetCount int `json:"domain"`
	VulCount         int `json:"vul"`
}

func (c *DashboardController) IndexAction() {
	// 用于前端展示当前用户的工作空间
	c.Data[Workspace] = c.GetSession(Workspace)

	c.TplName = "dashboard.html"
	c.Layout = "base.html"
}

func (c *DashboardController) GetUserAvailableWorkspaceAction() (workspaceList []WorkspaceList) {
	defer func(c *DashboardController, encoding ...bool) {
		c.Data["json"] = workspaceList
		_ = c.ServeJSON(encoding...)
	}(c)

	userName := c.GetCurrentUser()
	if len(userName) == 0 {
		return
	}
	workspaceList = GetUserAvailableWorkspaceList(userName)
	// 如果是超级管理员，并且没有设置默认的工作空间，则读取全部可用的工作空间
	if len(workspaceList) == 0 && c.GetSession(UserRole) == SuperAdmin {
		workspaceList = GetSuperAdminDefaultWorkspaceList()
	}

	return
}

func (c *DashboardController) ChangeUserWorkspaceAction() {
	defer func(c *DashboardController, encoding ...bool) {
		_ = c.ServeJSON(encoding...)
	}(c)

	workspaceId := c.GetString("workspaceId", "")
	if len(workspaceId) == 0 {
		c.FailedStatus("切换工作区失败，缺少参数")
		return
	}
	userName := c.GetCurrentUser()
	if len(userName) == 0 {
		c.FailedStatus("切换工作区失败，用户未登录")
		return
	}
	workspaceList := GetUserAvailableWorkspaceList(userName)
	// 如果是超级管理员，并且没有设置默认的工作空间，则读取全部可用的工作空间
	if len(workspaceList) == 0 && c.GetSession(UserRole) == SuperAdmin {
		workspaceList = GetSuperAdminDefaultWorkspaceList()
	}
	for _, workspace := range workspaceList {
		if workspace.WorkspaceId == workspaceId {
			_ = c.SetSession(Workspace, workspaceId)
			c.SucceededStatus("切换工作区成功")
			return
		}
	}
	c.FailedStatus("切换工作区失败，用户没有该工作区权限")
}

func (c *DashboardController) StatisticsAssetAction() {
	defer func(c *DashboardController, encoding ...bool) {
		_ = c.ServeJSON(encoding...)
	}(c)

	workspaceId := c.GetString("workspaceId", "")
	if len(workspaceId) == 0 {
		c.FailedStatus("统计资产失败，缺少参数")
		return
	}
	mongoClient, err := db.GetClient()
	if err != nil {
		c.FailedStatus("统计资产失败，数据库连接失败")
		return
	}
	defer db.CloseClient(mongoClient)

	focusAsset := db.NewAsset(workspaceId, db.GlobalAsset, "", mongoClient)
	results, err := focusAsset.Aggregate(bson.M{}, "category", 0, false)
	if err != nil {
		c.FailedStatus("统计资产失败，数据库查询失败")
		return
	}
	var assetStatistics AssetStatistics
	for _, result := range results {
		if result.Field == "ipv4" || result.Field == "ipv6" {
			assetStatistics.IpAssetCount += result.Count
		} else if result.Field == "domain" {
			assetStatistics.DomainAssetCount += result.Count
		}
	}
	assetStatistics.TotalAssetCount = assetStatistics.IpAssetCount + assetStatistics.DomainAssetCount
	// 统计漏洞数量
	assetStatistics.VulCount, _ = db.NewVul(workspaceId, db.GlobalVul, mongoClient).Count(bson.M{})

	c.Data["json"] = assetStatistics
}

func (c *DashboardController) AssetChartDataAction() {
	defer func(c *DashboardController, encoding ...bool) {
		_ = c.ServeJSON(encoding...)
	}(c)

	workspaceId := c.GetString("workspaceId", "")
	if len(workspaceId) == 0 {
		c.FailedStatus("生成资产统计图表失败，缺少参数")
		return
	}
	mongoClient, err := db.GetClient()
	if err != nil {
		c.FailedStatus("生成资产统计图表失败，数据库连接失败")
		return
	}
	defer db.CloseClient(mongoClient)

	asset := db.NewAsset(workspaceId, db.GlobalAsset, "", mongoClient)
	charData, err := asset.GenerateChartData(15)
	if err != nil {
		logging.RuntimeLog.Errorf("生成资产统计图表失败: %s", err)
		c.FailedStatus("生成资产统计图表失败")
		return
	}

	c.Data["json"] = charData
}

func (c *DashboardController) VulChartDataAction() {
	defer func(c *DashboardController, encoding ...bool) {
		_ = c.ServeJSON(encoding...)
	}(c)

	workspaceId := c.GetString("workspaceId", "")
	if len(workspaceId) == 0 {
		c.FailedStatus("生成资产统计图表失败，缺少参数")
		return
	}
	mongoClient, err := db.GetClient()
	if err != nil {
		c.FailedStatus("生成资产统计图表失败，数据库连接失败")
		return
	}
	defer db.CloseClient(mongoClient)

	vul := db.NewVul(workspaceId, db.GlobalVul, mongoClient)
	charData, err := vul.GetDailyStats(15)
	if err != nil {
		logging.RuntimeLog.Errorf("生成资产统计图表失败: %s", err)
		c.FailedStatus("生成资产统计图表失败")
		return
	}

	c.Data["json"] = charData
}

func (c *DashboardController) TaskChartDataAction() {
	defer func(c *DashboardController, encoding ...bool) {
		_ = c.ServeJSON(encoding...)
	}(c)

	workspaceId := c.GetString("workspaceId", "")
	if len(workspaceId) == 0 {
		c.FailedStatus("生成资产统计图表失败，缺少参数")
		return
	}
	mongoClient, err := db.GetClient()
	if err != nil {
		c.FailedStatus("生成资产统计图表失败，数据库连接失败")
		return
	}
	defer db.CloseClient(mongoClient)

	executorTask := db.NewExecutorTask(mongoClient)
	charData, err := executorTask.GetDailyStats(workspaceId, 15)
	if err != nil {
		logging.RuntimeLog.Errorf("生成资产统计图表失败: %s", err)
		c.FailedStatus("生成资产统计图表失败")
		return
	}

	c.Data["json"] = charData
}

func (c *DashboardController) UserChartDataAction() {
	defer func(c *DashboardController, encoding ...bool) {
		_ = c.ServeJSON(encoding...)
	}(c)

	mongoClient, err := db.GetClient()
	if err != nil {
		c.FailedStatus("生成资产统计图表失败，数据库连接失败")
		return
	}
	defer db.CloseClient(mongoClient)

	dbLog := db.NewRuntimeLog(mongoClient)
	charData, err := dbLog.GetLoginStats(15)
	if err != nil {
		logging.RuntimeLog.Errorf("生成资产统计图表失败: %s", err)
		c.FailedStatus("生成资产统计图表失败")
		return
	}

	c.Data["json"] = charData
}
