package fingerprint

import (
	"encoding/json"
	"fmt"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/domainscan"
	"github.com/hanc00l/nemo_go/pkg/task/portscan"
	"github.com/hanc00l/nemo_go/pkg/utils"
	"github.com/projectdiscovery/httpx/common/customheader"
	"github.com/projectdiscovery/httpx/runner"
	"github.com/remeh/sizedwaitgroup"
	"math"
	"os"
)

type Httpx struct {
	//Config           Config
	ResultPortScan   portscan.Result
	ResultDomainScan domainscan.Result
}

type HttpxResult struct {
	A           []string `json:"a,omitempty"`
	CNames      []string `json:"cnames,omitempty"`
	Url         string   `json:"url,omitempty"`
	Host        string   `json:"host,omitempty"`
	Title       string   `json:"title,omitempty"`
	WebServer   string   `json:"webserver,omitempty"`
	ContentType string   `json:"content-type,omitempty"`
	StatusCode  int      `json:"status-code,omitempty"`
	FinalUrl    string   `json:"final-url,omitempty"`
	TLSData     *TLS     `json:"tls-grab,omitempty"`
}

type TLS struct {
	DNSName            []string `json:"dns_names,omitempty"`
	CommonName         []string `json:"common_name,omitempty"`
	Organization       []string `json:"organization,omitempty"`
	IssuerCommonName   []string `json:"issuer_common_name,omitempty"`
	IssuerOrganization []string `json:"issuer_organization,omitempty"`
}

// NewHttpx 创建httpx对象
func NewHttpx() *Httpx {
	return &Httpx{}
}

// Do 执行httpx
func (x *Httpx) Do() {
	swg := sizedwaitgroup.New(fpHttpxThreadNumber)

	if x.ResultPortScan.IPResult != nil {
		bport := make(map[int]struct{})
		for _, p := range IgnorePort {
			bport[p] = struct{}{}
		}
		for ipName, ipResult := range x.ResultPortScan.IPResult {
			for portNumber, _ := range ipResult.Ports {
				if _, ok := bport[portNumber]; ok {
					continue
				}
				url := fmt.Sprintf("%v:%v", ipName, portNumber)
				swg.Add()
				go func(ip string, port int, u string) {
					fingerPrintResult := x.RunHttpx(u)
					if len(fingerPrintResult) > 0 {
						for _, fpa := range fingerPrintResult {
							par := portscan.PortAttrResult{
								Source:  "httpx",
								Tag:     fpa.Tag,
								Content: fpa.Content,
							}
							x.ResultPortScan.SetPortAttr(ip, port, par)
							if fpa.Tag == "status" {
								x.ResultPortScan.IPResult[ip].Ports[port].Status = fpa.Content
							}
						}
					}
					swg.Done()
				}(ipName, portNumber, url)
			}
		}
	}
	if x.ResultDomainScan.DomainResult != nil {
		for domain, _ := range x.ResultDomainScan.DomainResult {
			swg.Add()
			go func(d string) {
				fingerPrintResult := x.RunHttpx(d)
				if len(fingerPrintResult) > 0 {
					for _, fpa := range fingerPrintResult {
						dar := domainscan.DomainAttrResult{
							Source:  "httpx",
							Tag:     fpa.Tag,
							Content: fpa.Content,
						}
						x.ResultDomainScan.SetDomainAttr(d, dar)
					}
				}
				swg.Done()
			}(domain)
		}
	}
	swg.Wait()
}

// RunHttpx 调用httpx，获取一个domain的标题指纹
func (x *Httpx) RunHttpx(domain string) []FingerAttrResult {
	resultTempFile := utils.GetTempPathFileName()
	defer os.Remove(resultTempFile)
	inputTempFile := utils.GetTempPathFileName()
	defer os.Remove(inputTempFile)
	err := os.WriteFile(inputTempFile, []byte(domain), 0666)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		return nil
	}

	options := &runner.Options{
		CustomHeaders:       customheader.CustomHeaders{"User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/52.0.2743.116 Safari/537.36 Edge/15.15063"},
		Output:              resultTempFile,
		InputFile:           inputTempFile,
		Retries:             0,
		Threads:             50,
		Timeout:             5,
		ExtractTitle:        true,
		StatusCode:          true,
		FollowRedirects:     true,
		JSONOutput:          true,
		Silent:              true,
		NoColor:             true,
		OutputServerHeader:  true,
		OutputContentType:   true,
		TLSGrab:             true,
		MaxResponseBodySize: math.MaxInt32,
	}
	r, err := runner.New(options)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		return nil
	}
	r.RunEnumeration()
	r.Close()
	result := parseHttpxResult(resultTempFile)
	return result
}

// parseHttpxResult 解析httpx执行结果
func parseHttpxResult(outputTempFile string) (result []FingerAttrResult) {
	content, err := os.ReadFile(outputTempFile)
	if err != nil || len(content) == 0 {
		return result
	}

	resultJSON := HttpxResult{}
	err = json.Unmarshal(content, &resultJSON)
	if err != nil {
		return result
	}
	httpxResultMarshaled, err := json.Marshal(resultJSON)
	if err == nil {
		result = append(result, FingerAttrResult{
			Tag:     "httpx",
			Content: string(httpxResultMarshaled),
		})
	}
	if resultJSON.Title != "" {
		result = append(result, FingerAttrResult{
			Tag:     "title",
			Content: resultJSON.Title,
		})
	}
	if resultJSON.WebServer != "" {
		result = append(result, FingerAttrResult{
			Tag:     "server",
			Content: resultJSON.WebServer,
		})
	}
	if resultJSON.StatusCode != 0 {
		result = append(result, FingerAttrResult{
			Tag:     "status",
			Content: fmt.Sprintf("%v", resultJSON.StatusCode),
		})
	}
	return result
}
