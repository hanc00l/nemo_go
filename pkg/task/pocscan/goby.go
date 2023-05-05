package pocscan

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"io"
	"net/http"
	"strings"
	"time"
)

type Goby struct {
	Config Config
	Result []Result
	notice chan int
}

// GobyStartScanRequest 扫描任务请求参数
type GobyStartScanRequest struct {
	Asset struct {
		Ips      []string `json:"ips"`
		Ports    string   `json:"ports"`
		BlackIps []string `json:"blackIps"`
	} `json:"asset"`
	Vulnerability struct {
		//-vulmode int
		//vulmode: 	0 general(no bruteforce, but include webvulscan),
		//			1 only_burte_pocs,
		//			2 all_pocs,
		//			3 only_list_pocs,
		//			4 only_webvulscan,
		//			5 only_appvul,
		//			-1 disable
		Type string `json:"type"`
	} `json:"vulnerability"`
	Options struct {
		Random bool `json:"random,omitempty"`
		Rate   int  `json:"rate,omitempty"`
		//-hostListMode
		//    	whether enable hostListMode, which targets is hostinfo list, aka Fast mode
		HostListMode bool `json:"hostListMode,omitempty"`
		//-portscanmode intport scan mode: 0 is pcap, 1 is tcpudp
		Portscanmode  int  `json:"portscanmode,omitempty"`
		CheckHoneyPot bool `json:"CheckHoneyPot,omitempty"`
		PingFirst     bool `json:"pingFirst,omitempty"`
		//-pingCheckSize int
		//    	ping check size, only valid when set pingfirst (default 10)
		//  -pingConcurrent int
		//    	ping concurrent, only valid when set pingfirst (default 2)
		//  -pingSendCount int
		//    	ping senbd count, only valid when set pingfirst (default 2)
		PingCheckSize    int    `json:"pingCheckSize,omitempty"`
		PingConcurrent   int    `json:"pingConcurren,omitempty"`
		PingSendCount    int    `json:"pingSendCount,omitempty"`
		DefaultUserAgent string `json:"defaultUserAgent,omitempty"`
		SocketTimeout    int    `json:"socketTimeout,omitempty"`
		RetryTimes       int    `json:"retryTimes,omitempty"`
		CheckAliveMode   int    `json:"checkAliveMode,omitempty"`
	} `json:"options"`
}

// GobyStartScanResponse 扫描任务启动的返回结果
type GobyStartScanResponse struct {
	StatusCode int    `json:"statusCode"`
	Messages   string `json:"messages"`
	Data       struct {
		TaskId string `json:"taskId"`
	} `json:"data"`
}

// GobyProgessRequest 扫描进度请求参数
type GobyProgessRequest struct {
	Taskid string `json:"taskid"`
}

// GobyProgressResponse 扫描进度返回结果
type GobyProgressResponse struct {
	StatusCode int    `json:"statusCode"`
	Messages   string `json:"messages"`
	Data       struct {
		Logs     interface{} `json:"logs"`
		Progress int         `json:"progress"`
		State    int         `json:"state"`
	} `json:"data"`
}

// GobyAssetSearchRequest 获取扫描端口信息请求参数
type GobyAssetSearchRequest struct {
	Query   string `json:"query"`
	Options struct {
		Page struct {
			Page int `json:"page"`
			Size int `json:"size"`
		} `json:"page"`
	} `json:"options"`
}

// GobyAssetSearchResponse 扫描端口信息返回
type GobyAssetSearchResponse struct {
	StatusCode int    `json:"statusCode"`
	Messages   string `json:"messages"`
	Data       struct {
		Ips []struct {
			Ip       string `json:"ip"`
			Mac      string `json:"mac"`
			Os       string `json:"os"`
			Hostname string `json:"hostname"`
			Honeypot string `json:"honeypot,omitempty"`
			Ports    []struct {
				Port         string `json:"port"`
				Baseprotocol string `json:"baseprotocol"`
			} `json:"ports"`
			Protocols       map[string]AssertPortInfo `json:"protocols"`
			Vulnerabilities []struct {
				Hostinfo string `json:"hostinfo"`
				Name     string `json:"name"`
				Filename string `json:"filename"`
				Level    string `json:"level"`
				Vulurl   string `json:"vulurl"`
				Keymemo  string `json:"keymemo"`
				Hasexp   bool   `json:"hasexp "`
			} `json:"vulnerabilities"`
			Screenshots interface{} `json:"screenshots"`
			Favicons    interface{} `json:"favicons"`
			Hostnames   []string    `json:"hostnames"`
		} `json:"ips"`
	} `json:"data"`
}

// AssertPortInfo 端口信息
type AssertPortInfo struct {
	Port      string   `json:"port"`
	Hostinfo  string   `json:"hostinfo"`
	Url       string   `json:"url"`
	Product   string   `json:"product"`
	Protocol  string   `json:"protocol"`
	Json      string   `json:"json"`
	Fid       []string `json:"fid"`
	Products  []string `json:"products"`
	Protocols []string `json:"protocols"`
}

// GobyVulnerabilityRequest 扫描结果中的漏洞信息请求参数
type GobyVulnerabilityRequest struct {
	TaskId  string `json:"taskId"`
	Type    string `json:"type"`
	Query   string `json:"query"`
	Options struct {
		Page struct {
			Page int `json:"page"`
			Size int `json:"size"`
		} `json:"page"`
	} `json:"options"`
}

// GobyVulnerabilityResponse 扫描结果中的漏洞结果
type GobyVulnerabilityResponse struct {
	StatusCode int    `json:"statusCode"`
	Messages   string `json:"messages"`
	Data       struct {
		Total struct {
			Ips             int `json:"ips"`
			Vulnerabilities int `json:"vulnerabilities"`
		} `json:"total"`
		Lists []struct {
			Name  string `json:"name"`
			Nums  int    `json:"nums"`
			Lists []struct {
				Hostinfo string `json:"hostinfo"`
				Name     string `json:"name"`
				Filename string `json:"filename"`
				Level    string `json:"level,omitempty"`
				Vulurl   string `json:"vulurl,omitempty"`
				Keymemo  string `json:"keymemo,omitempty"`
				Hasexp   bool   `json:"hasexp,omitempty"`
			} `json:"lists"`
		} `json:"lists"`
	} `json:"data"`
}

const (
	APITaskList         = "/api/v1/tasks"
	APIStartScan        = "/api/v1/startScan"
	APIGetAsset         = "/api/v1/assetSearch"
	APIGetVulnerability = "/api/v1/vulnerabilitySearch"
	APIProgress         = "/api/v1/getProgress"
	// SleepDelayTimeSecond 任务执行出错时的休眠间隔
	SleepDelayTimeSecond = 10
	// CheckProgressTimeSecond 检查任务执行结果的时间
	CheckProgressTimeSecond = 10
)

// NewGoby 创建goby对象
func NewGoby(config Config) *Goby {
	g := Goby{Config: config}
	return &g
}

// Do 调用goby执行一次scan
func (g *Goby) Do() {
	/* 使用极速模式扫描，config.Target格式为ip:port,ip:port；传递到goby的参数格式为["ip:port","ip:port"]
	-hostListMode
	    	whether enable hostListMode, which targets is hostinfo list, aka Fast mode
	*/
	if len(conf.GlobalWorkerConfig().Pocscan.Goby.API) <= 0 {
		logging.CLILog.Error("no goby api set")
		logging.RuntimeLog.Error("no goby api set")
		return
	}
	ips := strings.Split(g.Config.Target, ",")
	taskId, api, err := g.StartScan(ips)
	if err != nil {
		logging.CLILog.Error(err)
		logging.RuntimeLog.Error(err)
		return
	}
	err = g.GetVulnerability(api, taskId)
	if err != nil {
		logging.CLILog.Error(err)
		logging.RuntimeLog.Error(err)
		return
	}
}

// GetTaskList 获取任务列表
func (g *Goby) GetTaskList(api string) (err error) {
	var respBody []byte
	respBody, err = g.postData("GET", fmt.Sprintf("%s%s", api, APITaskList), nil)
	if err != nil {
		logging.CLILog.Error(err)
		logging.RuntimeLog.Error(err)
		return
	}
	fmt.Println(string(respBody))
	if string(respBody) == "Not authorized" {
		err = errors.New("not authorized")
	}
	return
}

// StartScan 执行一次扫描，并等待扫描结束后返回
func (g *Goby) StartScan(ips []string) (taskId string, api string, err error) {
	reqBody := GobyStartScanRequest{}
	reqBody.Asset.Ips = ips
	reqBody.Vulnerability.Type = "2"
	reqBody.Options.Rate = 1000
	reqBody.Options.Random = true
	reqBody.Options.HostListMode = true
	reqBody.Options.DefaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.125 Safari/537.36"
	dataBytes, _ := json.Marshal(reqBody)
	var respBody []byte
	for {
		//从多个api接口中，选择一个可用的接口执行，如果接口不可用或失败则回一直等待
		for _, apiPath := range conf.GlobalWorkerConfig().Pocscan.Goby.API {
			// 当前使用的api
			api = apiPath
			respBody, err = g.postData("POST", fmt.Sprintf("%s%s", api, APIStartScan), dataBytes)
			// 连接失败、验证失败
			if err != nil {
				logging.CLILog.Error(err)
				logging.RuntimeLog.Error(err)
				//延迟后选择下一个接口执行
				time.Sleep(SleepDelayTimeSecond * time.Second)
				continue
			}
			// 读取json内容
			var result GobyStartScanResponse
			err = json.Unmarshal(respBody, &result)
			if err != nil {
				logging.CLILog.Error(err)
				logging.RuntimeLog.Error(err)
				time.Sleep(SleepDelayTimeSecond * time.Second)
				continue
			}
			// goby正在执行扫描任务，当前接口不可用
			//{"statusCode":500,"messages":"task launch failed, instance already running","data":null}
			if result.StatusCode == 500 {
				logging.CLILog.Infof("Goby api:%s is busy", api)
				err = errors.New(result.Messages)
				time.Sleep(3 * SleepDelayTimeSecond * time.Second)
				continue
			}
			// 任务成功执行，等待执行完成
			taskId = result.Data.TaskId
			logging.CLILog.Infof("Goby scan task:%s,ips:%s started", taskId, strings.Join(ips, ","))
			g.notice = make(chan int, 0)
			go g.tickListen(api, taskId)
			for {
				select {
				case <-g.notice:
					return
				}
			}
		}
	}
	return
}

// GetAsset 获取扫描结果
// 由于存在递归调用的情况，只能放弃goby的端口扫描信息
func (g *Goby) GetAsset(api string, taskId string) (err error) {
	req := GobyAssetSearchRequest{}
	req.Query = fmt.Sprintf("taskId=%s", taskId)
	req.Options.Page.Page = 1
	req.Options.Page.Size = 100000
	dataBytes, _ := json.Marshal(req)
	var respBody []byte
	respBody, err = g.postData("POST", fmt.Sprintf("%s%s", api, APIGetAsset), dataBytes)
	if err != nil {
		logging.CLILog.Error(err)
		logging.RuntimeLog.Error(err)
		return
	}
	err = g.parseAssertResult(respBody)
	return
}

// GetVulnerability 获取扫描漏洞结果
func (g *Goby) GetVulnerability(api string, taskId string) (err error) {
	req := GobyVulnerabilityRequest{}
	req.Query = fmt.Sprintf("taskId=%s", taskId)
	req.Type = "vulnerability"
	req.TaskId = taskId
	req.Options.Page.Page = 1
	req.Options.Page.Size = 100000
	dataBytes, _ := json.Marshal(req)
	var respBody []byte
	respBody, err = g.postData("POST", fmt.Sprintf("%s%s", api, APIGetVulnerability), dataBytes)
	if err != nil {
		logging.CLILog.Error(err)
		logging.RuntimeLog.Error(err)
		return
	}
	err = g.parseVulnerabilityResult(respBody)
	return
}

// postData 发送参数到goby-cmd api
func (g *Goby) postData(method string, apiUrl string, data []byte) (body []byte, err error) {
	var req *http.Request
	req, err = http.NewRequest(method, apiUrl, bytes.NewBuffer(data))
	if err != nil {
		return
	}
	apiAuth := base64.URLEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", conf.GlobalWorkerConfig().Pocscan.Goby.AuthUser, conf.GlobalWorkerConfig().Pocscan.Goby.AuthPass)))
	req.Header.Add("Authorization", fmt.Sprintf("Basic %s", apiAuth))
	req.Header.Add("Content-Type", "application/json")
	client := &http.Client{Timeout: 5 * time.Second}
	var resp *http.Response
	resp, err = client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	// 检查是否是验证错误
	if resp.StatusCode == http.StatusUnauthorized || string(body) == "Not authorized" {
		err = errors.New("not authorized")
	}
	return
}

// tickListen 定时器，监测任务状态
func (g *Goby) tickListen(api string, taskId string) {
	var progress int
	timer := time.NewTicker(CheckProgressTimeSecond * time.Second)
	defer timer.Stop()
	for {
		select {
		case <-timer.C:
			pNow, _ := g.checkGobyTaskProgress(api, taskId)
			if pNow-progress >= 10 {
				logging.CLILog.Infof("Goby scan task:%s,progress:%d%% ", taskId, pNow)
				progress = pNow
			}
			if progress >= 100 {
				logging.CLILog.Infof("Goby scan task:%s finish", taskId)
				g.notice <- 1
				return
			}
		}
	}
}

// checkGobyTaskProgress 获取任务进度
func (g *Goby) checkGobyTaskProgress(api string, taskId string) (progress int, err error) {
	req := GobyProgessRequest{Taskid: taskId}
	dataBytes, _ := json.Marshal(req)
	var respBody []byte
	respBody, err = g.postData("POST", fmt.Sprintf("%s%s", api, APIProgress), dataBytes)
	if err != nil {
		logging.CLILog.Error(err)
		logging.RuntimeLog.Error(err)
		return
	}
	var result GobyProgressResponse
	err = json.Unmarshal(respBody, &result)
	if err != nil {
		logging.CLILog.Error(err)
		logging.RuntimeLog.Error(err)
		return
	}
	progress = result.Data.Progress
	return
}

// parseAssertResult 获取资产扫描结果
func (g *Goby) parseAssertResult(content []byte) (err error) {
	var result GobyAssetSearchResponse
	err = json.Unmarshal(content, &result)
	if err != nil {
		return
	}
	for _, ipAsset := range result.Data.Ips {
		fmt.Println(ipAsset.Ip)
		for _, v := range ipAsset.Protocols {
			fmt.Println(v.Port, v.Protocol, v.Product)
		}
	}
	return nil
}

// parseVulnerabilityResult 获取漏洞扫描结果
func (g *Goby) parseVulnerabilityResult(content []byte) (err error) {
	var result GobyVulnerabilityResponse
	err = json.Unmarshal(content, &result)
	if err != nil {
		return
	}
	for _, vul := range result.Data.Lists {
		for _, l := range vul.Lists {
			ipPort := strings.Split(l.Hostinfo, ":")
			if len(ipPort) != 2 {
				continue
			}
			extra, _ := json.Marshal(l)
			g.Result = append(g.Result, Result{
				Target:      ipPort[0],
				Url:         l.Hostinfo,
				PocFile:     vul.Name,
				Source:      "goby",
				Extra:       string(extra),
				WorkspaceId: g.Config.WorkspaceId,
			})
		}
	}
	return nil
}
