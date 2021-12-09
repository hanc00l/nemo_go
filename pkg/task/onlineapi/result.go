package onlineapi

import (
	"encoding/json"
	"fmt"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/domainscan"
	"github.com/hanc00l/nemo_go/pkg/task/portscan"
	"github.com/hanc00l/nemo_go/pkg/utils"
	"strconv"
	"strings"
)

type FofaConfig struct {
	Target           string `json:"target"`
	OrgId            *int   `json:"orgId"`
	IsIPLocation     bool   `json:"ipLocation"`
	IsHttpx          bool   `json:"httpx"`
	IsWhatWeb        bool   `json:"whatweb"`
	IsScreenshot     bool   `json:"screenshot"`
	IsWappalyzer     bool   `json:"wappalyzer"`
	IsFingerprintHub bool   `json:"fingerprinthub"`
}

type ICPQueryConfig struct {
	Target string `json:"target"`
}

type fofaSearchResult struct {
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

// parseFofaSearchResult 转换FOFA搜索结果
func (ff *Fofa) parseFofaSearchResult(queryResult []byte) (result []fofaSearchResult) {
	r := fofaQueryResult{}
	err := json.Unmarshal(queryResult, &r)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		return result
	}
	for _, line := range r.Results {
		fsr := fofaSearchResult{
			Domain: line[0], Host: line[1], IP: line[2], Port: line[3], Title: line[4],
			Country: line[5], City: line[6], Server: line[7], Banner: line[8],
		}
		if strings.HasPrefix(fsr.Banner, "HTTP/1") {
			fsr.Banner = ""
		}
		result = append(result, fsr)
	}

	return result
}

// parseIpPort 解析搜索结果中的IP记录
func (ff *Fofa) parseIpPort(fsr fofaSearchResult) {
	if !utils.CheckIPV4(fsr.IP) {
		return
	}

	if !ff.IpResult.HasIP(fsr.IP) {
		ff.IpResult.SetIP(fsr.IP)
	}
	portNumber, _ := strconv.Atoi(fsr.Port)
	if !ff.IpResult.HasPort(fsr.IP, portNumber) {
		ff.IpResult.SetPort(fsr.IP, portNumber)
	}
	if fsr.Title != "" {
		ff.IpResult.SetPortAttr(fsr.IP, portNumber, portscan.PortAttrResult{
			Source:  "fofa",
			Tag:     "title",
			Content: fsr.Title,
		})
	}
	if fsr.Server != "" {
		ff.IpResult.SetPortAttr(fsr.IP, portNumber, portscan.PortAttrResult{
			Source:  "fofa",
			Tag:     "server",
			Content: fsr.Server,
		})
	}
	if fsr.Banner != "" {
		ff.IpResult.SetPortAttr(fsr.IP, portNumber, portscan.PortAttrResult{
			Source:  "fofa",
			Tag:     "banner",
			Content: fsr.Banner,
		})
	}
}

// parseDomainIP 解析搜索结果中的域名记录
func (ff *Fofa) parseDomainIP(fsr fofaSearchResult) {
	host := strings.Replace(fsr.Host, "https://", "", -1)
	host = strings.Replace(host, "http://", "", -1)
	host = strings.Replace(host, "/", "", -1)
	domain := strings.Split(host, ":")[0]
	if utils.CheckIPV4(domain) {
		return
	}

	if !ff.DomainResult.HasDomain(domain) {
		ff.DomainResult.SetDomain(domain)
	}
	ff.DomainResult.SetDomainAttr(domain, domainscan.DomainAttrResult{
		Source:  "fofa",
		Tag:     "A",
		Content: fsr.IP,
	})
	if fsr.Title != "" {
		ff.DomainResult.SetDomainAttr(domain, domainscan.DomainAttrResult{
			Source:  "fofa",
			Tag:     "title",
			Content: fsr.Title,
		})
	}
	if fsr.Server != "" {
		ff.DomainResult.SetDomainAttr(domain, domainscan.DomainAttrResult{
			Source:  "fofa",
			Tag:     "server",
			Content: fsr.Server,
		})
	}
	if fsr.Banner != "" {
		ff.DomainResult.SetDomainAttr(domain, domainscan.DomainAttrResult{
			Source:  "fofa",
			Tag:     "banner",
			Content: fsr.Banner,
		})
	}
}

// parseResult 解析搜索结果
func (ff *Fofa) parseResult() {
	ff.IpResult = portscan.Result{IPResult: make(map[string]*portscan.IPResult)}
	ff.DomainResult = domainscan.Result{DomainResult: make(map[string]*domainscan.DomainResult)}

	for _, fsr := range ff.Result {
		ff.parseIpPort(fsr)
		ff.parseDomainIP(fsr)
	}
}

// SaveResult 保存搜索的结果
func (ff *Fofa) SaveResult() string {
	if conf.GlobalWorkerConfig().API.Fofa.Key == "" || conf.GlobalWorkerConfig().API.Fofa.Name == "" {
		return "no fofa api"
	}
	ips := ff.IpResult.SaveResult(portscan.Config{OrgId: ff.Config.OrgId})
	domains := ff.DomainResult.SaveResult(domainscan.Config{OrgId: ff.Config.OrgId})

	return fmt.Sprintf("%s,%s", ips, domains)
}
