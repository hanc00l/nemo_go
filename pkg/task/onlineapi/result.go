package onlineapi

import (
	"encoding/json"
	"fmt"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/custom"
	"github.com/hanc00l/nemo_go/pkg/task/domainscan"
	"github.com/hanc00l/nemo_go/pkg/task/portscan"
	"github.com/hanc00l/nemo_go/pkg/utils"
	"regexp"
	"strconv"
	"strings"
)

type OnlineAPIConfig struct {
	Target             string `json:"target"`
	OrgId              *int   `json:"orgId"`
	IsIPLocation       bool   `json:"ipLocation"`
	IsHttpx            bool   `json:"httpx"`
	IsScreenshot       bool   `json:"screenshot"`
	IsFingerprintHub   bool   `json:"fingerprinthub"`
	IsIconHash         bool   `json:"iconhash"`
	IsIgnoreCDN        bool   `json:"ignorecdn"`
	IsIgnoreOutofChina bool   `json:"ignoreoutofchina"`
	SearchByKeyWord    bool   `json:"keywordsearch"`
	SearchLimitCount   int    `json:"searchlimitcount"`
	WorkspaceId        int    `json:"workspaceId"`
}

type ICPQueryConfig struct {
	Target string `json:"target"`
}

type WhoisQueryConfig struct {
	Target string `json:"target"`
}

type onlineSearchResult struct {
	Domain  string
	Host    string
	IP      string
	Port    string
	Title   string
	Country string
	City    string
	Server  string
	Banner  string
}

type fofaQueryResult struct {
	Results [][]string `json:"results"`
	Size    int        `json:"size"`
	Page    int        `json:"page"`
	Mode    string     `json:"mode"`
}

type icpQueryResult struct {
	StateCode int     `json:"StateCode"`
	Reason    string  `json:"Reason"`
	Result    ICPInfo `json:"Result"`
}

type ICPInfo struct {
	Domain      string `json:"Domain"`
	Owner       string `json:"Owner"`
	CompanyName string `json:"CompanyName"`
	CompanyType string `json:"CompanyType"`
	SiteLicense string `json:"SiteLicense"`
	SiteName    string `json:"SiteName"`
	MainPage    string `json:"MainPage"`
	VerifyTime  string `json:"VerifyTime"`
}

var (
	userAgent               = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.125 Safari/537.36"
	pageSize                = 10
	SameIpToDomainFilterMax = 100
)

// parseFofaSearchResult 转换FOFA搜索结果
func (ff *Fofa) parseFofaSearchResult(queryResult []byte) (result []onlineSearchResult, sizeTotal int) {
	r := fofaQueryResult{}
	err := json.Unmarshal(queryResult, &r)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		return
	}
	sizeTotal = r.Size
	for _, line := range r.Results {
		fsr := onlineSearchResult{
			Domain: line[0], Host: line[1], IP: line[2], Port: line[3], Title: line[4],
			Country: line[5], City: line[6], Server: line[7], Banner: line[8],
		}
		lowerBanner := strings.ToLower(fsr.Banner)
		//过滤部份无意义的banner
		if strings.HasPrefix(fsr.Banner, "HTTP/") || strings.HasPrefix(lowerBanner, "<html>") || strings.HasPrefix(lowerBanner, "<!doctype html>") {
			fsr.Banner = ""
		}
		//过滤所有的\x00 \x15等不可见字符
		re := regexp.MustCompile("\\\\x[0-9a-f]{2}")
		fsr.Banner = re.ReplaceAllString(fsr.Banner, "")
		result = append(result, fsr)
	}

	return
}

// parseResult 解析搜索结果
func (ff *Fofa) parseResult() {
	ff.IpResult = portscan.Result{IPResult: make(map[string]*portscan.IPResult)}
	ff.DomainResult = domainscan.Result{DomainResult: make(map[string]*domainscan.DomainResult)}

	blackDomain := custom.NewBlackDomain()
	blackIP := custom.NewBlackIP()
	for _, fsr := range ff.Result {
		parseIpPort(ff.IpResult, fsr, "fofa", blackIP)
		parseDomainIP(ff.DomainResult, fsr, "fofa", blackDomain)
	}

	checkDomainResult(ff.DomainResult.DomainResult)
}

// SaveResult 保存搜索的结果
func (ff *Fofa) SaveResult() string {
	if conf.GlobalWorkerConfig().API.Fofa.Key == "" || conf.GlobalWorkerConfig().API.Fofa.Name == "" {
		return "no fofa api"
	}
	ips := ff.IpResult.SaveResult(portscan.Config{OrgId: ff.Config.OrgId, WorkspaceId: ff.Config.WorkspaceId})
	domains := ff.DomainResult.SaveResult(domainscan.Config{OrgId: ff.Config.OrgId, WorkspaceId: ff.Config.WorkspaceId})

	return fmt.Sprintf("%s,%s", ips, domains)
}

// parseQuakeSearchResult 解析Quake搜索结果
func (q *Quake) parseQuakeSearchResult(queryResult []byte) (result []onlineSearchResult, finish bool) {
	var serviceInfo QuakeServiceInfo
	err := json.Unmarshal(queryResult, &serviceInfo)
	if err != nil {
		//json数据反序列化失败
		//如果是json: cannot unmarshal object into Go struct field QuakeServiceInfo.data of type []struct { Time time.Time "json:\"time\""; Transport string "json:\"transport\""; Service struct { HTTP struct
		//则基本上是API的key失效，或积分不足无法读取
		logging.CLILog.Println(err)
		return
	}
	if strings.HasPrefix(serviceInfo.Message, "Successful") == false {
		logging.CLILog.Printf("Quake Search Error:%s", serviceInfo.Message)
		return
	}
	for _, data := range serviceInfo.Data {
		qsr := onlineSearchResult{
			Host:   data.Service.HTTP.Host,
			IP:     data.IP,
			Port:   fmt.Sprintf("%d", data.Port),
			Title:  data.Service.HTTP.Title,
			Server: data.Service.HTTP.Server,
		}
		result = append(result, qsr)
	}
	// 如果是API有效、正确获取到数据，count为0，表示已是最后一页了
	if serviceInfo.Meta.Pagination.Count == 0 {
		finish = true
	}
	return
}

// parseResult 解析搜索结果
func (q *Quake) parseResult() {
	q.IpResult = portscan.Result{IPResult: make(map[string]*portscan.IPResult)}
	q.DomainResult = domainscan.Result{DomainResult: make(map[string]*domainscan.DomainResult)}

	blackDomain := custom.NewBlackDomain()
	blackIP := custom.NewBlackIP()
	for _, fsr := range q.Result {
		parseIpPort(q.IpResult, fsr, "quake", blackIP)
		parseDomainIP(q.DomainResult, fsr, "quake", blackDomain)
	}

	checkDomainResult(q.DomainResult.DomainResult)
}

// SaveResult 保存搜索结果
func (q *Quake) SaveResult() string {
	if conf.GlobalWorkerConfig().API.Quake.Key == "" {
		return "no quake api"
	}
	ips := q.IpResult.SaveResult(portscan.Config{OrgId: q.Config.OrgId, WorkspaceId: q.Config.WorkspaceId})
	domains := q.DomainResult.SaveResult(domainscan.Config{OrgId: q.Config.OrgId, WorkspaceId: q.Config.WorkspaceId})

	return fmt.Sprintf("%s,%s", ips, domains)
}

// parseResult
func (h *Hunter) parseResult() {
	h.IpResult = portscan.Result{IPResult: make(map[string]*portscan.IPResult)}
	h.DomainResult = domainscan.Result{DomainResult: make(map[string]*domainscan.DomainResult)}

	blackDomain := custom.NewBlackDomain()
	blackIP := custom.NewBlackIP()
	for _, fsr := range h.Result {
		parseIpPort(h.IpResult, fsr, "hunter", blackIP)
		parseDomainIP(h.DomainResult, fsr, "hunter", blackDomain)
	}

	checkDomainResult(h.DomainResult.DomainResult)
}

// SaveResult 保存搜索结果
func (h *Hunter) SaveResult() string {
	if conf.GlobalWorkerConfig().API.Hunter.Key == "" {
		return "no hunter api"
	}
	ips := h.IpResult.SaveResult(portscan.Config{OrgId: h.Config.OrgId, WorkspaceId: h.Config.WorkspaceId})
	domains := h.DomainResult.SaveResult(domainscan.Config{OrgId: h.Config.OrgId, WorkspaceId: h.Config.WorkspaceId})

	return fmt.Sprintf("%s,%s", ips, domains)
}

// parseIpPort 解析搜索结果中的IP记录
func parseIpPort(ipResult portscan.Result, fsr onlineSearchResult, source string, blackIP *custom.BlackIP) {
	if fsr.IP == "" || !utils.CheckIPV4(fsr.IP) {
		return
	}
	if blackIP != nil && blackIP.CheckBlack(fsr.IP) {
		return
	}
	if fsr.IP == "0.0.0.0" {
		//Protected data, Please contact (service@baimaohui.net)
		return
	}
	if !ipResult.HasIP(fsr.IP) {
		ipResult.SetIP(fsr.IP)
	}
	portNumber, _ := strconv.Atoi(fsr.Port)
	if !ipResult.HasPort(fsr.IP, portNumber) {
		ipResult.SetPort(fsr.IP, portNumber)
	}
	if fsr.Title != "" {
		ipResult.SetPortAttr(fsr.IP, portNumber, portscan.PortAttrResult{
			Source:  source,
			Tag:     "title",
			Content: fsr.Title,
		})
	}
	if fsr.Server != "" {
		ipResult.SetPortAttr(fsr.IP, portNumber, portscan.PortAttrResult{
			Source:  source,
			Tag:     "server",
			Content: fsr.Server,
		})
	}
	if fsr.Banner != "" {
		ipResult.SetPortAttr(fsr.IP, portNumber, portscan.PortAttrResult{
			Source:  source,
			Tag:     "banner",
			Content: fsr.Banner,
		})
	}
}

// parseDomainIP 解析搜索结果中的域名记录
func parseDomainIP(domainResult domainscan.Result, fsr onlineSearchResult, source string, blackDomain *custom.BlackDomain) {
	host := strings.Replace(fsr.Host, "https://", "", -1)
	host = strings.Replace(host, "http://", "", -1)
	host = strings.Replace(host, "/", "", -1)
	domain := strings.Split(host, ":")[0]
	if domain == "" || utils.CheckIPV4(domain) || utils.CheckDomain(domain) == false {
		return
	}
	if blackDomain != nil && blackDomain.CheckBlack(domain) {
		return
	}

	if !domainResult.HasDomain(domain) {
		domainResult.SetDomain(domain)
	}
	domainResult.SetDomainAttr(domain, domainscan.DomainAttrResult{
		Source:  source,
		Tag:     "A",
		Content: fsr.IP,
	})
	if fsr.Title != "" {
		domainResult.SetDomainAttr(domain, domainscan.DomainAttrResult{
			Source:  source,
			Tag:     "title",
			Content: fsr.Title,
		})
	}
	if fsr.Server != "" {
		domainResult.SetDomainAttr(domain, domainscan.DomainAttrResult{
			Source:  source,
			Tag:     "server",
			Content: fsr.Server,
		})
	}
	if fsr.Banner != "" {
		domainResult.SetDomainAttr(domain, domainscan.DomainAttrResult{
			Source:  source,
			Tag:     "banner",
			Content: fsr.Banner,
		})
	}
}

// checkDomainResult 对域名结果中进行过滤，
func checkDomainResult(result map[string]*domainscan.DomainResult) {
	ip2DomainMap := make(map[string]map[string]struct{})
	// 建立解析ip到domain的反向映射Map
	for domain, domainResult := range result {
		for _, attr := range domainResult.DomainAttrs {
			if attr.Tag == "A" {
				ip := attr.Content
				if _, ok := ip2DomainMap[ip]; !ok {
					ip2DomainMap[ip] = make(map[string]struct{})
					ip2DomainMap[ip][domain] = struct{}{}
				} else {
					if _, ok2 := ip2DomainMap[ip][domain]; !ok2 {
						ip2DomainMap[ip][domain] = struct{}{}
					}
				}
			}
		}
	}
	// 如果多个域名解析到同一个IP超过阈值，则过滤掉该结果
	for ip, domains := range ip2DomainMap {
		domainNumbers := len(domains)
		if domainNumbers > SameIpToDomainFilterMax {
			logging.RuntimeLog.Infof("the multiple domain for one same ip:%s,total:%d,ignored!", ip, domainNumbers)
			logging.CLILog.Infof("the multiple domain for one same ip:%s -- %s,ignored!", ip, utils.SetToString(domains))
			for domain := range domains {
				delete(result, domain)
			}
		}
	}
}
