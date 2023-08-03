package controllers

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/csv"
	"fmt"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/db"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/custom"
	"github.com/hanc00l/nemo_go/pkg/task/domainscan"
	"github.com/hanc00l/nemo_go/pkg/task/fingerprint"
	"github.com/hanc00l/nemo_go/pkg/task/onlineapi"
	"github.com/hanc00l/nemo_go/pkg/task/pocscan"
	"github.com/hanc00l/nemo_go/pkg/task/portscan"
	"github.com/hanc00l/nemo_go/pkg/utils"
	"net/http"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type IPController struct {
	BaseController
}

// ipRequestParam 请求参数
type ipRequestParam struct {
	DatableRequestParam
	OrgId                 int    `form:"org_id"`
	IPAddress             string `form:"ip_address"`
	DomainAddress         string `form:"domain_address"`
	Port                  string `form:"port"`
	Content               string `form:"content"`
	IPLocation            string `form:"iplocation"`
	PortStatus            string `form:"port_status"`
	ColorTag              string `form:"color_tag"`
	MemoContent           string `form:"memo_content"`
	DateDelta             int    `form:"date_delta"`
	AssertCreateDateDelta int    `form:"create_date_delta"`
	DisableFofa           bool   `form:"disable_fofa"`
	DisableBanner         bool   `form:"disable_banner"`
	DisableOutofChina     bool   `form:"disable_outof_china"`
	SelectOutofChina      bool   `form:"select_outof_china"`
	SelectNoOpenedPort    bool   `form:"select_no_openedport"`
	OrderByDate           bool   `form:"select_order_by_date"`
	IpHttp                string `form:"ip_http"`
}

// IPListData 列表中每一行显示的IP数据
type IPListData struct {
	Id             int      `json:"id"`
	Index          int      `json:"index"`
	IP             string   `json:"ip"`
	Location       string   `json:"location"`
	Port           []string `json:"port"`
	Title          string   `json:"title"`
	Banner         string   `json:"banner"`
	ColorTag       string   `json:"color_tag"`
	MemoContent    string   `json:"memo_content"`
	Vulnerability  string   `json:"vulnerability"`
	HoneyPot       string   `json:"honeypot"`
	ScreenshotFile []string `json:"screenshot"`
	IsCDN          bool     `json:"cdn"`
	IconImage      []string `json:"iconimage"`
	WorkspaceId    int      `json:"workspace"`
	WorkspaceGUID  string   `json:"workspace_guid"`
	PinIndex       int      `json:"pinindex"`
}

type IconHashWithFofa struct {
	IconHash  string
	IconImage string
	FofaUrl   string
}

// IPInfo IP的详细数据的集合
type IPInfo struct {
	Id            int
	IP            string
	Organization  string
	Status        string
	Location      string
	Port          []int
	Title         []string
	Banner        []string
	PortAttr      []PortAttrInfo
	Domain        []string
	ColorTag      string
	Memo          string
	Vulnerability []VulnerabilityInfo
	CreateTime    string
	UpdateTime    string
	Screenshot    []ScreenshotFileInfo
	DisableFofa   bool
	IconHashes    []IconHashWithFofa
	TlsData       []string
	Workspace     string
	WorkspaceGUID string
	PinIndex      string
}

// PortAttrInfo 每一个端口的详细数据
type PortAttrInfo struct {
	Id                 int
	PortId             int
	IP                 string
	Port               string
	Tag                string
	Content            string
	Source             string
	FofaLink           string
	CreateTime         string
	UpdateTime         string
	TableBackgroundSet bool
}

// ScreenshotFileInfo screenshot文件
type ScreenshotFileInfo struct {
	ScreenShotFile          string
	ScreenShotThumbnailFile string
	Tooltip                 string
}

// PortInfo 端口详细数据的集合
type PortInfo struct {
	PortNumbers      []int
	PortStatus       map[int]string
	TitleSet         map[string]struct{}
	BannerSet        map[string]struct{}
	PortAttr         []PortAttrInfo
	IconHashImageSet map[string]string
	TlsDataSet       map[string]struct{}
}

// IPStatisticInfo IP统计信息
type IPStatisticInfo struct {
	IP       map[string]int
	IPSubnet map[string]int
	Port     map[string]int
	Location map[string]int
}

// IPExportInfo 用于导出IP数据的信息
type IPExportInfo struct {
	IP         string
	Port       int
	Location   string
	StatusCode string
	TitleSet   map[string]struct{}
	BannerSet  map[string]struct{}
	FingerSet  map[string]struct{}
	TlsDataSet map[string]struct{}
	HttpxSet   map[string]struct{}
	SourceSet  map[string]struct{}
}

// IndexAction GET请求显示页面
func (c *IPController) IndexAction() {
	sessionData := c.GetGlobalSessionData()
	c.Data["data"] = sessionData
	c.Layout = "base.html"
	c.TplName = "ip-list.html"
}

// ListAction IP列表
func (c *IPController) ListAction() {
	defer c.ServeJSON()

	req := ipRequestParam{}
	err := c.ParseForm(&req)
	if err != nil {
		logging.RuntimeLog.Error(err)
		logging.CLILog.Error(err)
	}
	c.validateRequestParam(&req)
	//更新session
	if !c.IsServerAPI {
		c.setSessionData("ip_address_ip", req.IPAddress)
		c.setSessionData("domain_address", req.DomainAddress)
		c.setSessionData("port", req.Port)
		if req.OrgId == 0 {
			c.setSessionData("session_org_id", "")
		} else {
			c.setSessionData("session_org_id", fmt.Sprintf("%d", req.OrgId))
		}
	}
	resp := c.GetIPListData(req)
	c.Data["json"] = resp
}

// InfoAction 一个IP的详细情况
func (c *IPController) InfoAction() {
	var ipInfo IPInfo
	ipName := c.GetString("ip")
	workspaceId, err := c.GetInt("workspace")
	disableFofa, _ := c.GetBool("disable_fofa", false)
	if ipName != "" && err == nil && workspaceId > 0 {
		ipInfo = getIPInfo(workspaceId, ipName, disableFofa, false)
		// 修改背景色为交叉显示
		if len(ipInfo.PortAttr) > 0 {
			tableBackgroundSet := false
			for i := range ipInfo.PortAttr {
				if ipInfo.PortAttr[i].IP != "" && ipInfo.PortAttr[i].Port != "" {
					tableBackgroundSet = !tableBackgroundSet
				}
				ipInfo.PortAttr[i].TableBackgroundSet = tableBackgroundSet
			}
		}
	}
	if c.IsServerAPI {
		c.Data["json"] = ipInfo
		c.ServeJSON()
	} else {
		ipInfo.DisableFofa = disableFofa
		c.Data["ip_info"] = ipInfo
		c.Layout = "base.html"
		c.TplName = "ip-info.html"
	}
}

// DeleteIPAction 删除一个IP记录
func (c *IPController) DeleteIPAction() {
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
	ip := db.Ip{Id: id}
	if ip.Get() {
		workspace := db.Workspace{Id: ip.WorkspaceId}
		if workspace.Get() {
			ss := fingerprint.NewScreenShot()
			ss.Delete(workspace.WorkspaceGUID, ip.IpName)
		}
		c.MakeStatusResponse(ip.Delete())
	} else {
		c.MakeStatusResponse(false)
	}
}

// DeletePortAttrAction 删除一个Port属性值
func (c *IPController) DeletePortAttrAction() {
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
	portAttr := db.PortAttr{Id: id}
	c.MakeStatusResponse(portAttr.Delete())
}

// StatisticsAction IP的统计信息
func (c *IPController) StatisticsAction() {
	req := ipRequestParam{}
	err := c.ParseForm(&req)
	if err != nil {
		logging.RuntimeLog.Error(err)
		logging.CLILog.Error(err)
	}
	c.validateRequestParam(&req)
	r := c.getStatisticsData(req)
	//输出统计的内容
	var content []string
	// Port
	content = append(content, fmt.Sprintf("Port(%d):", len(r.Port)))
	var portList []string
	for _, pair := range utils.SortMapByValue(r.Port, true) {
		content = append(content, fmt.Sprintf("%-5s:%d", pair.Key, pair.Value))
		portList = append(portList, pair.Key)
	}
	content = append(content, strings.Join(portList, ","))
	// C段
	content = append(content, "")
	content = append(content, fmt.Sprintf("Subnet(%d):", len(r.IPSubnet)))
	for _, pair := range utils.SortMapByValue(r.IPSubnet, true) {
		content = append(content, fmt.Sprintf("%-16s:%d", pair.Key, pair.Value))
	}
	// IP
	content = append(content, "")
	content = append(content, fmt.Sprintf("IP(%d):", len(r.IP)))
	for _, pair := range utils.SortMapByValue(r.IP, false) {
		content = append(content, fmt.Sprintf("%-16s", pair.Key))
	}
	// Location
	content = append(content, "")
	content = append(content, fmt.Sprintf("Location(%d):", len(r.Location)))
	for _, pair := range utils.SortMapByValue(r.Location, true) {
		content = append(content, fmt.Sprintf("%s:%d", pair.Key, pair.Value))
	}
	rw := c.Ctx.ResponseWriter
	rw.Header().Set("Content-Disposition", "attachment; filename=ip-statistics.txt")
	rw.Header().Set("Content-Type", "application/octet-stream")
	rw.WriteHeader(http.StatusOK)
	http.ServeContent(rw, c.Ctx.Request, "ip-statistics.txt", time.Now(), strings.NewReader(strings.Join(content, "\n")))
}

// ExportMemoAction 导出IP备忘录信息
func (c *IPController) ExportMemoAction() {
	req := ipRequestParam{}
	err := c.ParseForm(&req)
	if err != nil {
		logging.RuntimeLog.Error(err)
		logging.CLILog.Error(err)
	}
	c.validateRequestParam(&req)
	content := c.getMemoData(req)
	rw := c.Ctx.ResponseWriter
	rw.Header().Set("Content-Disposition", "attachment; filename=ip-memo.txt")
	rw.Header().Set("Content-Type", "application/octet-stream")
	rw.WriteHeader(http.StatusOK)
	http.ServeContent(rw, c.Ctx.Request, "ip-memo.txt", time.Now(), strings.NewReader(strings.Join(content, "\n")))
}

// GetMemoAction 获取指定IP的备忘录信息
func (c *IPController) GetMemoAction() {
	defer c.ServeJSON()

	rid, err := c.GetInt("r_id")
	if err != nil {
		c.MakeStatusResponse(false)
		return
	}
	m := &db.IpMemo{RelatedId: rid}
	if m.GetByRelatedId() {
		c.SucceededStatus(m.Content)
		return
	}
	c.MakeStatusResponse(false)
}

// UpdateMemoAction 更新指定IP的备忘录信息
func (c *IPController) UpdateMemoAction() {
	defer c.ServeJSON()

	rid, err := c.GetInt("r_id")
	if err != nil {
		c.MakeStatusResponse(false)
		return
	}
	content := c.GetString("memo", "")
	var success bool
	m := &db.IpMemo{RelatedId: rid}
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
func (c *IPController) MarkColorTagAction() {
	defer c.ServeJSON()

	rid, err := c.GetInt("r_id")
	if err != nil {
		c.MakeStatusResponse(false)
		return
	}
	color := c.GetString("color")
	ct := db.IpColorTag{RelatedId: rid}
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

// PinTopAction 置顶/取消在列表中的置顶显示
func (c *IPController) PinTopAction() {
	defer c.ServeJSON()

	id, err1 := c.GetInt("id")
	pinIndex, err2 := c.GetInt("pin_index")
	if err1 != nil || err2 != nil {
		logging.RuntimeLog.Error("get id or pin_index error")
		c.FailedStatus("get id or pin_index error")
		return
	}
	ip := db.Ip{Id: id}
	if ip.Get() {
		updateMap := make(map[string]interface{})
		if pinIndex == 1 {
			updateMap["pin_index"] = 1
		} else {
			updateMap["pin_index"] = 0
		}
		c.MakeStatusResponse(ip.Update(updateMap))
		return
	}
	c.FailedStatus("ip not exist")
}

// InfoHttpAction 获取指定的http信息
func (c *IPController) InfoHttpAction() {
	defer c.ServeJSON()

	portId, err := c.GetInt("r_id")
	if err != nil {
		c.FailedStatus(err.Error())
		return
	}
	ipHttp := db.IpHttp{RelatedId: portId, Tag: "body"}
	if ipHttp.GetByRelatedIdAndTag() {
		c.SucceededStatus(ipHttp.Content)
		return
	}
	return
}

// ImportPortscanResultAction 导入portscan扫描结果
func (c *IPController) ImportPortscanResultAction() {
	defer c.ServeJSON()
	if c.CheckMultiAccessRequest([]RequestRole{SuperAdmin, Admin}, false) == false {
		c.FailedStatus("当前用户权限不允许！")
		return
	}

	workspaceId := c.GetCurrentWorkspace()
	if workspaceId <= 0 {
		c.FailedStatus("未选择当前的工作空间！")
		return
	}
	var fileContent []byte
	if c.IsServerAPI {
		fileContent = []byte(c.GetString("file"))
		if len(fileContent) == 0 {
			c.FailedStatus("empty file content")
			return
		}
	} else {
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
		// 读取文件内容
		fileContent = make([]byte, fileHeader.Size)
		_, err = file.Read(fileContent)
		if err != nil {
			c.FailedStatus(err.Error())
			return
		}
	}
	// 解析并保存
	bin := c.GetString("bin", "nmap")
	orgId, _ := c.GetInt("org_id", 0)
	config := portscan.Config{OrgId: &orgId, WorkspaceId: workspaceId}
	// 如果不指定所属于组织，将值为nil
	if orgId == 0 {
		config.OrgId = nil
	}
	var result string
	if bin == "nmap" {
		nmap := portscan.NewNmap(config)
		nmap.ParseXMLContentResult(fileContent)
		portscan.FilterIPHasTooMuchPort(&nmap.Result, true)
		result = nmap.Result.SaveResult(config)
	} else if bin == "masscan" {
		m := portscan.NewMasscan(config)
		m.ParseXMLContentResult(fileContent)
		portscan.FilterIPHasTooMuchPort(&m.Result, true)
		result = m.Result.SaveResult(config)
	} else if bin == "fscan" {
		f := portscan.NewFScan(config)
		f.ParseTxtContentResult(fileContent)
		portscan.FilterIPHasTooMuchPort(&f.Result, true)
		resultIpPort := f.Result.SaveResult(config)
		resultVul := pocscan.SaveResult(f.VulResult)
		result = fmt.Sprintf("%s,%s", resultIpPort, resultVul)
	} else if bin == "gogo" {
		g := portscan.NewGogo(config)
		g.ParseJsonContentResult(fileContent)
		portscan.FilterIPHasTooMuchPort(&g.Result, true)
		resultIpPort := g.Result.SaveResult(config)
		resultVul := pocscan.SaveResult(g.VulResult)
		result = fmt.Sprintf("%s,%s", resultIpPort, resultVul)
	} else if bin == "naabu" {
		n := portscan.NewNaabu(config)
		n.ParseTxtContentResult(fileContent)
		portscan.FilterIPHasTooMuchPort(&n.Result, true)
		resultIpPort := n.Result.SaveResult(config)
		result = fmt.Sprintf("%s", resultIpPort)
	} else if bin == "httpx" {
		n := fingerprint.NewHttpx()
		n.ParseJSONContentResult(fileContent)
		portscan.FilterIPHasTooMuchPort(&n.ResultPortScan, true)
		resultIpPort := n.ResultPortScan.SaveResult(config)
		result = fmt.Sprintf("%s", resultIpPort)
	} else if bin == "txportmap" {
		tx := portscan.NewTXPortMap(config)
		tx.ParseTxtContentResult(fileContent)
		portscan.FilterIPHasTooMuchPort(&tx.Result, true)
		resultIpPort := tx.Result.SaveResult(config)
		result = fmt.Sprintf("%s", resultIpPort)
	} else if bin == "0zone" {
		z := onlineapi.NewZeroZone(onlineapi.OnlineAPIConfig{})
		z.ParseCSVContentResult(fileContent)
		portscan.FilterIPHasTooMuchPort(&z.IpResult, true)
		resultIpPort := z.IpResult.SaveResult(config)
		resultDomain := z.DomainResult.SaveResult(domainscan.Config{OrgId: config.OrgId, WorkspaceId: workspaceId})
		result = fmt.Sprintf("%s,%s", resultDomain, resultIpPort)
	} else if bin == "fofa" {
		z := onlineapi.NewFofa(onlineapi.OnlineAPIConfig{})
		z.ParseCSVContentResult(fileContent)
		portscan.FilterIPHasTooMuchPort(&z.IpResult, true)
		resultIpPort := z.IpResult.SaveResult(config)
		resultDomain := z.DomainResult.SaveResult(domainscan.Config{OrgId: config.OrgId, WorkspaceId: workspaceId})
		result = fmt.Sprintf("%s,%s", resultDomain, resultIpPort)
	} else if bin == "hunter" {
		z := onlineapi.NewHunter(onlineapi.OnlineAPIConfig{})
		z.ParseCSVContentResult(fileContent)
		portscan.FilterIPHasTooMuchPort(&z.IpResult, true)
		resultIpPort := z.IpResult.SaveResult(config)
		resultDomain := z.DomainResult.SaveResult(domainscan.Config{OrgId: config.OrgId, WorkspaceId: workspaceId})
		result = fmt.Sprintf("%s,%s", resultDomain, resultIpPort)
	} else {
		c.FailedStatus("未知的扫描方法")
		return
	}
	c.SucceededStatus(result)
}

// validateRequestParam 校验请求的参数
func (c *IPController) validateRequestParam(req *ipRequestParam) {
	if req.Length <= 0 {
		req.Length = 50
	}
	if req.Start < 0 {
		req.Start = 0
	}
}

// getSearchMap 根据查询参数生成查询条件
func (c *IPController) getSearchMap(req ipRequestParam) (searchMap map[string]interface{}) {
	searchMap = make(map[string]interface{})

	workspaceId := c.GetCurrentWorkspace()
	if workspaceId > 0 {
		searchMap["workspace_id"] = workspaceId
	}
	if req.OrgId > 0 {
		searchMap["org_id"] = req.OrgId
	}
	if req.IPLocation != "" {
		searchMap["location"] = req.IPLocation
	}
	if req.DomainAddress != "" {
		searchMap["domain"] = req.DomainAddress
	}
	if req.IPAddress != "" {
		searchMap["ip"] = req.IPAddress
	}
	if req.Port != "" {
		searchMap["port"] = req.Port
	}
	if req.PortStatus != "" {
		searchMap["port_status"] = req.PortStatus
	}
	if req.Content != "" {
		searchMap["content"] = req.Content
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
	if req.AssertCreateDateDelta > 0 {
		searchMap["create_date_delta"] = req.AssertCreateDateDelta
	}
	if req.IpHttp != "" {
		searchMap["ip_http"] = req.IpHttp
	}
	return searchMap
}

// GetIPListData 获取IP列表显示的数据
func (c *IPController) GetIPListData(req ipRequestParam) (resp DataTableResponseData) {
	ip := db.Ip{}
	searchMap := c.getSearchMap(req)
	results, total := ip.Gets(searchMap, req.Start/req.Length+1, req.Length, req.OrderByDate)
	hp := custom.NewHoneyPot()
	ss := fingerprint.NewScreenShot()
	cdn := custom.NewCDNCheck()
	workspaceCacheMap := make(map[int]string)
	for i, ipRow := range results {
		// 筛选满足指定条件的IP
		// 只看国外的IP：
		if req.SelectOutofChina && (len(ipRow.Location) > 0 && utils.CheckIPLocationInChinaMainLand(ipRow.Location)) {
			continue
		}
		// 只看国内的IP：
		if req.DisableOutofChina && utils.CheckIPLocationInChinaMainLand(ipRow.Location) == false {
			continue
		}
		ipData := IPListData{}
		ipData.Index = req.Start + i + 1
		ipData.Id = ipRow.Id
		ipData.IP = ipRow.IpName
		ipData.Location = ipRow.Location
		ipData.PinIndex = ipRow.PinIndex
		ipInfo := getIPInfo(ipRow.WorkspaceId, ipRow.IpName, req.DisableFofa, req.DisableBanner)
		// 筛选满足指定条件的IP
		// 只看没有开放端口的IP：
		if req.SelectNoOpenedPort && len(ipInfo.Port) > 0 {
			continue
		}
		ipData.ColorTag = ipInfo.ColorTag
		ipData.MemoContent = ipInfo.Memo
		ipData.Title = strings.Join(ipInfo.Title, ", ")
		ipData.Banner = strings.Join(ipInfo.Banner, ", ")
		ipData.WorkspaceId = ipRow.WorkspaceId
		if _, ok := workspaceCacheMap[ipRow.WorkspaceId]; !ok {
			workspace := db.Workspace{Id: ipRow.WorkspaceId}
			if workspace.Get() {
				workspaceCacheMap[workspace.Id] = workspace.WorkspaceGUID
			}
		}
		if _, ok := workspaceCacheMap[ipRow.WorkspaceId]; ok {
			ipData.WorkspaceGUID = workspaceCacheMap[ipRow.WorkspaceId]
		}
		ipData.ScreenshotFile = ss.LoadScreenshotFile(ipData.WorkspaceGUID, ipRow.IpName)
		for _, ihm := range ipInfo.IconHashes {
			ipData.IconImage = append(ipData.IconImage, ihm.IconImage)
		}
		if ipData.ScreenshotFile == nil {
			ipData.ScreenshotFile = make([]string, 0)
		}
		var vulSet []string
		for _, v := range ipInfo.Vulnerability {
			vulSet = append(vulSet, fmt.Sprintf("%s/%s", v.PocFile, v.Source))
		}
		ipData.Vulnerability = strings.Join(vulSet, "\r\n")

		ipPortInfo := getPortInfo(ipData.WorkspaceGUID, ipRow.IpName, ipRow.Id, req.DisableFofa, req.DisableBanner)
		var ports []string
		for _, p := range ipPortInfo.PortNumbers {
			ports = append(ports, fmt.Sprintf("%d", p))
			if portStatus, ok := ipPortInfo.PortStatus[p]; ok {
				ipData.Port = append(ipData.Port, fmt.Sprintf("%d[%s]", p, portStatus))
			} else {
				ipData.Port = append(ipData.Port, fmt.Sprintf("%d", p))
			}
		}
		if ipData.Port == nil || len(ipData.Port) == 0 {
			ipData.Port = make([]string, 0)
		}
		isHoneypot, systemList := hp.CheckHoneyPot(ipRow.IpName, strings.Join(ports, ","))
		if isHoneypot && len(systemList) > 0 {
			ipData.HoneyPot = strings.Join(systemList, "\n")
		}
		if cdn.CheckIP(ipRow.IpName) || cdn.CheckASN(ipRow.IpName) {
			ipData.IsCDN = true
		}
		resp.Data = append(resp.Data, ipData)
	}
	resp.Draw = req.Draw
	resp.RecordsTotal = total
	resp.RecordsFiltered = total
	if resp.Data == nil {
		resp.Data = make([]interface{}, 0)
	}
	return
}

// getPortInfo 获取一个IP的所有端口信息集合
func getPortInfo(workspaceGUID string, ip string, ipId int, disableFofa, disableBanner bool) (r PortInfo) {
	r.PortStatus = make(map[int]string)
	r.BannerSet = make(map[string]struct{})
	r.TitleSet = make(map[string]struct{})
	r.TlsDataSet = make(map[string]struct{})
	r.IconHashImageSet = make(map[string]string)

	port := db.Port{IpId: ipId}
	portData := port.GetsByIPId()
	for _, pd := range portData {
		r.PortNumbers = append(r.PortNumbers, pd.PortNum)
		if pd.Status != "" {
			if _, err := strconv.Atoi(pd.Status); err == nil {
				r.PortStatus[pd.PortNum] = pd.Status
			}
		}
		portAttr := db.PortAttr{RelatedId: pd.Id}
		portAttrData := portAttr.GetsByRelatedId()
		FirstRow := true
		for _, pad := range portAttrData {
			if disableFofa && (pad.Source == "fofa" || pad.Source == "quake" || pad.Source == "hunter" || pad.Source == "0zone") {
				continue
			}
			pai := PortAttrInfo{}
			pai.Id = pad.Id
			pai.PortId = pd.Id
			pai.Tag = pad.Tag
			pai.Content = pad.Content
			pai.Source = pad.Source
			pai.CreateTime = FormatDateTime(pad.CreateDatetime)
			pai.UpdateTime = FormatDateTime(pad.UpdateDatetime)
			/* 每个端口的一个属性生成一行记录
			第一行记录显示IP和PORT，其它行保持为空（方便查看）*/
			if FirstRow {
				FirstRow = false
				pai.IP = ip
				pai.Port = fmt.Sprintf("%d", pd.PortNum)
			}
			if pad.Source == "fofa" {
				fofaSearch := fmt.Sprintf(`ip="%s" && port="%d"`, ip, pd.PortNum)
				pai.FofaLink = fmt.Sprintf("https://fofa.info/result?qbase64=%s", base64.URLEncoding.EncodeToString([]byte(fofaSearch)))
			}
			r.PortAttr = append(r.PortAttr, pai)
			if pad.Tag == "title" {
				if _, ok := r.TitleSet[pad.Content]; !ok {
					r.TitleSet[pad.Content] = struct{}{}
				}
			} else if pad.Tag == "banner" || pad.Tag == "server" || pad.Tag == "tag" || pad.Tag == "fingerprint" {
				if pad.Tag == "banner" && disableBanner {
					continue
				}
				if pad.Content == "unknown" || pad.Content == "" {
					continue
				}
				if _, ok := r.BannerSet[pad.Content]; !ok {
					r.BannerSet[pad.Content] = struct{}{}
				}

			} else if pad.Tag == "favicon" {
				hashAndUrls := strings.Split(pad.Content, "|")
				if len(hashAndUrls) == 2 {
					// icon hash
					hash := strings.TrimSpace(hashAndUrls[0])
					// icon hash image
					fileSuffix := utils.GetFaviconSuffixUrl(strings.TrimSpace(hashAndUrls[1]))
					if fileSuffix != "" {
						imageFile := fmt.Sprintf("%s.%s", utils.MD5(hash), fileSuffix)
						if utils.CheckFileExist(filepath.Join(conf.GlobalServerConfig().Web.WebFiles, workspaceGUID, "iconimage", imageFile)) {
							if _, ok := r.IconHashImageSet[hash]; !ok {
								r.IconHashImageSet[hash] = imageFile
							}
						}
					}
				}
			} else if pad.Tag == "tlsdata" {
				if _, ok := r.TlsDataSet[pad.Content]; !ok {
					r.TlsDataSet[pad.Content] = struct{}{}
				}
			}
		}
		// http header info
		httpInfo := db.IpHttp{RelatedId: pd.Id, Tag: "header"}
		if httpInfo.GetByRelatedIdAndTag() {
			httpPortAttr := PortAttrInfo{
				Id:         httpInfo.Id,
				PortId:     httpInfo.RelatedId,
				Tag:        "http_header",
				Content:    httpInfo.Content,
				Source:     httpInfo.Source,
				CreateTime: FormatDateTime(httpInfo.CreateDatetime),
				UpdateTime: FormatDateTime(httpInfo.UpdateDatetime),
			}
			/* 每个端口的一个属性生成一行记录
			第一行记录显示IP和PORT，其它行保持为空（方便查看）*/
			if FirstRow {
				FirstRow = false
				httpPortAttr.IP = ip
				httpPortAttr.Port = fmt.Sprintf("%d", pd.PortNum)
			}
			r.PortAttr = append(r.PortAttr, httpPortAttr)
		}
	}
	return
}

// getIPInfo 获取一个IP的信息集合
func getIPInfo(workspaceId int, ipName string, disableFofa, disableBanner bool) (r IPInfo) {
	ip := db.Ip{IpName: ipName, WorkspaceId: workspaceId}
	if !ip.GetByIp() {
		return r
	}
	r.IP = ipName
	r.Id = ip.Id
	r.Location = ip.Location
	r.Status = ip.Status
	r.CreateTime = FormatDateTime(ip.CreateDatetime)
	r.UpdateTime = FormatDateTime(ip.UpdateDatetime)
	r.PinIndex = fmt.Sprintf("%d", ip.PinIndex)
	r.Workspace = fmt.Sprintf("%d", ip.WorkspaceId)
	workspace := db.Workspace{Id: ip.WorkspaceId}
	if workspace.Get() {
		r.WorkspaceGUID = workspace.WorkspaceGUID
	}
	//screenshot
	for _, v := range fingerprint.NewScreenShot().LoadScreenshotFile(workspace.WorkspaceGUID, ipName) {
		sfp := fmt.Sprintf("/webfiles/%s/screenshot/%s/%s", r.WorkspaceGUID, ipName, v)
		filepathThumbnail := fmt.Sprintf("/webfiles/%s/screenshot/%s/%s", r.WorkspaceGUID, ipName, strings.ReplaceAll(v, ".png", "_thumbnail.png"))
		r.Screenshot = append(r.Screenshot, ScreenshotFileInfo{
			ScreenShotFile:          sfp,
			ScreenShotThumbnailFile: filepathThumbnail,
			Tooltip:                 v,
		})
	}
	if r.Screenshot == nil {
		r.Screenshot = make([]ScreenshotFileInfo, 0)
	}
	// orgId
	if ip.OrgId != nil {
		org := db.Organization{Id: *ip.OrgId}
		if org.Get() {
			r.Organization = org.OrgName
		}
	}
	// port
	portInfo := getPortInfo(r.WorkspaceGUID, ipName, ip.Id, disableFofa, disableBanner)
	r.PortAttr = portInfo.PortAttr
	r.Title = utils.SetToSlice(portInfo.TitleSet)
	r.Banner = utils.SetToSlice(portInfo.BannerSet)
	r.Port = portInfo.PortNumbers
	colorTag := db.IpColorTag{RelatedId: ip.Id}
	if colorTag.GetByRelatedId() {
		r.ColorTag = colorTag.Color
	}
	// memo
	memo := db.IpMemo{RelatedId: ip.Id}
	if memo.GetByRelatedId() {
		r.Memo = memo.Content
	}
	// vul
	vul := db.Vulnerability{Target: ipName}
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
	for hash, image := range portInfo.IconHashImageSet {
		r.IconHashes = append(r.IconHashes, IconHashWithFofa{
			IconHash:  hash,
			IconImage: image,
			FofaUrl: fmt.Sprintf("https://fofa.info/result?qbase64=%s",
				base64.URLEncoding.EncodeToString([]byte(fmt.Sprintf("icon_hash=%s", hash)))),
		})
	}
	r.TlsData = utils.SetToSlice(portInfo.TlsDataSet)
	r.Domain = getIpRelatedDomain(workspaceId, ipName)
	return
}

// getStatisticsData 获取IP统计数据
func (c *IPController) getStatisticsData(req ipRequestParam) IPStatisticInfo {
	r := IPStatisticInfo{
		IP:       make(map[string]int),
		IPSubnet: make(map[string]int),
		Port:     make(map[string]int),
		Location: make(map[string]int),
	}
	ip := db.Ip{}
	searchMap := c.getSearchMap(req)
	ipResult, _ := ip.Gets(searchMap, -1, -1, req.OrderByDate)
	for _, ipRow := range ipResult {
		// ip
		if _, ok := r.IP[ipRow.IpName]; !ok {
			r.IP[ipRow.IpName] = ipRow.IpInt
		}
		// C段
		ipArray := strings.Split(ipRow.IpName, ".")
		subnet := fmt.Sprintf("%s.%s.%s.0/24", ipArray[0], ipArray[1], ipArray[2])
		if _, ok := r.IPSubnet[subnet]; ok {
			r.IPSubnet[subnet]++
		} else {
			r.IPSubnet[subnet] = 1
		}
		// Location
		if ipRow.Location != "" {
			if _, ok := r.Location[ipRow.Location]; ok {
				r.Location[ipRow.Location]++
			} else {
				r.Location[ipRow.Location] = 1
			}
		}
		// Port
		port := db.Port{IpId: ipRow.Id}
		for _, portRow := range port.GetsByIPId() {
			portString := fmt.Sprintf("%d", portRow.PortNum)
			if _, ok := r.Port[portString]; ok {
				r.Port[portString]++
			} else {
				r.Port[portString] = 1
			}
		}
	}

	return r
}

// GetIPListData 获取备忘录数据
func (c *IPController) getMemoData(req ipRequestParam) (r []string) {
	ip := db.Ip{}

	searchMap := c.getSearchMap(req)
	ipResult, _ := ip.Gets(searchMap, -1, -1, req.OrderByDate)
	for _, ipRow := range ipResult {
		memo := db.IpMemo{RelatedId: ipRow.Id}
		if !memo.GetByRelatedId() || memo.Content == "" {
			continue
		}
		r = append(r, fmt.Sprintf("[+]%s:", ipRow.IpName))
		r = append(r, fmt.Sprintf("%s\n", memo.Content))
	}
	return
}

// getIpRelatedDomain 获取IP关联的域名
func getIpRelatedDomain(workspaceId int, ipName string) []string {
	domain := db.Domain{}
	searchMap := make(map[string]interface{})
	searchMap["ip"] = ipName
	searchMap["workspace_id"] = workspaceId
	rows, _ := domain.Gets(searchMap, -1, -1, false)
	domains := make(map[string]struct{})
	for _, r := range rows {
		if _, ok := domains[r.DomainName]; !ok {
			domains[r.DomainName] = struct{}{}
		}
	}
	return utils.SetToSlice(domains)
}

// BlackIPAction 一键拉黑一个IP
func (c *IPController) BlackIPAction() {
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
	ip := db.Ip{Id: id}
	if ip.Get() == false {
		c.FailedStatus("get ip fail")
		return
	}
	workspace := db.Workspace{Id: ip.WorkspaceId}
	if workspace.Get() == false {
		c.FailedStatus("get workspace fail")
		return
	}
	// 将IP追加到黑名单文件
	blackIP := custom.NewBlackIP()
	err = blackIP.AppendBlackIP(ip.IpName)
	if err != nil {
		c.FailedStatus(err.Error())
		return
	}
	// 删除IP
	if ip.Delete() == false {
		c.FailedStatus("删除IP失败！")
		return
	}
	ss := fingerprint.NewScreenShot()
	ss.Delete(workspace.WorkspaceGUID, ip.IpName)
	// 删除IP关联的域名记录的信息
	domains := getIpRelatedDomain(workspace.Id, ip.IpName)
	for _, d := range domains {
		domain := db.Domain{DomainName: d, WorkspaceId: workspace.Id}
		if domain.GetByDomain() {
			ss.Delete(workspace.WorkspaceGUID, domain.DomainName)
			domain.Delete()
		}
	}
	c.SucceededStatus("success")
}

// ExportIPResultAction 导出IP资产
func (c *IPController) ExportIPResultAction() {
	req := ipRequestParam{}
	err := c.ParseForm(&req)
	if err != nil {
		logging.RuntimeLog.Error(err)
		logging.CLILog.Error(err)
		return
	}
	c.validateRequestParam(&req)
	content := c.writeToCSVData(c.getIPExportData(req))
	rw := c.Ctx.ResponseWriter
	rw.Header().Set("Content-Disposition", "attachment; filename=ip-result.csv")
	rw.Header().Set("Content-Type", "text/csv; charset=utf-8")
	rw.WriteHeader(http.StatusOK)

	http.ServeContent(rw, c.Ctx.Request, "ip-result.csv", time.Now(), bytes.NewReader(content))
}

// getIPExportData 获取IP的资产
func (c *IPController) getIPExportData(req ipRequestParam) (result []IPExportInfo) {
	ip := db.Ip{}
	searchMap := c.getSearchMap(req)
	ipResult, _ := ip.Gets(searchMap, -1, -1, req.OrderByDate)
	for _, ipRow := range ipResult {
		port := db.Port{IpId: ipRow.Id}
		portData := port.GetsByIPId()
		for _, pd := range portData {
			eInfo := IPExportInfo{
				IP:         ipRow.IpName,
				Port:       pd.PortNum,
				Location:   ipRow.Location,
				StatusCode: pd.Status,
				TitleSet:   make(map[string]struct{}),
				BannerSet:  make(map[string]struct{}),
				FingerSet:  make(map[string]struct{}),
				TlsDataSet: make(map[string]struct{}),
				HttpxSet:   make(map[string]struct{}),
				SourceSet:  make(map[string]struct{}),
			}
			//port attr
			portAttr := db.PortAttr{RelatedId: pd.Id}
			portAttrData := portAttr.GetsByRelatedId()
			for _, pad := range portAttrData {
				if pad.Tag == "title" {
					if _, ok := eInfo.TitleSet[pad.Content]; !ok {
						eInfo.TitleSet[pad.Content] = struct{}{}
					}
				} else if pad.Tag == "fingerprint" {
					if _, ok := eInfo.FingerSet[pad.Content]; !ok {
						eInfo.FingerSet[pad.Content] = struct{}{}
					}

				} else if pad.Tag == "banner" || pad.Tag == "server" || pad.Tag == "tag" {
					if pad.Content == "unknown" || pad.Content == "" {
						continue
					}
					if _, ok := eInfo.BannerSet[pad.Content]; !ok {
						eInfo.BannerSet[pad.Content] = struct{}{}
					}

				} else if pad.Tag == "tlsdata" {
					if _, ok := eInfo.TlsDataSet[pad.Content]; !ok {
						eInfo.TlsDataSet[pad.Content] = struct{}{}
					}
				} else if pad.Tag == "httpx" {
					if _, ok := eInfo.HttpxSet[pad.Content]; !ok {
						eInfo.HttpxSet[pad.Content] = struct{}{}
					}
				}
				if _, ok := eInfo.SourceSet[pad.Source]; !ok {
					eInfo.SourceSet[pad.Source] = struct{}{}
				}
			}
			result = append(result, eInfo)
		}
	}
	return
}

// writeToCSVData 输出为csv格式
func (c *IPController) writeToCSVData(exportInfo []IPExportInfo) []byte {
	var buf bytes.Buffer
	bufWrite := bufio.NewWriter(&buf)
	csvWriter := csv.NewWriter(bufWrite)
	csvWriter.Write([]string{"index", "url", "ip", "port", "location", "status-code", "title", "finger", "tlsdata", "httpx", "source"})
	for i, v := range exportInfo {
		csvWriter.Write([]string{
			strconv.Itoa(i + 1),
			fmt.Sprintf("%s:%d", v.IP, v.Port),
			v.IP,
			strconv.Itoa(v.Port),
			v.Location,
			v.StatusCode,
			utils.SetToString(v.TitleSet),
			utils.SetToString(v.FingerSet),
			utils.SetToString(v.TlsDataSet),
			utils.SetToString(v.HttpxSet),
			utils.SetToString(v.SourceSet),
		})
	}
	csvWriter.Flush()
	bufWrite.Flush()
	return buf.Bytes()
}
