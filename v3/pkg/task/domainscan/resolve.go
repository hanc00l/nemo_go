package domainscan

import (
	"github.com/hanc00l/nemo_go/v3/pkg/conf"
	"github.com/hanc00l/nemo_go/v3/pkg/core"
	"github.com/hanc00l/nemo_go/v3/pkg/db"
	"github.com/hanc00l/nemo_go/v3/pkg/task/execute"
	"github.com/hanc00l/nemo_go/v3/pkg/utils"
	"github.com/remeh/sizedwaitgroup"
	"net"
	"strings"
)

type Resolve struct {
	Config execute.DomainscanConfig
}

func (r *Resolve) GetRequiredResources() (re []core.RequiredResource) {
	return
}

func (r *Resolve) IsExecuteFromCmd() bool {
	return false
}

func (r *Resolve) GetExecuteCmd() string {
	return ""
}

func (r *Resolve) GetExecuteArgs(inputTempFile, outputTempFile string) (cmdArgs []string) {
	return
}

func (r *Resolve) Run(target []string) (result Result) {
	result.DomainResult = make(map[string]*DomainResult)
	swg := sizedwaitgroup.New(resolveThreadNumber[conf.WorkerPerformanceMode])
	for _, line := range target {
		domain := strings.TrimSpace(line)
		if domain == "" || utils.CheckIPOrSubnet(domain) {
			continue
		}
		swg.Add()
		go func(d string, result *Result) {
			defer swg.Done()
			r.RunResolve(d, result)
		}(domain, &result)
	}

	swg.Wait()
	return
}

func (r *Resolve) ParseContentResult(content []byte) (result Result) {
	//TODO implement me
	panic("implement me")
}

// RunResolve 解析并保存一个域名结果
func (r *Resolve) RunResolve(domain string, result *Result) {
	if !result.HasDomain(domain) {
		result.SetDomain(domain)
	}
	_, host := ResolveDomain(domain)
	if len(host) > 0 {
		var ip db.IP
		for _, h := range host {
			if utils.CheckIPV4(h) {
				ip.IpV4 = append(ip.IpV4, db.IPV4{
					IPName: h,
				})
			} else if utils.CheckIPV6(h) {
				ip.IpV6 = append(ip.IpV6, db.IPV6{
					IPName: utils.GetIPV6ParsedFormat(h),
				})
			}
		}
		result.SetDomainIP(domain, ip)
	}
}

// ResolveDomain 解析一个域名的A记录和CNAME记录
func ResolveDomain(domain string) (CName string, Host []string) {
	CName, _ = net.LookupCNAME(domain)
	Host, _ = net.LookupHost(domain)
	return
}
