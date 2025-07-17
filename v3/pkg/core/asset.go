package core

import (
	"encoding/json"
	"github.com/hanc00l/nemo_go/v3/pkg/db"
	"github.com/hanc00l/nemo_go/v3/pkg/logging"
	"github.com/hanc00l/nemo_go/v3/pkg/task/llmapi"
	"github.com/hanc00l/nemo_go/v3/pkg/utils"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"os"
)

func SyncTaskAsset(workspaceId string, taskId string) (result string) {
	mongoClient, err := db.GetClient()
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		return
	}
	defer db.CloseClient(mongoClient)

	globalAsset := db.NewAsset(workspaceId, db.GlobalAsset, "", mongoClient)
	taskAsset := db.NewAsset(workspaceId, db.TaskAsset, taskId, mongoClient)
	//分页同步任务历史资产：
	totalCount, err := taskAsset.Count(bson.M{db.TaskId: taskId})
	if err != nil {
		logging.RuntimeLog.Errorf("获取任务资产数量失败, taskId:%s, %s", taskId, err.Error())
		return
	}
	//logging.RuntimeLog.Infof("同步任务历史资产, workspaceId:%s, taskId:%s，总数量:%d", workspaceId, taskId, totalCount)
	pageSize := 100
	pageCount := calculatePageCount(totalCount, pageSize)
	var assetSaveResult AssetSaveResultResp
	for i := 0; i < pageCount; i++ {
		taskAssetDocs, err := taskAsset.Find(bson.M{db.TaskId: taskId}, i+1, pageSize, true, false)
		if err != nil {
			logging.RuntimeLog.Errorf("获取任务资产失败, taskId:%s, err:%v", taskId, err.Error())
			return
		}
		for _, taskDoc := range taskAssetDocs {
			taskDoc.Id = bson.NewObjectID() // 重新生成id
			taskDoc.TaskId = ""             // 保存到全局focusAsset中必须要去掉task字段
			globalDss, err := globalAsset.InsertOrUpdate(taskDoc)
			if err != nil {
				logging.RuntimeLog.Errorf("同步任务资产失败, docId:%s, err:%v", taskDoc.Id, err)
				continue
			}
			if !globalDss.IsSuccess {
				logging.RuntimeLog.Errorf("同步任务资产失败, docId:%s", taskDoc.Id)
				continue
			}
			assetSaveResult.AssetTotal++
			if globalDss.IsNew {
				assetSaveResult.AssetNew++
			} else if globalDss.IsUpdated {
				assetSaveResult.AssetUpdate++
			}
		}
	}
	// 同步漏洞：
	assetSaveResult.VulNew, assetSaveResult.VulUpdate = SyncTaskHistoryVul(workspaceId, taskId, mongoClient)
	assetSaveResult.VulTotal = assetSaveResult.VulNew + assetSaveResult.VulUpdate

	return assetSaveResult.String()
}

func SyncTaskHistoryVul(workspaceId string, taskId string, mongoClient *mongo.Client) (newVul, updateVul int) {
	// 同步漏洞：
	globalVul := db.NewVul(workspaceId, db.GlobalVul, mongoClient)
	taskVul := db.NewVul(workspaceId, db.TaskVul, mongoClient)
	//logging.RuntimeLog.Infof("同步任务漏洞,workspaceId:%s, taskId:%s", workspaceId, taskId)

	taskVulDocs, err := taskVul.Find(bson.M{db.TaskId: taskId}, 0, 0)
	if err != nil {
		logging.RuntimeLog.Errorf("获取任务漏洞失败,taskId:%s,err:%v", taskId, err.Error())
		return
	}
	for _, newDoc := range taskVulDocs {
		oldDocs, _ := globalVul.Find(bson.M{"authority": newDoc.Authority, "url": newDoc.Url, "source": newDoc.Source, "pocfile": newDoc.PocFile}, 0, 0)
		newDoc.TaskId = "" // 保存到全局中必须要去掉task字段
		var isSuccess bool
		if len(oldDocs) == 0 {
			isSuccess, err = globalVul.Insert(newDoc)
		} else {
			newDoc.Id = oldDocs[0].Id
			newDoc.CreateTime = oldDocs[0].CreateTime
			isSuccess, err = globalVul.Update(oldDocs[0].Id.Hex(), newDoc)
		}
		if err != nil {
			logging.RuntimeLog.Errorf("保存漏洞失败, docId:%s,err:%v", newDoc.Authority, err.Error())
			continue
		}
		if !isSuccess {
			logging.RuntimeLog.Errorf("保存漏洞失败, docId:%s", newDoc.Authority)
			continue
		}
		if len(oldDocs) == 0 {
			newVul++
		} else {
			updateVul++
		}
	}

	return
}

func calculatePageCount(totalCount, pageSize int) int {
	if pageSize <= 0 {
		return 0 // 或根据业务需求返回其他值
	}
	if totalCount <= 0 {
		return 0
	}
	return (totalCount + pageSize - 1) / pageSize
}

func GenerateReport(workspaceId string, id, taskId string, llmapiName string) {
	mongoClient, err := db.GetClient()
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		return
	}
	defer db.CloseClient(mongoClient)

	assetPrompt := getTaskAssetCsvPrompt(workspaceId, taskId, mongoClient)
	if len(assetPrompt) == 0 {
		return
	}
	vulPrompt := getTaskVulJsonPrompt(workspaceId, taskId, mongoClient)
	// 调用接口生成报告
	systemPrompt, userPrompt := GetPrompt()
	userAllPrompt := userPrompt + assetPrompt + vulPrompt
	logging.RuntimeLog.Infof("调用LLMAPI生成任务报告, taskId:%s，请求长度:%d", taskId, len(systemPrompt+userAllPrompt))
	report, err := llmapi.DoCallAPI(llmapiName, systemPrompt, userAllPrompt)
	if err != nil {
		logging.RuntimeLog.Errorf("生成任务报告失败, taskId:%s, err:%v", taskId, err)
		return
	}
	//　更新任务资产报告
	task := db.NewMainTask(mongoClient)
	update := bson.M{}
	update["report"] = report
	isSuccess, err := task.Update(id, update)
	if err != nil {
		logging.RuntimeLog.Errorf("更新任务报告失败, taskId:%s, err:%v", taskId, err)
		return
	}
	if !isSuccess {
		logging.RuntimeLog.Errorf("更新任务报告失败, taskId:%s", taskId)
		return
	}
	logging.RuntimeLog.Infof("生成任务报告成功, taskId:%s，返回长度:%d", taskId, len(report))
}

func getReportFieldNames() []string {
	return []string{
		"authority", "host", "port", "category",
		"ip.ipv4.ip", "ip.ipv4.location", "ip.ipv6.ip", "ip.ipv6.location",
		"domain", "service", "server", "banner", "title", "app",
		"status",
	}
}

func GetPrompt() (systemPrompt, userPrompt string) {
	userPrompt = `
请根据本次执行的信息收集任务的结果，按指定模板生成一份报告；报告格式为独立完整的HTML文件，具有良好的可读性；返回的结果只包含纯HTML,不要包含其它信息和Markdown格式。
============
以下是报告模板的主要部份
一、域名及子域名情况
主域名分析
主要的子域名分析
二、IP情况
IP收集汇总（去重）
热点IP
三、端口开放情况
四、暴露的服务与应用
五、其他主要发现
六、漏洞情况（如果有，没有的话可以省略）
`
	systemPrompt = "你是一名网络安全研究员，在对授权的对象进行渗透测试，通过Nemo执行了资产的信息收集任务。"
	return
}

func getTaskAssetCsvPrompt(workspaceId string, taskId string, mongoClient *mongo.Client) (csvResult string) {
	templateFile := utils.GetTempPathFileName()
	defer func() {
		_ = os.Remove(templateFile)
	}()
	// 导出任务资产到csv文件
	taskAsset := db.NewAsset(workspaceId, db.TaskAsset, taskId, mongoClient)
	csvCount, err := taskAsset.ExportToCSV(bson.M{db.TaskId: taskId}, templateFile, getReportFieldNames())
	if err != nil {
		logging.RuntimeLog.Errorf("导出任务资产失败, taskId:%s, err:%v", taskId, err)
		return
	}
	if csvCount == 0 {
		logging.RuntimeLog.Warnf("导出任务资产失败, taskId:%s, 没有数据", taskId)
		return
	}
	csvData, err := os.ReadFile(templateFile)
	if err != nil {
		logging.RuntimeLog.Errorf("读取csv文件失败：%s", err.Error())
		return
	}
	prompt := `============
以下是csv格式的任务的结果数据：
	`
	return prompt + string(csvData)

}

func getTaskVulJsonPrompt(workspaceId string, taskId string, mongoClient *mongo.Client) (vulJsonResult string) {
	taskVul := db.NewVul(workspaceId, db.TaskVul, mongoClient)
	taskVulDocs, err := taskVul.Find(bson.M{db.TaskId: taskId}, 0, 0)
	if err != nil {
		logging.RuntimeLog.Errorf("获取任务漏洞失败,taskId:%s,err:%v", taskId, err.Error())
		return
	}
	if len(taskVulDocs) == 0 {
		return
	}
	// 去除一些不必要的字段
	for i := 0; i < len(taskVulDocs); i++ {
		taskVulDocs[i].TaskId = ""
		taskVulDocs[i].Extra = ""
	}
	jsonData, err := json.Marshal(taskVulDocs)
	if err != nil {
		logging.RuntimeLog.Errorf("序列化任务漏洞失败,taskId:%s,err:%v", taskId, err.Error())
		return
	}
	prompt := `============
以下是json格式的任务的漏洞数据：
`
	return prompt + string(jsonData)
}
