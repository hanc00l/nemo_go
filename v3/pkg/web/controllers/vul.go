package controllers

import (
	"github.com/hanc00l/nemo_go/v3/pkg/db"
	"github.com/hanc00l/nemo_go/v3/pkg/logging"
	"github.com/hanc00l/nemo_go/v3/pkg/task/execute"
	"github.com/hanc00l/nemo_go/v3/pkg/task/pocscan"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type VulController struct {
	BaseController
}

type vulRequestParam struct {
	DatableRequestParam
	Host     string `json:"host" form:"host"`
	PocFile  string `json:"pocfile" form:"pocfile"`
	Source   string `json:"source" form:"source"`
	Severity string `json:"severity" form:"severity"`
}

type VulData struct {
	Id             string `json:"id"`
	Index          int    `json:"index"`
	Authority      string `json:"authority"`
	Url            string `json:"url"`
	PocFile        string `json:"pocfile"`
	Severity       string `json:"severity"`
	Source         string `json:"source"`
	TaskId         string `json:"task_id"`
	CreateTime     string `json:"create_time"`
	UpdateDatetime string `json:"update_time"`
}

type VulDetailDataInfo struct {
	Id             string `json:"id"`
	Authority      string `json:"authority"`
	Url            string `json:"url"`
	PocFile        string `json:"pocfile"`
	Severity       string `json:"severity"`
	Name           string `json:"name"`
	Source         string `json:"source"`
	Extra          string `json:"extra"`
	CreateTime     string `json:"create_time"`
	UpdateDatetime string `json:"update_time"`
}

func (c *VulController) IndexAction() {
	c.TplName = "vul-list.html"
	c.Layout = "base.html"
}

// ListAction 获取列表显示的数据
func (c *VulController) ListAction() {
	defer func(c *VulController, encoding ...bool) {
		_ = c.ServeJSON(encoding...)
	}(c)

	req := vulRequestParam{}
	err := c.ParseForm(&req)
	if err != nil {
		logging.RuntimeLog.Error(err)
	}
	c.validateRequestParam(&req)
	resp := c.getListData(req)
	c.Data["json"] = resp
}

// validateRequestParam 校验请求的参数
func (c *VulController) validateRequestParam(req *vulRequestParam) {
	if req.Length <= 0 {
		req.Length = 50
	}
	if req.Start < 0 {
		req.Start = 0
	}
}

// getListData 获取列表数据
func (c *VulController) getListData(req vulRequestParam) (resp DataTableResponseData) {
	defer func() {
		resp.Draw = req.Draw
		if len(resp.Data) == 0 {
			resp.Data = make([]interface{}, 0)
		}
	}()
	workspaceId := c.GetWorkspace()
	if len(workspaceId) == 0 {
		c.FailedStatus("未选择当前的工作空间！")
		return
	}
	mongoClient, err := db.GetClient()
	if err != nil {
		logging.RuntimeLog.Error(err)
		c.FailedStatus(err.Error())
		return
	}
	defer db.CloseClient(mongoClient)

	filter := bson.M{}
	if req.Host != "" {
		filter["host"] = req.Host
	}
	if req.PocFile != "" {
		filter["pocfile"] = bson.M{"$regex": req.PocFile, "$options": "i"}
	}
	if req.Source != "" {
		filter["source"] = req.Source
	}
	if req.Severity != "" {
		filter["severity"] = req.Severity
	}
	taskId := c.GetString("task_id")
	var collectionName string
	if taskId != "" {
		filter["task_id"] = taskId
		collectionName = db.TaskVul
	} else {
		collectionName = db.GlobalVul
	}
	vul := db.NewVul(workspaceId, collectionName, mongoClient)
	startPage := req.Start/req.Length + 1
	results, err := vul.Find(filter, startPage, req.Length)
	if err != nil {
		logging.RuntimeLog.Error(err)
		c.FailedStatus(err.Error())
		return
	}
	for i, row := range results {
		task := VulData{
			Id:             row.Id.Hex(),
			Index:          req.Start + i + 1,
			Authority:      row.Authority,
			Url:            row.Url,
			PocFile:        row.PocFile,
			Source:         row.Source,
			Severity:       row.Severity,
			TaskId:         row.TaskId,
			CreateTime:     FormatDateTime(row.CreateTime),
			UpdateDatetime: FormatDateTime(row.UpdateTime),
		}
		resp.Data = append(resp.Data, task)
	}
	total, _ := vul.Count(filter)
	resp.RecordsTotal = total
	resp.RecordsFiltered = total

	return
}

// DeleteAction 删除一条记录
func (c *VulController) DeleteAction() {
	defer func(c *VulController, encoding ...bool) {
		_ = c.ServeJSON(encoding...)
	}(c)

	workspaceId := c.GetWorkspace()
	if len(workspaceId) == 0 {
		c.FailedStatus("未选择当前的工作空间！")
		return
	}
	id := c.GetString("id")
	if len(id) == 0 {
		logging.RuntimeLog.Error("empty id")
		c.FailedStatus("empty id")
		return
	}
	mongoClient, err := db.GetClient()
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		c.FailedStatus(err.Error())
		return
	}
	defer db.CloseClient(mongoClient)

	c.CheckErrorAndStatus(db.NewVul(workspaceId, db.GlobalVul, mongoClient).Delete(id))

	return
}

func (c *VulController) LoadPocFileAction() {
	defer func(c *VulController, encoding ...bool) {
		_ = c.ServeJSON(encoding...)
	}(c)

	pocBin := c.GetString("poc_bin")
	if len(pocBin) == 0 {
		logging.RuntimeLog.Error("empty poc bin")
		c.FailedStatus("empty poc bin")
		return
	}
	executor := pocscan.NewExecutor(pocBin, execute.PocscanConfig{}, false)
	if executor == nil {
		c.FailedStatus("invalid poc bin")
		return
	}
	pocFiles := executor.LoadPocFiles()
	type PocFile struct {
		Name string `json:"name"`
		Text string `json:"text"`
	}
	var pocList []PocFile
	for _, v := range pocFiles {
		pocList = append(pocList, PocFile{
			Name: v,
			Text: v,
		})
	}
	c.Data["json"] = pocList
}

func (c *VulController) InfoIndexAction() {
	c.TplName = "vul-info.html"
	c.Layout = "base.html"
}

func (c *VulController) InfoAction() {
	defer func(c *VulController, encoding ...bool) {
		_ = c.ServeJSON(encoding...)
	}(c)

	id := c.GetString("id")
	if len(id) == 0 {
		logging.RuntimeLog.Error("empty id")
		c.FailedStatus("empty id")
		return
	}
	workspaceId := c.GetWorkspace()
	if len(workspaceId) == 0 {
		c.FailedStatus("未选择当前的工作空间！")
		return
	}
	mongoClient, err := db.GetClient()
	if err != nil {
		logging.RuntimeLog.Error(err)
		c.FailedStatus(err.Error())
		return
	}
	defer db.CloseClient(mongoClient)
	taskId := c.GetString("task_id")
	var collectionName string
	if taskId != "" {
		collectionName = db.TaskVul
	} else {
		collectionName = db.GlobalVul
	}
	vul := db.NewVul(workspaceId, collectionName, mongoClient)
	doc, err := vul.Get(id)
	if err != nil {
		logging.RuntimeLog.Error(err)
		c.FailedStatus(err.Error())
		return
	}
	if doc.Id.Hex() != id {
		c.FailedStatus("未找到该漏洞详情！")
		return
	}
	infoData := VulDetailDataInfo{
		Id:             doc.Id.Hex(),
		Authority:      doc.Authority,
		Url:            doc.Url,
		PocFile:        doc.PocFile,
		Source:         doc.Source,
		Severity:       doc.Severity,
		Name:           doc.Name,
		Extra:          doc.Extra,
		CreateTime:     FormatDateTime(doc.CreateTime),
		UpdateDatetime: FormatDateTime(doc.UpdateTime),
	}
	c.Data["json"] = infoData
	return
}
