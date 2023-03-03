package portscan

import (
	"fmt"
	"github.com/hanc00l/nemo_go/pkg/db"
	"github.com/hanc00l/nemo_go/pkg/logging"
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
	CmdBin           string `json:"cmdBin"`
	IsLoadOpenedPort bool   `json:"loadOpenedPort"`
	IsPortscan       bool   `json:"isPortscan"`
	WorkspaceId      int    `json:"workspaceId"`
}

// PortAttrResult 端口属性结果
type PortAttrResult struct {
	RelatedId int
	Source    string
	Tag       string
	Content   string
}

// PortResult 端口结果
type PortResult struct {
	Status    string
	PortAttrs []PortAttrResult
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

// SaveResult 保存端口扫描的结果到数据库
func (r *Result) SaveResult(config Config) string {
	var resultIPCount, resultPortCount int
	var newIP, newPort int
	for ipName, ipResult := range r.IPResult {
		if len(ipResult.Ports) > IpOpenedPortFilterNumber {
			logging.RuntimeLog.Infof("ip:%s has too much open port:%d,discard to save!", ipName, len(ipResult.Ports))
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
					Content:   portAttrResult.Content,
				}
				portAttr.SaveOrUpdate()
			}
		}
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

// FilterIPHasTooMuchPort 过滤有安全防护、显示太多端口开放的IP
func FilterIPHasTooMuchPort(result *Result, isOnline bool) {
	MaxNumber := IpOpenedPortFilterNumber
	if isOnline {
		MaxNumber *= 2
	}
	for ipName, ipResult := range result.IPResult {
		if len(ipResult.Ports) > MaxNumber {
			logging.RuntimeLog.Infof("ip:%s has too much open port:%d,discard to save!", ipName, len(ipResult.Ports))
			delete(result.IPResult, ipName)
		}
	}
}
