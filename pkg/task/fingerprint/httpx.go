package fingerprint

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/custom"
	"github.com/hanc00l/nemo_go/pkg/task/domainscan"
	"github.com/hanc00l/nemo_go/pkg/task/portscan"
	"github.com/hanc00l/nemo_go/pkg/utils"
	"github.com/projectdiscovery/httpx/common/customheader"
	"github.com/projectdiscovery/httpx/runner"
	"github.com/remeh/sizedwaitgroup"
	"os"
	"strconv"
)

type Httpx struct {
	ResultPortScan   portscan.Result
	ResultDomainScan domainscan.Result
}

type HttpxResult struct {
	A           []string `json:"a,omitempty"`
	CNames      []string `json:"cnames,omitempty"`
	Url         string   `json:"url,omitempty"`
	Host        string   `json:"host,omitempty"`
	Port        string   `json:"port,omitempty"`
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
		CustomHeaders:      customheader.CustomHeaders{"User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/52.0.2743.116 Safari/537.36 Edge/15.15063"},
		Output:             resultTempFile,
		InputFile:          inputTempFile,
		Retries:            0,
		Threads:            50,
		Timeout:            5,
		ExtractTitle:       true,
		StatusCode:         true,
		FollowRedirects:    true,
		JSONOutput:         true,
		Silent:             true,
		NoColor:            true,
		OutputServerHeader: true,
		OutputContentType:  true,
		TLSGrab:            true,
	}
	r, err := runner.New(options)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		return nil
	}
	r.RunEnumeration()
	r.Close()
	result := x.parseHttpxResult(resultTempFile)
	return result
}

// ParseHttpxJson 解析一条httpx的JSON记录
func (x *Httpx) ParseHttpxJson(content []byte) (host string, port int, result []FingerAttrResult) {
	resultJSON := HttpxResult{}
	err := json.Unmarshal(content, &resultJSON)
	if err != nil {
		return
	}
	// 获取host与port
	host = resultJSON.Host
	port, err = strconv.Atoi(resultJSON.Port)
	if err != nil {
		return
	}
	// 获取全部的Httpx信息
	httpxResultMarshaled, err := json.Marshal(resultJSON)
	if err == nil {
		result = append(result, FingerAttrResult{
			Tag:     "httpx",
			Content: string(httpxResultMarshaled),
		})
	}
	// 解析字段
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
	if resultJSON.TLSData != nil {
		tlsData, err := json.Marshal(resultJSON.TLSData)
		if err == nil {
			result = append(result, FingerAttrResult{
				Tag:     "tlsdata",
				Content: string(tlsData),
			})
		}
	}
	return
}

// parseHttpxResult 解析httpx执行结果
func (x *Httpx) parseHttpxResult(outputTempFile string) (result []FingerAttrResult) {
	content, err := os.ReadFile(outputTempFile)
	if err != nil || len(content) == 0 {
		return result
	}
	// host与port这里不需要
	_, _, result = x.ParseHttpxJson(content)

	return
}

// ParseJSONContentResult 解析httpx扫描的JSON格式文件结果
func (x *Httpx) ParseJSONContentResult(content []byte) {
	s := custom.NewService()
	if x.ResultPortScan.IPResult == nil {
		x.ResultPortScan.IPResult = make(map[string]*portscan.IPResult)
	}
	scanner := bufio.NewScanner(bytes.NewReader(content))
	for scanner.Scan() {
		data := scanner.Bytes()
		host, port, fas := x.ParseHttpxJson(data)
		if host == "" || port == 0 || len(fas) == 0 || utils.CheckIPV4(host) == false {
			continue
		}
		if !x.ResultPortScan.HasIP(host) {
			x.ResultPortScan.SetIP(host)
		}
		if !x.ResultPortScan.HasPort(host, port) {
			x.ResultPortScan.SetPort(host, port)
		}
		service := s.FindService(port, "")
		x.ResultPortScan.SetPortAttr(host, port, portscan.PortAttrResult{
			Source:  "httpx",
			Tag:     "service",
			Content: service,
		})
		for _, fa := range fas {
			par := portscan.PortAttrResult{
				RelatedId: 0,
				Source:    "httpx",
				Tag:       fa.Tag,
				Content:   fa.Content,
			}
			x.ResultPortScan.SetPortAttr(host, port, par)
			if fa.Tag == "status" {
				x.ResultPortScan.IPResult[host].Ports[port].Status = fa.Content
			}
		}
	}
}
