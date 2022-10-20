package pocscan

import (
	"encoding/json"
	"fmt"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/poclib"
	"github.com/remeh/sizedwaitgroup"
	"os"
	"path"
)

type XrayPoc struct {
	ResultPortScan   PortscanVulResult
	ResultDomainScan DomainscanVulResult
	VulResult        []Result
	pocFiles         []Poc
}

type XrayPocConfig struct {
	IPPortResult map[string][]int
	DomainResult []string
}

type Poc struct {
	PocFileName string `json:"name"`
	PocString   string `json:"poc"`
}

// NewXrayPoc 创建xraypoc对象
func NewXrayPoc(config XrayPocConfig) *XrayPoc {
	p := &XrayPoc{
		ResultPortScan:   PortscanVulResult{IPResult: make(map[string]*IPResult)},
		ResultDomainScan: DomainscanVulResult{DomainResult: make(map[string]*DomainResult)},
	}
	for ip, ports := range config.IPPortResult {
		p.ResultPortScan.SetIP(ip)
		for _, port := range ports {
			p.ResultPortScan.SetPort(ip, port)
		}
	}
	for _, domain := range config.DomainResult {
		p.ResultDomainScan.SetDomain(domain)
	}
	p.loadCustomPoc()
	return p
}

// loadCustomPocs 从本地加载Poc
func (p *XrayPoc) loadCustomPoc() {
	pocJsonPathFile := path.Join(conf.GetRootPath(), "thirdparty/custom", "web_xraypoc_v1.json")
	pocContent, err := os.ReadFile(pocJsonPathFile)
	if err != nil {
		logging.CLILog.Error(err)
		return
	}
	if err = json.Unmarshal(pocContent, &p.pocFiles); err != nil {
		logging.CLILog.Error(err)
		return
	}
	logging.CLILog.Infof("Load xray poc total:%d", len(p.pocFiles))
}

// Do 执行poc扫描任务
func (p *XrayPoc) Do() {
	swg := sizedwaitgroup.New(pocMaxThreadNumber)

	if p.ResultPortScan.IPResult != nil {
		// 每一个IP
		for ipName, ipResult := range p.ResultPortScan.IPResult {
			// 每一个port
			for portNumber, _ := range ipResult.Ports {
				//检查该IP已命中的poc数量是否honeyport
				if p.checkPortscanVulResultForHoneyPort(ipName, 0) {
					break
				}
				url := fmt.Sprintf("%v:%v", ipName, portNumber)
				//测试每一个poc
				for _, poc := range p.pocFiles {
					//检查该端口命中的poc数量是否是honeyport
					if p.checkPortscanVulResultForHoneyPort(ipName, portNumber) {
						break
					}
					swg.Add()
					go func(ip string, port int, u string, poc Poc) {
						if success, _ := p.runXrayCheck(u, poc); success {
							p.ResultPortScan.SetPortVul(ip, port, poc.PocFileName)
						}
						swg.Done()
					}(ipName, portNumber, url, poc)
				}
			}
		}
	}

	if p.ResultDomainScan.DomainResult != nil {
		// 每一个域名
		for domain := range p.ResultDomainScan.DomainResult {
			//测试每一个poc
			for _, poc := range p.pocFiles {
				//检查域名命中的poc数量是否是honeyport
				if p.checkDomainscanVulResultForHoneyPort(domain) {
					break
				}
				swg.Add()
				go func(u string, poc Poc) {
					if success, _ := p.runXrayCheck(u, poc); success {
						p.ResultDomainScan.SetDomainVul(u, poc.PocFileName)
					}
					swg.Done()
				}(domain, poc)
			}
		}
	}
	swg.Wait()

	p.exportVulResult()
}

// runXrayCheck 调用xray poc测试代码
func (p *XrayPoc) runXrayCheck(url string, poc Poc) (status bool, name string) {
	u := fmt.Sprintf("http://%s", url)
	status, name = poclib.Execute(u, []byte(poc.PocString), poclib.Content{})
	return
}

// checkPortscanVulResultForHoneyPort 根据poc测试结果，检查是否中honeyport
func (p *XrayPoc) checkPortscanVulResultForHoneyPort(ip string, port int) bool {
	p.ResultPortScan.Lock()
	defer p.ResultPortScan.Unlock()

	if ip != "" && port > 0 {
		if len(p.ResultPortScan.IPResult[ip].Ports[port].Vuls) > portMaxPocForHoneypot {
			return true
		}
		return false
	}
	if ip != "" {
		var c int
		for _, p := range p.ResultPortScan.IPResult[ip].Ports {
			c += len(p.Vuls)
		}
		if c > ipMaxPocForHoneypot {
			return true
		}
	}
	return false
}

// checkDomainscanVulResultForHoneyPort 根据poc测试结果，检查是否中honeyport
func (p *XrayPoc) checkDomainscanVulResultForHoneyPort(domain string) bool {
	p.ResultDomainScan.Lock()
	defer p.ResultDomainScan.Unlock()

	if len(p.ResultDomainScan.DomainResult[domain].Vuls) > domainMaxPocForHoneypot {
		return true
	}
	return false
}

// exportVulResult 将测试结果导出为pocscan格式用于存入到数据库中
func (p *XrayPoc) exportVulResult() {
	if p.ResultPortScan.IPResult != nil {
		for ipName, ipResult := range p.ResultPortScan.IPResult {
			//检查该IP已命中的poc数量是否honeyport
			if p.checkPortscanVulResultForHoneyPort(ipName, 0) {
				logging.RuntimeLog.Warnf("%s has too much poc vul,discard!", ipName)
				logging.CLILog.Warnf("%s has too much poc vul,discard!", ipName)
				continue
			}
			for portNumber, portResult := range ipResult.Ports {
				//检查该port已命中的poc数量是否honeyport
				if p.checkPortscanVulResultForHoneyPort(ipName, portNumber) {
					logging.RuntimeLog.Warnf("%s:%d has too much poc vul,discard!", ipName, portNumber)
					logging.CLILog.Warnf("%s:%d has too much poc vul,discard!", ipName, portNumber)
					continue
				}
				for _, v := range portResult.Vuls {
					p.VulResult = append(p.VulResult, Result{
						Target:  fmt.Sprintf("%s:%d", ipName, portNumber),
						Url:     fmt.Sprintf("%s:%d", ipName, portNumber),
						PocFile: v,
						Source:  "xraypoc",
					})
				}
			}
		}
	}
	if p.ResultDomainScan.DomainResult != nil {
		for domain := range p.ResultDomainScan.DomainResult {
			//检查域名命中的poc数量是否是honeyport
			if p.checkDomainscanVulResultForHoneyPort(domain) {
				logging.RuntimeLog.Warnf("%s has too much poc vul,discard!", domain)
				logging.CLILog.Warnf("%s has too much poc vul,discard!", domain)
				continue
			}
			for _, v := range p.ResultDomainScan.DomainResult[domain].Vuls {
				p.VulResult = append(p.VulResult, Result{
					Target:  domain,
					Url:     domain,
					PocFile: v,
					Source:  "xraypoc",
				})
			}
		}
	}
}
