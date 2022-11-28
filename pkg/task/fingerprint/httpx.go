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
	"math"
	"os"
	"strconv"
)

type Httpx struct {
	ResultPortScan   portscan.Result
	ResultDomainScan domainscan.Result
	DomainTargetPort map[string]map[int]struct{}
	//保存响应的数据，用于自定义指纹匹配
	StoreResponse          bool
	StoreResponseDirectory string
	FingerPrintFunc        []func(domain string, ip string, port int, url string, result []FingerAttrResult) []string
}

/*
{"timestamp":"2022-11-28T09:45:09.742937+08:00",
"tls":{"host":"www.baidu.com","port":"443","probe_status":true,"tls_version":"tls12","cipher":"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
"not_before":"2022-07-05T05:16:02Z","not_after":"2023-08-06T05:16:01Z",
"subject_dn":"CN=baidu.com, O=Beijing Baidu Netcom Science Technology Co.\\, Ltd, OU=service operation department, L=beijing, ST=beijing, C=CN","subject_cn":"baidu.com",
"subject_org":["Beijing Baidu Netcom Science Technology Co., Ltd"],
"subject_an":["baidu.com","baifubao.com","www.baidu.cn","www.baidu.com.cn","mct.y.nuomi.com","apollo.auto","dwz.cn","*.baidu.com","*.baifubao.com","*.baidustatic.com","*.bdstatic.com","*.bdimg.com","*.hao123.com","*.nuomi.com","*.chuanke.com","*.trustgo.com","*.bce.baidu.com","*.eyun.baidu.com","*.map.baidu.com","*.mbd.baidu.com","*.fanyi.baidu.com","*.baidubce.com","*.mipcdn.com","*.news.baidu.com","*.baidupcs.com","*.aipage.com","*.aipage.cn","*.bcehost.com","*.safe.baidu.com","*.im.baidu.com","*.baiducontent.com","*.dlnel.com","*.dlnel.org","*.dueros.baidu.com","*.su.baidu.com","*.91.com","*.hao123.baidu.com","*.apollo.auto","*.xueshu.baidu.com","*.bj.baidubce.com","*.gz.baidubce.com","*.smartapps.cn","*.bdtjrcv.com","*.hao222.com","*.haokan.com","*.pae.baidu.com","*.vd.bdstatic.com","*.cloud.baidu.com","click.hm.baidu.com","log.hm.baidu.com","cm.pos.baidu.com","wn.pos.baidu.com","update.pan.baidu.com"],
"issuer_dn":"CN=GlobalSign RSA OV SSL CA 2018, O=GlobalSign nv-sa, C=BE","issuer_cn":"GlobalSign RSA OV SSL CA 2018",
"issuer_org":["GlobalSign nv-sa"],
"fingerprint_hash":{"md5":"ed1949098287a63d206f549a22918c38","sha1":"486aedd16852e5974fa09246b33c56463dd99cd5","sha256":"9ee66a02e0af04405c3e9570b039427af237cab3404d42d56dad235969ce626a"},"wildcard_certificate":true,"tls_connection":"ctls","sni":"www.baidu.com"},
"hash":{"body_md5":"d41d8cd98f00b204e9800998ecf8427e","body_mmh3":"-1840324437","body_sha256":"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855","body_simhash":"18446744073709551615","header_md5":"a211a24f0a8916bb3a9b405237f246fd","header_mmh3":"808001881","header_sha256":"236458af9411bdf5737baf03f8a7fe5ce2bdf47eddf8018c024a46a2c8262877","header_simhash":"11021017595241015215"},
"port":"443",
"url":"https://www.baidu.com:443",
"input":"www.baidu.com:443",
"title":"百度一下，你就知道",
"scheme":"https",
"webserver":"BWS/1.1",
"content_type":"text/html",
"method":"GET",
"host":"103.235.46.40",
"path":"/",
"time":"2.942597075s",
"a":["103.235.46.40"],
"cname":["www.a.shifen.com","www.wshifen.com"],
"words":1,"lines":1,
"status_code":200,"failed":false}
*/

type HttpxResult struct {
	A           []string `json:"a,omitempty"`
	CNames      []string `json:"cnames,omitempty"`
	Url         string   `json:"url,omitempty"`
	Host        string   `json:"host,omitempty"`
	Port        string   `json:"port,omitempty"`
	Title       string   `json:"title,omitempty"`
	WebServer   string   `json:"webserver,omitempty"`
	ContentType string   `json:"content_type,omitempty"`
	StatusCode  int      `json:"status_code,omitempty"`
	TLSData     *TLS     `json:"tls,omitempty"`
}

type TLS struct {
	SubjectDNSName           []string `json:"subject_an,omitempty"`
	SubjectCommonName        string   `json:"subject_cn,omitempty"`
	SubjectDistinguishedName string   `json:"subject_dn,omitempty"`
	SubjectOrganization      []string `json:"subject_org,omitempty"`
	IssuerDistinguishedName  string   `json:"issuer_dn,omitempty"`
	IssuerOrganization       []string `json:"issuer_org,omitempty"`
}

// NewHttpx 创建httpx对象
func NewHttpx() *Httpx {
	return &Httpx{}
}

// Do 执行httpx
func (x *Httpx) Do() {
	swg := sizedwaitgroup.New(fpHttpxThreadNumber)

	if x.ResultPortScan.IPResult != nil {
		for ipName, ipResult := range x.ResultPortScan.IPResult {
			for portNumber, _ := range ipResult.Ports {
				if _, ok := blankPort[portNumber]; ok {
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
						//处理自定义的finger
						if x.StoreResponse && len(x.FingerPrintFunc) > 0 {
							for _, f := range x.FingerPrintFunc {
								xfars := f("", ip, port, u, fingerPrintResult)
								if len(xfars) > 0 {
									for _, fps := range xfars {
										par := portscan.PortAttrResult{
											Source:  "httpxfinger",
											Tag:     "fingerprint",
											Content: fps,
										}
										x.ResultPortScan.SetPortAttr(ip, port, par)
									}
								}
							}
						}
					}
					swg.Done()
				}(ipName, portNumber, url)
			}
		}
	}
	if x.ResultDomainScan.DomainResult != nil {
		if x.DomainTargetPort == nil {
			x.DomainTargetPort = make(map[string]map[int]struct{})
		}
		for domain := range x.ResultDomainScan.DomainResult {
			//如果无域名对应的端口，默认80和443
			if _, ok := x.DomainTargetPort[domain]; !ok || len(x.DomainTargetPort[domain]) == 0 {
				x.DomainTargetPort[domain] = make(map[int]struct{})
				x.DomainTargetPort[domain][80] = struct{}{}
				x.DomainTargetPort[domain][443] = struct{}{}
			}
			for port := range x.DomainTargetPort[domain] {
				if _, ok := blankPort[port]; ok {
					continue
				}
				url := fmt.Sprintf("%s:%d", domain, port)
				swg.Add()
				go func(d string, u string) {
					fingerPrintResult := x.RunHttpx(u)
					if len(fingerPrintResult) > 0 {
						for _, fpa := range fingerPrintResult {
							dar := domainscan.DomainAttrResult{
								Source:  "httpx",
								Tag:     fpa.Tag,
								Content: fpa.Content,
							}
							x.ResultDomainScan.SetDomainAttr(d, dar)
						}
						//处理自定义的finger
						if x.StoreResponse && len(x.FingerPrintFunc) > 0 {
							for _, f := range x.FingerPrintFunc {
								xfars := f(d, "", 0, u, fingerPrintResult)
								if len(xfars) > 0 {
									for _, fps := range xfars {
										dar := domainscan.DomainAttrResult{
											Source:  "httpxfinger",
											Tag:     "fingerprint",
											Content: fps,
										}
										x.ResultDomainScan.SetDomainAttr(d, dar)
									}
								}
							}
						}
					}
					swg.Done()
				}(domain, url)
			}
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
		CustomHeaders:             customheader.CustomHeaders{"User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/52.0.2743.116 Safari/537.36 Edge/15.15063"},
		Output:                    resultTempFile,
		InputFile:                 inputTempFile,
		Retries:                   0,
		Threads:                   50,
		Timeout:                   5,
		ExtractTitle:              true,
		StatusCode:                true,
		FollowRedirects:           true,
		JSONOutput:                true,
		Silent:                    true,
		NoColor:                   true,
		OutputServerHeader:        true,
		OutputContentType:         true,
		TLSGrab:                   true,
		StoreResponse:             x.StoreResponse,
		StoreResponseDir:          x.StoreResponseDirectory,
		MaxResponseBodySizeToSave: math.MaxInt32,
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
