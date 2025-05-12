package core

import (
	"github.com/hanc00l/nemo_go/v3/pkg/db"
	"github.com/hanc00l/nemo_go/v3/pkg/logging"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
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
	//logging.RuntimeLog.Infof("同步任务历史资产, workspaceId:%s, taskId:%s", workspaceId, taskId)
	taskAssetDocs, err := taskAsset.Find(bson.M{db.TaskId: taskId}, 0, 0, false, false)
	if err != nil {
		logging.RuntimeLog.Errorf("获取任务资产失败,taskId:%s,err:%v", taskId, err.Error())
		return
	}
	var assetSaveResult AssetSaveResultResp
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
