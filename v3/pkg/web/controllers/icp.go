package controllers

import (
	"fmt"
	"github.com/hanc00l/nemo_go/v3/pkg/db"
	"github.com/hanc00l/nemo_go/v3/pkg/logging"
	"github.com/hanc00l/nemo_go/v3/pkg/task/execute"
	"github.com/hanc00l/nemo_go/v3/pkg/task/icp"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type ICPController struct {
	BaseController
}

type icpRequestParam struct {
	DatableRequestParam
	Domain   string `form:"domain"`
	UnitName string `form:"unit_name"`
}
type ICPData struct {
	Id             string `json:"id" form:"id"`
	Index          int    `json:"index" form:"-"`
	UnitName       string `json:"unit_name" form:"unit_name"`
	Domain         string `json:"domain" form:"domain"`
	CompanyType    string `json:"company_type" form:"company_type"`
	SiteLicense    string `json:"site_license" form:"site_license"`
	ServiceLicence string `json:"service_licence" form:"service_licence"`
	VerifyTime     string `json:"verify_time" form:"verify_time"`
	CreateDatetime string `json:"create_time" form:"-"`
	UpdateDatetime string `json:"update_time" form:"-"`
}

func (c *ICPController) IndexAction() {
	c.TplName = "icp-list.html"
	c.Layout = "base.html"
}

// ListAction 获取列表显示的数据
func (c *ICPController) ListAction() {
	defer func(c *ICPController, encoding ...bool) {
		_ = c.ServeJSON(encoding...)
	}(c)

	req := icpRequestParam{}
	err := c.ParseForm(&req)
	if err != nil {
		logging.RuntimeLog.Error(err)
	}
	c.validateRequestParam(&req)
	resp := c.getListData(req)
	c.Data["json"] = resp
}

// validateRequestParam 校验请求的参数
func (c *ICPController) validateRequestParam(req *icpRequestParam) {
	if req.Length <= 0 {
		req.Length = 50
	}
	if req.Start < 0 {
		req.Start = 0
	}
}

// getListData 获取列表数据
func (c *ICPController) getListData(req icpRequestParam) (resp DataTableResponseData) {
	defer func() {
		resp.Draw = req.Draw
		if len(resp.Data) == 0 {
			resp.Data = make([]interface{}, 0)
		}
	}()
	mongoClient, err := db.GetClient()
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		c.FailedStatus(err.Error())
		return
	}
	defer db.CloseClient(mongoClient)

	filter := bson.M{}
	if len(req.UnitName) > 0 {
		filter["unitName"] = bson.M{"$regex": req.UnitName, "$options": "i"}
	}
	if len(req.Domain) > 0 {
		filter["domain"] = bson.M{"$regex": req.Domain, "$options": "i"}
	}
	u := db.NewICP(mongoClient)
	startPage := req.Start/req.Length + 1
	results, err := u.Find(filter, startPage, req.Length)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		c.FailedStatus(err.Error())
		return
	}
	for i, row := range results {
		r := ICPData{
			Id:             row.Id.Hex(),
			Index:          i + 1,
			UnitName:       row.UnitName,
			Domain:         row.Domain,
			CompanyType:    row.CompanyType,
			SiteLicense:    row.SiteLicense,
			ServiceLicence: row.ServiceLicence,
			VerifyTime:     row.VerifyTime,
			CreateDatetime: row.CreateTime.Format("2006-01-02 15:04:05"),
			UpdateDatetime: row.UpdateTime.Format("2006-01-02 15:04:05"),
		}
		resp.Data = append(resp.Data, r)
	}
	total, _ := u.Count(filter)
	resp.RecordsTotal = total
	resp.RecordsFiltered = total

	return
}

// DeleteAction 删除一条记录
func (c *ICPController) DeleteAction() {
	defer func(c *ICPController, encoding ...bool) {
		_ = c.ServeJSON(encoding...)
	}(c)

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

	c.CheckErrorAndStatus(db.NewICP(mongoClient).Delete(id))

	return
}

func (c *ICPController) OnlineAPISearchAction() {
	defer func(c *ICPController, encoding ...bool) {
		_ = c.ServeJSON(encoding...)
	}(c)

	key := c.GetString("key")
	if len(key) == 0 {
		c.FailedStatus("请输入查询的关键字！")
		return
	}
	msg, err := doICPOnlineAPISearch(key)
	if err != nil {
		c.FailedStatus(msg)
		return
	}
	c.SucceededStatus(msg)
	return
}

func doICPOnlineAPISearch(key string) (msg string, err error) {
	executorConfig := execute.ExecutorConfig{
		ICP: map[string]execute.ICPConfig{
			"icp": {
				APIName: []string{"beianx"},
			},
		},
	}
	taskInfo := execute.ExecutorTaskInfo{
		MainTaskInfo: execute.MainTaskInfo{
			WorkspaceId:    db.GlobalDatabase,
			MainTaskId:     "icp_onlineapi",
			ExecutorConfig: executorConfig,
			Target:         key,
		},
		Executor: "icp",
	}
	result := icp.Do(taskInfo, true)
	if len(result) == 0 {
		return "未查询到相关信息！", fmt.Errorf("未查询到相关信息！")
	}
	mongoClient, err := db.GetClient()
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		return err.Error(), err
	}
	defer db.CloseClient(mongoClient)

	i := db.NewICP(mongoClient)
	for _, r := range result {
		doc := db.ICPDocument{
			Domain:         r.Domain,
			UnitName:       r.UnitName,
			CompanyType:    r.CompanyType,
			SiteLicense:    r.SiteLicense,
			ServiceLicence: r.ServiceLicence,
			VerifyTime:     r.VerifyTime,
		}
		_, err = i.InsertOrUpdate(doc)
		if err != nil {
			logging.RuntimeLog.Error(err.Error())
			return err.Error(), err
		}
	}
	return fmt.Sprintf("在线查询成功，共查询到%d条信息！", len(result)), nil
}
