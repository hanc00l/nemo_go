package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/hanc00l/nemo_go/pkg/db"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/custom"
	"github.com/hanc00l/nemo_go/pkg/task/domainscan"
	"github.com/hanc00l/nemo_go/pkg/task/fingerprint"
	"github.com/hanc00l/nemo_go/pkg/task/onlineapi"
	"github.com/hanc00l/nemo_go/pkg/utils"
	"net/http"
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
	OrgId         int    `form:"org_id"`
	IPAddress     string `form:"ip_address"`
	DomainAddress string `form:"domain_address"`
	ColorTag      string `form:"color_tag"`
	MemoContent   string `form:"memo_content"`
	DateDelta     int    `form:"date_delta"`
	DisableFofa   bool   `form:"disable_fofa"`
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
}

// DomainAttrInfo domain属性
type DomainAttrInfo struct {
	Id         int    `json:"id"`
	Tag        string `json:"tag"`
	Content    string `json:"content"`
	CreateTime string `json:"create_datetime"`
	UpdateTime string `json:"update_datetime"`
}

// DomainAttrFullInfo domain属性数据的聚合
type DomainAttrFullInfo struct {
	IP         map[string]struct{}
	TitleSet   map[string]struct{}
	BannerSet  map[string]struct{}
	DomainAttr []DomainAttrInfo
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
	c.UpdateOnlineUser()
	sessionData := c.GetGlobalSessionData()
	c.Data["data"] = sessionData
	c.Layout = "base.html"
	c.TplName = "domain-list.html"
}

// ListAction Datable列表数据
func (c *DomainController) ListAction() {
	c.UpdateOnlineUser()
	defer c.ServeJSON()

	req := domainRequestParam{}
	err := c.ParseForm(&req)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
	}
	c.validateRequestParam(&req)
	//更新session
	c.setSessionData("ip_address_domain", req.IPAddress)
	c.setSessionData("domain_address", req.DomainAddress)
	if req.OrgId == 0 {
		c.setSessionData("session_org_id", "")
	} else {
		c.setSessionData("session_org_id", fmt.Sprintf("%d", req.OrgId))
	}
	resp := c.getDomainListData(req)
	c.Data["json"] = resp
}

// InfoAction 一个域名的详细数据
func (c *DomainController) InfoAction() {
	var domainInfo DomainInfo

	domainName := c.GetString("domain")
	disableFofa, _ := c.GetBool("disable_fofa", false)
	if domainName != "" {
		domainInfo = getDomainInfo(domainName, disableFofa)
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

	id, err := c.GetInt("id")
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		c.FailedStatus(err.Error())
		return
	}
	domain := db.Domain{Id: id}
	if domain.Get() {
		ss := fingerprint.NewScreenShot()
		ss.Delete(domain.DomainName)
		c.MakeStatusResponse(domain.Delete())
		return
	}
	c.MakeStatusResponse(false)

}

// DeleteDomainAttrAction 删除一个域名的属性
func (c *DomainController) DeleteDomainAttrAction() {
	defer c.ServeJSON()

	id, err := c.GetInt("id")
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		c.MakeStatusResponse(false)
		return
	}
	domainAttr := db.DomainAttr{Id: id}
	c.MakeStatusResponse(domainAttr.Delete())
}

// DeleteDomainFofaAttrAction 删除fofa属性
func (c *DomainController) DeleteDomainFofaAttrAction() {
	defer c.ServeJSON()

	id, err := c.GetInt("id")
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		c.MakeStatusResponse(false)
		return
	}
	domainAttr := db.DomainAttr{RelatedId: id, Source: "fofa"}
	c.MakeStatusResponse(domainAttr.DeleteByRelatedIDAndSource())
}

// ExportMemoAction 导出备忘录信息
func (c *DomainController) ExportMemoAction() {
	req := domainRequestParam{}
	err := c.ParseForm(&req)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
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
		logging.RuntimeLog.Error(err.Error())
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

//validateRequestParam 校验请求的参数
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
	return
}

// getDomainListData 获取域名的Datable列表数据
func (c *DomainController) getDomainListData(req domainRequestParam) (resp DataTableResponseData) {
	domain := db.Domain{}
	searchMap := c.getSearchMap(req)
	results, total := domain.Gets(searchMap, req.Start/req.Length+1, req.Length)
	hp := custom.NewHoneyPot()
	ss := fingerprint.NewScreenShot()
	for i, domainRow := range results {
		domainData := DomainListData{}
		domainData.Id = domainRow.Id
		domainData.Index = req.Start + i + 1
		domainData.Domain = domainRow.DomainName
		domainData.ScreenshotFile = ss.LoadScreenshotFile(domainRow.DomainName)
		if domainData.ScreenshotFile == nil {
			domainData.ScreenshotFile = make([]string, 0)
		}
		domainInfo := getDomainInfo(domainRow.DomainName, req.DisableFofa)
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
		domainData.Banner = strings.Join(utils.RemoveDuplicationElement(append(domainInfo.Title, domainInfo.Banner...)), ", ")
		var vulSet []string
		for _, v := range domainInfo.Vulnerability {
			vulSet = append(vulSet, fmt.Sprintf("%s/%s", v.PocFile, v.Source))
		}
		domainData.Vulnerability = strings.Join(vulSet, "\r\n")

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
	domainResult, _ := domain.Gets(searchMap, -1, -1)
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
func getDomainInfo(domainName string, disableFofa bool) (r DomainInfo) {
	domain := db.Domain{DomainName: domainName}
	if !domain.GetByDomain() {
		return
	}
	r.Id = domain.Id
	r.Domain = domain.DomainName
	r.CreateTime = FormatDateTime(domain.CreateDatetime)
	r.UpdateTime = FormatDateTime(domain.UpdateDatetime)
	for _, v := range fingerprint.NewScreenShot().LoadScreenshotFile(domainName) {
		filepath := fmt.Sprintf("/screenshot/%s/%s", domainName, v)
		filepathThumbnail := fmt.Sprintf("/screenshot/%s/%s", domainName, strings.ReplaceAll(v, ".png", "_thumbnail.png"))
		r.Screenshot = append(r.Screenshot, ScreenshotFileInfo{
			ScreenShotFile:          filepath,
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
	domainAttrInfo := getDomainAttrFullInfo(domain.Id, disableFofa)
	//遍历域名关联的每一个IP，获取port,title,banner和PortAttrInfo
	for ipName, _ := range domainAttrInfo.IP {
		ip := db.Ip{IpName: ipName}
		if !ip.GetByIp() {
			continue
		}
		pi := getPortInfo(ipName, ip.Id, disableFofa)
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
	return
}

// getDomainAttrFullInfo 获取一个域名的属性集合
func getDomainAttrFullInfo(id int, disableFofa bool) DomainAttrFullInfo {
	r := DomainAttrFullInfo{
		IP:        make(map[string]struct{}),
		TitleSet:  make(map[string]struct{}),
		BannerSet: make(map[string]struct{}),
	}
	fofaInfo := make(map[string]string)
	domainAttr := db.DomainAttr{RelatedId: id}
	domainAttrData := domainAttr.GetsByRelatedId()
	for _, da := range domainAttrData {
		if disableFofa && da.Source == "fofa" {
			continue
		}
		if da.Source == "fofa" {
			fofaInfo[da.Tag] = da.Content
		}
		if da.Tag == "A" {
			if _, ok := r.IP[da.Content]; !ok {
				r.IP[da.Content] = struct{}{}
			}
		} else if da.Tag == "title" {
			if _, ok := r.TitleSet[da.Content]; !ok {
				r.TitleSet[da.Content] = struct{}{}
			}
		} else if da.Tag == "server" {
			if _, ok := r.BannerSet[da.Content]; !ok {
				r.BannerSet[da.Content] = struct{}{}
			}
		} else if da.Tag == "whatweb" {
			r.DomainAttr = append(r.DomainAttr, DomainAttrInfo{
				Id:         da.Id,
				Tag:        "whatweb",
				Content:    da.Content,
				CreateTime: FormatDateTime(da.CreateDatetime),
				UpdateTime: FormatDateTime(da.UpdateDatetime),
			})
		} else if da.Tag == "httpx" {
			r.DomainAttr = append(r.DomainAttr, DomainAttrInfo{
				Id:         da.Id,
				Tag:        "httpx",
				Content:    da.Content,
				CreateTime: FormatDateTime(da.CreateDatetime),
				UpdateTime: FormatDateTime(da.UpdateDatetime),
			})
		} else if da.Source == "wappalyzer" && da.Tag == "banner" {
			r.DomainAttr = append(r.DomainAttr, DomainAttrInfo{
				Id:         da.Id,
				Tag:        "wappalyzer",
				Content:    da.Content,
				CreateTime: FormatDateTime(da.CreateDatetime),
				UpdateTime: FormatDateTime(da.UpdateDatetime),
			})
			for _, b := range strings.Split(da.Content, ",") {
				if _, ok := r.BannerSet[b]; !ok {
					r.BannerSet[b] = struct{}{}
				}
			}
		}
	}
	if len(fofaInfo) > 0 {
		fofaContent, _ := json.Marshal(fofaInfo)
		r.DomainAttr = append(r.DomainAttr, DomainAttrInfo{
			Id:      id,
			Tag:     "fofa",
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
	domainResult, _ := domain.Gets(searchMap, -1, -1)
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
