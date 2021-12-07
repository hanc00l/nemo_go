package controllers

import (
	"encoding/base64"
	"fmt"
	"github.com/hanc00l/nemo_go/pkg/db"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/custom"
	"github.com/hanc00l/nemo_go/pkg/task/fingerprint"
	"github.com/hanc00l/nemo_go/pkg/utils"
	"net/http"
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
}

// PortAttrInfo 每一个端口的详细数据
type PortAttrInfo struct {
	Id                 int
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
	PortNumbers []int
	PortStatus  map[int]string
	TitleSet    map[string]struct{}
	BannerSet   map[string]struct{}
	PortAttr    []PortAttrInfo
}

// IPStatisticInfo IP统计信息
type IPStatisticInfo struct {
	IP       map[string]int
	IPSubnet map[string]int
	Port     map[string]int
	Location map[string]int
}

// IndexAction GET请求显示页面
func (c *IPController) IndexAction() {
	c.UpdateOnlineUser()
	sessionData := c.GetGlobalSessionData()
	c.Data["data"] = sessionData
	c.Layout = "base.html"
	c.TplName = "ip-list.html"
}

// ListAction IP列表
func (c *IPController) ListAction() {
	c.UpdateOnlineUser()
	defer c.ServeJSON()

	req := ipRequestParam{}
	err := c.ParseForm(&req)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
	}
	c.validateRequestParam(&req)
	//更新session
	c.setSessionData("ip_address_ip", req.IPAddress)
	c.setSessionData("domain_address", req.DomainAddress)
	c.setSessionData("port", req.Port)
	if req.OrgId == 0 {
		c.setSessionData("session_org_id", "")
	} else {
		c.setSessionData("session_org_id", fmt.Sprintf("%d", req.OrgId))
	}
	resp := c.getIPListData(req)
	c.Data["json"] = resp
}

// InfoAction 一个IP的详细情况
func (c *IPController) InfoAction() {
	var ipInfo IPInfo
	ipName := c.GetString("ip")
	disableFofa, _ := c.GetBool("disable_fofa", false)
	if ipName != "" {
		ipInfo = getIPInfo(ipName, disableFofa)
		// 修改背景色为交叉显示
		if len(ipInfo.PortAttr) > 0 {
			tableBackgroundSet := false
			for i, _ := range ipInfo.PortAttr {
				if ipInfo.PortAttr[i].IP != "" && ipInfo.PortAttr[i].Port != "" {
					tableBackgroundSet = !tableBackgroundSet
				}
				ipInfo.PortAttr[i].TableBackgroundSet = tableBackgroundSet
			}
		}
	}
	ipInfo.DisableFofa = disableFofa
	c.Data["ip_info"] = ipInfo
	c.Layout = "base.html"
	c.TplName = "ip-info.html"
}

// DeleteIPAction 删除一个IP记录
func (c *IPController) DeleteIPAction() {
	defer c.ServeJSON()

	id, err := c.GetInt("id")
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		c.FailedStatus(err.Error())
		return
	}
	ip := db.Ip{Id: id}
	if ip.Get() {
		ss := fingerprint.NewScreenShot()
		ss.Delete(ip.IpName)
		c.MakeStatusResponse(ip.Delete())
	} else {
		c.MakeStatusResponse(false)
	}
}

// DeletePortAttrAction 删除一个Port属性值
func (c *IPController) DeletePortAttrAction() {
	defer c.ServeJSON()

	id, err := c.GetInt("id")
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
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
		logging.RuntimeLog.Error(err.Error())
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
		logging.RuntimeLog.Error(err.Error())
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

//validateRequestParam 校验请求的参数
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
	return searchMap
}

// getIPListData 获取IP列表显示的数据
func (c *IPController) getIPListData(req ipRequestParam) (resp DataTableResponseData) {
	ip := db.Ip{}
	searchMap := c.getSearchMap(req)
	results, total := ip.Gets(searchMap, req.Start/req.Length+1, req.Length)
	hp := custom.NewHoneyPot()
	ss := fingerprint.NewScreenShot()
	for i, ipRow := range results {
		ipData := IPListData{}
		ipData.Index = req.Start + i + 1
		ipData.Id = ipRow.Id
		ipData.IP = ipRow.IpName
		ipData.Location = ipRow.Location
		ipInfo := getIPInfo(ipRow.IpName, req.DisableFofa)
		ipData.ColorTag = ipInfo.ColorTag
		ipData.MemoContent = ipInfo.Memo
		ipData.Banner = strings.Join(utils.RemoveDuplicationElement(append(ipInfo.Title, ipInfo.Banner...)), ", ")
		ipData.ScreenshotFile = ss.LoadScreenshotFile(ipRow.IpName)
		if ipData.ScreenshotFile == nil {
			ipData.ScreenshotFile = make([]string, 0)
		}
		var vulSet []string
		for _, v := range ipInfo.Vulnerability {
			vulSet = append(vulSet, fmt.Sprintf("%s/%s", v.PocFile, v.Source))
		}
		ipData.Vulnerability = strings.Join(vulSet, "\r\n")

		ipPortInfo := getPortInfo(ipRow.IpName, ipRow.Id, req.DisableFofa)
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
func getPortInfo(ip string, ipId int, disableFofa bool) (r PortInfo) {
	r.PortStatus = make(map[int]string)
	r.BannerSet = make(map[string]struct{})
	r.TitleSet = make(map[string]struct{})

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
			if disableFofa && pad.Source == "fofa" {
				continue
			}
			pai := PortAttrInfo{}
			pai.Id = pad.Id
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
				pai.FofaLink = fmt.Sprintf("https://fofa.so/result?qbase64=%s", base64.URLEncoding.EncodeToString([]byte(fofaSearch)))
			}
			r.PortAttr = append(r.PortAttr, pai)
			if pad.Tag == "title" {
				if _, ok := r.TitleSet[pad.Content]; !ok {
					r.TitleSet[pad.Content] = struct{}{}
				}
			} else if pad.Tag == "banner" || pad.Tag == "server" || pad.Tag == "tag" ||  pad.Tag == "fingerprint"{
				if pad.Content == "unknown" {
					continue
				}
				if pad.Source == "wappalyzer" {
					for _, b := range strings.Split(pad.Content, ",") {
						if _, ok := r.BannerSet[b]; !ok {
							r.BannerSet[b] = struct{}{}
						}
					}
				} else {
					if _, ok := r.BannerSet[pad.Content]; !ok {
						r.BannerSet[pad.Content] = struct{}{}
					}
				}
			}
		}
	}
	return
}

//getIPInfo 获取一个IP的信息集合
func getIPInfo(ipName string, disableFofa bool) (r IPInfo) {
	ip := db.Ip{IpName: ipName}
	if !ip.GetByIp() {
		return r
	}
	r.IP = ipName
	r.Id = ip.Id
	r.Location = ip.Location
	r.Status = ip.Status
	r.CreateTime = FormatDateTime(ip.CreateDatetime)
	r.UpdateTime = FormatDateTime(ip.UpdateDatetime)
	//screenshot
	for _, v := range fingerprint.NewScreenShot().LoadScreenshotFile(ipName) {
		filepath := fmt.Sprintf("/screenshot/%s/%s", ipName, v)
		filepathThumbnail := fmt.Sprintf("/screenshot/%s/%s", ipName, strings.ReplaceAll(v, ".png", "_thumbnail.png"))
		r.Screenshot = append(r.Screenshot, ScreenshotFileInfo{
			ScreenShotFile:          filepath,
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
	portInfo := getPortInfo(ipName, ip.Id, disableFofa)
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
	ipResult, _ := ip.Gets(searchMap, -1, -1)
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

// getIPListData 获取备忘录数据
func (c *IPController) getMemoData(req ipRequestParam) (r []string) {
	ip := db.Ip{}

	searchMap := c.getSearchMap(req)
	ipResult, _ := ip.Gets(searchMap, -1, -1)
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
