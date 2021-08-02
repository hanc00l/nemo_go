package fingerprint

import (
	"fmt"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/domainscan"
	"github.com/hanc00l/nemo_go/pkg/task/portscan"
	"github.com/hanc00l/nemo_go/pkg/utils"
	"github.com/remeh/sizedwaitgroup"
	"github.com/rverton/webanalyze"
	"os"
	"path/filepath"
	"sync"
)

const (
	technologiesDefault = "technologies.json"
	technologiesCustom  = "technologies_custom.json"
	fingerPrintFilterNumber = 20
)

type Wappalyzer struct {
	defaultWebAnalyzer *webanalyze.WebAnalyzer
	customWebAnalyzer  *webanalyze.WebAnalyzer
	ResultPortScan     portscan.Result
	ResultDomainScan   domainscan.Result
}

// NewWappalyzer 创建Wappalyzer对象
func NewWappalyzer() *Wappalyzer {
	return &Wappalyzer{}
}

var fingperMapMutex sync.Mutex

// Do 执行任务
func (w *Wappalyzer) Do() {
	technologiesDefaultFile := filepath.Join(conf.GetRootPath(), "thirdparty/wappalyzer", technologiesDefault)
	technologiesCustomFile := filepath.Join(conf.GetRootPath(), "thirdparty/wappalyzer", technologiesCustom)
	defaultAppsFile, err := os.Open(technologiesDefaultFile)
	if err == nil {
		defer defaultAppsFile.Close()
		w.defaultWebAnalyzer, _ = webanalyze.NewWebAnalyzer(defaultAppsFile, nil)
	}
	customAppsFile, err := os.Open(technologiesCustomFile)
	if err == nil {
		defer customAppsFile.Close()
		w.customWebAnalyzer, _ = webanalyze.NewWebAnalyzer(customAppsFile, nil)
	}
	if defaultAppsFile == nil && customAppsFile == nil {
		logging.RuntimeLog.Errorf("load %s and %s both fail", technologiesDefault, technologiesCustom)
		return
	}
	swg := sizedwaitgroup.New(fpWappalyzerThreadNumber)
	if w.ResultPortScan.IPResult != nil {
		bport := make(map[int]struct{})
		for _, p := range IgnorePort {
			bport[p] = struct{}{}
		}
		for ipName, ipResult := range w.ResultPortScan.IPResult {
			for portNumber, _ := range ipResult.Ports {
				if _, ok := bport[portNumber]; ok {
					continue
				}
				url := fmt.Sprintf("%v:%v", ipName, portNumber)
				swg.Add()
				go func(ip string, port int, u string) {
					fingerPrintResult := w.RunWappalyzer(u)
					if fingerPrintResult != "" {
						par := portscan.PortAttrResult{
							Source:  "wappalyzer",
							Tag:     "banner",
							Content: fingerPrintResult,
						}
						w.ResultPortScan.SetPortAttr(ip, port, par)
					}
					swg.Done()
				}(ipName, portNumber, url)
			}
		}
	}
	if w.ResultDomainScan.DomainResult != nil {
		for domain, _ := range w.ResultDomainScan.DomainResult {
			swg.Add()
			go func(d string) {
				fingerPrintResult := w.RunWappalyzer(d)
				if fingerPrintResult != "" {
					dar := domainscan.DomainAttrResult{
						Source:  "wappalyzer",
						Tag:     "banner",
						Content: fingerPrintResult,
					}
					w.ResultDomainScan.SetDomainAttr(d, dar)
				}
				swg.Done()
			}(domain)
		}
	}
	swg.Wait()
}

// RunWappalyzer 对缺省和自定义的指纹库进行匹配
func (w *Wappalyzer) RunWappalyzer(host string) string {
	fingerMap := make(map[string]struct{})
	swg := sizedwaitgroup.New(4)
	for _, schema := range []string{"http", "https"} {
		url := fmt.Sprintf("%s://%s", schema, host)
		if w.defaultWebAnalyzer != nil {
			swg.Add()
			go func(u string) {
				runWappalyzer(u, w.defaultWebAnalyzer, fingerMap)
				swg.Done()
			}(url)
		}
		if w.customWebAnalyzer != nil {
			swg.Add()
			go func(u string){
				runWappalyzer(u, w.customWebAnalyzer, fingerMap)
				swg.Done()
			}(url)
		}
	}
	swg.Wait()
	if len(fingerMap) > fingerPrintFilterNumber {
		logging.RuntimeLog.Errorf("%s has too many result:%d,discarded!", host, len(fingerMap))
		return ""
	}
	return utils.SetToString(fingerMap)
}

func runWappalyzer(host string, wa *webanalyze.WebAnalyzer, fingerMap map[string]struct{}) {
	job := webanalyze.NewOnlineJob(host, "", nil, 0, false, true)
	result, links := wa.Process(job)
	parseResult(fingerMap, result)
	for _, v := range links {
		crawlJob := webanalyze.NewOnlineJob(v, "", nil, 0, false, true)
		resultSubdomain, _ := wa.Process(crawlJob)
		parseResult(fingerMap, resultSubdomain)
	}
}

func parseResult(fingerMap map[string]struct{}, result webanalyze.Result) {
	if result.Error != nil {
		//logging.RuntimeLog.Errorf("%v error: %v\n", result.Host, result.Error)
		return
	}
	for _, a := range result.Matches {
		fingerPrint := fmt.Sprintf("%s", a.AppName)
		if a.Version != "" {
			fingerPrint = fmt.Sprintf("%s/%s", a.AppName, a.Version)
		}
		fingperMapMutex.Lock()
		if _, ok := fingerMap[fingerPrint]; !ok {
			fingerMap[fingerPrint] = struct{}{}
		}
		fingperMapMutex.Unlock()
	}
}
