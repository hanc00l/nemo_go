package onlineapi

import (
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/custom"
	"github.com/hanc00l/nemo_go/pkg/task/domainscan"
	"github.com/hanc00l/nemo_go/pkg/task/portscan"
	"github.com/hanc00l/nemo_go/pkg/utils"
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
	SearchPageSize     int    `json:"searchpagesize"`
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
	Results      [][]string `json:"results"`
	Size         int        `json:"size"`
	Page         int        `json:"page"`
	Mode         string     `json:"mode"`
	IsError      bool       `json:"error"`
	ErrorMessage string     `json:"errmsg"`
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
	userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.125 Safari/537.36"
	// pageSizeDefault 缺省的每次API查询的分页数量
	pageSizeDefault = 100
	// SameIpToDomainFilterMax 对结果中domain关联的IP进行统计后，当同一IP被域名关联的数量超过阈值后则过滤掉该掉关联的域名
	SameIpToDomainFilterMax = 100
)

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
