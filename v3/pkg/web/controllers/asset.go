package controllers

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/hanc00l/nemo_go/v3/pkg/conf"
	"github.com/hanc00l/nemo_go/v3/pkg/core"
	"github.com/hanc00l/nemo_go/v3/pkg/db"
	"github.com/hanc00l/nemo_go/v3/pkg/logging"
	"github.com/hanc00l/nemo_go/v3/pkg/task/custom"
	"github.com/hanc00l/nemo_go/v3/pkg/task/onlineapi"
	"github.com/hanc00l/nemo_go/v3/pkg/utils"
	"go.mongodb.org/mongo-driver/v2/bson"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type AssetController struct {
	BaseController
}

type AssetListData struct {
	Id        string `json:"id"`
	Index     int    `json:"index"`
	Authority string `json:"authority"`

	Host string `json:"host"`
	//Domain         string   `json:"domain"`
	IP       []string `json:"ip"`
	Port     string   `json:"port"`
	Status   string   `json:"status"`
	Location []string `json:"location"`
	Org      string   `json:"org"`

	Service    string   `json:"service"`
	Title      string   `json:"title"`
	Header     string   `json:"header"`
	Cert       string   `json:"cert"`
	Banner     string   `json:"banner"`
	App        []string `json:"app"`
	Memo       string   `json:"memo"`
	IconHash   string   `json:"icon_hash"`
	IsNew      bool     `json:"new"`
	IsUpdate   bool     `json:"update"`
	IsCDN      bool     `json:"cdn"`
	IsHoneypot bool     `json:"honeypot"`

	ScreenshotFile []string `json:"screenshot"` //base64编码的图片文件
	IconImage      []string `json:"iconimage"`  //base64编码的icon图片文件

	Vul []string `json:"vul"` // 漏洞列表

	Icp        string `json:"icp"` // 备案信息
	IcpCompany string `json:"icp_company"`
	Whois      string `json:"whois"`

	WorkspaceId string `json:"workspace"`
	UpdateTime  string `json:"update_time"`
}

type assetRequestParam struct {
	DatableRequestParam
	TaskId         string `form:"task_id"`
	Query          string `form:"query"`
	OrgId          string `form:"org_id"`
	IsSelectNew    bool   `form:"new"`
	IsSelectUpdate bool   `form:"update"`
	IsOrderByDate  bool   `form:"order_by_date"`
}

func (c *AssetController) IndexAction() {
	c.Layout = "base.html"
	c.TplName = "asset-list.html"
}

// ListAction 获取列表显示的数据
func (c *AssetController) ListAction() {
	defer func(c *AssetController, encoding ...bool) {
		_ = c.ServeJSON(encoding...)
	}(c)
	var workspaceId, taskId string
	workspaceId = c.GetWorkspace()
	if len(workspaceId) == 0 {
		c.FailedStatus("未选择当前的工作空间！")
		return
	}
	taskId = c.GetString("task_id", "")

	req := assetRequestParam{}
	err := c.ParseForm(&req)
	if err != nil {
		logging.RuntimeLog.Error(err)
	}
	c.validateRequestParam(&req)
	resp := c.getListData(workspaceId, taskId, req)
	c.Data["json"] = resp
}

func (c *AssetController) AssetHttpContentAction() {
	defer func(c *AssetController, encoding ...bool) {
		_ = c.ServeJSON(encoding...)
	}(c)

	doc := c.getAssetDocument()
	if doc == nil {
		return
	}
	c.SucceededStatus(doc.HttpBody)
}

func (c *AssetController) AssetMemoAction() {
	defer func(c *AssetController, encoding ...bool) {
		_ = c.ServeJSON(encoding...)
	}(c)

	doc := c.getAssetDocument()
	if doc == nil {
		return
	}
	c.SucceededStatus(doc.Memo)
}

func (c *AssetController) UpdateMemoAction() {
	defer func(c *AssetController, encoding ...bool) {
		_ = c.ServeJSON(encoding...)
	}(c)
	var workspaceId, taskId string
	workspaceId = c.GetWorkspace()
	if len(workspaceId) == 0 {
		c.FailedStatus("未选择当前的工作空间！")
		return
	}
	id := c.GetString("id", "")
	taskId = c.GetString("task_id", "")
	if len(id) == 0 {
		c.FailedStatus("id参数错误！")
		return
	}
	memoContent := c.GetString("memo", "")

	mongoClient, err := db.GetClient()
	if err != nil {
		logging.RuntimeLog.Error(err)
		c.FailedStatus(err.Error())
		return
	}
	defer db.CloseClient(mongoClient)

	var asset *db.Asset
	if len(taskId) > 0 {
		asset = db.NewAsset(workspaceId, db.TaskAsset, taskId, mongoClient)
	} else {
		asset = db.NewAsset(workspaceId, db.GlobalAsset, "", mongoClient)
	}
	c.CheckErrorAndStatus(asset.Update(id, bson.M{"memo": memoContent}))
	return
}

func (c *AssetController) getAssetDocument() (doc *db.AssetDocument) {
	var workspaceId, taskId string
	workspaceId = c.GetWorkspace()
	if len(workspaceId) == 0 {
		c.FailedStatus("未选择当前的工作空间！")
		return
	}
	id := c.GetString("id", "")
	taskId = c.GetString("task_id", "")
	if len(id) == 0 {
		c.FailedStatus("id参数错误！")
		return
	}
	mongoClient, err := db.GetClient()
	if err != nil {
		logging.RuntimeLog.Error(err)
		c.FailedStatus(err.Error())
		return
	}
	defer db.CloseClient(mongoClient)

	var asset *db.Asset
	if len(taskId) > 0 {
		asset = db.NewAsset(workspaceId, db.TaskAsset, taskId, mongoClient)
	} else {
		asset = db.NewAsset(workspaceId, db.GlobalAsset, "", mongoClient)
	}
	result, err := asset.Get(id)
	if err != nil {
		logging.RuntimeLog.Error(err)
		c.FailedStatus(err.Error())
		return
	}
	return &result
}

func (c *AssetController) AssetStatisticsAction() {
	defer func(c *AssetController, encoding ...bool) {
		_ = c.ServeJSON(encoding...)
	}(c)
	//
	var workspaceId, taskId string
	workspaceId = c.GetWorkspace()
	if len(workspaceId) == 0 {
		c.FailedStatus("未选择当前的工作空间！")
		return
	}
	taskId = c.GetString("task_id", "")
	req := assetRequestParam{}
	err := c.ParseForm(&req)
	if err != nil {
		logging.RuntimeLog.Error(err)
	}
	c.validateRequestParam(&req)
	//
	mongoClient, err := db.GetClient()
	if err != nil {
		logging.RuntimeLog.Error(err)
		c.FailedStatus(err.Error())
		return
	}
	defer db.CloseClient(mongoClient)
	colName := db.GlobalAsset
	if len(taskId) > 0 {
		colName = db.TaskAsset
	}
	filter, err := c.parseQueryParam(&req)
	if err != nil {
		c.FailedStatus(err.Error())
	}
	focusAsset := db.NewAsset(workspaceId, colName, taskId, mongoClient)
	limit := 5
	//
	f := func(s string, unwind bool, ignore ...string) (r []db.StatisticData, err error) {
		i := 0
		data, err := focusAsset.Aggregate(filter, s, limit, unwind)
		if err != nil {
			return nil, err
		}
		for _, v := range data {
			if i >= limit {
				break
			}
			if len(v.Field) == 0 {
				continue
			}
			ignored := false
			for _, ignoreValue := range ignore {
				if v.Field == ignoreValue {
					ignored = true
					break
				}
			}
			if !ignored {
				i++
				r = append(r, v)
			}
		}
		return r, nil
	}
	// port忽略为0的端口（表示没有）
	result := make(map[string][]db.StatisticData)
	portData, err := f("port", false, "0")
	if err != nil {
		logging.RuntimeLog.Error(err)
		c.FailedStatus(err.Error())
		return
	}
	result["port"] = portData
	// icon需要转换为base64编码
	iconData, err := f("icon_hash_bytes", true)
	if err != nil {
		logging.RuntimeLog.Error(err)
		c.FailedStatus(err.Error())
		return
	}
	for _, v := range iconData {
		base64Img := base64.StdEncoding.EncodeToString([]byte(v.Field))
		result["icon_hash_bytes"] = append(result["icon_hash_bytes"], db.StatisticData{Field: base64Img, Count: v.Count})
	}
	// 统计app,需要unwind
	appData, err := f("app", true)
	if err != nil {
		logging.RuntimeLog.Error(err)
		c.FailedStatus(err.Error())
		return
	}
	result["app"] = appData
	// 其它统计数据
	for _, v := range []string{"title", "service"} {
		data, err := f(v, false, "", "unknown")
		if err != nil {
			logging.RuntimeLog.Error(err)
			c.FailedStatus(err.Error())
			return
		}
		result[v] = data
	}

	c.Data["json"] = result
	return
}

func (c *AssetController) DeleteAction() {
	defer func(c *AssetController, encoding ...bool) {
		_ = c.ServeJSON(encoding...)
	}(c)
	if !c.CheckMultiAccessRequest([]RequestRole{SuperAdmin, Admin}, false) {
		c.FailedStatus("无权访问！")
		return
	}

	var workspaceId, taskId string
	workspaceId = c.GetWorkspace()
	if len(workspaceId) == 0 {
		c.FailedStatus("未选择当前的工作空间！")
		return
	}
	taskId = c.GetString("task_id", "")
	id := c.GetString("id", "")
	if len(id) == 0 {
		c.FailedStatus("id参数错误！")
		return
	}
	mongoClient, err := db.GetClient()
	if err != nil {
		logging.RuntimeLog.Error(err)
		c.FailedStatus(err.Error())
		return
	}
	defer db.CloseClient(mongoClient)
	var asset *db.Asset
	if len(taskId) > 0 {
		asset = db.NewAsset(workspaceId, db.TaskAsset, taskId, mongoClient)
	} else {
		asset = db.NewAsset(workspaceId, db.GlobalAsset, "", mongoClient)
	}
	// 先读取Asset并删除
	assetDoc, err := asset.Get(id)
	if err != nil {
		logging.RuntimeLog.Error(err)
		c.FailedStatus(err.Error())
		return
	}
	if assetDoc.Id.Hex() != id {
		c.FailedStatus("id参数错误或不存在！")
		return
	}
	_, err = asset.Delete(id)
	if err != nil {
		logging.RuntimeLog.Error(err)
		c.FailedStatus(err.Error())
		return
	}
	//删除screenshot文件
	deleteScreenshotFile(workspaceId, assetDoc.Host, assetDoc.Port)

	c.SucceededStatus("删除成功！")
}

func (c *AssetController) ImportIndexAction() {
	c.Layout = "base.html"
	c.TplName = "asset-import.html"
}

func (c *AssetController) ImportSaveAction() {
	defer func(c *AssetController, encoding ...bool) {
		_ = c.ServeJSON(encoding...)
	}(c)
	if !c.CheckMultiAccessRequest([]RequestRole{SuperAdmin, Admin}, false) {
		c.FailedStatus("无权访问！")
		return
	}
	var workspaceId string
	workspaceId = c.GetWorkspace()
	if len(workspaceId) == 0 {
		c.FailedStatus("未选择当前的工作空间！")
		return
	}
	file, fileHeader, err := c.GetFile("file")
	if err != nil {
		c.FailedStatus(err.Error())
		return
	}
	// 文件后缀检查
	ext := path.Ext(fileHeader.Filename)
	if ext != ".json" && ext != ".xml" && ext != ".txt" && ext != ".csv" && ext != ".dat" {
		c.FailedStatus("只允许.json、.xml、.csv、.dat或.txt文件")
		return
	}
	// 文件转存
	templatePathPath := utils.GetTempPathFileName()
	data := make([]byte, fileHeader.Size)
	n, err := file.Read(data)
	if err != nil {
		c.FailedStatus(err.Error())
		return
	}
	if n != int(fileHeader.Size) {
		c.FailedStatus("文件读取失败,读写文件大小不一致")
		return
	}
	err = os.WriteFile(templatePathPath, data, 0666)
	if err != nil {
		c.FailedStatus("转存文件失败")
		return
	}
	defer func() {
		_ = os.RemoveAll(templatePathPath)
	}()
	// 检查导入类型
	bin := c.GetString("bin", "")
	var recordToDocFunc db.RecordToDocFunc
	if bin == "nemo" {
		recordToDocFunc = db.RecordToDoc
	} else if bin == "fofa" {
		recordToDocFunc = recordToDocForFofa
	} else {
		c.FailedStatus("未设置导入类型或不支持的导入格式")
		return
	}
	insertAndUpdate, err := c.GetBool("insert_and_update", false)
	if err != nil {
		c.FailedStatus(err.Error())
		return
	}
	orgId := c.GetString("org_id", "")
	mongoClient, err := db.GetClient()
	if err != nil {
		c.FailedStatus(err.Error())
		return
	}
	defer db.CloseClient(mongoClient)
	focusAsset := db.NewAsset(workspaceId, db.GlobalAsset, "", mongoClient)
	count, err := focusAsset.ImportFromCSV(templatePathPath, orgId, recordToDocFunc, insertAndUpdate)
	if err != nil {
		c.FailedStatus(err.Error())
		return
	}
	_ = os.RemoveAll(fileHeader.Filename)
	c.SucceededStatus(fmt.Sprintf("导入成功，共导入%d条数据", count))

	return
}

func (c *AssetController) ExportAction() {
	failFunc := func(message string) {
		logging.RuntimeLog.Error(message)
		c.FailedStatus(message)
		_ = c.ServeJSON()
	}

	if !c.CheckMultiAccessRequest([]RequestRole{SuperAdmin, Admin}, false) {
		failFunc("无权访问！")
		return
	}

	var workspaceId, taskId string
	workspaceId = c.GetWorkspace()
	if len(workspaceId) == 0 {
		failFunc("未选择当前的工作空间！")
		return
	}
	taskId = c.GetString("task_id", "")

	req := assetRequestParam{}
	err := c.ParseForm(&req)
	if err != nil {
		failFunc(err.Error())
		return
	}

	mongoClient, err := db.GetClient()
	if err != nil {
		failFunc(err.Error())
		return
	}
	defer db.CloseClient(mongoClient)
	// 根据taskId判断是全局还是任务的资产
	colName := db.GlobalAsset
	if len(taskId) > 0 {
		colName = db.TaskAsset
	}
	// 处理查询条件
	filter, err := c.parseQueryParam(&req)
	if err != nil {
		failFunc(err.Error())
		return
	}

	templateFile := fmt.Sprintf("/static/download/%s.csv", uuid.New().String())
	localFilePath := path.Join(conf.GetRootPath(), "web", templateFile)
	focusAsset := db.NewAsset(workspaceId, colName, taskId, mongoClient)
	_, err = focusAsset.ExportToCSV(filter, localFilePath, []string{})
	if err != nil {
		failFunc(err.Error())
		return
	}
	c.SucceededStatus(templateFile)
	_ = c.ServeJSON()
}

func (c *AssetController) BlockAssetAction() {
	defer func(c *AssetController, encoding ...bool) {
		_ = c.ServeJSON(encoding...)
	}(c)
	if !c.CheckMultiAccessRequest([]RequestRole{SuperAdmin, Admin}, false) {
		c.FailedStatus("无权访问！")
		return
	}
	var workspaceId, taskId string
	workspaceId = c.GetWorkspace()
	if len(workspaceId) == 0 {
		c.FailedStatus("未选择当前的工作空间！")
		return
	}
	taskId = c.GetString("task_id", "")
	id := c.GetString("id", "")
	if len(id) == 0 {
		c.FailedStatus("id参数错误！")
		return
	}
	mongoClient, err := db.GetClient()
	if err != nil {
		logging.RuntimeLog.Error(err)
		c.FailedStatus(err.Error())
		return
	}
	defer db.CloseClient(mongoClient)
	var asset *db.Asset
	if len(taskId) > 0 {
		asset = db.NewAsset(workspaceId, db.TaskAsset, taskId, mongoClient)
	} else {
		asset = db.NewAsset(workspaceId, db.GlobalAsset, "", mongoClient)
	}
	doc, err := asset.Get(id)
	if err != nil {
		logging.RuntimeLog.Error(err)
		c.FailedStatus(err.Error())
		return
	}
	if doc.Id.Hex() != id {
		c.FailedStatus("id参数错误或不存在！")
		return
	}
	// 如果不在黑名单中，则加入黑名单
	blacklist := core.NewBlacklist()
	if blacklist.LoadBlacklist(workspaceId) {
		if !blacklist.IsHostBlocked(doc.Host) {
			// 加入到增加黑名单功能
			dbCustom := db.NewCustomData(workspaceId, mongoClient)
			customDocs, err := dbCustom.Find(db.CategoryBlacklist)
			if err != nil {
				logging.RuntimeLog.Error(err)
				c.FailedStatus(err.Error())
				return
			}
			if len(customDocs) == 0 {
				c.FailedStatus("未设置黑名单，请先设置黑名单！")
				return
			}
			customDoc := customDocs[0]
			customDoc.Data = fmt.Sprintf("%s\n%s", customDoc.Data, doc.Host)
			_, err = dbCustom.Update(bson.M{"_id": customDoc.Id}, bson.M{"$set": customDoc})
			if err != nil {
				logging.RuntimeLog.Error(err)
				c.FailedStatus(err.Error())
				return
			}
		}
	}
	// 删除资产
	c.CheckErrorAndStatus(asset.DeleteByHost(doc.Host))
}

// validateRequestParam 校验请求的参数
func (c *AssetController) validateRequestParam(req *assetRequestParam) {
	if req.Length <= 0 {
		req.Length = 50
	}
	if req.Start < 0 {
		req.Start = 0
	}
}

// parseQueryParam 解析查询条件
func (c *AssetController) parseQueryParam(req *assetRequestParam) (query bson.M, err error) {
	query = bson.M{}
	if len(req.Query) > 0 {
		query, err = db.ParseQuery(req.Query)
		if err != nil {
			logging.RuntimeLog.Error(err)
			return
		}
	}
	if req.IsSelectNew {
		query["new"] = true
	}
	if req.IsSelectUpdate {
		query["update"] = true
	}
	if len(req.TaskId) > 0 {
		query["taskId"] = req.TaskId
	}
	return query, err
}

// getListData 获取列表数据
func (c *AssetController) getListData(workspaceId, taskId string, req assetRequestParam) (resp DataTableResponseData) {
	defer func() {
		resp.Draw = req.Draw
		if len(resp.Data) == 0 {
			resp.Data = make([]interface{}, 0)
		}
	}()
	mongoClient, err := db.GetClient()
	if err != nil {
		logging.RuntimeLog.Error(err)
		c.FailedStatus(err.Error())
		return
	}
	defer db.CloseClient(mongoClient)
	// 根据taskId判断是全局还是任务的资产
	colName := db.GlobalAsset
	if len(taskId) > 0 {
		colName = db.TaskAsset
	}
	// 处理查询条件
	filter, err := c.parseQueryParam(&req)
	if err != nil {
		return
	}
	focusAsset := db.NewAsset(workspaceId, colName, taskId, mongoClient)
	results, err := focusAsset.Find(filter, req.Start/req.Length+1, req.Length, req.IsOrderByDate, true)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return
	}
	orgNameMap := make(map[string]string)
	org := db.NewOrg(workspaceId, mongoClient)
	honeypot := core.NewHoneypot(workspaceId)
	vul := db.NewVul(workspaceId, db.GlobalVul, mongoClient)
	icpStringMap := make(map[string]string)
	icpCompanyMap := make(map[string]string)
	queryData := db.NewQueryData(db.GlobalDatabase, mongoClient)
	whoisMap := make(map[string]string)

	for i, result := range results {
		asset := AssetListData{
			Id:          result.Id.Hex(),
			Index:       req.Start + 1 + i,
			Authority:   result.Authority,
			Host:        result.Host,
			Status:      result.HttpStatus,
			Service:     result.Service,
			Title:       result.Title,
			Header:      result.HttpHeader,
			Cert:        result.Cert,
			Banner:      result.Banner,
			App:         result.App,
			Memo:        result.Memo,
			IsNew:       result.IsNewAsset,
			IsUpdate:    result.IsUpdated,
			IsCDN:       result.IsCDN,
			IconHash:    result.IconHash,
			IsHoneypot:  honeypot.IsHoneypot(result.Host),
			WorkspaceId: workspaceId,
			UpdateTime:  result.UpdateTime.In(conf.LocalTimeLocation).Format("2006-01-02 15:04:05"),
		}
		if len(result.OrgId) > 0 {
			if _, ok := orgNameMap[result.OrgId]; !ok {
				orgDoc, err := org.Get(result.OrgId)
				if err != nil {
					orgNameMap[result.OrgId] = ""
				} else {
					orgNameMap[result.OrgId] = orgDoc.Name
				}
			}
			asset.Org = orgNameMap[result.OrgId]
		}
		// 处理IP和Location
		locationMap := make(map[string]struct{})
		if len(result.Ip.IpV4) > 0 {
			for _, ip := range result.Ip.IpV4 {
				asset.IP = append(asset.IP, ip.IPName)
				// 对IP进行honeypot检测，进一步确保host为域名的时候不漏过检测
				if honeypot.IsHoneypot(ip.IPName) {
					asset.IsHoneypot = true
				}
				if len(ip.Location) > 0 {
					locationMap[ip.Location] = struct{}{}
				}
			}
		}
		if len(result.Ip.IpV6) > 0 {
			for _, ip := range result.Ip.IpV6 {
				asset.IP = append(asset.IP, ip.IPName)
				// 对IP进行honeypot检测，进一步确保host为域名的时候不漏过检测
				if honeypot.IsHoneypot(ip.IPName) {
					asset.IsHoneypot = true
				}
				if len(ip.Location) > 0 {
					locationMap[ip.Location] = struct{}{}
				}
			}
		}
		if len(locationMap) > 0 {
			for location := range locationMap {
				asset.Location = append(asset.Location, location)
			}
		}
		if result.Port > 0 {
			asset.Port = strconv.Itoa(result.Port)
		}
		// 处理icon图片
		if len(result.IconHashBytes) > 0 {
			base64Img := base64.StdEncoding.EncodeToString(result.IconHashBytes)
			asset.IconImage = append(asset.IconImage, base64Img)
		}
		// 处理screenshot图片
		asset.ScreenshotFile = loadScreenshotFile(workspaceId, result.Host, result.Port)
		// 处理漏洞列表
		vulDocs, _ := vul.Find(bson.M{"authority": result.Authority}, 0, 0)
		for _, v := range vulDocs {
			asset.Vul = append(asset.Vul, v.PocFile)
		}
		//icp:
		if result.Category == db.CategoryDomain && len(result.Domain) > 0 {
			if _, ok := icpStringMap[result.Domain]; !ok {
				icpResult, errIcp := queryData.GetByDomain(result.Domain, db.QueryICP)
				if errIcp == nil {
					var icpInfo onlineapi.ICPInfo
					errIcp = json.Unmarshal([]byte(icpResult.Content), &icpInfo)
					if errIcp == nil && len(icpInfo.Domain) > 0 {
						icpStringMap[result.Domain] = icpResult.Content
						icpCompanyMap[result.Domain] = icpInfo.CompanyName
					}
				}
			}
			if _, ok := icpCompanyMap[result.Domain]; ok {
				asset.IcpCompany = icpCompanyMap[result.Domain]
			}
			if _, ok := icpStringMap[result.Domain]; ok {
				asset.Icp = icpStringMap[result.Domain]
			}
		}
		//whois
		if result.Category == db.CategoryDomain && len(result.Domain) > 0 {
			if _, ok := whoisMap[result.Domain]; !ok {
				whoisResult, errWhois := queryData.GetByDomain(result.Domain, db.QueryWhois)
				if errWhois == nil {
					whoisMap[result.Domain] = whoisResult.Content
				}
			}
			if _, ok := whoisMap[result.Domain]; ok {
				asset.Whois = whoisMap[result.Domain]
			}
		}
		resp.Data = append(resp.Data, asset)
	}
	count, _ := focusAsset.Count(filter)
	resp.RecordsTotal = count
	resp.RecordsFiltered = count

	return
}

// LoadScreenshotFile 获取screenshot文件
func loadScreenshotFile(workspaceId, host string, port int) (r []string) {
	if port <= 0 {
		return // 没有端口，不加载screenshot
	} else {
		// host:port
		files, _ := filepath.Glob(filepath.Join(conf.GlobalServerConfig().Web.WebFiles, workspaceId, "screenshot", host, fmt.Sprintf("%d_*.png", port)))
		for _, file := range files {
			_, f := filepath.Split(file)
			if !strings.HasSuffix(f, "_thumbnail.png") {
				r = append(r, f)
			}
		}
	}
	return
}

// LoadScreenshotFile 获取screenshot文件
func deleteScreenshotFile(workspaceId, host string, port int) (r []string) {
	if port <= 0 {
		return
	} else {
		// ip:port
		files, _ := filepath.Glob(filepath.Join(conf.GlobalServerConfig().Web.WebFiles, workspaceId, "screenshot", host, fmt.Sprintf("%d_*.png", port)))
		for _, file := range files {
			err := os.Remove(file)
			if err != nil {
				logging.RuntimeLog.Error(err)
			}
		}
	}
	return
}

// recordToDocForFofa 将CSV记录转换为AssetDocument
func recordToDocForFofa(headers, record []string, orgId string) (*db.AssetDocument, error) {
	now := time.Now()
	doc := &db.AssetDocument{
		Id:         bson.NewObjectID(),
		OrgId:      orgId,
		CreateTime: now,
		UpdateTime: now,
	}
	ipl := custom.NewIPv4Location("")

	for i, header := range headers {
		if i >= len(record) {
			continue // 跳过不完整的记录
		}

		value := record[i]
		if value == "" {
			continue // 跳过空值
		}
		// 这里巨坑，fofa的csv文件前三个字节是 0xefbbbf，0xefbbbf 是 UTF-8 编码中的一个特殊字节序列，它表示的是 UTF-8 编码的字节序标记（BOM，Byte Order Mark）
		if i == 0 {
			header = "host"
		}
		switch header {
		case "host":
			doc.Host = utils.ParseHost(value)
		case "ip":
			ipv4 := db.IPV4{
				IPName:   value,
				IPInt:    utils.IPV4ToUInt32(value),
				Location: ipl.FindPublicIP(value),
			}
			doc.Ip.IpV4 = append(doc.Ip.IpV4, ipv4)
		case "port":
			if port, err := strconv.Atoi(value); err == nil {
				doc.Port = port
			}
		case "protocol":
			doc.Service = value
		case "title":
			doc.Title = value
		case "domain":
			doc.Domain = value
		case "server":
			doc.Server = value
		case "header":
			doc.HttpHeader = value
		case "cert":
			doc.Cert = value
		case "product":
			doc.App = append(doc.App, strings.Split(value, ",")...)
		case "product_category":
			doc.App = append(doc.App, strings.Split(value, ",")...)
		}
	}
	if len(doc.Host) == 0 {
		return nil, errors.New("host is empty")
	}
	if len(doc.Ip.IpV4) == 0 && len(doc.Ip.IpV6) == 0 {
		return nil, errors.New("ip is empty")
	}
	if utils.CheckIPV4(doc.Host) {
		doc.Category = db.CategoryIPv4
	} else if utils.CheckIPV6(doc.Host) {
		doc.Category = db.CategoryIPv6
	} else {
		doc.Category = db.CategoryDomain
	}
	doc.Authority = fmt.Sprintf("%s:%d", doc.Host, doc.Port)

	return doc, nil
}
