package domainscan

import (
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/task/custom"
	"github.com/hanc00l/nemo_go/pkg/utils"
	"github.com/remeh/sizedwaitgroup"
	"net"
	"strings"
)

type Resolve struct {
	Config Config
	Result Result
}

// NewResolve 创建resolve对象
func NewResolve(config Config) *Resolve {
	return &Resolve{Config: config}
}

// Do 执行域名解析
func (r *Resolve) Do() {
	swg := sizedwaitgroup.New(resolveThreadNumber[conf.WorkerPerformanceMode])
	// 如果Result中已有map[domain]*DomainResult，则遍历并解析域名
	if r.Result.DomainResult != nil {
		for domain, _ := range r.Result.DomainResult {
			swg.Add()
			go func(d string) {
				defer swg.Done()
				r.RunResolve(d)
			}(domain)
		}
	} else {
		// 解析Config中的域名
		r.Result.DomainResult = make(map[string]*DomainResult)
		for _, line := range strings.Split(r.Config.Target, ",") {
			domain := strings.TrimSpace(line)
			if domain == "" || utils.CheckIPV4(domain) || utils.CheckIPV4Subnet(domain) {
				continue
			}
			swg.Add()
			go func(d string) {
				defer swg.Done()
				r.RunResolve(d)
			}(domain)
		}
	}
	swg.Wait()
}

// RunResolve 解析并保存一个域名结果
func (r *Resolve) RunResolve(domain string) {
	if !r.Result.HasDomain(domain) {
		r.Result.SetDomain(domain)
	}
	_, host := ResolveDomain(domain)
	if len(host) > 0 {
		for _, h := range host {
			r.Result.SetDomainAttr(domain, DomainAttrResult{
				Source:  "domainscan",
				Tag:     "A",
				Content: h,
			})
		}
	}
	//CDN检查
	cdnCheck := custom.NewCDNCheck()
	isCDN, cdnName, CName := cdnCheck.CheckCName(domain)
	if isCDN {
		r.Result.SetDomainAttr(domain, DomainAttrResult{
			Source:  "domainscan",
			Tag:     "CDN",
			Content: cdnName,
		})
	}
	if CName != "" {
		r.Result.SetDomainAttr(domain, DomainAttrResult{
			Source:  "domainscan",
			Tag:     "CNAME",
			Content: CName,
		})
	}
}

// ResolveDomain 解析一个域名的A记录和CNAME记录
func ResolveDomain(domain string) (CName string, Host []string) {
	//CName, _ = net.LookupCNAME(domain)
	host, _ := net.LookupHost(domain)
	for _, h := range host {
		if utils.CheckIPV4(h) {
			Host = append(Host, h)
		}
	}
	return
}
