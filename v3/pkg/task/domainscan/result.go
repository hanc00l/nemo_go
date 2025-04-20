package domainscan

import (
	"fmt"
	"github.com/hanc00l/nemo_go/v3/pkg/conf"
	"github.com/hanc00l/nemo_go/v3/pkg/db"
	"github.com/hanc00l/nemo_go/v3/pkg/logging"
	"github.com/hanc00l/nemo_go/v3/pkg/task/custom"
	"github.com/hanc00l/nemo_go/v3/pkg/task/execute"
	"github.com/hanc00l/nemo_go/v3/pkg/utils"
	"sync"
)

var (
	resolveThreadNumber   = make(map[string]int)
	subfinderThreadNumber = make(map[string]int)
	massdnsThreadNumber   = make(map[string]int)
	massdnsRunnerThreads  = make(map[string]int)
)

// DomainAttrResult 域名属性结果
type DomainAttrResult struct {
	Source  string
	Tag     string
	Content string
}

// DomainResult 域名结果
type DomainResult struct {
	Org                 string
	Ip                  db.IP
	DomainAttrs         []DomainAttrResult
	DomainAttrsWithPort map[int][]DomainAttrResult
}

// Result 域名结果
type Result struct {
	sync.RWMutex
	DomainResult map[string]*DomainResult
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
	r.DomainResult[domain].DomainAttrsWithPort = make(map[int][]DomainAttrResult)
}

func (r *Result) SetDomainIP(domain string, ip db.IP) {
	r.Lock()
	defer r.Unlock()

	r.DomainResult[domain].Ip = ip
}

func (r *Result) SetDomainAttr(domain string, dar DomainAttrResult) {
	r.Lock()
	defer r.Unlock()

	r.DomainResult[domain].DomainAttrs = append(r.DomainResult[domain].DomainAttrs, dar)
}

func (r *Result) SetDomainPortAttr(domain string, port int, portAttr []DomainAttrResult) {
	r.Lock()
	defer r.Unlock()

	r.DomainResult[domain].DomainAttrsWithPort[port] = portAttr
}

func (r *Result) ParseResult(config execute.ExecutorTaskInfo) (docs []db.AssetDocument) {
	var domainscanConfig execute.DomainscanConfig
	for _, v := range config.DomainScan {
		domainscanConfig = v
		break
	}
	// 过滤同一个IP解析到多个域名超过阈值的结果
	if domainscanConfig.MaxResolvedDomainPerIP > 0 {
		FilterDomainResult(domainscanConfig.MaxResolvedDomainPerIP, r)
	}

	cdn := custom.NewCDNCheck()
	tld := NewTldExtract()
	for domainName, domainResult := range r.DomainResult {
		rootDomain := tld.ExtractFLD(domainName)
		if rootDomain == "" {
			logging.RuntimeLog.Warningf("从子域名:%s中提取根域名失败", domainName)
			continue
		}
		doc := db.AssetDocument{
			Authority: domainName,
			Host:      utils.GetDomainName(domainName),
			Port:      utils.GetDomainPort(domainName),
			Category:  db.CategoryDomain,
			Ip:        domainResult.Ip,
			Domain:    rootDomain,
			OrgId:     config.OrgId,
			TaskId:    config.MainTaskId,
		}
		// cdn check
		isCDN, cdnName, CName := cdn.Check(doc.Host)
		if domainscanConfig.IsIgnoreCDN && isCDN {
			logging.RuntimeLog.Warningf("忽略CDN域名:%s，CDNName:%s，CName:%s", doc.Host, cdnName, CName)
			continue
		}
		doc.IsCDN = isCDN
		doc.CName = CName
		for _, portAttrResult := range domainResult.DomainAttrs {
			switch portAttrResult.Tag {
			case db.FingerTitle:
				doc.Title = portAttrResult.Content
			case db.FingerApp:
				doc.App = append(doc.App, portAttrResult.Content)
			case db.FingerService:
				doc.Service = portAttrResult.Content
			case db.FingerBanner:
				doc.Banner = portAttrResult.Content
			case db.FingerServer:
				doc.Server = portAttrResult.Content
			}
		}
		docs = append(docs, doc)
		// 处理带端口的域名
		for port, dpAttr := range domainResult.DomainAttrsWithPort {
			docWithPort := db.AssetDocument{
				Authority: fmt.Sprintf("%s:%d", utils.GetDomainName(domainName), port),
				Host:      utils.GetDomainName(domainName),
				Port:      port,
				Category:  db.CategoryDomain,
				Ip:        domainResult.Ip,
				Domain:    rootDomain,
				OrgId:     config.OrgId,
				TaskId:    config.TaskId,
			}
			docWithPort.IsCDN = isCDN
			docWithPort.CName = CName
			for _, portAttr := range dpAttr {
				switch portAttr.Tag {
				case db.FingerTitle:
					docWithPort.Title = portAttr.Content
				case db.FingerApp:
					docWithPort.App = append(docWithPort.App, portAttr.Content)
				case db.FingerService:
					docWithPort.Service = portAttr.Content
				case db.FingerBanner:
					docWithPort.Banner = portAttr.Content
				case db.FingerServer:
					docWithPort.Server = portAttr.Content
				}
			}
			docs = append(docs, docWithPort)
		}
	}

	return
}

// FilterDomainResult 反向对域名结果进行过滤
func FilterDomainResult(maxDomainPerIp int, result *Result) {
	ip2DomainMap := make(map[string]map[string]struct{})
	// 建立解析ip到domain的反向映射Map
	for domain, domainResult := range result.DomainResult {
		for _, ipv4 := range domainResult.Ip.IpV4 {
			if _, ok := ip2DomainMap[ipv4.IPName]; !ok {
				ip2DomainMap[ipv4.IPName] = make(map[string]struct{})
				ip2DomainMap[ipv4.IPName][domain] = struct{}{}
			} else {
				if _, ok2 := ip2DomainMap[ipv4.IPName][domain]; !ok2 {
					ip2DomainMap[ipv4.IPName][domain] = struct{}{}
				}
			}
		}
		for _, ipv6 := range domainResult.Ip.IpV6 {
			if _, ok := ip2DomainMap[ipv6.IPName]; !ok {
				ip2DomainMap[ipv6.IPName] = make(map[string]struct{})
				ip2DomainMap[ipv6.IPName][domain] = struct{}{}
			} else {
				if _, ok2 := ip2DomainMap[ipv6.IPName][domain]; !ok2 {
					ip2DomainMap[ipv6.IPName][domain] = struct{}{}
				}
			}
		}
	}
	//// 如果多个域名解析到同一个IP超过阈值，则过滤掉该结果
	for ip, domains := range ip2DomainMap {
		domainNumbers := len(domains)
		if domainNumbers > maxDomainPerIp {
			logging.RuntimeLog.Infof("多个域名解析到同一个IP超过阈值，ip:%s,total:%d,ignored!", ip, domainNumbers)
			for domain := range domains {
				delete(result.DomainResult, domain)
			}
		}
	}
}
