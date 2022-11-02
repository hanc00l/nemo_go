package workerapi

import (
	"context"
	"errors"
	"fmt"
	"github.com/hanc00l/nemo_go/pkg/comm"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/db"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/domainscan"
	"github.com/hanc00l/nemo_go/pkg/task/onlineapi"
	"github.com/hanc00l/nemo_go/pkg/task/pocscan"
	"github.com/hanc00l/nemo_go/pkg/task/portscan"
	"github.com/remeh/sizedwaitgroup"
	"strings"
)

type XScanConfig struct {
	OrgId *int `json:"orgid,omitempty"`
	// orgscan
	IsOrgIP     bool   `json:"orgip,omitempty"`     //XOrganizaiton：IP资产
	IsOrgDomain bool   `json:"orgdomain,omitempty"` //XOrganizaiton：domain资产
	OrgIPPort   string `json:"orgport,omitempty"`   // Org扫描时，是否指定IP的端口
	// fofa
	FofaTarget      string `json:"fofatarget,omitempty"`
	FofaKeyword     string `json:"fofaKeyword,omitempty"`
	FofaSearchLimit int    `json:"fofaSearchLimit,omitempty"`
	// portscan
	IPPort       map[string][]int  `json:"ipport,omitempty"`       //IP:PORT列表
	IPPortString map[string]string `json:"ipportstring,omitempty"` //格式为ip列表，port可以为多种形式，如"80,443,8000-9000"、"--top-port 100"
	// domainscan
	Domain             map[string]struct{} `json:"domain,omitempty"`
	IsSubDomainFinder  bool                `json:"subfinder,omitempty"`
	IsSubDomainBrute   bool                `json:"subdomainBrute,omitempty"`
	IsIgnoreCDN        bool                `json:"ignorecdn,omitempty"`
	IsIgnoreOutofChina bool                `json:"ignoreoutofchina,omitempty"`
	// fingerprint
	IsHttpx          bool `json:"httpx,omitempty"`
	IsScreenshot     bool `json:"screenshot,omitempty"`
	IsFingerprintHub bool `json:"fingerprinthub,omitempty"`
	IsIconHash       bool `json:"iconhash,omitempty"`
	// xraypoc
	IsXrayPocScan bool   `json:"xraypocscan,omitempty"`
	XrayPocFile   string `json:"xraypocfile,omitempty"`
}

type XScan struct {
	Config XScanConfig

	ResultIP     portscan.Result
	ResultDomain domainscan.Result
}

const (
	portscanMaxThreadNum   = 4
	domainscanMaxThreadNum = 4
)

func NewXScan(config XScanConfig) *XScan {
	x := XScan{Config: config}

	return &x
}

// XOrganization 根据组织ID获取资产，并进行IP和域名的任务
func XOrganization(taskId, configJSON string) (result string, err error) {
	// 检查任务状态
	var ok bool
	if ok, result, err = CheckTaskStatus(taskId); !ok {
		return result, err
	}
	// 解析任务参数
	config := XScanConfig{}
	if err = ParseConfig(configJSON, &config); err != nil {
		return FailedTask(err.Error()), err
	}
	// 执行任务
	if config.OrgId == nil || *config.OrgId == 0 {
		logging.RuntimeLog.Error("no org id")
		return FailedTask("no org id"), errors.New("no org id")
	}
	scan := NewXScan(config)
	// 根据ID读取资产
	if scan.Config.IsOrgIP && scan.Config.IsOrgDomain {
		LoadIpAndDomainByOrgId(*config.OrgId, &scan.ResultIP, &scan.ResultDomain)
		result = fmt.Sprintf("ip:%d,domain:%d", len(scan.ResultIP.IPResult), len(scan.ResultDomain.DomainResult))
	} else if scan.Config.IsOrgIP {
		LoadIpAndDomainByOrgId(*config.OrgId, &scan.ResultIP, nil)
		result = fmt.Sprintf("ip:%d", len(scan.ResultIP.IPResult))
	} else if scan.Config.IsOrgDomain {
		LoadIpAndDomainByOrgId(*config.OrgId, nil, &scan.ResultDomain)
		result = fmt.Sprintf("domain:%d", len(scan.ResultDomain.DomainResult))
	}
	// 执行portscan与domainscan
	ipPortMap, domainMap := MakeSubTaskTarget(&scan.ResultIP, &scan.ResultDomain)
	if len(ipPortMap) > 0 {
		// 指定了扫描的端口
		if len(config.OrgIPPort) > 0 {
			var ipPortMapString []map[string]string
			for _, ipm := range ipPortMap {
				ipp := make(map[string]string)
				for ip := range ipm {
					ipp[ip] = config.OrgIPPort
				}
				ipPortMapString = append(ipPortMapString, ipp)
			}
			_, err = scan.NewPortScan(nil, ipPortMapString)
		} else {
			_, err = scan.NewPortScan(ipPortMap, nil)
		}
		if err != nil {
			logging.RuntimeLog.Error(err)
			return FailedTask(err.Error()), err
		}
	}
	// 域名任务只执行解析不进行子域名任务
	_, err = scan.NewDomainScan(domainMap, false, false)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return FailedTask(err.Error()), err
	}

	return SucceedTask(result), nil
}

// XFofa Fofa任务
func XFofa(taskId, configJSON string) (result string, err error) {
	// 检查任务状态
	var ok bool
	if ok, result, err = CheckTaskStatus(taskId); !ok {
		return result, err
	}
	// 解析任务参数
	config := XScanConfig{}
	if err = ParseConfig(configJSON, &config); err != nil {
		return FailedTask(err.Error()), err
	}
	// 执行任务
	scan := NewXScan(config)
	result, err = scan.FofaSearch(taskId)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return FailedTask(err.Error()), err
	}
	// 执行portscan与domainscan
	ipPortMap, domainMap := MakeSubTaskTarget(&scan.ResultIP, &scan.ResultDomain)
	_, err = scan.NewPortScan(ipPortMap, nil)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return FailedTask(err.Error()), err
	}
	//域名任务只执行解析不进行子域名任务
	_, err = scan.NewDomainScan(domainMap, false, false)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return FailedTask(err.Error()), err
	}

	return SucceedTask(result), nil
}

// XPortScan 端口扫描任务
func XPortScan(taskId, configJSON string) (result string, err error) {
	// 检查任务状态
	var ok bool
	if ok, result, err = CheckTaskStatus(taskId); !ok {
		return result, err
	}
	// 解析任务参数
	config := XScanConfig{}
	if err = ParseConfig(configJSON, &config); err != nil {
		return FailedTask(err.Error()), err
	}
	// 执行任务
	scan := NewXScan(config)
	result, err = scan.Portscan(taskId)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return FailedTask(err.Error()), err
	}
	// 启动指纹识别任务：
	_, err = scan.NewFingerprintScan()
	if err != nil {
		logging.RuntimeLog.Error(err)
		return FailedTask(err.Error()), err
	}
	//
	return SucceedTask(result), nil
}

// XDomainscan 域名任务
func XDomainscan(taskId, configJSON string) (result string, err error) {
	// 检查任务状态
	var ok bool
	if ok, result, err = CheckTaskStatus(taskId); !ok {
		return result, err
	}
	// 解析任务参数
	config := XScanConfig{}
	if err = ParseConfig(configJSON, &config); err != nil {
		return FailedTask(err.Error()), err
	}
	// 执行任务
	scan := NewXScan(config)
	result, err = scan.Domainscan(taskId)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return FailedTask(err.Error()), err
	}
	// 启动指纹识别任务：
	_, err = scan.NewFingerprintScan()
	if err != nil {
		logging.RuntimeLog.Error(err)
		return FailedTask(err.Error()), err
	}
	return SucceedTask(result), nil
}

// XFingerPrint指纹识别任务
func XFingerPrint(taskId, configJSON string) (result string, err error) {
	// 检查任务状态
	var ok bool
	if ok, result, err = CheckTaskStatus(taskId); !ok {
		return result, err
	}
	// 解析任务参数
	config := XScanConfig{}
	if err = ParseConfig(configJSON, &config); err != nil {
		return FailedTask(err.Error()), err
	}
	// 执行任务
	scan := NewXScan(config)
	result, err = scan.FingerPrint(taskId)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return FailedTask(err.Error()), err
	}
	// 启动XrayPoc任务
	if config.IsXrayPocScan {
		_, err = scan.NewXrayPoc()
		if err != nil {
			return FailedTask(err.Error()), err
		}
	}
	return SucceedTask(result), nil
}

// XXrayPocScan XrayPoc扫描任务
func XXrayPocScan(taskId, configJSON string) (result string, err error) {
	// 检查任务状态
	var ok bool
	if ok, result, err = CheckTaskStatus(taskId); !ok {
		return result, err
	}
	// 解析任务参数
	config := XScanConfig{}
	if err = ParseConfig(configJSON, &config); err != nil {
		return FailedTask(err.Error()), err
	}
	// 执行任务
	scan := NewXScan(config)
	result, err = scan.XrayPocScan(taskId)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return FailedTask(err.Error()), err
	}
	//
	return SucceedTask(result), nil
}

// Portscan 执行端口扫描，通过协程并发执行
func (x *XScan) Portscan(taskId string) (result string, err error) {
	x.ResultIP.IPResult = make(map[string]*portscan.IPResult)

	swg := sizedwaitgroup.New(portscanMaxThreadNum)
	// 生成扫描参数
	defaultConf := conf.GlobalWorkerConfig().Portscan
	config := portscan.Config{
		OrgId:        x.Config.OrgId,
		Rate:         defaultConf.Rate,
		IsPing:       defaultConf.IsPing,
		Tech:         defaultConf.Tech,
		IsIpLocation: true,
		CmdBin:       defaultConf.Cmdbin,
	}
	if len(x.Config.IPPortString) > 0 {
		for ip, ports := range x.Config.IPPortString {
			if len(ports) <= 0 {
				continue
			}
			runConfig := config
			runConfig.Target = ip
			runConfig.Port = ports
			swg.Add()
			//执行扫描
			go x.doPortscan(&swg, runConfig)
		}
	}
	if len(x.Config.IPPort) > 0 {
		for ip, ports := range x.Config.IPPort {
			if len(ports) <= 0 {
				continue
			}
			//按IP执行扫描任务
			var ps []string
			for _, p := range ports {
				ps = append(ps, fmt.Sprintf("%d", p))
			}
			runConfig := config
			runConfig.Target = ip
			runConfig.Port = strings.Join(ps, ",")
			swg.Add()
			//执行扫描
			go x.doPortscan(&swg, runConfig)
		}
	}
	swg.Wait()
	// 保存结果
	resultArgs := comm.ScanResultArgs{
		TaskID:   taskId,
		IPConfig: &portscan.Config{OrgId: config.OrgId},
		IPResult: x.ResultIP.IPResult,
	}
	xc := comm.NewXClient()
	err = xc.Call(context.Background(), "SaveScanResult", &resultArgs, &result)
	if err != nil {
		logging.RuntimeLog.Error(err)
	}
	return
}

// doPortscan 调用一次端口扫描
func (x *XScan) doPortscan(swg *sizedwaitgroup.SizedWaitGroup, config portscan.Config) {
	var result portscan.Result
	if config.CmdBin == "masnmap" {
		result.IPResult = doMasscanPlusNmap(config).IPResult
	} else if config.CmdBin == "masscan" {
		m := portscan.NewMasscan(config)
		m.Do()
		result.IPResult = m.Result.IPResult
	} else {
		m := portscan.NewNmap(config)
		m.Do()
		result.IPResult = m.Result.IPResult
	}
	//合并结果
	x.ResultIP.Lock()
	for k, v := range result.IPResult {
		x.ResultIP.IPResult[k] = v
	}
	x.ResultIP.Unlock()

	swg.Done()
}

// doDomainscan 调用执行一次域名任务
func (x *XScan) doDomainscan(swg *sizedwaitgroup.SizedWaitGroup, config domainscan.Config) {
	var result domainscan.Result
	//扫描
	result = doDomainScan(config)
	//合并结果
	x.ResultDomain.Lock()
	for k, v := range result.DomainResult {
		x.ResultDomain.DomainResult[k] = v
	}
	x.ResultDomain.Unlock()

	swg.Done()
}

// FofaSearch 执行fofa搜索任务
func (x *XScan) FofaSearch(taskId string) (result string, err error) {
	config := onlineapi.OnlineAPIConfig{
		OrgId:              x.Config.OrgId,
		IsIPLocation:       true,
		IsHttpx:            x.Config.IsHttpx,
		IsScreenshot:       x.Config.IsScreenshot,
		IsFingerprintHub:   x.Config.IsFingerprintHub,
		IsIconHash:         x.Config.IsIconHash,
		IsIgnoreCDN:        x.Config.IsIgnoreCDN,
		IsIgnoreOutofChina: x.Config.IsIgnoreOutofChina,
	}
	//fofa任务支持两种模式：
	//一种是关键词，需设置SearchByKeyWord为true
	//另一种是ip/domain
	if len(x.Config.FofaKeyword) > 0 {
		config.SearchByKeyWord = true
		config.Target = x.Config.FofaKeyword
		config.SearchLimitCount = x.Config.FofaSearchLimit
	} else if len(x.Config.FofaTarget) > 0 {
		config.Target = x.Config.FofaTarget
	}
	x.ResultIP, x.ResultDomain, result, err = doFofaAndSave(taskId, "fofa", config)

	return
}

// Domainscan 执行域名任务
func (x *XScan) Domainscan(taskId string) (result string, err error) {
	x.ResultDomain.DomainResult = make(map[string]*domainscan.DomainResult)
	swg := sizedwaitgroup.New(domainscanMaxThreadNum)

	config := domainscan.Config{
		OrgId:              x.Config.OrgId,
		IsSubDomainFinder:  x.Config.IsSubDomainFinder,
		IsSubDomainBrute:   x.Config.IsSubDomainBrute,
		IsIgnoreCDN:        x.Config.IsIgnoreCDN,
		IsIgnoreOutofChina: x.Config.IsIgnoreOutofChina,
	}
	for domain := range x.Config.Domain {
		runConfig := config
		runConfig.Target = domain
		swg.Add()
		go x.doDomainscan(&swg, runConfig)

	}
	swg.Wait()
	// 如果有端口扫描的选项
	if config.IsIPPortScan || config.IsIPSubnetPortScan {
		doPortScanByDomainscan(config, &x.ResultDomain)
	}
	// 保存结果
	xc := comm.NewXClient()
	resultArgs := comm.ScanResultArgs{
		TaskID:       taskId,
		DomainConfig: &domainscan.Config{OrgId: config.OrgId},
		DomainResult: x.ResultDomain.DomainResult,
	}
	err = xc.Call(context.Background(), "SaveScanResult", &resultArgs, &result)
	if err != nil {
		logging.RuntimeLog.Error(err)
	}
	return
}

// NewPortScan 根据IP/port列表，生成端口扫描任务
func (x *XScan) NewPortScan(ipPortMap []map[string][]int, ipPortMapString []map[string]string) (result string, err error) {
	config := XScanConfig{
		OrgId:              x.Config.OrgId,
		IsIgnoreCDN:        x.Config.IsIgnoreCDN,
		IsIgnoreOutofChina: x.Config.IsIgnoreOutofChina,
		IsHttpx:            x.Config.IsHttpx,
		IsScreenshot:       x.Config.IsScreenshot,
		IsFingerprintHub:   x.Config.IsFingerprintHub,
		IsIconHash:         x.Config.IsIconHash,
		IsXrayPocScan:      x.Config.IsXrayPocScan,
	}
	for _, t := range ipPortMap {
		configRun := config
		configRun.IPPort = t
		result, err = sendTask(configRun, "xportscan")
		if err != nil {
			logging.RuntimeLog.Error(err)
			return
		}
	}
	for _, t := range ipPortMapString {
		configRun := config
		configRun.IPPortString = t
		result, err = sendTask(configRun, "xportscan")
		if err != nil {
			logging.RuntimeLog.Error(err)
			return
		}
	}
	return
}

// NewDomainScan 根据域名列表，生成域名任务
func (x *XScan) NewDomainScan(domainMap []map[string]struct{}, isSubDomainFinder, isSubDomainBrute bool) (result string, err error) {
	config := XScanConfig{
		OrgId:             x.Config.OrgId,
		IsSubDomainFinder: isSubDomainFinder,
		IsSubDomainBrute:  isSubDomainBrute,
		//
		IsIgnoreCDN:        x.Config.IsIgnoreCDN,
		IsIgnoreOutofChina: x.Config.IsIgnoreOutofChina,
		IsHttpx:            x.Config.IsHttpx,
		IsScreenshot:       x.Config.IsScreenshot,
		IsFingerprintHub:   x.Config.IsFingerprintHub,
		IsIconHash:         x.Config.IsIconHash,
		IsXrayPocScan:      x.Config.IsXrayPocScan,
	}
	for _, t := range domainMap {
		configRun := config
		configRun.Domain = t
		result, err = sendTask(configRun, "xdomainscan")
		if err != nil {
			logging.RuntimeLog.Error(err)
			return
		}
	}
	return
}

// FingerPrint 执行指纹识别任务
func (x *XScan) FingerPrint(taskId string) (result string, err error) {
	config := FingerprintTaskConfig{
		IsHttpx:          x.Config.IsHttpx,
		IsFingerprintHub: x.Config.IsFingerprintHub,
		IsIconHash:       x.Config.IsIconHash,
		IsScreenshot:     x.Config.IsScreenshot,
		IPTargetMap:      x.Config.IPPort,
		DomainTargetMap:  x.Config.Domain,
	}
	x.ResultIP, x.ResultDomain, result, err = doFingerPrintAndSave(taskId, config)

	return
}

// NewFingerprintScan 生成指纹识别任务
func (x *XScan) NewFingerprintScan() (result string, err error) {
	if x.Config.IsHttpx == false && x.Config.IsFingerprintHub == false && x.Config.IsIconHash == false && x.Config.IsScreenshot == false {
		return
	}
	config := XScanConfig{
		IsHttpx:          x.Config.IsHttpx,
		IsScreenshot:     x.Config.IsScreenshot,
		IsFingerprintHub: x.Config.IsFingerprintHub,
		IsIconHash:       x.Config.IsIconHash,
		IsXrayPocScan:    x.Config.IsXrayPocScan,
	}
	//拆分子任务
	ipTarget, domainTarget := MakeSubTaskTarget(&x.ResultIP, &x.ResultDomain)
	//生成任务
	for _, t := range ipTarget {
		newConfig := config
		newConfig.IPPort = t
		result, err = sendTask(newConfig, "xfingerprint")
		if err != nil {
			return
		}
	}
	for _, t := range domainTarget {
		newConfig := config
		newConfig.Domain = t
		result, err = sendTask(newConfig, "xfingerprint")
		if err != nil {
			return
		}
	}
	return
}

// NewXrayPoc 生成xraypoc任务
func (x *XScan) NewXrayPoc() (result string, err error) {
	//拆分子任务
	ipTarget, domainTarget := MakeSubTaskTarget(&x.ResultIP, &x.ResultDomain)
	for _, t := range ipTarget {
		newConfig := XScanConfig{IPPort: t, IsXrayPocScan: true}
		result, err = sendTask(newConfig, "xxraypoc")
		if err != nil {
			return
		}
	}
	for _, t := range domainTarget {
		newConfig := XScanConfig{Domain: t, IsXrayPocScan: true}
		result, err = sendTask(newConfig, "xxraypoc")
		if err != nil {
			return
		}
	}
	return
}

// XrayPocScan 调用执行xraypoc扫描任务
func (x *XScan) XrayPocScan(taskId string) (result string, err error) {
	config := pocscan.XrayPocConfig{
		IPPort: x.Config.IPPort,
		Domain: x.Config.Domain,
	}
	result, err = doXrayPocScanAndSave(taskId, config)
	return
}

// LoadIpAndDomainByOrgId 根据组织ID获取ip与域名资产
func LoadIpAndDomainByOrgId(orgId int, portScanResult *portscan.Result, domainScanResult *domainscan.Result) {
	searchMap := make(map[string]interface{})
	searchMap["org_id"] = orgId

	if portScanResult != nil {
		portScanResult.IPResult = make(map[string]*portscan.IPResult)
		ipDb := db.Ip{}
		ipResults, _ := ipDb.Gets(searchMap, 1, 1000000, false)
		for _, ipRow := range ipResults {
			portScanResult.SetIP(ipRow.IpName)
			portDb := db.Port{IpId: ipRow.Id}
			portResults := portDb.GetsByIPId()
			for _, port := range portResults {
				portScanResult.SetPort(ipRow.IpName, port.PortNum)
			}
		}
	}
	if domainScanResult != nil {
		domainScanResult.DomainResult = make(map[string]*domainscan.DomainResult)
		domainDb := db.Domain{}
		domainResults, _ := domainDb.Gets(searchMap, 1, 1000000, false)
		for _, domainRow := range domainResults {
			domainScanResult.SetDomain(domainRow.DomainName)
		}
	}
	return
}
