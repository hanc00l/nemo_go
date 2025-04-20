package controllers

import (
	"github.com/hanc00l/nemo_go/v3/pkg/db"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type CustomController struct {
	BaseController
}
type CustomData struct {
	Id          string `json:"id" form:"id"`
	Category    string `json:"category" form:"category"`
	Description string `json:"description" form:"description"`
	Data        string `json:"data" form:"data"`
}

func (c *CustomController) IndexAction() {
	c.CheckMultiAccessRequest([]RequestRole{SuperAdmin, Admin}, true)

	c.Layout = "base.html"
	c.TplName = "custom-info.html"
}

func (c *CustomController) LoadAction() {
	defer func(c *CustomController, encoding ...bool) {
		_ = c.ServeJSON(encoding...)
	}(c)

	if c.CheckMultiAccessRequest([]RequestRole{SuperAdmin, Admin}, false) == false {
		c.FailedStatus("当前用户权限不允许！")
		return
	}
	workspaceId := c.GetWorkspace()
	if len(workspaceId) == 0 {
		c.FailedStatus("未选择当前的工作空间！")
		return
	}
	category := c.GetString("category")
	if len(category) == 0 {
		c.FailedStatus("未指定分类！")
		return
	}
	data, err := c.loadCustomData(workspaceId, category)
	if err != nil {
		c.FailedStatus("加载自定义数据失败！")
		return
	}
	c.Data["json"] = data

}

func (c *CustomController) SaveAction() {
	defer func(c *CustomController, encoding ...bool) {
		_ = c.ServeJSON(encoding...)
	}(c)
	if c.CheckMultiAccessRequest([]RequestRole{SuperAdmin, Admin}, false) == false {
		c.FailedStatus("当前用户权限不允许！")
		return
	}
	workspaceId := c.GetWorkspace()
	if len(workspaceId) == 0 {
		c.FailedStatus("未选择当前的工作空间！")
		return
	}
	var data CustomData
	if err := c.ParseForm(&data); err != nil {
		c.FailedStatus("解析参数失败！")
		return
	}
	mongoClient, err := db.GetClient()
	if err != nil {
		c.FailedStatus("数据库连接失败！")
		return
	}
	defer db.CloseClient(mongoClient)

	cd := db.NewCustomData(workspaceId, mongoClient)
	if len(data.Id) == 0 {
		if len(data.Category) == 0 {
			c.FailedStatus("未指定分类！")
			return
		}
		doc := db.CustomDataDocument{
			Category:    data.Category,
			Description: data.Description,
			Data:        data.Data,
		}
		c.CheckErrorAndStatus(cd.Insert(doc))
	} else {
		_idd, err := bson.ObjectIDFromHex(data.Id)
		if err != nil {
			c.FailedStatus("参数错误！")
			return
		}
		filter := bson.M{"_id": _idd}
		update := bson.M{"$set": bson.M{"data": data.Data, "description": data.Description}}
		c.CheckErrorAndStatus(cd.Update(filter, update))
	}

}

func (c *CustomController) loadCustomData(workspaceId, category string) (data CustomData, err error) {
	mongoClient, err := db.GetClient()
	if err != nil {
		return data, err
	}
	defer db.CloseClient(mongoClient)

	cd := db.NewCustomData(workspaceId, mongoClient)
	results, err := cd.Find(category)
	if err != nil {
		return data, err
	}
	if len(results) == 0 {
		return data, nil
	}
	data = CustomData{
		Id:          results[0].Id.Hex(),
		Category:    results[0].Category,
		Description: results[0].Description,
		Data:        results[0].Data,
	}

	return data, nil
}
