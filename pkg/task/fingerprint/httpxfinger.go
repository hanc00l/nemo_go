package fingerprint

import (
	"encoding/json"
	"fmt"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/domainscan"
	"github.com/hanc00l/nemo_go/pkg/task/portscan"
	"github.com/hanc00l/nemo_go/pkg/utils"
	"github.com/hanc00l/nemo_go/pkg/xraypocv1"
	"github.com/projectdiscovery/urlutil"
	"os"
	"path"
	"strings"
)

// HttpxFinger 基于httpx实现的web应用的fingerprint功能
// 通过httpx获取web的指纹，并保存返回的信息实现自定义扩展
type HttpxFinger struct {
	Httpx
	fpWebFingerprintHub []WebFingerPrint
	fpCustom            []CustomFingerPrint
}

const (
	// The maximum file length is 251 (255 - 4 bytes for ".ext" suffix)
	maxFileNameLength = 251
)

// WebFingerPrint 匹配web_fingerprint_v3.json的指纹结构
// 通过借鉴afrog代码获取fingerprinthub定义的指纹信息
type WebFingerPrint struct {
	Name           string            `json:"name"`
	Path           string            `json:"path"`
	RequestMethod  string            `json:"request_method"`
	RequestHeaders map[string]string `json:"request_headers"`
	RequestData    string            `json:"request_data"`
	StatusCode     int               `json:"status_code"`
	Headers        map[string]string `json:"headers"`
	Keyword        []string          `json:"keyword"`
	FaviconHash    []string          `json:"favicon_hash"`
	Priority       int               `json:"priority"`
}

type CustomFingerPrint struct {
	Id      int     `json:"id"`
	App     string  `json:"app"`
	Rule    string  `json:"rule"`
	Company *string `json:"company,omitempty"`
	RuleId  *int    `json:"rule_id,omitempty"`
}

// NewHttpxFinger 创建对象
func NewHttpxFinger() *HttpxFinger {
	h := &HttpxFinger{}
	h.StoreResponse = true
	//加载自定义指纹及回调函数
	//h.loadFingerprintHub()
	h.loadCustomFingerprint()
	//保存HTTP的header与body到数据库：
	h.FingerPrintFunc = append(h.FingerPrintFunc, h.fingerPrintFuncForSaveHttpHeaderAndBody)

	return h
}

// loadFingerprintHub 加载finerprinthub的指纹及处理函数
func (h *HttpxFinger) loadFingerprintHub() {
	fingerprintJsonPathFile := path.Join(conf.GetRootPath(), "thirdparty/fingerprinthub", "web_fingerprint_v3.json")
	fingerContent, err := os.ReadFile(fingerprintJsonPathFile)
	if err != nil {
		logging.CLILog.Error(err)
		return
	}
	if err = json.Unmarshal(fingerContent, &h.fpWebFingerprintHub); err != nil {
		logging.CLILog.Error(err)
		return
	}
	if len(h.fpWebFingerprintHub) > 0 {
		h.FingerPrintFunc = append(h.FingerPrintFunc, h.fingerPrintFuncForFingerprintHub)
		logging.CLILog.Infof("Load fingerprinthub total:%d", len(h.fpWebFingerprintHub))
	}
}

// loadCustomFingerprint 加载自定义指纹
func (h *HttpxFinger) loadCustomFingerprint() {
	fingerprintJsonPathFile := path.Join(conf.GetRootPath(), "thirdparty/custom", "web_fingerprint.json")
	fingerContent, err := os.ReadFile(fingerprintJsonPathFile)
	if err != nil {
		logging.CLILog.Error(err)
		return
	}
	if err = json.Unmarshal(fingerContent, &h.fpCustom); err != nil {
		logging.CLILog.Error(err)
		return
	}
	if len(h.fpCustom) > 0 {
		h.FingerPrintFunc = append(h.FingerPrintFunc, h.fingerPrintFuncForCustom)
		logging.CLILog.Infof("Load custom web finger total:%d", len(h.fpCustom))
	}
}

// DoHttpxAndFingerPrint 执行指纹识别
func (h *HttpxFinger) DoHttpxAndFingerPrint() {
	// 保存响应结果，用于自定义的指纹分析
	h.StoreResponseDirectory = utils.GetTempPathDirName()
	defer os.RemoveAll(h.StoreResponseDirectory)
	//调用httpx识别指纹
	h.Do()
}

// fingerPrintFuncForFingerprintHub 回调函数，用于处理自己的指纹识别
func (h *HttpxFinger) fingerPrintFuncForFingerprintHub(domain string, ip string, port int, url string, result []FingerAttrResult) (fingers []string) {
	// 读取httpx保存的response内容，并解析为body和headers
	body, _, headers := h.parseHttpHeaderAndBody(h.getStoredResponseContent(url))
	// 指纹匹配
	for _, v := range h.fpWebFingerprintHub {
		flag := false

		hflag := true
		if len(v.Headers) > 0 {
			hflag = false
			for k, hh := range v.Headers {
				if len(headers[strings.ToLower(k)]) == 0 {
					hflag = false
					break
				}
				if len(headers[strings.ToLower(k)]) > 0 {
					if !strings.Contains(headers[strings.ToLower(k)][0], hh) {
						hflag = false
						break
					}
					hflag = true
				}
			}
		}
		if len(v.Headers) > 0 && hflag {
			flag = true
		}

		kflag := true
		if len(v.Keyword) > 0 {
			kflag = false
			for _, k := range v.Keyword {
				if !strings.Contains(body, k) {
					kflag = false
					break
				}
				kflag = true
			}
		}
		if len(v.Keyword) > 0 && kflag {
			flag = true
		}
		//是否需要匹配多个指纹？
		if flag {
			fingers = append(fingers, v.Name)
			//break
		}
	}
	//fmt.Println(fingers)
	return
}

// fingerPrintFuncForIceMoon 回调函数，用于处理自己的指纹识别
func (h *HttpxFinger) fingerPrintFuncForCustom(domain string, ip string, port int, url string, result []FingerAttrResult) (fingers []string) {
	body, header, _ := h.parseHttpHeaderAndBody(h.getStoredResponseContent(url))

	content := xraypocv1.Content{
		Port:   fmt.Sprintf("%d", port),
		Body:   body,
		Header: header,
	}
	for _, fa := range result {
		if fa.Tag == "title" {
			content.Title = fa.Content
		} else if fa.Tag == "server" {
			content.Server = fa.Content
		} else if fa.Tag == "tlsdata" {
			content.Cert = fa.Content
		}
	}
	//fmt.Println(content)
	for _, v := range h.fpCustom {
		rule := xraypocv1.ParseRules(v.Rule)
		if xraypocv1.MatchRules(*rule, content) {
			//fmt.Println(v)
			fingers = append(fingers, v.App)
		}
	}
	return
}

// fingerPrintFuncForSaveHttpHeaderAndBody 回调函数，用于保存http协议的raw信息
func (h *HttpxFinger) fingerPrintFuncForSaveHttpHeaderAndBody(domain string, ip string, port int, url string, result []FingerAttrResult) (fingers []string) {
	body, header, _ := h.parseHttpHeaderAndBody(h.getStoredResponseContent(url))
	if len(ip) > 0 && port > 0 {
		if len(header) > 0 {
			h.ResultPortScan.SetPortHttpInfo(ip, port, portscan.HttpResult{
				Source:  "httpx",
				Tag:     "header",
				Content: header,
			})
		}
		if len(body) > 0 {
			h.ResultPortScan.SetPortHttpInfo(ip, port, portscan.HttpResult{
				Source:  "httpx",
				Tag:     "body",
				Content: body,
			})
		}
	}
	if len(domain) > 0 && port > 0 {
		if len(header) > 0 {
			h.ResultDomainScan.SetHttpInfo(domain, domainscan.HttpResult{
				Source:  "httpx",
				Port:    port,
				Tag:     "header",
				Content: header,
			})
		}
		if len(body) > 0 {
			h.ResultDomainScan.SetHttpInfo(domain, domainscan.HttpResult{
				Source:  "httpx",
				Port:    port,
				Tag:     "body",
				Content: body,
			})
		}
	}
	return
}

// getStoredResponseContent 读取httpx保存的response内容
func (h *HttpxFinger) getStoredResponseContent(url string) string {
	domainFile := strings.ReplaceAll(urlutil.TrimScheme(url), ":", ".")

	// On various OS the file max file name length is 255 - https://serverfault.com/questions/9546/filename-length-limits-on-linux
	// Truncating length at 255
	if len(domainFile) >= maxFileNameLength {
		// leaving last 4 bytes free to append ".txt"
		domainFile = domainFile[:maxFileNameLength]
	}

	domainFile = strings.ReplaceAll(domainFile, "/", "[slash]") + ".txt"
	// store response
	responsePath := path.Join(h.StoreResponseDirectory, domainFile)
	content, err := os.ReadFile(responsePath)
	if err != nil || len(content) == 0 {
		return ""
	}

	return string(content)
}

// parseHttpHeaderAndBody 分离、解析http的header与body
func (h *HttpxFinger) parseHttpHeaderAndBody(content string) (body string, header string, headerMap map[string][]string) {
	headerMap = make(map[string][]string)
	if len(content) <= 0 {
		return
	}
	headerAndBodyArrays := strings.Split(content, "\r\n\r\n")
	if len(headerAndBodyArrays) >= 1 {
		header = headerAndBodyArrays[0]
		respHeaderSlice := strings.Split(header, "\r\n")
		for _, hh := range respHeaderSlice {
			hslice := strings.SplitN(hh, ":", 2)
			if len(hslice) != 2 {
				continue
			}
			k := strings.ToLower(hslice[0])
			v := strings.TrimLeft(hslice[1], " ")
			if len(headerMap[k]) > 0 {
				headerMap[k] = append(headerMap[k], v)
			} else {
				headerMap[k] = []string{v}
			}
		}
	}
	if len(headerAndBodyArrays) >= 2 {
		body = strings.Join(headerAndBodyArrays[1:], "\r\n\r\n")
	}
	return
}
