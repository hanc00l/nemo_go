package domainscan

import (
	"fmt"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/db"
	"github.com/hanc00l/nemo_go/pkg/task/custom"
	"strings"
	"sync"
)

var (
	resolveThreadNumber   = make(map[string]int)
	subfinderThreadNumber = make(map[string]int)
	massdnsThreadNumber   = make(map[string]int)
	massdnsRunnerThreads  = make(map[string]int)
	crawlerThreadNumber   = make(map[string]int)
)

// Config 端口扫描的参数配置
type Config struct {
	Target             string `json:"target"`
	OrgId              *int   `json:"orgId"`
	IsSubDomainFinder  bool   `json:"subfinder"`
	IsSubDomainBrute   bool   `json:"subdomainBrute"`
	IsCrawler          bool   `json:"crawler"`
	IsHttpx            bool   `json:"httpx"`
	IsIPPortScan       bool   `json:"portscan"`
	IsIPSubnetPortScan bool   `json:"subnetPortscan"`
	IsScreenshot       bool   `json:"screenshot"`
	IsFingerprintHub   bool   `json:"fingerprinthub"`
	IsIconHash         bool   `json:"iconhash"`
	PortTaskMode       int    `json:"portTaskMode"`
	IsIgnoreCDN        bool   `json:"ignorecdn"`
	IsIgnoreOutofChina bool   `json:"ignoreoutofchina"`
	WorkspaceId        int    `json:"workspaceId"`
}

// DomainAttrResult 域名属性结果
type DomainAttrResult struct {
	RelatedId int
	Source    string
	Tag       string
	Content   string
}

type HttpResult struct {
	RelatedId int
	Port      int
	Source    string
	Tag       string
	Content   string
}

// DomainResult 域名结果
type DomainResult struct {
	OrgId       *int
	DomainAttrs []DomainAttrResult
	HttpInfo    []HttpResult
}

// Result 域名结果
type Result struct {
	sync.RWMutex
	DomainResult    map[string]*DomainResult
	ReqResponseList []UrlResponse
}

func init() {
	resolveThreadNumber[conf.HighPerformance] = 100
	resolveThreadNumber[conf.NormalPerformance] = 50
	//
	subfinderThreadNumber[conf.HighPerformance] = 4
	subfinderThreadNumber[conf.NormalPerformance] = 2
	//
	massdnsThreadNumber[conf.HighPerformance] = 1
	massdnsThreadNumber[conf.NormalPerformance] = 1
	//
	massdnsRunnerThreads[conf.HighPerformance] = 600
	massdnsRunnerThreads[conf.NormalPerformance] = 300
	//
	crawlerThreadNumber[conf.HighPerformance] = 2
	crawlerThreadNumber[conf.NormalPerformance] = 1

}

func (r *Result) HasDomain(domain string) bool {
	r.RLock()
	defer r.RUnlock()

	_, ok := r.DomainResult[domain]
	return ok
}

func (r *Result) SetDomain(domain string) {
	r.Lock()
	defer r.Unlock()

	r.DomainResult[domain] = &DomainResult{DomainAttrs: []DomainAttrResult{}}
}

func (r *Result) SetDomainAttr(domain string, dar DomainAttrResult) {
	r.Lock()
	defer r.Unlock()

	r.DomainResult[domain].DomainAttrs = append(r.DomainResult[domain].DomainAttrs, dar)
}

func (r *Result) SetHttpInfo(domain string, result HttpResult) {
	r.Lock()
	defer r.Unlock()

	r.DomainResult[domain].HttpInfo = append(r.DomainResult[domain].HttpInfo, result)
}

// SaveResult 保存域名结果
func (r *Result) SaveResult(config Config) string {
	var resultDomainCount int
	var newDomain int
	blackDomain := custom.NewBlackDomain()
	for domainName, domainResult := range r.DomainResult {
		if blackDomain.CheckBlack(domainName) {
			continue
		}
		domain := &db.Domain{
			DomainName:  domainName,
			OrgId:       config.OrgId,
			WorkspaceId: config.WorkspaceId,
		}
		if ok, isNew := domain.SaveOrUpdate(); !ok {
			continue
		} else {
			if isNew {
				newDomain++
			}
		}
		resultDomainCount++
		// save domain attr
		for _, domainAttrResult := range domainResult.DomainAttrs {
			domainAttr := &db.DomainAttr{
				RelatedId: domain.Id,
				Source:    domainAttrResult.Source,
				Tag:       domainAttrResult.Tag,
				Content:   domainAttrResult.Content,
			}
			if len(domainAttrResult.Content) > db.AttrContentSize {
				domainAttr.Content = domainAttrResult.Content[:db.AttrContentSize]
			} else {
				domainAttr.Content = domainAttrResult.Content
			}
			domainAttr.SaveOrUpdate()
		}
		//save http info
		for _, httpInfoResult := range domainResult.HttpInfo {
			httpInfo := &db.DomainHttp{
				RelatedId: domain.Id,
				Port:      httpInfoResult.Port,
				Source:    httpInfoResult.Source,
				Tag:       httpInfoResult.Tag,
			}
			if len(httpInfoResult.Content) > db.HttpBodyContentSize {
				httpInfo.Content = httpInfoResult.Content[:db.HttpBodyContentSize]
			} else {
				httpInfo.Content = httpInfoResult.Content
			}
			httpInfo.SaveOrUpdate()
		}
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("domain:%d", resultDomainCount))
	if newDomain > 0 {
		sb.WriteString(fmt.Sprintf(",domainNew:%d", newDomain))
	}
	return sb.String()
}
