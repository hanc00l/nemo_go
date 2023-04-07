package fingerprint

import (
	"encoding/json"
	"fmt"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/domainscan"
	"github.com/hanc00l/nemo_go/pkg/task/portscan"
	"github.com/hanc00l/nemo_go/pkg/utils"
	"github.com/remeh/sizedwaitgroup"
	"os"
	"os/exec"
	"path/filepath"
)

type FingerprintHub struct {
	ResultPortScan   portscan.Result
	ResultDomainScan domainscan.Result
	DomainTargetPort map[string]map[int]struct{}
}

type FingerprintHubReult struct {
	Url        string   `json:"url"`
	Name       []string `json:"name"`
	Priority   int      `json:"priority"`
	Length     int      `json:"length"`
	Title      string   `json:"title"`
	StatusCode int      `json:"status_code"`
	Plugins    []string `json:"plugins"`
}

// NewFingerprintHub NNewFingerprintHub 创建FingerprintHub对象
func NewFingerprintHub() *FingerprintHub {
	return &FingerprintHub{}
}

// Do 调用ObserverWard，获取指纹
func (f *FingerprintHub) Do() {
	swg := sizedwaitgroup.New(fpObserverWardThreadNumber[conf.WorkerPerformanceMode])

	if f.ResultPortScan.IPResult != nil {
		for ipName, ipResult := range f.ResultPortScan.IPResult {
			for portNumber := range ipResult.Ports {
				if _, ok := blankPort[portNumber]; ok {
					continue
				}
				url := fmt.Sprintf("%v:%v", ipName, portNumber)
				swg.Add()
				go func(ip string, port int, u string) {
					defer swg.Done()
					fingerPrintResult := f.RunObserverWard(u)
					if len(fingerPrintResult) > 0 {
						for _, fpa := range fingerPrintResult {
							for _, name := range fpa.Name {
								par := portscan.PortAttrResult{
									Source:  "ObserverWard",
									Tag:     "fingerprint",
									Content: name,
								}
								f.ResultPortScan.SetPortAttr(ip, port, par)
							}
						}
					}
				}(ipName, portNumber, url)
			}
		}
	}
	if f.ResultDomainScan.DomainResult != nil {
		if f.DomainTargetPort == nil {
			f.DomainTargetPort = make(map[string]map[int]struct{})
		}
		for domain := range f.ResultDomainScan.DomainResult {
			//如果无域名对应的端口，默认80和443
			if _, ok := f.DomainTargetPort[domain]; !ok || len(f.DomainTargetPort[domain]) == 0 {
				f.DomainTargetPort[domain] = make(map[int]struct{})
				f.DomainTargetPort[domain][80] = struct{}{}
				f.DomainTargetPort[domain][443] = struct{}{}
			}
			for port := range f.DomainTargetPort[domain] {
				if _, ok := blankPort[port]; ok {
					continue
				}
				url := fmt.Sprintf("%s:%d", domain, port)
				swg.Add()
				go func(d string, u string) {
					defer swg.Done()
					fingerPrintResult := f.RunObserverWard(u)
					if len(fingerPrintResult) > 0 {
						for _, fpa := range fingerPrintResult {
							for _, name := range fpa.Name {
								dar := domainscan.DomainAttrResult{
									Source:  "ObserverWard",
									Tag:     "fingerprint",
									Content: name,
								}
								f.ResultDomainScan.SetDomainAttr(d, dar)
							}
						}
					}
				}(domain, url)
			}
		}
	}
	swg.Wait()
}

// RunObserverWard 调用ObserverWard，获取一个目标的指纹
func (f *FingerprintHub) RunObserverWard(url string) []FingerprintHubReult {
	resultTempFile := utils.GetTempPathFileName()
	defer os.Remove(resultTempFile)

	//Fix：要指定绝对路径
	observerWardBinPath := filepath.Join(conf.GetAbsRootPath(), "thirdparty/fingerprinthub", utils.GetThirdpartyBinNameByPlatform(utils.ObserverWard))
	var cmdArgs []string
	cmdArgs = append(cmdArgs, "-t", url, "-j", resultTempFile)
	cmd := exec.Command(observerWardBinPath, cmdArgs...)
	//Fix:指定当前路径，这样才会正确调用web_fingerprint_v3.json
	//Fix:必须指定绝对路径
	cmd.Dir = filepath.Join(conf.GetAbsRootPath(), "thirdparty/fingerprinthub")
	_, err := cmd.CombinedOutput()
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		return nil
	}
	return parseObserverWardResult(resultTempFile)
}

// parseObserverWardResult 解析结果
func parseObserverWardResult(outputTempFile string) (result []FingerprintHubReult) {
	content, err := os.ReadFile(outputTempFile)
	if err != nil || len(content) == 0 {
		return
	}
	json.Unmarshal(content, &result)
	return
}
