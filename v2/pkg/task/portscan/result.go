package portscan

import (
	"encoding/json"
	"fmt"
	"github.com/hanc00l/nemo_go/v2/pkg/conf"
	"github.com/hanc00l/nemo_go/v2/pkg/db"
	"github.com/hanc00l/nemo_go/v2/pkg/logging"
	"github.com/hanc00l/nemo_go/v2/pkg/task/custom"
	"strings"
	"sync"
)

const (
	IpOpenedPortFilterNumber = 50 // IP开放端口数量，超过该数量则认为存在安全设备过滤
)

// Config 端口扫描的参数配置
type Config struct {
	Target           string `json:"target"`
	ExcludeTarget    string `json:"executeTarget"`
	Port             string `json:"port"`
	OrgId            *int   `json:"orgId"`
	Rate             int    `json:"rate"`
	IsPing           bool   `json:"ping"`
	Tech             string `json:"tech"`
	IsIpLocation     bool   `json:"ipLocation"`
	IsHttpx          bool   `json:"httpx"`
	IsScreenshot     bool   `json:"screenshot"`
	IsFingerprintHub bool   `json:"fingerprinthub"`
	IsIconHash       bool   `json:"iconhash"`
	IsFingerprintx   bool   `json:"fingerprintx"`
	CmdBin           string `json:"cmdBin"`
	IsLoadOpenedPort bool   `json:"loadOpenedPort"`
	IsPortscan       bool   `json:"isPortscan"`
	WorkspaceId      int    `json:"workspaceId"`
	IsProxy          bool   `json:"proxy"`
}

// PortAttrResult 端口属性结果
type PortAttrResult struct {
	RelatedId int
	Source    string
	Tag       string
	Content   string
}

type HttpResult struct {
	RelatedId int
	Source    string
	Tag       string
	Content   string
}

// PortResult 端口结果
type PortResult struct {
	Status    string
	PortAttrs []PortAttrResult
	HttpInfo  []HttpResult
}

// IPResult IP结果
type IPResult struct {
	OrgId    *int
	Location string
	Status   string
	Ports    map[int]*PortResult
}

// Result 端口扫描结果
type Result struct {
	sync.RWMutex
	IPResult map[string]*IPResult
}

type OfflineResult interface {
	ParseContentResult(content []byte) (ipResult Result)
}

type ImportOfflineResult struct {
	resultType       string
	offlineInterface OfflineResult
	IpResult         Result
}

func NewImportOfflineResult(resultType string) *ImportOfflineResult {
	i := &ImportOfflineResult{resultType: resultType}
	switch resultType {
	case "nmap":
		i.offlineInterface = new(Nmap)
	case "masscan":
		i.offlineInterface = new(Masscan)
	case "fscan":
		i.offlineInterface = new(FScan)
	case "gogo":
		i.offlineInterface = new(Gogo)
	case "goby":
		i.offlineInterface = new(Goby)
	}
	return i
}

func NewImportOfflineResultWithInterface(resultType string, resultInterface OfflineResult) *ImportOfflineResult {
	i := &ImportOfflineResult{resultType: resultType}
	i.offlineInterface = resultInterface

	return i
}

func (i *ImportOfflineResult) Parse(content []byte) {
	if i.offlineInterface == nil {
		logging.RuntimeLog.Errorf("invalid offline result:%s", i.resultType)
		return
	}
	i.IpResult = i.offlineInterface.ParseContentResult(content)
}

func (r *Result) HasIP(ip string) bool {
	r.RLock()
	defer r.RUnlock()

	_, ok := r.IPResult[ip]
	return ok
}

func (r *Result) SetIP(ip string) {
	r.Lock()
	defer r.Unlock()

	r.IPResult[ip] = &IPResult{Ports: make(map[int]*PortResult)}
}

func (r *Result) HasPort(ip string, port int) bool {
	r.RLock()
	defer r.RUnlock()

	_, ok := r.IPResult[ip].Ports[port]
	return ok
}

func (r *Result) SetPort(ip string, port int) {
	r.Lock()
	defer r.Unlock()

	r.IPResult[ip].Ports[port] = &PortResult{PortAttrs: []PortAttrResult{}}
}

func (r *Result) SetPortAttr(ip string, port int, par PortAttrResult) {
	r.Lock()
	defer r.Unlock()

	r.IPResult[ip].Ports[port].PortAttrs = append(r.IPResult[ip].Ports[port].PortAttrs, par)
}

func (r *Result) SetPortHttpInfo(ip string, port int, result HttpResult) {
	r.Lock()
	defer r.Unlock()

	r.IPResult[ip].Ports[port].HttpInfo = append(r.IPResult[ip].Ports[port].HttpInfo, result)
}

// SaveResult 保存端口扫描的结果到数据库
func (r *Result) SaveResult(config Config) string {
	var resultIPCount, resultPortCount int
	var newIP, newPort int
	blackIP := custom.NewBlackTargetCheck(custom.CheckIP)
	// 用于同步到es的IP
	var ElasticAssets []db.Ip
	for ipName, ipResult := range r.IPResult {
		if blackIP.CheckBlack(ipName) {
			logging.RuntimeLog.Warningf("%s is in blacklist,skip...", ipName)
			continue
		}
		if len(ipResult.Ports) > IpOpenedPortFilterNumber {
			logging.RuntimeLog.Warningf("ip:%s has too much open port:%d,discard to save!", ipName, len(ipResult.Ports))
			continue
		}
		//save ip
		ip := &db.Ip{
			IpName:      ipName,
			OrgId:       config.OrgId,
			Location:    ipResult.Location,
			Status:      ipResult.Status,
			WorkspaceId: config.WorkspaceId,
		}
		if ok, isNew := ip.SaveOrUpdate(); !ok {
			continue
		} else {
			if isNew {
				newIP++
			}
		}
		resultIPCount++
		// elastic assets
		if conf.ElasticSyncAssetsChan != nil {
			ElasticAssets = append(ElasticAssets, *ip)
		}
		for portNumber, portResult := range ipResult.Ports {
			//save port
			port := &db.Port{
				IpId:    ip.Id,
				PortNum: portNumber,
				Status:  portResult.Status,
			}
			if ok, isNew := port.SaveOrUpdate(); !ok {
				continue
			} else {
				if isNew {
					newPort++
				}
			}
			resultPortCount++
			//save port attribute
			for _, portAttrResult := range portResult.PortAttrs {
				portAttr := &db.PortAttr{
					RelatedId: port.Id,
					Source:    portAttrResult.Source,
					Tag:       portAttrResult.Tag,
				}
				if len(portAttrResult.Content) > db.AttrContentSize {
					portAttr.Content = portAttrResult.Content[:db.AttrContentSize]
				} else {
					portAttr.Content = portAttrResult.Content
				}
				portAttr.SaveOrUpdate()
			}
			//save http info
			for _, httpInfoResult := range portResult.HttpInfo {
				httpInfo := &db.IpHttp{
					RelatedId: port.Id,
					Source:    httpInfoResult.Source,
					Tag:       httpInfoResult.Tag,
				}
				if len(httpInfoResult.Content) > db.HttpBodyContentSize {
					httpInfo.Content = httpInfoResult.Content[:db.HttpBodyContentSize]
				} else {
					httpInfo.Content = httpInfoResult.Content
				}
				httpInfo.SaveOrUpdate()
			}
		}
	}
	// 将IP资产同步到Elastic
	if conf.ElasticSyncAssetsChan != nil && len(ElasticAssets) > 0 {
		ElasticAssetsIPByte, _ := json.Marshal(ElasticAssets)
		syncArgs := conf.ElasticSyncAssetsArgs{
			Contents:       ElasticAssetsIPByte,
			SyncOp:         conf.SyncOpNew,
			SyncAssetsType: conf.SyncAssetsTypeIP,
		}
		conf.ElasticSyncAssetsChan <- syncArgs
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("ip:%d", resultIPCount))
	if newIP > 0 {
		sb.WriteString(fmt.Sprintf(",ipNew:%d", newIP))
	}
	sb.WriteString(fmt.Sprintf(",port:%d", resultPortCount))
	if newPort > 0 {
		sb.WriteString(fmt.Sprintf(",portNew:%d", resultPortCount))
	}
	return sb.String()
}

// FilterIPResult 过滤IP结果
func FilterIPResult(result *Result, isOnline bool) {
	MaxNumber := conf.GlobalWorkerConfig().Filter.MaxPortPerIp
	if MaxNumber <= 0 {
		MaxNumber = IpOpenedPortFilterNumber
	}
	if isOnline {
		MaxNumber *= 2
	}
	titleFilter := strings.Split(conf.GlobalWorkerConfig().Filter.Title, "|")
	for ipName, ipResult := range result.IPResult {
		// 过滤开放端口数量
		if len(ipResult.Ports) > MaxNumber {
			logging.RuntimeLog.Warningf("ip:%s has too much open port:%d,discard to save!", ipName, len(ipResult.Ports))
			delete(result.IPResult, ipName)
			continue
		}
		// 根据标题进行过滤
		// strings.split始终返回len()>=1，即使是空字符串
		if titleFilter[0] != "" {
			ipHadFiltered := false
			for _, portResult := range ipResult.Ports {
				if ipHadFiltered {
					break
				}
				for _, portAttrResult := range portResult.PortAttrs {
					if ipHadFiltered {
						break
					}
					if portAttrResult.Tag == "title" {
						for _, title := range titleFilter {
							if titleTrim := strings.TrimSpace(title); titleTrim != "" {
								if strings.Contains(portAttrResult.Content, titleTrim) {
									logging.RuntimeLog.Warningf("ip:%s has filter title:%s,discard to save!", ipName, title)
									delete(result.IPResult, ipName)
									ipHadFiltered = true
									break
								}
							}
						}
					}
				}
			}
		}
	}
}
