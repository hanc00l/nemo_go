package portscan

import (
	"fmt"
	"github.com/hanc00l/nemo_go/v3/pkg/db"
	"github.com/hanc00l/nemo_go/v3/pkg/logging"
	"github.com/hanc00l/nemo_go/v3/pkg/task/execute"
	"sync"
)

const (
	IpOpenedPortFilterNumber = 50 // IP开放端口数量，超过该数量则认为存在安全设备过滤
)

// PortAttrResult 端口属性结果
type PortAttrResult struct {
	Source  string
	Tag     string
	Content string
}

// PortResult 端口结果
type PortResult struct {
	HttpStatus string
	HttpBody   string
	HttpHeader string
	PortAttrs  []PortAttrResult
}

// IPResult IP结果
type IPResult struct {
	Org      string
	Location string
	Category string
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

func (r *Result) ParseResult(config execute.ExecutorTaskInfo) (docs []db.AssetDocument) {
	var portscanConfig execute.PortscanConfig
	for _, v := range config.PortScan {
		portscanConfig = v
		break
	}
	for ipName, ipResult := range r.IPResult {
		if portscanConfig.MaxOpenedPortPerIp > 0 && len(ipResult.Ports) > portscanConfig.MaxOpenedPortPerIp {
			logging.RuntimeLog.Warningf("IP:%s开放端口数量:%d，超过允许的最大数量，已过滤!", ipName, len(ipResult.Ports))
			continue
		}
		for portNumber, portResult := range ipResult.Ports {
			doc := db.AssetDocument{
				Authority: fmt.Sprintf("%s:%d", ipName, portNumber),
				Host:      ipName,
				Port:      portNumber,
				Category:  db.CategoryIPv4,
				Ip: db.IP{
					IpV4: []db.IPV4{{
						IPName: ipName,
					}},
				},
				HttpStatus: portResult.HttpStatus,
				HttpBody:   portResult.HttpBody,
				HttpHeader: portResult.HttpHeader,
				OrgId:      config.OrgId,
				TaskId:     config.MainTaskId,
			}
			for _, portAttrResult := range portResult.PortAttrs {
				switch portAttrResult.Tag {
				case db.FingerServer:
					doc.Server = portAttrResult.Content
				case db.FingerApp:
					doc.App = append(doc.App, portAttrResult.Content)
				case db.FingerBanner:
					doc.Banner = portAttrResult.Content
				case db.FingerTitle:
					doc.Title = portAttrResult.Content
				case db.FingerService:
					doc.Service = portAttrResult.Content
				}
			}
			docs = append(docs, doc)
		}
	}
	return
}
