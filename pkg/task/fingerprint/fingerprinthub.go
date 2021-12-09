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
	"runtime"
)

type FingerprintHub struct {
	ResultPortScan   portscan.Result
	ResultDomainScan domainscan.Result
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

// NNewFingerprintHub 创建FingerprintHub对象
func NewFingerprintHub() *FingerprintHub {
	return &FingerprintHub{}
}

// Do 调用ObserverWard，获取指纹
func (f *FingerprintHub) Do() {
	swg := sizedwaitgroup.New(fpObserverWardThreadNumber)

	if f.ResultPortScan.IPResult != nil {
		bport := make(map[int]struct{})
		for _, p := range IgnorePort {
			bport[p] = struct{}{}
		}
		for ipName, ipResult := range f.ResultPortScan.IPResult {
			for portNumber, _ := range ipResult.Ports {
				if _, ok := bport[portNumber]; ok {
					continue
				}
				url := fmt.Sprintf("%v:%v", ipName, portNumber)
				swg.Add()
				go func(ip string, port int, u string) {
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
					swg.Done()
				}(ipName, portNumber, url)
			}
		}
	}
	if f.ResultDomainScan.DomainResult != nil {
		for domain, _ := range f.ResultDomainScan.DomainResult {
			swg.Add()
			go func(d string) {
				fingerPrintResult := f.RunObserverWard(d)
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
				swg.Done()
			}(domain)
		}
	}
	swg.Wait()
}

// RunObserverWard 调用ObserverWard，获取一个目标的指纹
func (f *FingerprintHub) RunObserverWard(url string) []FingerprintHubReult {
	resultTempFile := utils.GetTempPathFileName()
	defer os.Remove(resultTempFile)

	observerWardBinPath := filepath.Join(conf.GetRootPath(), "thirdparty/fingerprinthub", "observer_ward_darwin")
	if runtime.GOOS == "linux" {
		observerWardBinPath = filepath.Join(conf.GetRootPath(), "thirdparty/fingerprinthub", "observer_ward_amd64")
	}
	var cmdArgs []string
	cmdArgs = append(cmdArgs, "-t", url, "-j", resultTempFile)
	cmd := exec.Command(observerWardBinPath, cmdArgs...)
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
