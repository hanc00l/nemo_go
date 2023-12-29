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
	IsFingerprintx     bool   `json:"fingerprintx"`
	IsIconHash         bool   `json:"iconhash"`
	IsIgnoreCDN        bool   `json:"ignorecdn"`
	IsIgnoreOutofChina bool   `json:"ignoreoutofchina"`
	SearchByKeyWord    bool   `json:"keywordsearch"`
	SearchStartTime    string `json:"searchstarttime"`
	SearchLimitCount   int    `json:"searchlimitcount"`
	SearchPageSize     int    `json:"searchpagesize"`
	WorkspaceId        int    `json:"workspaceId"`
	IsProxy            bool   `json:"proxy"`
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
)

// parseIpPort 解析搜索结果中的IP记录
func parseIpPort(ipResult *portscan.Result, fsr onlineSearchResult, source string, btc *custom.BlackTargetCheck) {
	if fsr.IP == "" || !utils.CheckIP(fsr.IP) {
		return
	}
	if btc != nil && btc.CheckBlack(fsr.IP) {
		logging.RuntimeLog.Warningf("%s is in blacklist,skip...", fsr.IP)
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
func parseDomainIP(domainResult *domainscan.Result, fsr onlineSearchResult, source string, btc *custom.BlackTargetCheck) {
	host := strings.Replace(fsr.Host, "https://", "", -1)
	host = strings.Replace(host, "http://", "", -1)
	host = strings.Replace(host, "/", "", -1)
	domain := strings.Split(host, ":")[0]
	if domain == "" || utils.CheckIP(domain) || !utils.CheckDomain(domain) {
		return
	}
	if btc != nil && btc.CheckBlack(domain) {
		logging.RuntimeLog.Warningf("%s is in blacklist,skip...", domain)
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
