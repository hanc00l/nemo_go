package domainscan

import (
	"fmt"
	"github.com/hanc00l/nemo_go/pkg/db"
	"sync"
)

const (
	resolveThreadNumber   = 100
	subfinderThreadNumber = 4
	massdnsThreadNumber   = 1
)

// Config 端口扫描的参数配置
type Config struct {
	Target             string `json:"target"`
	OrgId              *int   `json:"orgId"`
	IsSubDomainFinder  bool   `json:"subfinder"`
	IsSubDomainBrute   bool   `json:"subdomainBrute"`
	IsJSFinder         bool   `json:"jsfinder"`
	IsHttpx            bool   `json:"httpx"`
	IsWhatWeb          bool   `json:"whatweb"`
	IsIPPortScan       bool   `json:"portscan"`
	IsIPSubnetPortScan bool   `json:"subnetPortscan"`
	IsScreenshot       bool   `json:"screenshot"`
	IsWappalyzer  bool   `json:"wappalyzer"`
}

// DomainAttrResult 域名属性结果
type DomainAttrResult struct {
	RelatedId int
	Source    string
	Tag       string
	Content   string
}

//DomainResult 域名结果
type DomainResult struct {
	OrgId       *int
	DomainAttrs []DomainAttrResult
}

//Result 域名结果
type Result struct {
	sync.RWMutex
	DomainResult map[string]*DomainResult
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

// SaveResult 保存域名结果
func (r *Result) SaveResult(config Config) string {
	var resultDomainCount int
	for domainName, domainResult := range r.DomainResult {
		domain := &db.Domain{
			DomainName: domainName,
			OrgId:      config.OrgId,
		}
		if !domain.SaveOrUpdate() {
			continue
		}
		resultDomainCount++
		for _, domainAttrResult := range domainResult.DomainAttrs {
			domainAttr := &db.DomainAttr{
				RelatedId: domain.Id,
				Source:    domainAttrResult.Source,
				Tag:       domainAttrResult.Tag,
				Content:   domainAttrResult.Content,
			}
			domainAttr.SaveOrUpdate()
		}
	}

	return fmt.Sprintf("domain:%d", resultDomainCount)
}
