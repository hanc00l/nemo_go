package fingerprint

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/chainreactors/fingers"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/custom"
	"github.com/hanc00l/nemo_go/pkg/task/domainscan"
	"github.com/hanc00l/nemo_go/pkg/task/portscan"
	"github.com/hanc00l/nemo_go/pkg/utils"
	"github.com/remeh/sizedwaitgroup"
	"image"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

var HttpxOutputDirectory string //全局的httpx保存响应的数据，用于自定义指纹匹配

// 全局变量，只加载一次
var fpMutex sync.Mutex

// 被动指纹引擎
var chainreactorsEngine *fingers.Engine

// 是否保存http协议的header与body
var isSaveHTTPHeaderAndBody = true

type HttpxAll struct {
	IsScreenshot bool
	IsIconHash   bool

	ResultPortScan     *portscan.Result
	ResultDomainScan   *domainscan.Result
	DomainTargetPort   map[string]map[int]struct{}
	ResultScreenShot   *ScreenshotResult
	ResultIconHashInfo *IconHashInfoResult
	FingerPrintFunc    []func(domain string, ip string, port int, url string, result []FingerAttrResult, storedResponsePathFile string) []string
	//保存响应的数据，用于自定义指纹匹配
	httpxOutputDir string

	IsProxy bool
}

type HttpxResult struct {
	A                  []string `json:"a,omitempty"`
	CNames             []string `json:"cnames,omitempty"`
	Scheme             string   `json:"scheme,omitempty"`
	Url                string   `json:"url,omitempty"`
	Host               string   `json:"host,omitempty"`
	Port               string   `json:"port,omitempty"`
	Title              string   `json:"title,omitempty"`
	WebServer          string   `json:"webserver,omitempty"`
	ContentType        string   `json:"content_type,omitempty"`
	StatusCode         int      `json:"status_code,omitempty"`
	TLSData            *TLS     `json:"tls,omitempty"`
	Jarm               string   `json:"jarm,omitempty"`
	StoredResponsePath string   `json:"stored_response_path,omitempty"`
	IconHash           string   `json:"favicon,omitempty"`
	FaviconPath        string   `json:"favicon_path,omitempty"`
	ScreenShotPath     string   `json:"screenshot_path,omitempty"`
	Tech               []string `json:"tech,omitempty"`
}

type TLS struct {
	SubjectDNSName           []string `json:"subject_an,omitempty"`
	SubjectCommonName        string   `json:"subject_cn,omitempty"`
	SubjectDistinguishedName string   `json:"subject_dn,omitempty"`
	SubjectOrganization      []string `json:"subject_org,omitempty"`
	IssuerDistinguishedName  string   `json:"issuer_dn,omitempty"`
	IssuerOrganization       []string `json:"issuer_org,omitempty"`
}

// NewHttpxAll 创建对象
func NewHttpxAll() *HttpxAll {
	h := &HttpxAll{
		IsScreenshot:   true,
		IsIconHash:     true,
		httpxOutputDir: HttpxOutputDirectory,
	}
	//加载自定义指纹及回调函数
	fpMutex.Lock()
	defer fpMutex.Unlock()
	h.loadChainReactorsFingers()

	return h
}

func (x *HttpxAll) Do() {
	// 如果没有全局的output目录，使用临时目录（在直接进行调试时）
	if x.httpxOutputDir == "" {
		x.httpxOutputDir = utils.GetTempPathDirName()
		logging.CLILog.Infof("httpx output tempdir is %s", x.httpxOutputDir)
		defer os.RemoveAll(x.httpxOutputDir)
	}
	x.ResultScreenShot = &ScreenshotResult{Result: make(map[string][]ScreenshotInfo)}
	x.ResultIconHashInfo = &IconHashInfoResult{}
	//
	swg := sizedwaitgroup.New(fpHttpxThreadNumber[conf.WorkerPerformanceMode])
	btc := custom.NewBlackTargetCheck(custom.CheckAll)
	if x.ResultPortScan != nil && x.ResultPortScan.IPResult != nil {
		for ipName, ipResult := range x.ResultPortScan.IPResult {
			if btc.CheckBlack(ipName) {
				logging.RuntimeLog.Warningf("%s is in blacklist,skip...", ipName)
				continue
			}
			for portNumber, _ := range ipResult.Ports {
				if _, ok := blankPort[portNumber]; ok {
					continue
				}
				url := utils.FormatHostUrl("", ipName, portNumber) //fmt.Sprintf("%v:%v", ipName, portNumber)
				swg.Add()
				go func(ip string, port int, u string) {
					defer swg.Done()
					fingerPrintResult, urlResponse, storedResponsePathFile, storedFaviconPathFile, storedScreenshotPathFile := x.RunHttpx(u)
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
						if isSaveHTTPHeaderAndBody {
							x.saveHttpHeaderAndBody("", ip, port, storedResponsePathFile)
						}
						//处理自定义的finger
						if len(storedResponsePathFile) > 0 && len(x.FingerPrintFunc) > 0 {
							x.doFinger("", ip, port, urlResponse, fingerPrintResult, storedResponsePathFile)
						}
						//screenshot
						if x.IsScreenshot && len(storedScreenshotPathFile) > 0 {
							x.doScreenshot(ip, port, urlResponse, storedScreenshotPathFile)
						}
						//iconhash
						if x.IsIconHash && len(storedFaviconPathFile) > 0 {
							iconHashResult := x.doFavicon(urlResponse, storedFaviconPathFile)
							if iconHashResult != nil {
								par := portscan.PortAttrResult{
									Source:  "iconhash",
									Tag:     "favicon",
									Content: fmt.Sprintf("%s | %s", iconHashResult.Hash, iconHashResult.Url),
								}
								x.ResultPortScan.SetPortAttr(ip, port, par)
							}
						}
					}
				}(ipName, portNumber, url)
			}
		}
	}
	if x.ResultDomainScan != nil && x.ResultDomainScan.DomainResult != nil {
		if x.DomainTargetPort == nil {
			x.DomainTargetPort = make(map[string]map[int]struct{})
		}
		for domain := range x.ResultDomainScan.DomainResult {
			if btc.CheckBlack(domain) {
				logging.RuntimeLog.Warningf("%s is in blacklist,skip...", domain)
				continue
			}
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
				url := utils.FormatHostUrl("", domain, port) //fmt.Sprintf("%s:%d", domain, port)
				swg.Add()
				go func(d string, p int, u string) {
					defer swg.Done()
					fingerPrintResult, urlResponse, storedResponsePathFile, storedFaviconPathFile, storedScreenshotPathFile := x.RunHttpx(u)
					if len(fingerPrintResult) > 0 {
						for _, fpa := range fingerPrintResult {
							dar := domainscan.DomainAttrResult{
								Source:  "httpx",
								Tag:     fpa.Tag,
								Content: fpa.Content,
							}
							x.ResultDomainScan.SetDomainAttr(d, dar)
						}
						if isSaveHTTPHeaderAndBody {
							x.saveHttpHeaderAndBody(domain, "", port, storedResponsePathFile)
						}
						//处理自定义的finger
						if len(storedResponsePathFile) > 0 && len(x.FingerPrintFunc) > 0 {
							x.doFinger(domain, "", port, urlResponse, fingerPrintResult, storedResponsePathFile)
						}
						//screenshot
						if x.IsScreenshot && len(storedScreenshotPathFile) > 0 {
							x.doScreenshot(domain, port, urlResponse, storedScreenshotPathFile)
						}
						//iconhash
						if x.IsIconHash && len(storedFaviconPathFile) > 0 {
							iconHashResult := x.doFavicon(urlResponse, storedFaviconPathFile)
							if iconHashResult != nil {
								dar := domainscan.DomainAttrResult{
									Source:  "iconhash",
									Tag:     "favicon",
									Content: fmt.Sprintf("%s | %s", iconHashResult.Hash, iconHashResult.Url),
								}
								x.ResultDomainScan.SetDomainAttr(d, dar)
							}
						}
					}
				}(domain, port, url)
			}
		}
	}
	swg.Wait()
	// 过滤任务的结果，主要是对标题字段过滤
	if x.ResultPortScan != nil && x.ResultPortScan.IPResult != nil {
		portscan.FilterIPResult(x.ResultPortScan, false)
	}
	if x.ResultDomainScan != nil && x.ResultDomainScan.DomainResult != nil {
		domainscan.FilterDomainResult(x.ResultDomainScan)
	}
}

// RunHttpx 调用httpx，获取一个domain的标题指纹
func (x *HttpxAll) RunHttpx(domain string) (result []FingerAttrResult, urlResponse, storedResponsePathFile, storedFaviconPathFile, storedScreenshotPathFile string) {
	resultTempFile := utils.GetTempPathFileName()
	defer os.Remove(resultTempFile)
	inputTempFile := utils.GetTempPathFileName()
	defer os.Remove(inputTempFile)
	err := os.WriteFile(inputTempFile, []byte(domain), 0666)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		logging.CLILog.Error(err)
		return
	}
	var cmdArgs []string
	cmdArgs = append(cmdArgs,
		"-random-agent", "-l", inputTempFile, "-o", resultTempFile,
		"-retries", "0", "-threads", "50", "-timeout", "5", "-disable-update-check",
		"-title", "-server", "-status-code", "-content-type", "-follow-redirects", "-json", "-silent", "-no-color", "-tls-grab", "-jarm",
		"-favicon", "-screenshot", "--system-chrome", "-esb", "-ehb",
		// -esb, -exclude-screenshot-bytes  enable excluding screenshot bytes from json output
		// -ehb, -exclude-headless-body     enable excluding headless header from json output
		"-store-response", "-store-response-dir", x.httpxOutputDir,
	)
	// 由于chrome不支持带认证的socks5代理，因此httpx及chrome使用本地的socks5转发
	if x.IsProxy {
		if proxy := conf.GetProxyConfig(); proxy != "" {
			if conf.Socks5ForwardAddr != "" {
				cmdArgs = append(cmdArgs, "-http-proxy", fmt.Sprintf("socks5://%s", conf.Socks5ForwardAddr))
			}
		} else {
			logging.RuntimeLog.Warning("get proxy config fail or disabled by worker,skip proxy!")
			logging.CLILog.Warning("get proxy config fail or disabled by worker,skip proxy!")
		}
	}
	binPath := filepath.Join(conf.GetRootPath(), "thirdparty/httpx", utils.GetThirdpartyBinNameByPlatform(utils.Httpx))
	cmd := exec.Command(binPath, cmdArgs...)
	var stderr bytes.Buffer
	cmd.Stdout = os.Stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		logging.RuntimeLog.Error(err, stderr)
		logging.CLILog.Error(err, stderr)
		return
	}
	result, urlResponse, storedResponsePathFile, storedFaviconPathFile, storedScreenshotPathFile = x.parseHttpxResult(resultTempFile)
	return
}

// ParseHttpxJson 解析一条httpx的JSON记录
func (x *HttpxAll) ParseHttpxJson(content []byte) (host string, port int, result []FingerAttrResult, urlResponse, storedResponsePathFile, storedFaviconPathFile, storedScreenshotPathFile string) {
	resultJSON := HttpxResult{}
	err := json.Unmarshal(content, &resultJSON)
	if err != nil {
		logging.RuntimeLog.Error(err)
		logging.CLILog.Error(err)
		return
	}
	// 获取host与port
	host = resultJSON.Host
	port, err = strconv.Atoi(resultJSON.Port)
	if err != nil {
		logging.RuntimeLog.Error(err)
		logging.CLILog.Error(err)
		return
	}
	urlResponse = resultJSON.Url
	// 保存stored_response_path等供fingerprint功能及其它使用，但返回的JSON字符串不需要
	storedScreenshotPathFile = resultJSON.ScreenShotPath
	storedResponsePathFile = resultJSON.StoredResponsePath
	// 只有httpx的JSON结果有favicon的hash，才表示有favicon图像存在
	if resultJSON.IconHash != "" {
		storedFaviconPathFile = resultJSON.FaviconPath
		resultJSON.FaviconPath = ""
	}
	resultJSON.StoredResponsePath = ""
	resultJSON.ScreenShotPath = ""
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
	if resultJSON.Scheme != "" {
		result = append(result, FingerAttrResult{
			Tag:     "service",
			Content: resultJSON.Scheme,
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
func (x *HttpxAll) parseHttpxResult(outputTempFile string) (result []FingerAttrResult, urlResponse, storedResponsePathFile, storedFaviconPathFile, storedScreenshotPathFile string) {
	content, err := os.ReadFile(outputTempFile)
	if err != nil || len(content) == 0 {
		if err != nil {
			logging.RuntimeLog.Error(err)
			logging.CLILog.Error(err)
		}
		return
	}
	// host与port这里不需要
	_, _, result, urlResponse, storedResponsePathFile, storedFaviconPathFile, storedScreenshotPathFile = x.ParseHttpxJson(content)

	return
}

// ParseContentResult 解析httpx扫描的JSON格式文件结果
func (x *HttpxAll) ParseContentResult(content []byte) (result portscan.Result) {
	result.IPResult = make(map[string]*portscan.IPResult)
	s := custom.NewService()
	scanner := bufio.NewScanner(bytes.NewReader(content))
	for scanner.Scan() {
		data := scanner.Bytes()
		host, port, fas, _, _, _, _ := x.ParseHttpxJson(data)
		if host == "" || port == 0 || len(fas) == 0 || !utils.CheckIP(host) {
			continue
		}
		if !result.HasIP(host) {
			result.SetIP(host)
		}
		if !result.HasPort(host, port) {
			result.SetPort(host, port)
		}
		service := s.FindService(port, "")
		result.SetPortAttr(host, port, portscan.PortAttrResult{
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
			result.SetPortAttr(host, port, par)
			if fa.Tag == "status" {
				result.IPResult[host].Ports[port].Status = fa.Content
			}
		}
	}
	return
}

// loadChainReactorsFingers 加载ChainReactorsFingers指纹
func (x *HttpxAll) loadChainReactorsFingers() {
	if chainreactorsEngine != nil {
		return
	}
	var err error
	chainreactorsEngine, err = fingers.NewEngine(fingers.FingersEngine)
	if err != nil {
		logging.RuntimeLog.Errorf("load chainreactors fingers err:%v", err)
		return
	}
	x.FingerPrintFunc = append(x.FingerPrintFunc, x.fingerPrintFuncForChainReactorsFingers)
	logging.CLILog.Info("Load chainreactors fingers engine ok")
}

// doFinger 执行指纹识别
func (x *HttpxAll) doFinger(domain, ip string, port int, url string, fingerPrintResult []FingerAttrResult, storedResponsePathFile string) {
	if ip != "" {
		for _, f := range x.FingerPrintFunc {
			xfars := f("", ip, port, url, fingerPrintResult, storedResponsePathFile)
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
	if domain != "" {
		for _, f := range x.FingerPrintFunc {
			xfars := f(domain, "", port, url, fingerPrintResult, storedResponsePathFile)
			if len(xfars) > 0 {
				for _, fps := range xfars {
					dar := domainscan.DomainAttrResult{
						Source:  "httpxfinger",
						Tag:     "fingerprint",
						Content: fps,
					}
					x.ResultDomainScan.SetDomainAttr(domain, dar)
				}
			}
		}
	}
}

func (x *HttpxAll) doScreenshot(domain string, port int, u string, storedScreenshotPathFile string) {
	var protocol string
	uu, err := url.Parse(u)
	if err != nil {
		protocol = "http"
	} else {
		protocol = uu.Scheme
	}
	fileResized := utils.GetTempPNGPathFileName()
	if utils.ReSizeAndCropPicture(storedScreenshotPathFile, fileResized, SavedWidth, SavedHeight) {
		si := ScreenshotInfo{
			Port:         port,
			Protocol:     protocol,
			FilePathName: fileResized,
		}
		x.ResultScreenShot.SetScreenshotInfo(domain, si)
	}
}

func (x *HttpxAll) doFavicon(u string, storedFaviconPathFile string) (result *IconHashInfo) {
	urlFavicon := fmt.Sprintf("%s%s", u, storedFaviconPathFile)
	logging.CLILog.Info(urlFavicon)
	//获取icon
	request, err := http.NewRequest(http.MethodGet, urlFavicon, nil)
	if err != nil {
		return
	}
	resp, err := utils.GetProxyHttpClient(x.IsProxy).Do(request)
	if err != nil {
		return
	}
	content, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil || len(content) == 0 {
		return
	}
	// 检查是否是图像格式
	if isSVG(content) {
		// Special handling for svg, which golang can't decode with
		// image.DecodeConfig. Fill in an absurdly large width/height so SVG always
		//// wins size contests.
		//i.Format = "svg"
		//i.Width = 9999
		//i.Height = 9999
	} else {
		_, _, e := image.DecodeConfig(bytes.NewReader(content))
		if e != nil {
			logging.CLILog.Errorf("unknown image format: %s", e)
			return
		}
		// jpeg => jpg
		//if format == "jpeg" {
		//	format = "jpg"
		//}
		//i.Width = cfg.Width
		//i.Height = cfg.Height
		//i.Format = format
	}
	//计算哈希值
	hash := mmh3Hash32(standBase64(content))
	result = &IconHashInfo{
		Url:       urlFavicon,
		Hash:      hash,
		ImageData: content,
	}
	x.ResultIconHashInfo.Lock()
	x.ResultIconHashInfo.Result = append(x.ResultIconHashInfo.Result, *result)
	x.ResultIconHashInfo.Unlock()

	return
}

// fingerPrintFuncForChainReactorsFingers 回调函数，用于处理ChainReactorsFingers指纹识别
func (x *HttpxAll) fingerPrintFuncForChainReactorsFingers(domain string, ip string, port int, url string, result []FingerAttrResult, storedResponsePathFile string) (fingers []string) {
	if chainreactorsEngine == nil {
		return
	}
	body, header := x.parseHttpHeaderAndBody(x.getStoredResponseContent(storedResponsePathFile))
	c := strings.Join([]string{header, body}, "\r\n\r\n")
	frames, err := chainreactorsEngine.DetectContent([]byte(c))
	if err != nil {
		logging.RuntimeLog.Errorf("run chainreactors fingers err:%v", err)
		return
	}
	for _, f := range frames.List() {
		fingers = append(fingers, f.String())
	}
	return
}

// saveHttpHeaderAndBody 保存http协议的raw信息
func (x *HttpxAll) saveHttpHeaderAndBody(domain string, ip string, port int, storedResponsePathFile string) {
	body, header := x.parseHttpHeaderAndBody(x.getStoredResponseContent(storedResponsePathFile))
	if len(ip) > 0 && port > 0 {
		if len(header) > 0 {
			x.ResultPortScan.SetPortHttpInfo(ip, port, portscan.HttpResult{
				Source:  "httpx",
				Tag:     "header",
				Content: header,
			})
		}
		if len(body) > 0 {
			x.ResultPortScan.SetPortHttpInfo(ip, port, portscan.HttpResult{
				Source:  "httpx",
				Tag:     "body",
				Content: body,
			})
		}
	}
	if len(domain) > 0 && port > 0 {
		if len(header) > 0 {
			x.ResultDomainScan.SetHttpInfo(domain, domainscan.HttpResult{
				Source:  "httpx",
				Port:    port,
				Tag:     "header",
				Content: header,
			})
		}
		if len(body) > 0 {
			x.ResultDomainScan.SetHttpInfo(domain, domainscan.HttpResult{
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
func (x *HttpxAll) getStoredResponseContent(storedResponsePathFile string) string {
	content, err := os.ReadFile(storedResponsePathFile)
	if err != nil || len(content) == 0 {
		if err != nil {
			logging.RuntimeLog.Error(err)
			logging.CLILog.Error(err)
		}
		return ""
	}
	return string(content)
}

// parseHttpHeaderAndBody 分离、解析http的header与body
func (x *HttpxAll) parseHttpHeaderAndBody(content string) (body string, header string) {
	if len(content) <= 0 {
		return
	}
	/* httpx保存的文件格式为(v1.2.9）：
	GET / HTTP/1.1
	Request ->header

	Response->header

	Response->body



	URL
	*/
	headerAndBodyArrays := strings.Split(content, "\r\n\r\n")
	if len(headerAndBodyArrays) >= 2 {
		header = headerAndBodyArrays[1]
	}
	if len(headerAndBodyArrays) >= 3 {
		body = strings.Join(headerAndBodyArrays[2:], "\r\n\r\n")
	}
	return
}
