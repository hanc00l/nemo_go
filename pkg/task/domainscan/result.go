package domainscan

import (
	"encoding/json"
	"fmt"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/db"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/custom"
	"github.com/hanc00l/nemo_go/pkg/utils"
	"strings"
	"sync"
)

const (
	SameIpToDomainFilterMax = 100
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
	IsFingerprintx     bool   `json:"fingerprintx"`
	PortTaskMode       int    `json:"portTaskMode"`
	IsIgnoreCDN        bool   `json:"ignorecdn"`
	IsIgnoreOutofChina bool   `json:"ignoreoutofchina"`
	WorkspaceId        int    `json:"workspaceId"`
	IsProxy            bool   `json:"proxy"`
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
	resolveThreadNumber[conf.HighPerformance] = 200
	resolveThreadNumber[conf.NormalPerformance] = 100
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
	blackDomain := custom.NewBlackTargetCheck(custom.CheckDomain)
	// 用于同步到es的域名
	var ElasticAssets []db.Domain
	for domainName, domainResult := range r.DomainResult {
		if blackDomain.CheckBlack(domainName) {
			logging.RuntimeLog.Warningf("%s is in blacklist,skip...", domainName)
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
		// elastic assets
		if conf.ElasticSyncAssetsChan != nil {
			ElasticAssets = append(ElasticAssets, *domain)
		}
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
	// 将资产同步到Elastic
	if conf.ElasticSyncAssetsChan != nil && len(ElasticAssets) > 0 {
		ElasticAssetsByte, _ := json.Marshal(ElasticAssets)
		syncArgs := conf.ElasticSyncAssetsArgs{
			Contents:       ElasticAssetsByte,
			SyncOp:         conf.SyncOpNew,
			SyncAssetsType: conf.SyncAssetsTypeDomain,
		}
		conf.ElasticSyncAssetsChan <- syncArgs
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("domain:%d", resultDomainCount))
	if newDomain > 0 {
		sb.WriteString(fmt.Sprintf(",domainNew:%d", newDomain))
	}
	return sb.String()
}

// FilterDomainResult 对域名结果进行过滤
func FilterDomainResult(result *Result) { //result map[string]*DomainResult) {
	ip2DomainMap := make(map[string]map[string]struct{})
	// 建立解析ip到domain的反向映射Map
	for domain, domainResult := range result.DomainResult {
		for _, attr := range domainResult.DomainAttrs {
			if attr.Tag == "A" || attr.Tag == "AAAA" {
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
	MaxDomainPerIp := conf.GlobalWorkerConfig().Filter.MaxDomainPerIp
	if MaxDomainPerIp <= 0 {
		MaxDomainPerIp = SameIpToDomainFilterMax
	}
	for ip, domains := range ip2DomainMap {
		domainNumbers := len(domains)
		if domainNumbers > MaxDomainPerIp {
			logging.RuntimeLog.Infof("the multiple domain for one same ip:%s,total:%d,ignored!", ip, domainNumbers)
			logging.CLILog.Infof("the multiple domain for one same ip:%s -- %s,ignored!", ip, utils.SetToString(domains))
			for domain := range domains {
				delete(result.DomainResult, domain)
			}
		}
	}
	//根据标题进行过滤
	titleFilter := strings.Split(conf.GlobalWorkerConfig().Filter.Title, "|")
	// strings.split始终返回len()>=1，即使是空字符串
	if titleFilter[0] != "" {
		for domain, domainResult := range result.DomainResult {
			domainHadFiltered := false
			for _, attr := range domainResult.DomainAttrs {
				if domainHadFiltered {
					break
				}
				if attr.Tag == "title" {
					for _, title := range titleFilter {
						if titleTrim := strings.TrimSpace(title); titleTrim != "" {
							if strings.Contains(attr.Content, titleTrim) {
								logging.RuntimeLog.Warningf("domain:%s has filter title:%s,discard to save!", domain, title)
								delete(result.DomainResult, domain)
								domainHadFiltered = true
								break
							}
						}
					}
				}
			}
		}
	}
}
