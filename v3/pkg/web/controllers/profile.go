package controllers

import (
	"encoding/json"
	"errors"
	"github.com/hanc00l/nemo_go/v3/pkg/db"
	"github.com/hanc00l/nemo_go/v3/pkg/logging"
	"github.com/hanc00l/nemo_go/v3/pkg/task/execute"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type ProfileController struct {
	BaseController
}

type llmapiData struct {
	IsEnable bool                 `json:"enabled" form:"enabled"`
	Qwen     bool                 `json:"qwen" form:"qwen"`
	Kimi     bool                 `json:"kimi" form:"kimi"`
	Deepseek bool                 `json:"deepseek" form:"deepseek"`
	ICPPlus  bool                 `json:"icpPlus" form:"icpPlus"`
	Config   execute.LLMAPIConfig `json:"config" form:"config"`
}

type portscanData struct {
	IsEnable bool                   `json:"enabled" form:"enabled"`
	Nmap     bool                   `json:"nmap" form:"nmap"`
	Masscan  bool                   `json:"masscan" form:"masscan"`
	Gogo     bool                   `json:"gogo" form:"gogo"`
	Config   execute.PortscanConfig `json:"config" form:"config"`
}

type domainscanData struct {
	IsEnable  bool                     `json:"enabled" form:"enabled"`
	Massdns   bool                     `json:"massdns" form:"massdns"`
	Subfinder bool                     `json:"subfinder" form:"subfinder"`
	Config    execute.DomainscanConfig `json:"config" form:"config"`
}
type onlineapiData struct {
	IsEnable bool                    `json:"enabled" form:"enabled"`
	Fofa     bool                    `json:"fofa" form:"fofa"`
	Hunter   bool                    `json:"hunter" form:"hunter"`
	Quake    bool                    `json:"quake" form:"quake"`
	Whois    bool                    `json:"whois" form:"whois"`
	ICP      bool                    `json:"icp" form:"icp"`
	Config   execute.OnlineAPIConfig `json:"config" form:"config"`
}

type pocscanData struct {
	IsEnable bool                  `json:"enabled" form:"enabled"`
	PocBin   string                `json:"pocbin" form:"pocbin"`
	Config   execute.PocscanConfig `json:"config" form:"config"`
}

type standaloneData struct {
	Config execute.StandaloneScanConfig `json:"config" form:"config"`
}

type profileRequestParam struct {
	DatableRequestParam
}

type fingerprintData struct {
	IsEnable bool                      `json:"enabled" form:"enabled"`
	Config   execute.FingerprintConfig `json:"config" form:"config"`
}

type ProfileInfoData struct {
	Id              string          `json:"id" form:"id"`
	Name            string          `json:"name" form:"name"`
	Description     string          `json:"description" form:"description"`
	Status          string          `json:"status" form:"status"`
	SortNumber      int             `json:"sort_number" form:"sort_number"`
	ConfigType      string          `json:"config_type" form:"config_type"`
	LLMAPIData      llmapiData      `json:"llmapi" form:"llmapi"`
	PortscanData    portscanData    `json:"portscan" form:"portscan"`
	DomainscanData  domainscanData  `json:"domainscan" form:"domainscan"`
	OnlineapiData   onlineapiData   `json:"onlineapi" form:"onlineapi"`
	FingerprintData fingerprintData `json:"fingerprint" form:"fingerprint"`
	PocscanData     pocscanData     `json:"pocscan" form:"pocscan"`
	StandaloneData  standaloneData  `json:"standalone" form:"standalone"`
}

type ProfileData struct {
	Id             string   `json:"id" form:"id"`
	Index          int      `json:"index" form:"-"`
	Name           string   `json:"name" form:"name"`
	Description    string   `json:"description" form:"description"`
	Status         string   `json:"status" form:"status"`
	SortNumber     int      `json:"sort_number" form:"sort_number"`
	Executors      []string `json:"executors" form:"executors"`
	CreateDatetime string   `json:"create_time" form:"-"`
	UpdateDatetime string   `json:"update_time" form:"-"`
}

// IndexAction 显示列表页面
func (c *ProfileController) IndexAction() {
	c.CheckMultiAccessRequest([]RequestRole{SuperAdmin, Admin}, true)

	c.Layout = "base.html"
	c.TplName = "profile-list.html"
}

// ListAction 获取列表显示的数据
func (c *ProfileController) ListAction() {
	defer func(c *ProfileController, encoding ...bool) {
		_ = c.ServeJSON(encoding...)
	}(c)
	if c.CheckMultiAccessRequest([]RequestRole{SuperAdmin, Admin}, false) == false {
		c.FailedStatus("当前用户权限不允许！")
		return
	}

	req := profileRequestParam{}
	err := c.ParseForm(&req)
	if err != nil {
		logging.RuntimeLog.Error(err)
	}
	c.validateRequestParam(&req)
	resp := c.getListData(req)
	c.Data["json"] = resp
}

// validateRequestParam 校验请求的参数
func (c *ProfileController) validateRequestParam(req *profileRequestParam) {
	if req.Length <= 0 {
		req.Length = 50
	}
	if req.Start < 0 {
		req.Start = 0
	}
}

// getListData 获取列表数据
func (c *ProfileController) getListData(req profileRequestParam) (resp DataTableResponseData) {
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

	profile := db.NewProfile(workspaceId, mongoClient)
	startPage := req.Start/req.Length + 1
	results, err := profile.Find(bson.M{}, startPage, req.Length)
	if err != nil {
		logging.RuntimeLog.Error(err)
		c.FailedStatus(err.Error())
		return
	}
	for i, row := range results {
		u := ProfileData{}
		u.Id = row.Id.Hex()
		u.Status = row.Status
		u.Index = req.Start + i + 1
		u.Name = row.ProfileName
		u.Description = row.Description
		var executorConfigArgs execute.ExecutorConfig
		err = json.Unmarshal([]byte(row.Args), &executorConfigArgs)
		if err != nil {
			logging.RuntimeLog.Error(err)
		} else {
			if executorConfigArgs.LLMAPI != nil {
				for executorName, _ := range executorConfigArgs.LLMAPI {
					u.Executors = append(u.Executors, executorName)
				}
			}
			if executorConfigArgs.PortScan != nil {
				for executorName, _ := range executorConfigArgs.PortScan {
					u.Executors = append(u.Executors, executorName)
				}
			}
			if executorConfigArgs.DomainScan != nil {
				for executorName, _ := range executorConfigArgs.DomainScan {
					u.Executors = append(u.Executors, executorName)
				}
			}
			if executorConfigArgs.OnlineAPI != nil {
				for executorName, _ := range executorConfigArgs.OnlineAPI {
					u.Executors = append(u.Executors, executorName)
				}
			}
			if executorConfigArgs.FingerPrint != nil {
				for executorName, _ := range executorConfigArgs.FingerPrint {
					u.Executors = append(u.Executors, executorName)
				}
			}
			if executorConfigArgs.PocScan != nil {
				for executorName, _ := range executorConfigArgs.PocScan {
					u.Executors = append(u.Executors, executorName)
				}
			}
			if executorConfigArgs.Standalone != nil {
				u.Executors = append(u.Executors, "standalone")
			}
		}
		u.SortNumber = row.SortNumber
		u.UpdateDatetime = FormatDateTime(row.UpdateTime)
		u.CreateDatetime = FormatDateTime(row.CreateTime)
		resp.Data = append(resp.Data, u)
	}
	total, _ := profile.Count(bson.M{})
	resp.RecordsTotal = total
	resp.RecordsFiltered = total

	return
}

// AddIndexAction 新增页面显示
func (c *ProfileController) AddIndexAction() {
	c.CheckMultiAccessRequest([]RequestRole{SuperAdmin, Admin}, true)

	c.Data["profileId"] = ""
	c.Layout = "base.html"
	c.TplName = "profile-edit.html"
}

// EditIndexAction 编辑页面显示
func (c *ProfileController) EditIndexAction() {
	c.CheckMultiAccessRequest([]RequestRole{SuperAdmin, Admin}, true)

	id := c.GetString("id")
	if len(id) == 0 {
		logging.RuntimeLog.Error("empty id")
	}
	c.Data["profileId"] = id

	c.Layout = "base.html"
	c.TplName = "profile-edit.html"
}

// SaveAction 保存新记录、更新原记录
func (c *ProfileController) SaveAction() {
	defer func(c *ProfileController, encoding ...bool) {
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
	doc, err := c.processProfileInfoData()
	if err != nil {
		return
	}
	mongoClient, err := db.GetClient()
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		c.FailedStatus(err.Error())
		return
	}
	defer db.CloseClient(mongoClient)

	var isSuccess bool
	var saveError error
	if doc.Id.IsZero() {
		doc.Id = bson.NewObjectID()
		isSuccess, saveError = db.NewProfile(workspaceId, mongoClient).Insert(doc)
	} else {
		// 更新
		profile := db.NewProfile(workspaceId, mongoClient)
		docUpdate, err := profile.Get(doc.Id.Hex())
		if err != nil {
			logging.RuntimeLog.Error(err.Error())
			c.FailedStatus(err.Error())
			return
		}
		if docUpdate.Id != doc.Id {
			c.FailedStatus("记录不存在！")
			return
		}
		docUpdate.Status = doc.Status
		docUpdate.ProfileName = doc.ProfileName
		docUpdate.Description = doc.Description
		docUpdate.SortNumber = doc.SortNumber
		docUpdate.Args = doc.Args
		isSuccess, saveError = profile.Update(docUpdate.Id, docUpdate)
	}
	if saveError != nil {
		logging.RuntimeLog.Error(err)
		c.FailedStatus(err.Error())
	}
	if isSuccess {
		c.SucceededStatus(doc.Id.Hex())
	} else {
		c.FailedStatus("失败")
	}

	return
}

// GetAction 根据ID获取一个记录
func (c *ProfileController) GetAction() {
	defer func(c *ProfileController, encoding ...bool) {
		_ = c.ServeJSON(encoding...)
	}(c)
	if c.CheckMultiAccessRequest([]RequestRole{SuperAdmin, Admin}, false) == false {
		c.FailedStatus("当前用户权限不允许！")
		return
	}

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

	profileInfo, err := getProfileInfoData(id, workspaceId)
	if err != nil {
		c.FailedStatus(err.Error())
		return
	}
	c.Data["json"] = profileInfo

	return
}

// DeleteAction 删除一条记录
func (c *ProfileController) DeleteAction() {
	defer func(c *ProfileController, encoding ...bool) {
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

	_ = c.CheckErrorAndStatus(db.NewProfile(workspaceId, mongoClient).Delete(id))

	return
}

func getProfileInfoData(id, workspaceId string) (profileInfo ProfileInfoData, err error) {
	mongoClient, err := db.GetClient()
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		return profileInfo, err
	}
	defer db.CloseClient(mongoClient)

	doc, err := db.NewProfile(workspaceId, mongoClient).Get(id)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		return profileInfo, err
	}
	if doc.Id.Hex() != id {
		err = errors.New("记录不存在！")
		return
	}
	profileInfo, err = parseProfileInfoData(doc)
	return
}

func parseProfileInfoData(doc db.ProfileDocument) (profileInfo ProfileInfoData, err error) {
	var executorConfigArgs execute.ExecutorConfig
	err = json.Unmarshal([]byte(doc.Args), &executorConfigArgs)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		return
	}
	if len(executorConfigArgs.Standalone) > 0 {
		profileInfo.ConfigType = "standalone"
		if _, ok := executorConfigArgs.Standalone["standalone"]; ok {
			profileInfo.StandaloneData.Config = executorConfigArgs.Standalone["standalone"]
		}
	} else {
		profileInfo.ConfigType = "staged"
		if len(executorConfigArgs.LLMAPI) > 0 {
			profileInfo.LLMAPIData.IsEnable = true
			if _, ok := executorConfigArgs.LLMAPI["qwen"]; ok {
				profileInfo.LLMAPIData.Qwen = true
			}
			if _, ok := executorConfigArgs.LLMAPI["kimi"]; ok {
				profileInfo.LLMAPIData.Kimi = true
			}
			if _, ok := executorConfigArgs.LLMAPI["deepseek"]; ok {
				profileInfo.LLMAPIData.Deepseek = true
			}
			// icpPlus 接口比较特殊，和LLMAPI一样都是根据组织机构查询，所以先放到llmapi的配置里
			if _, ok := executorConfigArgs.LLMAPI["icpPlus"]; ok {
				profileInfo.LLMAPIData.ICPPlus = true
			}
			for _, v := range executorConfigArgs.LLMAPI {
				profileInfo.LLMAPIData.Config = v
				break
			}
		}
		if len(executorConfigArgs.PortScan) > 0 {
			profileInfo.PortscanData.IsEnable = true
			if _, ok := executorConfigArgs.PortScan["nmap"]; ok {
				profileInfo.PortscanData.Nmap = true
			}
			if _, ok := executorConfigArgs.PortScan["masscan"]; ok {
				profileInfo.PortscanData.Masscan = true
			}
			if _, ok := executorConfigArgs.PortScan["gogo"]; ok {
				profileInfo.PortscanData.Gogo = true
			}
			for _, v := range executorConfigArgs.PortScan {
				profileInfo.PortscanData.Config = v
				break
			}
		}
		if len(executorConfigArgs.DomainScan) > 0 {
			profileInfo.DomainscanData.IsEnable = true
			if _, ok := executorConfigArgs.DomainScan["massdns"]; ok {
				profileInfo.DomainscanData.Massdns = true
			}
			if _, ok := executorConfigArgs.DomainScan["subfinder"]; ok {
				profileInfo.DomainscanData.Subfinder = true
			}
			for _, v := range executorConfigArgs.DomainScan {
				profileInfo.DomainscanData.Config = v
				break
			}
		}
		if len(executorConfigArgs.OnlineAPI) > 0 {
			profileInfo.OnlineapiData.IsEnable = true
			if _, ok := executorConfigArgs.OnlineAPI["fofa"]; ok {
				profileInfo.OnlineapiData.Fofa = true
			}
			if _, ok := executorConfigArgs.OnlineAPI["hunter"]; ok {
				profileInfo.OnlineapiData.Hunter = true
			}
			if _, ok := executorConfigArgs.OnlineAPI["quake"]; ok {
				profileInfo.OnlineapiData.Quake = true
			}
			if _, ok := executorConfigArgs.OnlineAPI["whois"]; ok {
				profileInfo.OnlineapiData.Whois = true
			}
			if _, ok := executorConfigArgs.OnlineAPI["icp"]; ok {
				profileInfo.OnlineapiData.ICP = true
			}
			for _, v := range executorConfigArgs.OnlineAPI {
				profileInfo.OnlineapiData.Config = v
				break
			}
		}
		if len(executorConfigArgs.FingerPrint) > 0 {
			profileInfo.FingerprintData.IsEnable = true
			if v, ok := executorConfigArgs.FingerPrint["fingerprint"]; ok {
				if v.IsHttpx {
					profileInfo.FingerprintData.Config.IsHttpx = true
				}
				if v.IsFingerprintx {
					profileInfo.FingerprintData.Config.IsFingerprintx = true
				}
				if v.IsIconHash {
					v.IsIconHash = true
				}
				if v.IsScreenshot {
					profileInfo.FingerprintData.Config.IsScreenshot = true
				}
				profileInfo.FingerprintData.Config = v
			}
		}
		if len(executorConfigArgs.PocScan) > 0 {
			profileInfo.PocscanData.IsEnable = true
			if _, ok := executorConfigArgs.PocScan["nuclei"]; ok {
				profileInfo.PocscanData.PocBin = "nuclei"
				profileInfo.PocscanData.Config = executorConfigArgs.PocScan["nuclei"]
			}
		}
	}
	profileInfo.Id = doc.Id.Hex()
	profileInfo.Name = doc.ProfileName
	profileInfo.Description = doc.Description
	profileInfo.SortNumber = doc.SortNumber
	profileInfo.Status = doc.Status

	return profileInfo, nil
}

func (c *ProfileController) processProfileInfoData() (doc db.ProfileDocument, err error) {
	// 前端 form 表单提交的数据
	var profileInfo ProfileInfoData
	err = json.Unmarshal(c.Ctx.Input.RequestBody, &profileInfo)
	if err != nil {
		logging.RuntimeLog.Error(err)
		c.FailedStatus(err.Error())
		return
	}
	// 处理数据
	doc = db.ProfileDocument{
		ProfileName: profileInfo.Name,
		Description: profileInfo.Description,
		SortNumber:  profileInfo.SortNumber,
		Status:      profileInfo.Status,
	}
	if len(profileInfo.Id) > 0 {
		doc.Id, err = bson.ObjectIDFromHex(profileInfo.Id)
		if err != nil {
			logging.RuntimeLog.Error(err.Error())
			c.FailedStatus(err.Error())
			return
		}
	}
	// 构造执行器配置
	var executorConfigArgs execute.ExecutorConfig
	if profileInfo.ConfigType == "staged" {
		// staged模式
		if profileInfo.LLMAPIData.IsEnable && (profileInfo.LLMAPIData.Qwen || profileInfo.LLMAPIData.Kimi || profileInfo.LLMAPIData.Deepseek || profileInfo.LLMAPIData.ICPPlus) {
			executorConfigArgs.LLMAPI = make(map[string]execute.LLMAPIConfig)
			if profileInfo.LLMAPIData.Qwen {
				executorConfigArgs.LLMAPI["qwen"] = profileInfo.LLMAPIData.Config
			}
			if profileInfo.LLMAPIData.Kimi {
				executorConfigArgs.LLMAPI["kimi"] = profileInfo.LLMAPIData.Config
			}
			if profileInfo.LLMAPIData.Deepseek {
				executorConfigArgs.LLMAPI["deepseek"] = profileInfo.LLMAPIData.Config
			}
			// icpPlus 接口比较特殊，和LLMAPI一样都是根据组织机构查询，所以先放到llmapi的配置里
			if profileInfo.LLMAPIData.ICPPlus {
				executorConfigArgs.LLMAPI["icpPlus"] = profileInfo.LLMAPIData.Config
			}
		}
		if profileInfo.PortscanData.IsEnable && (profileInfo.PortscanData.Nmap || profileInfo.PortscanData.Masscan || profileInfo.PortscanData.Gogo) {
			executorConfigArgs.PortScan = make(map[string]execute.PortscanConfig)
			if profileInfo.PortscanData.Nmap {
				executorConfigArgs.PortScan["nmap"] = profileInfo.PortscanData.Config
			}
			if profileInfo.PortscanData.Masscan {
				executorConfigArgs.PortScan["masscan"] = profileInfo.PortscanData.Config
			}
			if profileInfo.PortscanData.Gogo {
				executorConfigArgs.PortScan["gogo"] = profileInfo.PortscanData.Config
			}
		}
		if profileInfo.DomainscanData.IsEnable && (profileInfo.DomainscanData.Massdns || profileInfo.DomainscanData.Subfinder) {
			// 如果域名不执行结果的IP端口扫描，则清空端口扫描配置
			if !profileInfo.DomainscanData.Config.IsIPPortScan && !profileInfo.DomainscanData.Config.IsIPSubnetPortScan {
				profileInfo.DomainscanData.Config.ResultPortscanBin = ""
				profileInfo.DomainscanData.Config.ResultPortscanConfig = nil
			}
			executorConfigArgs.DomainScan = make(map[string]execute.DomainscanConfig)
			if profileInfo.DomainscanData.Massdns {
				executorConfigArgs.DomainScan["massdns"] = profileInfo.DomainscanData.Config
			}
			if profileInfo.DomainscanData.Subfinder {
				executorConfigArgs.DomainScan["subfinder"] = profileInfo.DomainscanData.Config
			}
		}
		if profileInfo.OnlineapiData.IsEnable && (profileInfo.OnlineapiData.Fofa || profileInfo.OnlineapiData.Hunter || profileInfo.OnlineapiData.Quake || profileInfo.OnlineapiData.Whois || profileInfo.OnlineapiData.ICP) {
			executorConfigArgs.OnlineAPI = make(map[string]execute.OnlineAPIConfig)
			if profileInfo.OnlineapiData.Fofa {
				executorConfigArgs.OnlineAPI["fofa"] = profileInfo.OnlineapiData.Config
			}
			if profileInfo.OnlineapiData.Hunter {
				executorConfigArgs.OnlineAPI["hunter"] = profileInfo.OnlineapiData.Config
			}
			if profileInfo.OnlineapiData.Quake {
				executorConfigArgs.OnlineAPI["quake"] = profileInfo.OnlineapiData.Config
			}
			if profileInfo.OnlineapiData.Whois {
				executorConfigArgs.OnlineAPI["whois"] = profileInfo.OnlineapiData.Config
			}
			if profileInfo.OnlineapiData.ICP {
				executorConfigArgs.OnlineAPI["icp"] = profileInfo.OnlineapiData.Config
			}
		}
		if profileInfo.FingerprintData.IsEnable && (profileInfo.FingerprintData.Config.IsFingerprintx || profileInfo.FingerprintData.Config.IsHttpx || profileInfo.FingerprintData.Config.IsIconHash || profileInfo.FingerprintData.Config.IsScreenshot) {
			executorConfigArgs.FingerPrint = make(map[string]execute.FingerprintConfig)
			executorConfigArgs.FingerPrint["fingerprint"] = profileInfo.FingerprintData.Config
		}
		if profileInfo.PocscanData.IsEnable && profileInfo.PocscanData.PocBin == "nuclei" {
			executorConfigArgs.PocScan = make(map[string]execute.PocscanConfig)
			executorConfigArgs.PocScan[profileInfo.PocscanData.PocBin] = profileInfo.PocscanData.Config
		}
	} else if profileInfo.ConfigType == "standalone" {
		// standalone模式
		executorConfigArgs.Standalone = make(map[string]execute.StandaloneScanConfig)
		executorConfigArgs.Standalone["standalone"] = profileInfo.StandaloneData.Config
	} else {
		c.FailedStatus("配置类型错误！")
		return
	}
	argsBytes, err := json.Marshal(executorConfigArgs)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		c.FailedStatus(err.Error())
		return
	}
	doc.Args = string(argsBytes)

	return doc, nil
}
