package controllers

import (
	"encoding/csv"
	"fmt"
	"github.com/google/uuid"
	"github.com/hanc00l/nemo_go/v3/pkg/conf"
	"github.com/hanc00l/nemo_go/v3/pkg/db"
	"github.com/hanc00l/nemo_go/v3/pkg/logging"
	"github.com/hanc00l/nemo_go/v3/pkg/task/unit"
	"github.com/hanc00l/nemo_go/v3/pkg/utils"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"os"
	"path"
	"strings"
	"time"
)

type UnitController struct {
	BaseController
}

type unitRequestParam struct {
	DatableRequestParam
	ParentUnitName string `form:"parent_unit_name"`
	UnitName       string `form:"unit_name"`
	IsFuzzy        bool   `form:"fuzzy"`
}

type UnitData struct {
	Id             string `json:"id" form:"id"`
	Index          int    `json:"index" form:"-"`
	UnitName       string `json:"unit_name" form:"unit_name"`
	ParentUnitName string `json:"parent_unit_name" form:"parent_unit_name"`
	Status         string `json:"status" form:"status"`
	EntId          string `json:"ent_id" form:"ent_id"`
	Type           string `json:"type" form:"type"`
	InvestRation   string `json:"invest_ration" form:"invest_ration"`
	ICPData        int    `json:"icp_data" form:"icp_data"`
	CreateDatetime string `json:"create_time" form:"-"`
	UpdateDatetime string `json:"update_time" form:"-"`
}

type UnitOnlineSearchData struct {
	UnitName     string `json:"unit_name" form:"unit_name"`
	IsBranch     bool   `json:"is_branch" form:"is_branch"`
	IsInvest     bool   `json:"is_invest" form:"is_invest"`
	InvestRation string `json:"invest_ration" form:"invest_ration"`
	IsICPOnline  bool   `json:"is_icp_online" form:"is_icp_online"`
	MaxDepth     int    `json:"max_depth" form:"max_depth"`
	Cookie       string `json:"cookie" form:"cookie"`
}

func (c *UnitController) IndexAction() {
	c.TplName = "unit-list.html"
	c.Layout = "base.html"
}

// ListAction 获取列表显示的数据
func (c *UnitController) ListAction() {
	defer func(c *UnitController, encoding ...bool) {
		_ = c.ServeJSON(encoding...)
	}(c)

	req := unitRequestParam{}
	err := c.ParseForm(&req)
	if err != nil {
		logging.RuntimeLog.Error(err)
	}
	c.validateRequestParam(&req)
	resp := c.getListData(req)
	c.Data["json"] = resp
}

// validateRequestParam 校验请求的参数
func (c *UnitController) validateRequestParam(req *unitRequestParam) {
	if req.Length <= 0 {
		req.Length = 50
	}
	if req.Start < 0 {
		req.Start = 0
	}
}

// getListData 获取列表数据
func (c *UnitController) getListData(req unitRequestParam) (resp DataTableResponseData) {
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
	if req.IsFuzzy {
		if len(req.ParentUnitName) > 0 {
			filter["parentUnitName"] = bson.M{"$regex": req.ParentUnitName, "$options": "i"}
		}
		if len(req.UnitName) > 0 {
			filter["unitName"] = bson.M{"$regex": req.UnitName, "$options": "i"}
		}
	} else {
		if len(req.ParentUnitName) > 0 {
			filter["parentUnitName"] = req.ParentUnitName
		}
		if len(req.UnitName) > 0 {
			filter["unitName"] = req.UnitName
		}
	}
	u := db.NewUnit(mongoClient)
	startPage := req.Start/req.Length + 1
	results, err := u.Find(filter, startPage, req.Length)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		c.FailedStatus(err.Error())
		return
	}
	icpApp := db.NewICP(mongoClient)
	for i, row := range results {
		r := UnitData{
			Id:             row.Id.Hex(),
			Index:          req.Start + i + 1,
			UnitName:       row.UnitName,
			ParentUnitName: row.ParentUnitName,
			Status:         row.Status,
			EntId:          row.EntId,
			CreateDatetime: FormatDateTime(row.UpdateTime),
			UpdateDatetime: FormatDateTime(row.CreateTime),
		}
		if row.IsBranch {
			r.Type = "分支机构"
		} else if row.IsInvest {
			r.Type = "对外投资"
			if row.InvestRation > 0 {
				r.InvestRation = fmt.Sprintf("%.2f%%", row.InvestRation)
			}
		}
		icpCount, err := icpApp.Count(bson.M{"unitName": row.UnitName})
		if err == nil && icpCount > 0 {
			r.ICPData = icpCount
		}
		resp.Data = append(resp.Data, r)
	}
	total, _ := u.Count(filter)
	resp.RecordsTotal = total
	resp.RecordsFiltered = total

	return
}

// DeleteAction 删除一条记录
func (c *UnitController) DeleteAction() {
	defer func(c *UnitController, encoding ...bool) {
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

	c.CheckErrorAndStatus(db.NewUnit(mongoClient).Delete(id))

	return
}

func (c *UnitController) OnlineAPISearchAction() {
	defer func(c *UnitController, encoding ...bool) {
		_ = c.ServeJSON(encoding...)
	}(c)

	req := UnitOnlineSearchData{}
	err := c.ParseForm(&req)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return
	}
	if len(req.UnitName) == 0 {
		c.FailedStatus("请输入单位名称！")
		return
	}
	if len(req.Cookie) == 0 {
		c.FailedStatus("请输入cookie！")
		return
	}
	if req.MaxDepth <= 0 {
		req.MaxDepth = 1
	}
	if req.MaxDepth > 3 {
		req.MaxDepth = 3
	}
	investRation := 100.0
	if req.IsInvest {
		investRation, err = unit.ParsePercentage(req.InvestRation)
		if err != nil {
			c.FailedStatus(err.Error())
			return
		}
	}
	entryList := make([]unit.UnitEntry, 0)
	entryList = append(entryList, unit.UnitEntry{
		CompanyName: req.UnitName,
		Depth:       unit.Root,
	})
	errGet := unit.GetAllUnit(&entryList, unit.Root, req.MaxDepth, req.IsBranch, req.IsInvest, investRation, req.Cookie)
	if errGet != nil {
		logging.RuntimeLog.Error(errGet.Error())
	}
	count, unitSet, err := saveUnitData("", &entryList)
	if err != nil {
		c.FailedStatus(err.Error())
		return
	}
	if errGet != nil {
		c.SucceededStatus(fmt.Sprintf("共获取数据:%d，发生错误:%s！", count, errGet.Error()))
	} else {
		c.SucceededStatus(fmt.Sprintf("共获取数据:%d", count))
	}
	if req.IsICPOnline && len(unitSet) > 0 {
		logging.RuntimeLog.Infof("开始获取ICP在线数据,共%d个单位...", len(unitSet))
		go func() {
			msg, err := doICPOnlineAPISearch(strings.Join(utils.SetToSlice(unitSet), ","))
			if err != nil {
				logging.RuntimeLog.Error(err.Error())
			} else {
				if len(msg) > 0 {
					logging.RuntimeLog.Info(msg)
				}
			}
		}()
	}
}

func (c *UnitController) GetAction() {
	defer func(c *UnitController, encoding ...bool) {
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
		logging.RuntimeLog.Error(err)
		c.FailedStatus(err.Error())
		return
	}
	defer db.CloseClient(mongoClient)

	u := db.NewUnit(mongoClient)
	doc, err := u.Get(id)
	if err != nil {
		logging.RuntimeLog.Error(err)
		c.FailedStatus(err.Error())
		return
	}
	if doc.Id.Hex() != id {
		c.FailedStatus("未找到对象")
		return
	}
	infoData := UnitData{
		Id:             doc.Id.Hex(),
		UnitName:       doc.UnitName,
		ParentUnitName: doc.ParentUnitName,
	}
	if doc.IsBranch {
		infoData.Type = "branch"
	} else if doc.IsInvest {
		infoData.Type = "invest"
		if doc.InvestRation > 0 {
			infoData.InvestRation = fmt.Sprintf("%.2f", doc.InvestRation)
		}
	}
	c.Data["json"] = infoData
	return
}

func (c *UnitController) UpdateAction() {
	defer func(c *UnitController, encoding ...bool) {
		_ = c.ServeJSON(encoding...)
	}(c)
	reqData := UnitData{}
	err := c.ParseForm(&reqData)
	if err != nil {
		logging.RuntimeLog.Error(err)
		c.FailedStatus(err.Error())
		return
	}
	if len(reqData.UnitName) == 0 {
		c.FailedStatus("请输入单位名称！")
		return
	}
	if len(reqData.Id) == 0 {
		c.FailedStatus("请输入单位ID！")
		return
	}
	investRation := 100.0
	if reqData.Type == "invest" {
		if len(reqData.InvestRation) == 0 {
			c.FailedStatus("请输入投资比例！")
			return
		}
		investRation, err = unit.ParsePercentage(reqData.InvestRation)
		if err != nil {
			c.FailedStatus(err.Error())
			return
		}
	}
	mongoClient, err := db.GetClient()
	if err != nil {
		logging.RuntimeLog.Error(err)
		c.FailedStatus(err.Error())
		return
	}
	defer db.CloseClient(mongoClient)

	u := db.NewUnit(mongoClient)
	doc, err := u.Get(reqData.Id)
	if err != nil {
		logging.RuntimeLog.Error(err)
		c.FailedStatus(err.Error())
		return
	}
	if doc == nil || doc.Id.Hex() != reqData.Id {
		c.FailedStatus("未找到对象")
		return
	}
	doc.UnitName = reqData.UnitName
	doc.ParentUnitName = reqData.ParentUnitName
	doc.IsBranch = reqData.Type == "branch"
	doc.IsInvest = reqData.Type == "invest"
	if doc.IsInvest {
		doc.InvestRation = investRation
	}
	_, err = u.Update(doc.Id, *doc)
	if err != nil {
		logging.RuntimeLog.Error(err)
		c.FailedStatus(err.Error())
		return
	}
	c.SucceededStatus("更新成功！")
}

func saveUnitData(parentUnitName string, unitList *[]unit.UnitEntry) (count int, unitSet map[string]struct{}, err error) {
	mongoClient, err := db.GetClient()
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		return 0, nil, err
	}
	defer db.CloseClient(mongoClient)
	u := db.NewUnit(mongoClient)
	now := time.Now()
	savedUnitMap := make(map[string]struct{})
	unitSet = make(map[string]struct{})
	for _, unitEntry := range *unitList {
		// 有效的公司
		if unitEntry.CompanyName == unitEntry.Company.EntName {
			unitSet[unitEntry.CompanyName] = struct{}{}
			// 避免重复保存
			if _, ok := savedUnitMap[fmt.Sprintf("%s-%s", unitEntry.ParentName, unitEntry.CompanyName)]; !ok {
				unitDoc := db.UnitDocument{
					UnitName:       unitEntry.CompanyName,
					ParentUnitName: unitEntry.ParentName,
					Status:         unitEntry.Company.EntStatus,
					EntId:          unitEntry.Company.Entid,
					CreateTime:     now,
					UpdateTime:     now,
				}
				// 重定位到上级单位
				if unitEntry.ParentName == "" {
					unitDoc.ParentUnitName = parentUnitName
				}
				_, err = u.InsertOrUpdate(unitDoc)
				if err != nil {
					logging.RuntimeLog.Errorf("保存单位数据失败：%s", err.Error())
					return
				}
				count++
			}
			//   保存分支机构
			for _, branch := range unitEntry.BranchList {
				branchDoc := db.UnitDocument{
					UnitName:       branch.BrName,
					ParentUnitName: unitEntry.CompanyName,
					Status:         branch.EntStatus,
					EntId:          branch.Entid,
					IsBranch:       true,
					CreateTime:     now,
					UpdateTime:     now,
				}
				_, err = u.InsertOrUpdate(branchDoc)
				if err != nil {
					logging.RuntimeLog.Errorf("保存分支机构数据失败：%s", err.Error())
					return
				}
				count++
				savedUnitMap[fmt.Sprintf("%s-%s", branchDoc.ParentUnitName, branchDoc.UnitName)] = struct{}{}
				unitSet[branchDoc.UnitName] = struct{}{}
			}
			// 保存投资机构
			for _, invest := range unitEntry.InvestList {
				percent, _ := unit.ParsePercentage(invest.FundedRatio)
				investDoc := db.UnitDocument{
					UnitName:       invest.EntJgName,
					ParentUnitName: unitEntry.CompanyName,
					Status:         invest.EntStatus,
					EntId:          invest.Entid,
					IsInvest:       true,
					InvestRation:   percent,
					CreateTime:     now,
					UpdateTime:     now,
				}
				_, err = u.InsertOrUpdate(investDoc)
				if err != nil {
					logging.RuntimeLog.Errorf("保存投资机构数据失败：%s", err.Error())
					return
				}
				count++
				savedUnitMap[fmt.Sprintf("%s-%s", investDoc.ParentUnitName, investDoc.UnitName)] = struct{}{}
				unitSet[investDoc.UnitName] = struct{}{}
			}
		}
	}

	return count, unitSet, nil
}

func (c *UnitController) ExportAction() {
	defer func(c *UnitController, encoding ...bool) {
		_ = c.ServeJSON(encoding...)
	}(c)

	req := unitRequestParam{}
	err := c.ParseForm(&req)
	if err != nil {
		logging.RuntimeLog.Error(err)
		c.FailedStatus(err.Error())
		return
	}
	filter := bson.M{}
	if req.IsFuzzy {
		if len(req.ParentUnitName) > 0 {
			filter["parentUnitName"] = bson.M{"$regex": req.ParentUnitName, "$options": "i"}
		}
		if len(req.UnitName) > 0 {
			filter["unitName"] = bson.M{"$regex": req.UnitName, "$options": "i"}
		}
	} else {
		if len(req.ParentUnitName) > 0 {
			filter["parentUnitName"] = req.ParentUnitName
		}
		if len(req.UnitName) > 0 {
			filter["unitName"] = req.UnitName
		}
	}
	templateFile := fmt.Sprintf("/static/download/%s.csv", uuid.New().String())
	localFilePath := path.Join(conf.GetRootPath(), "web", templateFile)
	_, err = exportToCSV(filter, localFilePath)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		c.FailedStatus(err.Error())
	} else {
		c.SucceededStatus(templateFile)
	}
	return
}

func exportToCSV(filter bson.M, outputFile string) (int64, error) {
	mongoClient, err := db.GetClient()
	if err != nil {
		return 0, err
	}
	defer db.CloseClient(mongoClient)
	// 当前查询条件的记录
	u := db.NewUnit(mongoClient)
	docs, err := u.Find(filter, 0, 0)
	if err != nil {
		return 0, err
	}
	// 递归查询所有子公司记录
	var exportDocs []db.UnitDocument
	exportDocs = append(exportDocs, docs...)
	err = exportRecursiveToCSV(mongoClient, &docs, &exportDocs)
	//
	i := db.NewICP(mongoClient)
	count, err := docsToRecord(exportDocs, outputFile, i, getAllFieldNames())
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		return 0, err
	}

	return count, nil
}

func exportRecursiveToCSV(mongoClient *mongo.Client, currentDocs *[]db.UnitDocument, outputDocs *[]db.UnitDocument) error {
	for _, doc := range *currentDocs {
		nextDocs, err := db.NewUnit(mongoClient).Find(bson.M{"parentUnitName": doc.UnitName}, 0, 0)
		if err != nil {
			return err
		}
		if len(nextDocs) > 0 {
			*outputDocs = append(*outputDocs, nextDocs...)
			err = exportRecursiveToCSV(mongoClient, &nextDocs, outputDocs)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func docsToRecord(docs []db.UnitDocument, outputFile string, i *db.ICP, fields []string) (int64, error) {
	// 3. 创建CSV文件
	file, err := os.Create(outputFile)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 4. 处理字段选择
	selectedFields := getAllFieldNames()

	// 5. 写入CSV头
	if err := writer.Write(selectedFields); err != nil {
		return 0, err
	}

	// 6. 遍历结果并写入CSV
	var count int64
	for _, doc := range docs {
		record, err := docToRecord(&doc, i, selectedFields)
		if err != nil {
			return count, err
		}

		if err := writer.Write(record); err != nil {
			return count, err
		}
		count++
	}

	return count, nil
}

// docToRecord 将文档转换为CSV记录
func docToRecord(doc *db.UnitDocument, i *db.ICP, fields []string) ([]string, error) {
	record := make([]string, len(fields))
	for p, field := range fields {
		switch field {
		case "parentUnitName":
			record[p] = doc.ParentUnitName
		case "unitName":
			record[p] = doc.UnitName
		case "status":
			record[p] = doc.Status
		case "entId":
			record[p] = doc.EntId
		case "type":
			if doc.IsBranch {
				record[p] = "分支机构"
			} else if doc.IsInvest {
				record[p] = "对外投资"
			} else {
				record[p] = ""
			}
		case "investRation":
			if doc.IsInvest {
				record[p] = fmt.Sprintf("%.2f", doc.InvestRation)
			}
		case "icp":
			icpDocs, err := i.GetByCompany(doc.UnitName)
			if err == nil && len(icpDocs) > 0 {
				domains := make([]string, 0)
				for _, icp := range icpDocs {
					domains = append(domains, icp.Domain)
				}
				record[p] = strings.Join(domains, ",")
			} else {
				record[p] = ""
			}
		case "create_time":
			record[p] = FormatDateTime(doc.CreateTime)
		case "update_time":
			record[p] = FormatDateTime(doc.UpdateTime)
		default:
			return nil, fmt.Errorf("unknown field: %s", field)
		}
	}

	return record, nil
}

// getAllFieldNames 返回所有可能的字段名
func getAllFieldNames() []string {
	return []string{
		"parentUnitName", "unitName", "status", "entId",
		"type", "investRation", "icp",
		"create_time", "update_time",
	}
}
