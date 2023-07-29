package controllers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/db"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/custom"
	"github.com/hanc00l/nemo_go/pkg/task/domainscan"
	"github.com/hanc00l/nemo_go/pkg/task/fingerprint"
	"github.com/hanc00l/nemo_go/pkg/task/onlineapi"
	"github.com/hanc00l/nemo_go/pkg/utils"
	"net/http"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type DomainController struct {
	BaseController
}

// domainRequestParam domain的请求参数
type domainRequestParam struct {
	DatableRequestParam
	OrgId              int    `form:"org_id"`
	IPAddress          string `form:"ip_address"`
	DomainAddress      string `form:"domain_address"`
	ColorTag           string `form:"color_tag"`
	MemoContent        string `form:"memo_content"`
	DateDelta          int    `form:"date_delta"`
	CreateDateDelta    int    `form:"create_date_delta"`
	DisableFofa        bool   `form:"disable_fofa"`
	DisableBanner      bool   `form:"disable_banner"`
	Content            string `form:"content"`
	SelectNoResolvedIP bool   `form:"select_no_ip"`
	OrderByDate        bool   `form:"select_order_by_date"`
	DomainHttp         string `form:"domain_http"`
}

// DomainListData datable显示的每一行数据
type DomainListData struct {
	Id             int      `json:"id"`
	Index          int      `json:"index"`
	Domain         string   `json:"domain"`
	IP             []string `json:"ip"`
	Title          string   `json:"title"`
	Banner         string   `json:"banner"`
	ColorTag       string   `json:"color_tag"`
	MemoContent    string   `json:"memo_content"`
	Vulnerability  string   `json:"vulnerability"`
	HoneyPot       string   `json:"honeypot"`
	ScreenshotFile []string `json:"screenshot"`
	DomainCDN      string   `json:"domaincdn"`
	DomainCNAME    string   `json:"domaincname"`
	IsIPCDN        bool     `json:"ipcdn"`
	IconImage      []string `json:"iconimage"`
	WorkspaceId    int      `json:"workspace"`
	WorkspaceGUID  string   `json:"workspace_guid"`
	PinIndex       int      `json:"pinindex"`
}

// DomainInfo domain详细数据聚合
type DomainInfo struct {
	Id            int
	Domain        string
	Organization  string
	IP            []string
	Port          []int
	PortAttr      []PortAttrInfo
	Title         []string
	Banner        []string
	ColorTag      string
	Memo          string
	Vulnerability []VulnerabilityInfo
	CreateTime    string
	UpdateTime    string
	Screenshot    []ScreenshotFileInfo
	DomainAttr    []DomainAttrInfo
	DisableFofa   bool
	IconHashes    []IconHashWithFofa
	TlsData       []string
	DomainCDN     string
	DomainCNAME   string
	Workspace     string
	WorkspaceGUID string
	PinIndex      string
}

// DomainAttrInfo domain属性
type DomainAttrInfo struct {
	Id         int    `json:"id"`
	DomainId   int    `json:"domainId"`
	Port       int    `json:"port"`
	Tag        string `json:"tag"`
	Content    string `json:"content"`
	CreateTime string `json:"create_datetime"`
	UpdateTime string `json:"update_datetime"`
}

// DomainAttrFullInfo domain属性数据的聚合
type DomainAttrFullInfo struct {
	IP           map[string]struct{}
	TitleSet     map[string]struct{}
	BannerSet    map[string]struct{}
	DomainAttr   []DomainAttrInfo
	IconImageSet map[string]string
	TlsData      map[string]struct{}
	DomainCDN    string
	DomainCNAME  string
}

// DomainStatisticInfo domain统计信息
type DomainStatisticInfo struct {
	Domain    map[string][]string
	DomainIP  map[string]string
	Subdomain map[string]int
	IP        map[string]int
	IPSubnet  map[string]int
}

// IndexAction index
func (c *DomainController) IndexAction() {
	sessionData := c.GetGlobalSessionData()
	c.Data["data"] = sessionData
	c.Layout = "base.html"
	c.TplName = "domain-list.html"
}

// ListAction Datable列表数据
func (c *DomainController) ListAction() {
	defer c.ServeJSON()

	req := domainRequestParam{}
	err := c.ParseForm(&req)
	if err != nil {
		logging.RuntimeLog.Error(err)
		logging.CLILog.Error(err)
	}
	c.validateRequestParam(&req)
	if !c.IsServerAPI {
		//更新session
		c.setSessionData("ip_address_domain", req.IPAddress)
		c.setSessionData("domain_address", req.DomainAddress)
		if req.OrgId == 0 {
			c.setSessionData("session_org_id", "")
		} else {
			c.setSessionData("session_org_id", fmt.Sprintf("%d", req.OrgId))
		}
	}
	resp := c.getDomainListData(req)
	c.Data["json"] = resp
}

// InfoAction 一个域名的详细数据
func (c *DomainController) InfoAction() {
	var domainInfo DomainInfo

	domainName := c.GetString("domain")
	workspaceId, err := c.GetInt("workspace")
	disableFofa, _ := c.GetBool("disable_fofa", false)
	if domainName != "" && err == nil && workspaceId > 0 {
		domainInfo = getDomainInfo(workspaceId, domainName, disableFofa, false)
		if len(domainInfo.PortAttr) > 0 {
			tableBackgroundSet := false
			for i, _ := range domainInfo.PortAttr {
				if domainInfo.PortAttr[i].IP != "" && domainInfo.PortAttr[i].Port != "" {
					tableBackgroundSet = !tableBackgroundSet
				}
				domainInfo.PortAttr[i].TableBackgroundSet = tableBackgroundSet
			}
		}
	}
	domainInfo.DisableFofa = disableFofa
	c.Data["domain_info"] = domainInfo
	c.Layout = "base.html"
	c.TplName = "domain-info.html"
}

// DeleteDomainAction 删除一个记录
func (c *DomainController) DeleteDomainAction() {
	defer c.ServeJSON()
	if c.CheckMultiAccessRequest([]RequestRole{SuperAdmin, Admin}, false) == false {
		c.FailedStatus("当前用户权限不允许！")
		return
	}

	id, err := c.GetInt("id")
	if err != nil {
		logging.RuntimeLog.Error(err)
		logging.CLILog.Error(err)
		c.FailedStatus(err.Error())
		return
	}
	domain := db.Domain{Id: id}
	if domain.Get() {
		workspace := db.Workspace{Id: domain.WorkspaceId}
		if workspace.Get() {
			ss := fingerprint.NewScreenShot()
			ss.Delete(workspace.WorkspaceGUID, domain.DomainName)
		}
		c.MakeStatusResponse(domain.Delete())
	} else {
		c.MakeStatusResponse(false)
	}
}

// DeleteDomainAttrAction 删除一个域名的属性
func (c *DomainController) DeleteDomainAttrAction() {
	defer c.ServeJSON()
	if c.CheckMultiAccessRequest([]RequestRole{SuperAdmin, Admin}, false) == false {
		c.FailedStatus("当前用户权限不允许！")
		return
	}

	id, err := c.GetInt("id")
	if err != nil {
		logging.RuntimeLog.Error(err)
		logging.CLILog.Error(err)
		c.MakeStatusResponse(false)
		return
	}
	domainAttr := db.DomainAttr{Id: id}
	c.MakeStatusResponse(domainAttr.Delete())
}

// DeleteDomainOnlineAPIAttrAction 删除fofa等属性
func (c *DomainController) DeleteDomainOnlineAPIAttrAction() {
	defer c.ServeJSON()
	if c.CheckMultiAccessRequest([]RequestRole{SuperAdmin, Admin}, false) == false {
		c.FailedStatus("当前用户权限不允许！")
		return
	}

	id, err := c.GetInt("id")
	if err != nil {
		logging.RuntimeLog.Error(err)
		logging.CLILog.Error(err)
		c.MakeStatusResponse(false)
		return
	}
	for _, source := range []string{"fofa", "hunter", "quake", "0zone"} {
		domainAttr := db.DomainAttr{RelatedId: id, Source: source}
		c.MakeStatusResponse(domainAttr.DeleteByRelatedIDAndSource())
	}
}

// ExportMemoAction 导出备忘录信息
func (c *DomainController) ExportMemoAction() {
	req := domainRequestParam{}
	err := c.ParseForm(&req)
	if err != nil {
		logging.RuntimeLog.Error(err)
		logging.CLILog.Error(err)
	}
	c.validateRequestParam(&req)
	content := c.getMemoData(req)
	rw := c.Ctx.ResponseWriter
	rw.Header().Set("Content-Disposition", "attachment; filename=domain-memo.txt")
	rw.Header().Set("Content-Type", "application/octet-stream")
	rw.WriteHeader(http.StatusOK)
	http.ServeContent(rw, c.Ctx.Request, "domain-memo.txt", time.Now(), strings.NewReader(strings.Join(content, "\n")))
}

// GetMemoAction 获取指定IP的备忘录信息
func (c *DomainController) GetMemoAction() {
	defer c.ServeJSON()

	rid, err := c.GetInt("r_id")
	if err != nil {
		c.MakeStatusResponse(false)
		return
	}
	m := &db.DomainMemo{RelatedId: rid}
	if m.GetByRelatedId() {
		c.SucceededStatus(m.Content)
		return
	}
	c.MakeStatusResponse(true)
}

// UpdateMemoAction 更新指定IP的备忘录信息
func (c *DomainController) UpdateMemoAction() {
	defer c.ServeJSON()

	rid, err := c.GetInt("r_id")
	if err != nil {
		c.FailedStatus(err.Error())
		return
	}
	content := c.GetString("memo", "")
	var success bool
	m := &db.DomainMemo{RelatedId: rid}
	if m.GetByRelatedId() {
		updateMap := make(map[string]interface{})
		updateMap["content"] = content
		success = m.Update(updateMap)
	} else {
		m.Content = content
		success = m.Add()
	}
	c.MakeStatusResponse(success)
}

// MarkColorTagAction 颜色标记
func (c *DomainController) MarkColorTagAction() {
	defer c.ServeJSON()

	rid, err := c.GetInt("r_id")
	if err != nil {
		c.FailedStatus(err.Error())
		return
	}
	color := c.GetString("color")
	ct := db.DomainColorTag{RelatedId: rid}
	if color == "" || color == "DELETE" {
		c.MakeStatusResponse(ct.DeleteByRelatedId())
		return
	}
	var success bool
	if ct.GetByRelatedId() {
		updateMap := make(map[string]interface{})
		updateMap["color"] = color
		success = ct.Update(updateMap)
	} else {
		ct.Color = color
		success = ct.Add()
	}
	c.MakeStatusResponse(success)
}

// StatisticsAction 域名的统计信息
func (c *DomainController) StatisticsAction() {
	req := domainRequestParam{}
	err := c.ParseForm(&req)
	if err != nil {
		logging.RuntimeLog.Error(err)
		logging.CLILog.Error(err)
	}
	c.validateRequestParam(&req)
	r := c.getDomainStatisticsData(req)
	//输出统计的内容
	var content []string
	// domain
	content = append(content, fmt.Sprintf("Domain(%d):", len(r.Domain)))
	for k, _ := range r.Domain {
		content = append(content, k)
	}
	// domain detail
	content = append(content, "")
	for k, _ := range r.Domain {
		content = append(content, fmt.Sprintf("%s(%d):", k, len(r.Domain[k])))
		for _, v := range r.Domain[k] {
			hostDomainReversed := strings.Split(v, ",")
			for i, j := 0, len(hostDomainReversed)-1; i < j; i, j = i+1, j-1 {
				hostDomainReversed[i], hostDomainReversed[j] = hostDomainReversed[j], hostDomainReversed[i]
			}
			domainFullName := fmt.Sprintf("%s.%s", strings.Join(hostDomainReversed, "."), k)
			content = append(content, domainFullName)
		}
		content = append(content, "")
	}
	// subdomain
	content = append(content, "")
	content = append(content, fmt.Sprintf("Subname(%d):", len(r.Subdomain)))
	subs := utils.SortMapByValue(r.Subdomain, true)
	for _, v := range subs {
		content = append(content, fmt.Sprintf("%s: %d", v.Key, v.Value))
	}
	//domain subnet
	content = append(content, "")
	content = append(content, fmt.Sprintf("Subnet(%d):", len(r.IPSubnet)))
	ipSubnetSorted := utils.SortMapByValue(r.IPSubnet, true)
	for _, v := range ipSubnetSorted {
		content = append(content, fmt.Sprintf("%-16s:%d", v.Key, v.Value))
	}
	//domain ip
	content = append(content, "")
	content = append(content, fmt.Sprintf("IP(%d):", len(r.IP)))
	ipSorted := utils.SortMapByValue(r.IP, true)
	for _, v := range ipSorted {
		content = append(content, fmt.Sprintf("%-16s:%d", v.Key, v.Value))
	}
	rw := c.Ctx.ResponseWriter
	rw.Header().Set("Content-Disposition", "attachment; filename=domain-statistics.txt")
	rw.Header().Set("Content-Type", "application/octet-stream")
	rw.WriteHeader(http.StatusOK)
	http.ServeContent(rw, c.Ctx.Request, "domain-statistics.txt", time.Now(), strings.NewReader(strings.Join(content, "\n")))
}

// PinTopAction 置顶/取消在列表中的置顶显示
func (c *DomainController) PinTopAction() {
	defer c.ServeJSON()

	id, err1 := c.GetInt("id")
	pinIndex, err2 := c.GetInt("pin_index")
	if err1 != nil || err2 != nil {
		logging.RuntimeLog.Error("get id or pin_index error")
		c.FailedStatus("get id or pin_index error")
		return
	}
	domain := db.Domain{Id: id}
	if domain.Get() {
		updateMap := make(map[string]interface{})
		if pinIndex == 1 {
			updateMap["pin_index"] = 1
		} else {
			updateMap["pin_index"] = 0
		}
		c.MakeStatusResponse(domain.Update(updateMap))
		return
	}
	c.FailedStatus("domain not exist")
}

// InfoHttpAction 获取指定的http信息
func (c *DomainController) InfoHttpAction() {
	defer c.ServeJSON()

	domainId, err := c.GetInt("r_id")
	port, err2 := c.GetInt("port")
	if err != nil {
		c.FailedStatus(err.Error())
		return
	}
	if err2 != nil {
		c.FailedStatus(err2.Error())
		return
	}
	domainHttp := db.DomainHttp{RelatedId: domainId, Port: port, Tag: "body"}
	if domainHttp.GetByRelatedIdAndPortAndTag() {
		c.SucceededStatus(domainHttp.Content)
		return
	}
	return
}

// validateRequestParam 校验请求的参数
func (c *DomainController) validateRequestParam(req *domainRequestParam) {
	if req.Length <= 0 {
		req.Length = 50
	}
	if req.Start < 0 {
		req.Start = 0
	}
}

// getSearchMap 根据查询参数生成查询条件
func (c *DomainController) getSearchMap(req domainRequestParam) (searchMap map[string]interface{}) {
	searchMap = make(map[string]interface{})

	workspaceId := c.GetCurrentWorkspace()
	if workspaceId > 0 {
		searchMap["workspace_id"] = workspaceId
	}
	if req.OrgId > 0 {
		searchMap["org_id"] = req.OrgId
	}
	if req.DomainAddress != "" {
		searchMap["domain"] = req.DomainAddress
	}
	if req.IPAddress != "" {
		searchMap["ip"] = req.IPAddress
	}
	if req.ColorTag != "" {
		searchMap["color_tag"] = req.ColorTag
	}
	if req.MemoContent != "" {
		searchMap["memo_content"] = req.MemoContent
	}
	if req.DateDelta > 0 {
		searchMap["date_delta"] = req.DateDelta
	}
	if req.CreateDateDelta > 0 {
		searchMap["create_date_delta"] = req.CreateDateDelta
	}
	if req.Content != "" {
		searchMap["content"] = req.Content
	}
	if req.DomainHttp != "" {
		searchMap["domain_http"] = req.DomainHttp
	}
	return
}

// getDomainListData 获取域名的Datable列表数据
func (c *DomainController) getDomainListData(req domainRequestParam) (resp DataTableResponseData) {
	domain := db.Domain{}
	searchMap := c.getSearchMap(req)
	results, total := domain.Gets(searchMap, req.Start/req.Length+1, req.Length, req.OrderByDate)
	hp := custom.NewHoneyPot()
	ss := fingerprint.NewScreenShot()
	cdn := custom.NewCDNCheck()
	workspaceCacheMap := make(map[int]string)
	for i, domainRow := range results {
		domainData := DomainListData{}
		domainData.Id = domainRow.Id
		domainData.Index = req.Start + i + 1
		domainData.Domain = domainRow.DomainName
		domainData.PinIndex = domainRow.PinIndex
		domainData.WorkspaceId = domainRow.WorkspaceId
		if _, ok := workspaceCacheMap[domainData.WorkspaceId]; !ok {
			workspace := db.Workspace{Id: domainData.WorkspaceId}
			if workspace.Get() {
				workspaceCacheMap[workspace.Id] = workspace.WorkspaceGUID
			}
		}
		if _, ok := workspaceCacheMap[domainData.WorkspaceId]; ok {
			domainData.WorkspaceGUID = workspaceCacheMap[domainData.WorkspaceId]
		}
		domainInfo := getDomainInfo(domainRow.WorkspaceId, domainRow.DomainName, req.DisableFofa, req.DisableBanner)
		// 筛选没有域名的解析IP的记录：
		if req.SelectNoResolvedIP && len(domainInfo.IP) > 0 {
			continue
		}
		domainData.IP = domainInfo.IP
		if domainData.IP == nil {
			domainData.IP = make([]string, 0)
		}
		var systemList []string
		isDomainHoneypot, domainSystemList := hp.CheckHoneyPot(domainRow.DomainName, "")
		if isDomainHoneypot && len(domainSystemList) > 0 {
			systemList = append(systemList, domainSystemList...)
		}
		for _, ip := range domainInfo.IP {
			isDomainHoneypot, domainSystemList = hp.CheckHoneyPot(ip, "")
			if isDomainHoneypot && len(domainSystemList) > 0 {
				systemList = append(systemList, domainSystemList...)
			}
		}
		if len(systemList) > 0 {
			domainData.HoneyPot = strings.Join(systemList, "\n")
		}
		domainData.MemoContent = domainInfo.Memo
		domainData.ColorTag = domainInfo.ColorTag
		//domainData.Banner = strings.Join(utils.RemoveDuplicationElement(append(domainInfo.Title, domainInfo.Banner...)), ", ")
		domainData.Title = strings.Join(domainInfo.Title, ", ")
		domainData.Banner = strings.Join(domainInfo.Banner, ", ")
		domainData.ScreenshotFile = ss.LoadScreenshotFile(domainData.WorkspaceGUID, domainRow.DomainName)
		if domainData.ScreenshotFile == nil {
			domainData.ScreenshotFile = make([]string, 0)
		}
		var vulSet []string
		for _, v := range domainInfo.Vulnerability {
			vulSet = append(vulSet, fmt.Sprintf("%s/%s", v.PocFile, v.Source))
		}
		domainData.Vulnerability = strings.Join(vulSet, "\r\n")
		domainData.DomainCDN = domainInfo.DomainCDN
		domainData.DomainCNAME = domainInfo.DomainCNAME
		for _, ip := range domainData.IP {
			if cdn.CheckIP(ip) || cdn.CheckASN(ip) {
				domainData.IsIPCDN = true
				break
			}
		}
		for _, ihm := range domainInfo.IconHashes {
			domainData.IconImage = append(domainData.IconImage, ihm.IconImage)
		}
		resp.Data = append(resp.Data, domainData)
	}
	resp.Draw = req.Draw
	resp.RecordsTotal = total
	resp.RecordsFiltered = total
	if resp.Data == nil {
		resp.Data = make([]interface{}, 0)
	}
	return
}

// DomainController 获取备忘录信息
func (c *DomainController) getMemoData(req domainRequestParam) (r []string) {
	domain := db.Domain{}
	searchMap := c.getSearchMap(req)
	domainResult, _ := domain.Gets(searchMap, -1, -1, req.OrderByDate)
	for _, domainRow := range domainResult {
		memo := db.DomainMemo{RelatedId: domainRow.Id}
		if !memo.GetByRelatedId() || memo.Content == "" {
			continue
		}
		r = append(r, fmt.Sprintf("[+]%s:", domainRow.DomainName))
		r = append(r, fmt.Sprintf("%s\n", memo.Content))
	}
	return
}

// getDomainInfo获取一个域名的数据集合
func getDomainInfo(workspaceId int, domainName string, disableFofa, disableBanner bool) (r DomainInfo) {
	domain := db.Domain{DomainName: domainName, WorkspaceId: workspaceId}
	if !domain.GetByDomain() {
		return
	}
	r.Id = domain.Id
	r.Domain = domain.DomainName
	r.CreateTime = FormatDateTime(domain.CreateDatetime)
	r.UpdateTime = FormatDateTime(domain.UpdateDatetime)
	r.PinIndex = fmt.Sprintf("%d", domain.PinIndex)
	r.Workspace = fmt.Sprintf("%d", domain.WorkspaceId)
	workspace := db.Workspace{Id: workspaceId}
	if workspace.Get() {
		r.WorkspaceGUID = workspace.WorkspaceGUID
	}
	for _, v := range fingerprint.NewScreenShot().LoadScreenshotFile(workspace.WorkspaceGUID, domainName) {
		screenFilePath := fmt.Sprintf("/webfiles/%s/screenshot/%s/%s", r.WorkspaceGUID, domainName, v)
		filepathThumbnail := fmt.Sprintf("/webfiles/%s/screenshot/%s/%s", r.WorkspaceGUID, domainName, strings.ReplaceAll(v, ".png", "_thumbnail.png"))
		r.Screenshot = append(r.Screenshot, ScreenshotFileInfo{
			ScreenShotFile:          screenFilePath,
			ScreenShotThumbnailFile: filepathThumbnail,
			Tooltip:                 v,
		})
	}
	if r.Screenshot == nil {
		r.Screenshot = make([]ScreenshotFileInfo, 0)
	}
	if domain.OrgId != nil {
		org := db.Organization{Id: *domain.OrgId}
		if org.Get() {
			r.Organization = org.OrgName
		}
	}
	portSet := make(map[int]struct{})
	//域名的属性
	domainAttrInfo := getDomainAttrFullInfo(r.WorkspaceGUID, domain.Id, disableFofa, disableBanner)
	//遍历域名关联的每一个IP，获取port,title,banner和PortAttrInfo
	for ipName := range domainAttrInfo.IP {
		ip := db.Ip{IpName: ipName, WorkspaceId: workspaceId}
		if !ip.GetByIp() {
			continue
		}
		pi := getPortInfo(r.WorkspaceGUID, ipName, ip.Id, disableFofa, disableBanner)
		for _, portNumber := range pi.PortNumbers {
			if _, ok := portSet[portNumber]; !ok {
				portSet[portNumber] = struct{}{}
			}
		}
		for k, _ := range pi.TitleSet {
			if _, ok := domainAttrInfo.TitleSet[k]; !ok {
				domainAttrInfo.TitleSet[k] = struct{}{}
			}
		}
		for k, _ := range pi.BannerSet {
			if _, ok := domainAttrInfo.BannerSet[k]; !ok {
				domainAttrInfo.BannerSet[k] = struct{}{}
			}
		}
		r.PortAttr = append(r.PortAttr, pi.PortAttr...)
	}
	r.Port = utils.SetToSliceInt(portSet)
	r.IP = utils.SetToSlice(domainAttrInfo.IP)
	r.Title = utils.SetToSlice(domainAttrInfo.TitleSet)
	r.Banner = utils.SetToSlice(domainAttrInfo.BannerSet)
	r.DomainAttr = domainAttrInfo.DomainAttr
	icp := onlineapi.NewICPQuery(onlineapi.ICPQueryConfig{})
	if icpInfo := icp.LookupICP(domainName); icpInfo != nil {
		icpContent, _ := json.Marshal(*icpInfo)
		r.DomainAttr = append(r.DomainAttr, DomainAttrInfo{
			Tag:     "ICP",
			Content: string(icpContent),
		})
	}
	whois := onlineapi.NewWhois(onlineapi.WhoisQueryConfig{})
	if whoisInfo := whois.LookupWhois(domainName); whoisInfo != nil {
		whoisContent, _ := json.Marshal(*whoisInfo)
		r.DomainAttr = append(r.DomainAttr, DomainAttrInfo{
			Tag:     "Whois",
			Content: string(whoisContent),
		})
	}
	colorTag := db.DomainColorTag{RelatedId: domain.Id}
	if colorTag.GetByRelatedId() {
		r.ColorTag = colorTag.Color
	}
	memo := db.DomainMemo{RelatedId: domain.Id}
	if memo.GetByRelatedId() {
		r.Memo = memo.Content
	}
	vul := db.Vulnerability{Target: domainName}
	vulData := vul.GetsByTarget()
	for _, v := range vulData {
		r.Vulnerability = append(r.Vulnerability, VulnerabilityInfo{
			Id:         v.Id,
			Target:     v.Target,
			Url:        v.Url,
			PocFile:    v.PocFile,
			Source:     v.Source,
			UpdateTime: FormatDateTime(v.UpdateDatetime),
		})
	}
	//
	r.TlsData = utils.SetToSlice(domainAttrInfo.TlsData)
	r.DomainCDN = domainAttrInfo.DomainCDN
	r.DomainCNAME = domainAttrInfo.DomainCNAME
	for hash, image := range domainAttrInfo.IconImageSet {
		r.IconHashes = append(r.IconHashes, IconHashWithFofa{
			IconHash:  hash,
			IconImage: image,
			FofaUrl: fmt.Sprintf("https://fofa.info/result?qbase64=%s",
				base64.URLEncoding.EncodeToString([]byte(fmt.Sprintf("icon_hash=%s", hash)))),
		})
	}
	// http info
	domainHttp := db.DomainHttp{RelatedId: domain.Id, Tag: "header"}
	domainHttpInfos := domainHttp.GetsByRelatedIdAndTag()
	for _, info := range domainHttpInfos {
		dai := DomainAttrInfo{
			Id:         info.Id,
			DomainId:   info.RelatedId,
			Port:       info.Port,
			Tag:        "http_header",
			Content:    info.Content,
			CreateTime: FormatDateTime(info.CreateDatetime),
			UpdateTime: FormatDateTime(info.UpdateDatetime),
		}
		r.DomainAttr = append(r.DomainAttr, dai)
	}
	return
}

// getDomainAttrFullInfo 获取一个域名的属性集合
func getDomainAttrFullInfo(workspaceGUID string, id int, disableFofa, disableBanner bool) DomainAttrFullInfo {
	r := DomainAttrFullInfo{
		IP:           make(map[string]struct{}),
		TitleSet:     make(map[string]struct{}),
		BannerSet:    make(map[string]struct{}),
		TlsData:      make(map[string]struct{}),
		IconImageSet: make(map[string]string),
	}
	fofaInfo := make(map[string]string)
	domainAttr := db.DomainAttr{RelatedId: id}
	domainAttrData := domainAttr.GetsByRelatedId()
	for _, da := range domainAttrData {
		if disableFofa && (da.Source == "fofa" || da.Source == "quake" || da.Source == "hunter" || da.Source == "0zone") {
			continue
		}
		if da.Source == "fofa" || da.Source == "quake" || da.Source == "hunter" || da.Source == "0zone" {
			fofaInfo[da.Tag] = da.Content
		}
		if da.Tag == "A" {
			if _, ok := r.IP[da.Content]; !ok {
				r.IP[da.Content] = struct{}{}
			}
		} else if da.Tag == "CDN" {
			r.DomainCDN = da.Content
		} else if da.Tag == "CNAME" {
			r.DomainCNAME = da.Content
		} else if da.Tag == "title" {
			if _, ok := r.TitleSet[da.Content]; !ok {
				r.TitleSet[da.Content] = struct{}{}
			}
		} else if da.Tag == "server" || da.Tag == "fingerprint" {
			// banner信息：server、fingerpinter
			if _, ok := r.BannerSet[da.Content]; !ok {
				r.BannerSet[da.Content] = struct{}{}
			}
			if da.Tag == "fingerprint" {
				r.DomainAttr = append(r.DomainAttr, DomainAttrInfo{
					Id:         da.Id,
					Tag:        da.Tag,
					Content:    da.Content,
					CreateTime: FormatDateTime(da.CreateDatetime),
					UpdateTime: FormatDateTime(da.UpdateDatetime),
				})
			}
		} else if da.Tag == "httpx" {
			r.DomainAttr = append(r.DomainAttr, DomainAttrInfo{
				Id:         da.Id,
				Tag:        da.Tag,
				Content:    da.Content,
				CreateTime: FormatDateTime(da.CreateDatetime),
				UpdateTime: FormatDateTime(da.UpdateDatetime),
			})
		} else if da.Tag == "favicon" {
			hashAndUrls := strings.Split(da.Content, "|")
			if len(hashAndUrls) == 2 {
				hash := strings.TrimSpace(hashAndUrls[0])
				r.DomainAttr = append(r.DomainAttr, DomainAttrInfo{
					Id:         da.Id,
					Tag:        "favicon",
					Content:    da.Content,
					CreateTime: FormatDateTime(da.CreateDatetime),
					UpdateTime: FormatDateTime(da.UpdateDatetime),
				})
				// icon hash image
				fileSuffix := utils.GetFaviconSuffixUrl(strings.TrimSpace(hashAndUrls[1]))
				if fileSuffix != "" {
					imageFile := fmt.Sprintf("%s.%s", utils.MD5(hash), fileSuffix)
					if utils.CheckFileExist(filepath.Join(conf.GlobalServerConfig().Web.WebFiles, workspaceGUID, "iconimage", imageFile)) {
						if _, ok := r.IconImageSet[hash]; !ok {
							r.IconImageSet[hash] = imageFile
						}
					}
				}
			}
		} else if da.Tag == "tlsdata" {
			if _, ok := r.TlsData[da.Content]; !ok {
				r.TlsData[da.Content] = struct{}{}
			}
			r.DomainAttr = append(r.DomainAttr, DomainAttrInfo{
				Id:         da.Id,
				Tag:        "tlsdata",
				Content:    da.Content,
				CreateTime: FormatDateTime(da.CreateDatetime),
				UpdateTime: FormatDateTime(da.UpdateDatetime),
			})
		}
	}
	if len(fofaInfo) > 0 {
		fofaContent, _ := json.Marshal(fofaInfo)
		r.DomainAttr = append(r.DomainAttr, DomainAttrInfo{
			Id:      id,
			Tag:     "OnlineAPI",
			Content: string(fofaContent),
		})
	}
	return r
}

// getDomainStatisticsData 获取域名的统计信息
func (c *DomainController) getDomainStatisticsData(req domainRequestParam) DomainStatisticInfo {
	dsi := DomainStatisticInfo{
		Domain:    make(map[string][]string),
		DomainIP:  make(map[string]string),
		Subdomain: make(map[string]int),
		IP:        make(map[string]int),
		IPSubnet:  make(map[string]int),
	}
	domain := db.Domain{}
	searchMap := c.getSearchMap(req)
	domainResult, _ := domain.Gets(searchMap, -1, -1, req.OrderByDate)
	for _, domainRow := range domainResult {
		fldDomain, hostDomainReversed := reverseDomainHost(domainRow.DomainName)
		if fldDomain == "" || len(hostDomainReversed) == 0 {
			continue
		}
		//域名的fld与host
		dsi.Domain[fldDomain] = append(dsi.Domain[fldDomain], strings.Join(hostDomainReversed, "."))
		//subdoman
		for _, s := range hostDomainReversed {
			if _, ok := dsi.Subdomain[s]; !ok {
				dsi.Subdomain[s] = 1
			} else {
				dsi.Subdomain[s]++
			}
		}
		//domainIP、IP
		domainAttr := db.DomainAttr{RelatedId: domainRow.Id}
		domainAttrInfo := domainAttr.GetsByRelatedId()
		domainIP := make(map[string]struct{})
		for _, dai := range domainAttrInfo {
			if dai.Tag == "A" {
				domainIP[dai.Content] = struct{}{}
				if _, ok := dsi.IP[dai.Content]; !ok {
					dsi.IP[dai.Content] = 1
				} else {
					dsi.IP[dai.Content]++
				}
			}
		}
		dsi.DomainIP[domainRow.DomainName] = utils.SetToString(domainIP)
	}
	for k, _ := range dsi.IP {
		ipArray := strings.Split(k, ".")
		if len(ipArray) == 4 {
			subnet := fmt.Sprintf("%s.0/24", strings.Join(ipArray[:3], "."))
			if _, ok := dsi.IPSubnet[subnet]; !ok {
				dsi.IPSubnet[subnet] = 1
			} else {
				dsi.IPSubnet[subnet]++
			}
		}
	}
	// 对二级域名排充
	for k, _ := range dsi.Domain {
		sort.Strings(dsi.Domain[k])
	}
	return dsi
}

// reverseDomainHost 将domain提取为fld和host，并将host反向
func reverseDomainHost(domain string) (fldDomain string, hostDomainReversed []string) {
	tld := domainscan.NewTldExtract()
	fldDomain = tld.ExtractFLD(domain)
	if fldDomain == "" {
		return
	}
	domainSepList := strings.Split(domain, ".")
	dotCount := strings.Count(fldDomain, ".")
	if len(domainSepList) <= dotCount+1 {
		return
	}
	hostDomainReversed = domainSepList[:len(domainSepList)-dotCount-1]
	//reverse
	for i, j := 0, len(hostDomainReversed)-1; i < j; i, j = i+1, j-1 {
		hostDomainReversed[i], hostDomainReversed[j] = hostDomainReversed[j], hostDomainReversed[i]
	}
	return
}

// BlockDomainAction 一键拉黑域名
func (c *DomainController) BlockDomainAction() {
	defer c.ServeJSON()
	if c.CheckMultiAccessRequest([]RequestRole{SuperAdmin, Admin}, false) == false {
		c.FailedStatus("当前用户权限不允许！")
		return
	}

	id, err := c.GetInt("id")
	if err != nil {
		logging.RuntimeLog.Error(err)
		logging.CLILog.Error(err)
		c.FailedStatus(err.Error())
		return
	}
	domain := db.Domain{Id: id}
	if domain.Get() == false {
		c.FailedStatus("get domain fail")
		return
	}
	workspace := db.Workspace{Id: domain.WorkspaceId}
	if workspace.Get() == false {
		c.FailedStatus("get workspace fail")
		return
	}
	//  域提取名参数的主域，比如www.images.qq.com的主域名为.qq.com
	tld := domainscan.NewTldExtract()
	fldDomain := tld.ExtractFLD(domain.DomainName)
	if len(fldDomain) == 0 {
		c.FailedStatus("err domain format")
		return
	}
	if strings.HasPrefix(fldDomain, ".") == false {
		fldDomain = "." + fldDomain
	}
	// 将主域名增加到黑名单文件中
	blackDomain := custom.NewBlackDomain()
	if err = blackDomain.AppendBlackDomain(fldDomain); err != nil {
		c.FailedStatus(err.Error())
		return
	}
	domainRelatedIP := make(map[string]struct{})
	// 从数据中获取主域的所有子域名记录
	domainDb := db.Domain{}
	domainResult := domainDb.GetsForBlackListDomain(fldDomain, workspace.Id)
	for _, d := range domainResult {
		// 获取域名关联的IP解析记录
		domainAttr := db.DomainAttr{RelatedId: d.Id}
		domainAttrData := domainAttr.GetsByRelatedId()
		for _, da := range domainAttrData {
			if da.Tag == "A" {
				if _, ok := domainRelatedIP[da.Content]; !ok {
					domainRelatedIP[da.Content] = struct{}{}
				}
			}
		}
		// 删除screenshot
		ss := fingerprint.NewScreenShot()
		ss.Delete(workspace.WorkspaceGUID, d.DomainName)
		// 删除domain记录
		d.Delete()
	}
	// 删除关联的IP记录
	for ip := range domainRelatedIP {
		// 删除数据库中IP记录
		ipDB := db.Ip{IpName: ip, WorkspaceId: workspace.Id}
		if ipDB.GetByIp() {
			ipDB.Delete()
		}
		ss := fingerprint.NewScreenShot()
		ss.Delete(workspace.WorkspaceGUID, ip)
	}
	c.SucceededStatus("success")
}
