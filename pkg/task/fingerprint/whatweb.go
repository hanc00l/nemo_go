package fingerprint

import (
	"fmt"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/utils"
	"github.com/hanc00l/nemo_go/pkg/task/domainscan"
	"github.com/hanc00l/nemo_go/pkg/task/portscan"
	"github.com/remeh/sizedwaitgroup"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

type Whatweb struct {
	Config           Config
	ResultPortScan   portscan.Result
	ResultDomainScan domainscan.Result
}

// NewWhatweb 创建whatweb对象
func NewWhatweb(config Config) *Whatweb {
	return &Whatweb{Config: config}
}

// Do 执行whatweb
func (w *Whatweb) Do() {
	swg := sizedwaitgroup.New(fpWhatwebThreadNumber)
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
					fingerPrintResult := w.RunWhatweb(u)
					if len(fingerPrintResult) > 0 {
						for _, fpa := range fingerPrintResult {
							w.ResultPortScan.SetPortAttr(ip, port, portscan.PortAttrResult{
								Source:  "whatweb",
								Tag:     fpa.Tag,
								Content: fpa.Content,
							})
							if fpa.Tag == "status" {
								w.ResultPortScan.IPResult[ip].Ports[port].Status = fpa.Content
							}
						}
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
				fingerPrintResult := w.RunWhatweb(d)
				if len(fingerPrintResult) > 0 {
					for _, fpa := range fingerPrintResult {
						w.ResultDomainScan.SetDomainAttr(d, domainscan.DomainAttrResult{
							Source:  "whatweb",
							Tag:     fpa.Tag,
							Content: fpa.Content,
						})
					}
				}
				swg.Done()
			}(domain)
		}
	}
	swg.Wait()
}

// RunWhatweb 调用whatweb获取一个url的标题指纹
func (w *Whatweb) RunWhatweb(url string) []FingerAttrResult {
	resultTempFile := utils.GetTempPathFileName()
	defer os.Remove(resultTempFile)
	var cmdArgs []string
	cmdArgs = append(
		cmdArgs,
		"-q", "--color=never", "--log-brief", resultTempFile, "--max-threads", "4",
		"--open-timeout", "5", "--read-timeout", "10",
		"-U=Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/52.0.2743.116 Safari/537.36 Edge/15.15063",
	)
	cmdArgs = append(cmdArgs, url)
	cmd := exec.Command("whatweb", cmdArgs...)
	_, err := cmd.CombinedOutput()
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		return nil
	}
	faResult := parseWhatwebResult(resultTempFile)
	return faResult
}

// parseWhatwebResult 解析whatweb结果
func parseWhatwebResult(outputTempFile string) (result []FingerAttrResult) {
	content, err := os.ReadFile(outputTempFile)
	if err != nil || len(content) == 0 {
		return result
	}
	result = append(result, FingerAttrResult{
		Tag:     "whatweb",
		Content: string(content),
	})
	keys := map[string]string{"Title": "title", "HTTPServer": "server"}
	for key, tag := range keys {
		fieldRegx := regexp.MustCompile(fmt.Sprintf("%s\\[(.*?)\\]", key))
		m := fieldRegx.FindAllStringSubmatch(string(content), -1)
		values := make(map[string]struct{})
		for _, mm := range m {
			if _, ok := values[mm[1]]; !ok {
				values[mm[1]] = struct{}{}
			}
		}
		var valueList []string
		for k, _ := range values {
			valueList = append(valueList, k)
		}
		txt := strings.Join(valueList, ",")
		if txt != "" {
			result = append(result, FingerAttrResult{
				Tag:     tag,
				Content: txt,
			})
		}
	}
	//get status-code:
	//http://47.98.181.116:80 [200 OK] Bootstrap, Country[CANADA][CA], ...
	statusRegx := regexp.MustCompile(`\[(\d{3})\s.*?\]`)
	m := statusRegx.FindAllStringSubmatch(string(content), -1)
	if len(m) > 0 {
		result = append(result, FingerAttrResult{
			Tag:     "status",
			Content: m[len(m)-1][1],
		})
	}
	return result
}
