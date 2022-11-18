package pocscan

import (
	"encoding/json"
	"fmt"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/utils"
	"github.com/hanc00l/nemo_go/pkg/xraypocv1"
	xv2 "github.com/hanc00l/nemo_go/pkg/xraypocv2"
	"os"
	"path"
	"path/filepath"
	"sync"
)

var singleMutex sync.Mutex

// XrayPocV2版本，在加载POC时多线程使用存在冲突（具体原因无果）
// 目前在加载时用锁控制

type XrayPoc struct {
	ResultPortScan   PortscanVulResult
	ResultDomainScan DomainscanVulResult
	VulResult        []Result
	pocFiles         []Poc
	pocV2Bytes       [][]byte
}

type XrayPocConfig struct {
	IPPort      map[string][]int
	Domain      map[string]struct{}
	XrayPocFile string
}

type Poc struct {
	PocFileName string `json:"name"`
	PocString   string `json:"poc"`
}

// NewXrayPoc 创建xraypoc对象
func NewXrayPoc(config XrayPocConfig) *XrayPoc {
	singleMutex.Lock()
	defer singleMutex.Unlock()

	p := &XrayPoc{
		ResultPortScan:   PortscanVulResult{IPResult: make(map[string]*IPResult)},
		ResultDomainScan: DomainscanVulResult{DomainResult: make(map[string]*DomainResult)},
	}
	for ip, ports := range config.IPPort {
		p.ResultPortScan.SetIP(ip)
		for _, port := range ports {
			p.ResultPortScan.SetPort(ip, port)
		}
	}
	for domain := range config.Domain {
		p.ResultDomainScan.SetDomain(domain)
	}
	if len(config.XrayPocFile) > 0 {
		p.loadOneXrayPocV2(config.XrayPocFile)
	} else {
		p.loadXrayPocV2()
	}
	return p
}

// loadXrayPocV1 从本地加载Poc
func (p *XrayPoc) loadXrayPocV1() {
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

// loadXrayPocV2 从本地加载Poc
func (p *XrayPoc) loadXrayPocV2() {
	files, _ := filepath.Glob(filepath.Join(conf.GetRootPath(), conf.GlobalWorkerConfig().Pocscan.Xray.PocPath, "*.yml"))
	for _, file := range files {
		pocContent, err := os.ReadFile(file)
		if err != nil {
			logging.CLILog.Error(err)
			continue
		}
		p.pocV2Bytes = append(p.pocV2Bytes, pocContent)

	}
	logging.CLILog.Infof("Load xray poc v2 total:%d", len(p.pocV2Bytes))
}

// loadOneXrayPocV2 从本地加载一个Poc
func (p *XrayPoc) loadOneXrayPocV2(pocFile string) {

	filePathName := filepath.Join(conf.GetRootPath(), conf.GlobalWorkerConfig().Pocscan.Xray.PocPath, pocFile)
	pocContent, err := os.ReadFile(filePathName)
	if err != nil {
		logging.CLILog.Error(err)
		return
	}
	p.pocV2Bytes = append(p.pocV2Bytes, pocContent)
	logging.CLILog.Infof("Load xray poc v2 total:%d", len(p.pocV2Bytes))
}

// Do 执行poc扫描任务
func (p *XrayPoc) Do() {
	if p.ResultPortScan.IPResult != nil {
		// 每一个IP
		for ipName, ipResult := range p.ResultPortScan.IPResult {
			// 每一个port
			for portNumber := range ipResult.Ports {
				//检查该IP已命中的poc数量是否honeyport
				if p.checkPortscanVulResultForHoneyPort(ipName, 0) {
					break
				}
				url := fmt.Sprintf("%v:%v", ipName, portNumber)
				results := p.runXrayCheckV2(url)
				for _, vul := range results {
					//检查该IP的端口已命中的poc数量是否honeyport
					if p.checkPortscanVulResultForHoneyPort(ipName, portNumber) {
						break
					}
					p.ResultPortScan.SetPortVul(ipName, portNumber, vul)
				}
			}
		}
	}

	if p.ResultDomainScan.DomainResult != nil {
		// 每一个域名
		for domain := range p.ResultDomainScan.DomainResult {
			results := p.runXrayCheckV2(domain)
			for _, vul := range results {
				//检查该域名已命中的poc数量是否honeyport
				if p.checkDomainscanVulResultForHoneyPort(domain) {
					break
				}
				p.ResultDomainScan.SetDomainVul(domain, vul)
			}
		}
	}

	p.exportVulResult()
}

// runXrayCheckV1 调用xray poc测试代码
func (p *XrayPoc) runXrayCheckV1(url string, poc Poc) (status bool, name string) {
	protocol := utils.GetProtocol(url, 5)
	status, name = xraypocv1.Execute(fmt.Sprintf("%s://%s", protocol, url), []byte(poc.PocString), xraypocv1.Content{})

	return
}

// runXrayCheckV2 调用xray poc测试代码
func (p *XrayPoc) runXrayCheckV2(url string) (result []string) {
	protocol := utils.GetProtocol(url, 5)
	x := xv2.InitXrayV2Poc("", "", "")
	result = x.RunXrayMultiPocByQuery(fmt.Sprintf("%s://%s", protocol, url), p.pocV2Bytes, []xv2.Content{})

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
