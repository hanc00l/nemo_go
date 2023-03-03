package runner

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/db"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/custom"
	"github.com/hanc00l/nemo_go/pkg/task/domainscan"
	"github.com/hanc00l/nemo_go/pkg/task/onlineapi"
	"github.com/hanc00l/nemo_go/pkg/task/pocscan"
	"github.com/hanc00l/nemo_go/pkg/task/portscan"
	"github.com/hanc00l/nemo_go/pkg/task/serverapi"
	"github.com/hanc00l/nemo_go/pkg/task/workerapi"
	"github.com/hanc00l/nemo_go/pkg/utils"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// StartPortScanTask 端口扫描任务
func StartPortScanTask(req PortscanRequestParam, mainTaskId string, workspaceId int) (taskId string, err error) {
	// 解析参数
	ts := utils.NewTaskSlice()
	ts.TaskMode = req.TaskMode
	ts.IpTarget = formatIpTarget(req.Target, req.OrgId)
	ts.Port = req.Port
	tc := conf.GlobalServerConfig().Task
	ts.IpSliceNumber = tc.IpSliceNumber
	ts.PortSliceNumber = tc.PortSliceNumber
	targets, ports := ts.DoIpSlice()
	for _, t := range targets {
		for _, p := range ports {
			// 端口扫描
			if taskId, err = doPortscan(workspaceId, mainTaskId, t, p, req); err != nil {
				return
			}
			// IP归属地：如果有端口执行任务，则IP归属地任务在端口扫描中执行，否则单独执行
			// 如果IP地址是带掩码的子网（如192.168.1.0/24）则不进行归属地查询（在实际中容易出现误操作，导致整段IP地址无意义地进行归属地查询）
			if !req.IsPortScan && req.IsIPLocation && utils.CheckIPV4Subnet(t) == false {
				if taskId, err = doIPLocation(mainTaskId, t, &req.OrgId); err != nil {
					return
				}
			}
			// FOFA
			if req.IsFofa {
				if taskId, err = doOnlineAPISearch(workspaceId, mainTaskId, "fofa", t, &req.OrgId, req.IsIPLocation, req.IsHttpx, req.IsFingerprintHub, req.IsScreenshot, req.IsIconHash, req.IsIgnoreCDN, req.IsIgnoreOutofChina); err != nil {
					return
				}
			}
			// Quake
			if req.IsQuake {
				if taskId, err = doOnlineAPISearch(workspaceId, mainTaskId, "quake", t, &req.OrgId, req.IsIPLocation, req.IsHttpx, req.IsFingerprintHub, req.IsScreenshot, req.IsIconHash, req.IsIgnoreCDN, req.IsIgnoreOutofChina); err != nil {
					return
				}
			}
			// Hunter
			if req.IsHunter {
				if taskId, err = doOnlineAPISearch(workspaceId, mainTaskId, "hunter", t, &req.OrgId, req.IsIPLocation, req.IsHttpx, req.IsFingerprintHub, req.IsScreenshot, req.IsIconHash, req.IsIgnoreCDN, req.IsIgnoreOutofChina); err != nil {
					return
				}
			}
		}
	}
	return taskId, nil
}

// StartBatchScanTask 探测+扫描任务
func StartBatchScanTask(req PortscanRequestParam, mainTaskId string, workspaceId int) (taskId string, err error) {
	ts := utils.NewTaskSlice()
	ts.TaskMode = req.TaskMode
	ts.IpTarget = formatIpTarget(req.Target, req.OrgId)
	ts.Port = req.Port
	tc := conf.GlobalServerConfig().Task
	ts.IpSliceNumber = tc.IpSliceNumber
	ts.PortSliceNumber = tc.PortSliceNumber
	targets, ports := ts.DoIpSlice()
	for _, t := range targets {
		for _, p := range ports {
			// 端口扫描
			if taskId, err = doBatchScan(workspaceId, mainTaskId, t, p, req); err != nil {
				return
			}
		}
	}
	return taskId, nil
}

// StartDomainScanTask 域名任务
func StartDomainScanTask(req DomainscanRequestParam, mainTaskId string, workspaceId int) (taskId string, err error) {
	ts := utils.NewTaskSlice()
	domainTargetList := formatDomainTarget(req.Target)
	// 域名的FLD
	if req.IsFldDomain {
		ts.DomainTarget = getDomainFLD(domainTargetList)
	} else {
		ts.DomainTarget = domainTargetList
	}
	ts.TaskMode = req.TaskMode
	targets := ts.DoDomainSlice()
	for _, t := range targets {
		// 每个获取子域名的方式采用独立任务，以提高速度
		var taskStarted bool
		if req.IsSubfinder {
			subConfig := req
			subConfig.IsSubdomainBrute = false
			subConfig.IsCrawler = false
			if taskId, err = doDomainscan(workspaceId, mainTaskId, t, subConfig); err != nil {
				return
			}
			taskStarted = true
		}
		if req.IsSubdomainBrute {
			subConfig := req
			subConfig.IsSubfinder = false
			subConfig.IsCrawler = false
			if taskId, err = doDomainscan(workspaceId, mainTaskId, t, subConfig); err != nil {
				return
			}
			taskStarted = true
		}
		if req.IsCrawler {
			subConfig := req
			subConfig.IsSubfinder = false
			subConfig.IsSubdomainBrute = false
			if taskId, err = doDomainscan(workspaceId, mainTaskId, t, subConfig); err != nil {
				return
			}
			taskStarted = true
		}
		// 如果没有子域名任务，则至少启动一个域名解析任务
		if !taskStarted {
			if taskId, err = doDomainscan(workspaceId, mainTaskId, t, req); err != nil {
				return
			}
		}
		if req.IsFofa {
			if taskId, err = doOnlineAPISearch(workspaceId, mainTaskId, "fofa", t, &req.OrgId, true, req.IsHttpx, req.IsFingerprintHub, req.IsScreenshot, req.IsIconHash, req.IsIgnoreCDN, req.IsIgnoreOutofChina); err != nil {
				return
			}
		}
		if req.IsQuake {
			if taskId, err = doOnlineAPISearch(workspaceId, mainTaskId, "quake", t, &req.OrgId, true, req.IsHttpx, req.IsFingerprintHub, req.IsScreenshot, req.IsIconHash, req.IsIgnoreCDN, req.IsIgnoreOutofChina); err != nil {
				return
			}
		}
		if req.IsHunter {
			if taskId, err = doOnlineAPISearch(workspaceId, mainTaskId, "hunter", t, &req.OrgId, true, req.IsHttpx, req.IsFingerprintHub, req.IsScreenshot, req.IsIconHash, req.IsIgnoreCDN, req.IsIgnoreOutofChina); err != nil {
				return
			}
		}
		if req.IsICPQuery {
			if taskId, err = doICPQuery(mainTaskId, t); err != nil {
				return
			}
		}
		if req.IsWhoisQuery {
			if taskId, err = doWhoisQuery(mainTaskId, t); err != nil {
				return
			}
		}
	}
	return taskId, nil
}

// StartPocScanTask pocscan任务
func StartPocScanTask(req PocscanRequestParam, mainTaskId string, workspaceId int) (taskId string, err error) {
	var targetList []string
	for _, t := range strings.Split(req.Target, "\n") {
		if tt := strings.TrimSpace(t); tt != "" {
			targetList = append(targetList, tt)
		}
	}
	if req.IsXrayVerify && req.XrayPocFile != "" {
		config := pocscan.Config{Target: strings.Join(targetList, ","), PocFile: req.XrayPocFile, CmdBin: "xray", IsLoadOpenedPort: req.IsLoadOpenedPort, WorkspaceId: workspaceId}
		configJSON, _ := json.Marshal(config)
		taskId, err = serverapi.NewRunTask("xray", string(configJSON), mainTaskId, "")
		if err != nil {
			return
		}
	}
	if req.IsNucleiVerify && req.NucleiPocFile != "" {
		config := pocscan.Config{Target: strings.Join(targetList, ","), PocFile: req.NucleiPocFile, CmdBin: "nuclei", IsLoadOpenedPort: req.IsLoadOpenedPort, WorkspaceId: workspaceId}
		configJSON, _ := json.Marshal(config)
		taskId, err = serverapi.NewRunTask("nuclei", string(configJSON), mainTaskId, "")
		if err != nil {
			return
		}
	}
	if req.IsDirsearch && req.DirsearchExtName != "" {
		config := pocscan.Config{Target: strings.Join(targetList, ","), PocFile: req.DirsearchExtName, CmdBin: "dirsearch", IsLoadOpenedPort: req.IsLoadOpenedPort, WorkspaceId: workspaceId}
		configJSON, _ := json.Marshal(config)
		taskId, err = serverapi.NewRunTask("dirsearch", string(configJSON), mainTaskId, "")
		if err != nil {
			return
		}
	}
	return taskId, nil
}

// StartXFofaKeywordTask xscan任务，根据fofa关键字查询资产
func StartXFofaKeywordTask(req XScanRequestParam, mainTaskId string, workspaceId int) (taskId string, err error) {
	config := workerapi.XScanConfig{
		OrgId:              &req.OrgId,
		IsIgnoreCDN:        false,
		IsIgnoreOutofChina: req.IsCn,
		IsXrayPoc:          req.IsXrayPocscan,
		XrayPocFile:        req.XrayPocFile,
		WorkspaceId:        workspaceId,
	}
	// config.OrgId 为int，默认为0
	// db.Organization.OrgId为指针，默认nil
	if *config.OrgId == 0 {
		config.OrgId = nil
	}
	if req.IsFingerprint {
		config.IsHttpx = conf.GlobalWorkerConfig().Fingerprint.IsHttpx
		config.IsFingerprintHub = conf.GlobalWorkerConfig().Fingerprint.IsFingerprintHub
		config.IsScreenshot = conf.GlobalWorkerConfig().Fingerprint.IsScreenshot
		config.IsIconHash = conf.GlobalWorkerConfig().Fingerprint.IsIconHash
	}
	// 生成查询语法
	keywords := searchKeyword(req)
	for keyword, count := range keywords {
		configRun := config
		configRun.FofaKeyword = keyword
		configRun.FofaSearchLimit = count
		configJSONRun, _ := json.Marshal(configRun)
		taskId, err = serverapi.NewRunTask("xfofa", string(configJSONRun), mainTaskId, "")
		if err != nil {
			logging.RuntimeLog.Errorf("start xfofa fail:%s", err.Error())
			return "", err
		}
	}
	return
}

// StartXDomainScanTask xscan任务，域名任务
func StartXDomainScanTask(req XScanRequestParam, mainTaskId string, workspaceId int) (taskId string, err error) {
	config := workerapi.XScanConfig{
		OrgId:              &req.OrgId,
		IsIgnoreCDN:        false,
		IsIgnoreOutofChina: req.IsCn,
		IsXrayPoc:          req.IsXrayPocscan,
		XrayPocFile:        req.XrayPocFile,
		WorkspaceId:        workspaceId,
	}
	// config.OrgId 为int，默认为0
	// db.Organization.OrgId为指针，默认nil
	if *config.OrgId == 0 {
		config.OrgId = nil
	}
	if req.IsFingerprint {
		config.IsHttpx = conf.GlobalWorkerConfig().Fingerprint.IsHttpx
		config.IsFingerprintHub = conf.GlobalWorkerConfig().Fingerprint.IsFingerprintHub
		config.IsScreenshot = conf.GlobalWorkerConfig().Fingerprint.IsScreenshot
		config.IsIconHash = conf.GlobalWorkerConfig().Fingerprint.IsIconHash
	}
	targetList := formatDomainTarget(req.Target)
	for _, target := range targetList {
		// 忽略IP
		if utils.CheckIPV4(target) || utils.CheckIPV4Subnet(target) {
			continue
		}
		configRun := config
		configRun.Domain = make(map[string]struct{})
		configRun.Domain[target] = struct{}{}
		// 子域名枚举和爆破拆分成两个任务并行执行
		configRun.IsSubDomainFinder = true
		configRun.IsSubDomainBrute = false
		configJSON, _ := json.Marshal(configRun)
		taskId, err = serverapi.NewRunTask("xdomainscan", string(configJSON), mainTaskId, "")
		if err != nil {
			logging.RuntimeLog.Errorf("start xdomainscan fail:%s", err.Error())
			return "", err
		}
		configRun.IsSubDomainFinder = false
		configRun.IsSubDomainBrute = true
		configJSON, _ = json.Marshal(configRun)
		taskId, err = serverapi.NewRunTask("xdomainscan", string(configJSON), mainTaskId, "")
		if err != nil {
			logging.RuntimeLog.Errorf("start xdomainscan fail:%s", err.Error())
			return "", err
		}
		if req.IsFofaSearch {
			configRunFofa := config
			configRunFofa.FofaTarget = target
			configJSONFofa, _ := json.Marshal(configRunFofa)
			taskId, err = serverapi.NewRunTask("xfofa", string(configJSONFofa), mainTaskId, "")
			if err != nil {
				logging.RuntimeLog.Errorf("start xfofa fail:%s", err.Error())
				return "", err
			}
		}
	}
	return taskId, nil
}

// StartXPortScanTask xscan的IP任务
func StartXPortScanTask(req XScanRequestParam, mainTaskId string, workspaceId int) (taskId string, err error) {
	config := workerapi.XScanConfig{
		OrgId:              &req.OrgId,
		IsIgnoreCDN:        false,
		IsIgnoreOutofChina: req.IsCn,
		IsXrayPoc:          req.IsXrayPocscan,
		XrayPocFile:        req.XrayPocFile,
		WorkspaceId:        workspaceId,
	}
	// config.OrgId 为int，默认为0
	// db.Organization.OrgId为指针，默认nil
	if *config.OrgId == 0 {
		config.OrgId = nil
	}
	if req.IsFingerprint {
		config.IsHttpx = conf.GlobalWorkerConfig().Fingerprint.IsHttpx
		config.IsFingerprintHub = conf.GlobalWorkerConfig().Fingerprint.IsFingerprintHub
		config.IsScreenshot = conf.GlobalWorkerConfig().Fingerprint.IsScreenshot
		config.IsIconHash = conf.GlobalWorkerConfig().Fingerprint.IsIconHash
	}
	ts := utils.NewTaskSlice()
	ts.TaskMode = utils.SliceByIP
	ts.IpTarget = formatIpTarget(req.Target, req.OrgId)
	ts.Port = req.Port
	tc := conf.GlobalServerConfig().Task
	ts.IpSliceNumber = tc.IpSliceNumber
	ts.PortSliceNumber = tc.PortSliceNumber
	targets, _ := ts.DoIpSlice()
	for _, target := range targets {
		configRun := config
		configRun.IPPortString = make(map[string]string)
		configRun.IPPortString[target] = req.Port
		configJSON, _ := json.Marshal(configRun)
		taskId, err = serverapi.NewRunTask("xportscan", string(configJSON), mainTaskId, "")
		if err != nil {
			logging.RuntimeLog.Errorf("start xportscan fail:%s", err.Error())
			return "", err
		}
		if req.IsFofaSearch {
			configRunFofa := config
			configRunFofa.FofaTarget = target
			configJSONFofa, _ := json.Marshal(configRunFofa)
			taskId, err = serverapi.NewRunTask("xfofa", string(configJSONFofa), mainTaskId, "")
			if err != nil {
				logging.RuntimeLog.Errorf("start xfofa fail:%s", err.Error())
				return "", err
			}
		}
	}
	return taskId, nil
}

// StartXOrgScanTask xscan任务，获取指定组织的资产并开始扫描任务
func StartXOrgScanTask(req XScanRequestParam, mainTaskId string, workspaceId int) (taskId string, err error) {
	config := workerapi.XScanConfig{
		OrgId:              &req.OrgId,
		IsOrgIP:            req.IsOrgIP,
		IsOrgDomain:        req.IsOrgDomain,
		OrgIPPort:          req.Port,
		IsIgnoreCDN:        false,
		IsIgnoreOutofChina: req.IsCn,
		IsXrayPoc:          req.IsXrayPocscan,
		XrayPocFile:        req.XrayPocFile,
		WorkspaceId:        workspaceId,
	}
	if req.IsFingerprint {
		config.IsHttpx = conf.GlobalWorkerConfig().Fingerprint.IsHttpx
		config.IsFingerprintHub = conf.GlobalWorkerConfig().Fingerprint.IsFingerprintHub
		config.IsScreenshot = conf.GlobalWorkerConfig().Fingerprint.IsScreenshot
		config.IsIconHash = conf.GlobalWorkerConfig().Fingerprint.IsIconHash
	}
	configJSON, _ := json.Marshal(config)
	taskId, err = serverapi.NewRunTask("xorgscan", string(configJSON), mainTaskId, "")
	if err != nil {
		logging.RuntimeLog.Errorf("start xorgscan fail:%s", err.Error())
		return "", err
	}
	return
}

// doPortscan 端口扫描
func doPortscan(workspaceId int, mainTaskId string, target string, port string, req PortscanRequestParam) (taskId string, err error) {
	config := portscan.Config{
		Target:           target,
		ExcludeTarget:    req.ExcludeIP,
		Port:             port,
		OrgId:            &req.OrgId,
		Rate:             req.Rate,
		IsPing:           req.IsPing,
		Tech:             req.NmapTech,
		IsIpLocation:     req.IsIPLocation,
		IsHttpx:          req.IsHttpx,
		IsScreenshot:     req.IsScreenshot,
		IsFingerprintHub: req.IsFingerprintHub,
		IsIconHash:       req.IsIconHash,
		CmdBin:           req.CmdBin,
		IsPortscan:       req.IsPortScan,
		IsLoadOpenedPort: req.IsLoadOpenedPort,
		WorkspaceId:      workspaceId,
	}
	if req.CmdBin == "" {
		config.CmdBin = conf.GlobalWorkerConfig().Portscan.Cmdbin
	}
	if config.Port == "" {
		config.Port = conf.GlobalWorkerConfig().Portscan.Port
	}
	if config.Rate == 0 {
		config.Rate = conf.GlobalWorkerConfig().Portscan.Rate
	}
	if config.Tech == "" {
		config.Target = conf.GlobalWorkerConfig().Portscan.Tech
	}
	// config.OrgId 为int，默认为0
	// db.Organization.OrgId为指针，默认nil
	if *config.OrgId == 0 {
		config.OrgId = nil
	}
	configJSON, err := json.Marshal(config)
	if err != nil {
		logging.RuntimeLog.Errorf("start portscan fail:%s", err.Error())
		return "", err
	}
	taskId, err = serverapi.NewRunTask("portscan", string(configJSON), mainTaskId, "")
	if err != nil {
		logging.RuntimeLog.Errorf("start portscan fail:%s", err.Error())
		return "", err
	}
	return taskId, nil
}

// doBatchScan 探测+端口扫描
func doBatchScan(workspaceId int, mainTaskId string, target string, port string, req PortscanRequestParam) (taskId string, err error) {
	config := portscan.Config{
		Target:           target,
		ExcludeTarget:    req.ExcludeIP,
		Port:             port,
		OrgId:            &req.OrgId,
		Rate:             req.Rate,
		IsPing:           req.IsPing,
		Tech:             req.NmapTech,
		IsIpLocation:     req.IsIPLocation,
		IsHttpx:          req.IsHttpx,
		IsScreenshot:     req.IsScreenshot,
		IsFingerprintHub: req.IsFingerprintHub,
		IsIconHash:       req.IsIconHash,
		CmdBin:           "masscan",
		WorkspaceId:      workspaceId,
	}
	if req.CmdBin == "nmap" {
		config.CmdBin = "nmap"
	}
	if config.Port == "" {
		config.Port = "80,443,8080|" + conf.GlobalWorkerConfig().Portscan.Port
	}
	if config.Rate == 0 {
		config.Rate = conf.GlobalWorkerConfig().Portscan.Rate
	}
	if config.Tech == "" {
		config.Target = conf.GlobalWorkerConfig().Portscan.Tech
	}
	// config.OrgId 为int，默认为0
	// db.Organization.OrgId为指针，默认nil
	if *config.OrgId == 0 {
		config.OrgId = nil
	}
	configJSON, err := json.Marshal(config)
	if err != nil {
		logging.RuntimeLog.Errorf("start batchscan fail:%s", err.Error())
		return "", err
	}
	taskId, err = serverapi.NewRunTask("batchscan", string(configJSON), mainTaskId, "")
	if err != nil {
		logging.RuntimeLog.Errorf("start batchscan fail:%s", err.Error())
		return "", err
	}
	return taskId, nil
}

// doDomainscan 域名任务
func doDomainscan(workspaceId int, mainTaskId string, target string, req DomainscanRequestParam) (taskId string, err error) {
	config := domainscan.Config{
		Target:             target,
		OrgId:              &req.OrgId,
		IsSubDomainFinder:  req.IsSubfinder,
		IsSubDomainBrute:   req.IsSubdomainBrute,
		IsCrawler:          req.IsCrawler,
		IsHttpx:            req.IsHttpx,
		IsIPPortScan:       req.IsIPPortscan,
		IsIPSubnetPortScan: req.IsSubnetPortscan,
		IsScreenshot:       req.IsScreenshot,
		IsFingerprintHub:   req.IsFingerprintHub,
		IsIconHash:         req.IsIconHash,
		PortTaskMode:       req.PortTaskMode,
		WorkspaceId:        workspaceId,
	}
	// config.OrgId 为int，默认为0
	// db.Organization.OrgId为指针，默认nil
	if *config.OrgId == 0 {
		config.OrgId = nil
	}
	configJSON, err := json.Marshal(config)
	if err != nil {
		logging.RuntimeLog.Errorf("start domainscan fail:%s", err.Error())
		return "", err
	}
	taskId, err = serverapi.NewRunTask("domainscan", string(configJSON), mainTaskId, "")
	if err != nil {
		logging.RuntimeLog.Errorf("start domainscan fail:%s", err.Error())
		return "", err
	}
	return taskId, nil
}

// doOnlineAPISearch Fofa,hunter,quaker的查询
func doOnlineAPISearch(workspaceId int, mainTaskId string, apiName string, target string, orgId *int, isIplocation, isHttp, isFingerprintHub, isScreenshot, isIconHash, isIgnoreCDN, isIgnorOutofChina bool) (taskId string, err error) {
	config := onlineapi.OnlineAPIConfig{
		Target:             target,
		OrgId:              orgId,
		IsIPLocation:       isIplocation,
		IsHttpx:            isHttp,
		IsFingerprintHub:   isFingerprintHub,
		IsScreenshot:       isScreenshot,
		IsIconHash:         isIconHash,
		IsIgnoreCDN:        isIgnoreCDN,
		IsIgnoreOutofChina: isIgnorOutofChina,
		WorkspaceId:        workspaceId,
	}
	// config.OrgId 为int，默认为0
	// db.Organization.OrgId为指针，默认nil
	if *config.OrgId == 0 {
		config.OrgId = nil
	}
	configJSON, err := json.Marshal(config)
	if err != nil {
		logging.RuntimeLog.Errorf("start %s fail:%s", apiName, err.Error())
		return "", err
	}
	taskId, err = serverapi.NewRunTask(apiName, string(configJSON), mainTaskId, "")
	if err != nil {
		logging.RuntimeLog.Errorf("start %s fail:%s", apiName, err.Error())
		return "", err
	}
	return taskId, nil
}

// doICPQuery ICP备案信息查询
func doICPQuery(mainTaskId string, target string) (taskId string, err error) {
	config := onlineapi.ICPQueryConfig{Target: target}
	configJSON, err := json.Marshal(config)
	if err != nil {
		logging.RuntimeLog.Errorf("start icpquery fail:%s", err.Error())
		return "", err
	}
	taskId, err = serverapi.NewRunTask("icpquery", string(configJSON), mainTaskId, "")
	if err != nil {
		logging.RuntimeLog.Errorf("start icpquery fail:%s", err.Error())
		return "", err
	}
	return taskId, nil
}

// doWhoisQuery Whois信息查询
func doWhoisQuery(mainTaskId string, target string) (taskId string, err error) {
	config := onlineapi.WhoisQueryConfig{Target: target}
	configJSON, err := json.Marshal(config)
	if err != nil {
		logging.RuntimeLog.Errorf("start whoisquery fail:%s", err.Error())
		return "", err
	}
	taskId, err = serverapi.NewRunTask("whoisquery", string(configJSON), mainTaskId, "")
	if err != nil {
		logging.RuntimeLog.Errorf("start whoisquery fail:%s", err.Error())
		return "", err
	}
	return taskId, nil
}

// doIPLocation IP归属地
func doIPLocation(mainTaskId string, target string, orgId *int) (taskId string, err error) {
	config := custom.Config{Target: target, OrgId: orgId}
	// config.OrgId 为int，默认为0
	// db.Organization.OrgId为指针，默认nil
	if *config.OrgId == 0 {
		config.OrgId = nil
	}
	configJSON, err := json.Marshal(config)
	if err != nil {
		logging.RuntimeLog.Errorf("start portscan fail:%s", err.Error())
		return "", err
	}
	taskId, err = serverapi.NewRunTask("iplocation", string(configJSON), mainTaskId, "")
	if err != nil {
		logging.RuntimeLog.Errorf("start iplocation fail:%s", err.Error())
		return "", err
	}
	return taskId, nil
}

// getDomainFLD 提取域名的FLD
func getDomainFLD(domainTargetList []string) (fldDomain []string) {
	domains := make(map[string]struct{})
	tld := domainscan.NewTldExtract()
	for _, domain := range domainTargetList {
		fld := tld.ExtractFLD(domain)
		if fld == "" {
			continue
		}
		if _, ok := domains[fld]; !ok {
			domains[fld] = struct{}{}
		}
	}
	fldDomain = utils.SetToSlice(domains)
	return
}

// formatIpTarget 将从web端传入的ip参数（以\n分隔）转换为ip列表，对域名进行解析转换为，并保存域名及A记录到数据库中
func formatIpTarget(target string, orgId int) (ipTargetList []string) {
	for _, t := range strings.Split(target, "\n") {
		if tt := strings.TrimSpace(t); tt != "" {
			//192.168.1.1  192.168.1.0/24
			if utils.CheckIPV4(tt) || utils.CheckIPV4Subnet(tt) {
				ipTargetList = append(ipTargetList, tt)
				continue
			}
			//192.168.1.1-192.168.1.5
			address := strings.Split(tt, "-")
			if len(address) == 2 && utils.CheckIPV4(address[0]) && utils.CheckIPV4(address[1]) {
				ipTargetList = append(ipTargetList, tt)
				continue
			}
			//域名，将域名转成ip地址
			_, hosts := domainscan.ResolveDomain(tt)
			if len(hosts) > 0 {
				domainResult := domainscan.Result{DomainResult: make(map[string]*domainscan.DomainResult)}
				domainResult.SetDomain(tt)
				for _, h := range hosts {
					ipTargetList = append(ipTargetList, h)
					domainResult.SetDomainAttr(tt, domainscan.DomainAttrResult{
						Source:  "portscan",
						Tag:     "A",
						Content: h,
					})
				}
				config := domainscan.Config{OrgId: &orgId}
				// config.OrgId 为int，默认为0
				// db.Organization.OrgId为指针，默认nil
				if *config.OrgId == 0 {
					config.OrgId = nil
				}
				domainResult.SaveResult(config)
			}
		}
	}

	return
}

// formatDomainTarget 将前端web的域名，转换为列表；同时去除非域名的IP地址
func formatDomainTarget(target string) (domainTargetList []string) {
	for _, t := range strings.Split(target, "\n") {
		if tt := strings.TrimSpace(t); tt != "" {
			//192.168.1.1  192.168.1.0/24
			if utils.CheckIPV4(tt) || utils.CheckIPV4Subnet(tt) {
				continue
			}
			//192.168.1.1-192.168.1.5
			address := strings.Split(tt, "-")
			if len(address) == 2 && utils.CheckIPV4(address[0]) && utils.CheckIPV4(address[1]) {
				continue
			}
			domainTargetList = append(domainTargetList, tt)
		}
	}
	return
}

// ParseTargetFromKwArgs 从经过JSON序列化的参数中单独提取出target
func ParseTargetFromKwArgs(taskName, args string) (target string) {
	const displayedLength = 100
	type TargetStrut struct {
		Target string `json:"target"`
	}
	type FingerTargetStrut struct {
		IPTargetMap     *map[string][]int    `json:"IPTargetMap"`
		DomainTargetMap *map[string]struct{} `json:"DomainTargetMap"`
	}
	type XrayPocStrut struct {
		IPPortResult map[string][]int
		DomainResult []string
	}
	type XScanConfig struct {
		OrgId        *int                `json:"orgid"`
		FofaTarget   string              `json:"fofatarget"`
		FofaKeyword  string              `json:"fofaKeyword"`
		IPPort       map[string][]int    `json:"ipport"`
		IPPortString map[string]string   `json:"ipportstring"`
		Domain       map[string]struct{} `json:"domain"`
		Target       string              `json:"target"`
	}
	if taskName == "fingerprint" {
		var t FingerTargetStrut
		err := json.Unmarshal([]byte(args), &t)
		if err != nil {
			target = args
		} else {
			var allTarget []string
			if t.IPTargetMap != nil {
				for ip := range *t.IPTargetMap {
					allTarget = append(allTarget, ip)
				}
			}
			if t.DomainTargetMap != nil {
				for domain := range *t.DomainTargetMap {
					allTarget = append(allTarget, domain)
				}
			}
			target = strings.Join(allTarget, ",")
		}
	} else if taskName == "xraypoc" {
		var t XrayPocStrut
		err := json.Unmarshal([]byte(args), &t)
		if err != nil {
			target = args
		} else {
			var allTarget []string
			for ip, ports := range t.IPPortResult {
				for _, port := range ports {
					allTarget = append(allTarget, fmt.Sprintf("%s:%d", ip, port))
				}
			}
			for _, domain := range t.DomainResult {
				allTarget = append(allTarget, domain)
			}
			target = strings.Join(allTarget, ",")
		}
	} else if taskName == "xportscan" || taskName == "xdomainscan" || taskName == "xfofa" || taskName == "xxraypoc" || taskName == "xxray" || taskName == "xfingerprint" || taskName == "xorgscan" {
		var t XScanConfig
		err := json.Unmarshal([]byte(args), &t)
		if err != nil {
			target = args
		} else {
			var allTarget []string
			if len(t.FofaKeyword) > 0 {
				allTarget = append(allTarget, t.FofaKeyword)
			}
			if len(t.FofaTarget) > 0 {
				allTarget = append(allTarget, t.FofaTarget)
			}
			if len(t.IPPort) > 0 {
				for tip := range t.IPPort {
					allTarget = append(allTarget, tip)
				}
			}
			if len(t.IPPortString) > 0 {
				for tip := range t.IPPortString {
					allTarget = append(allTarget, tip)
				}
			}
			if len(t.Domain) > 0 {
				for td := range t.Domain {
					allTarget = append(allTarget, td)
				}
			}
			if len(t.Target) > 0 {
				allTarget = append(allTarget, t.Target)
			}
			if taskName == "xorgscan" {
				orgDb := db.Organization{Id: *t.OrgId}
				if orgDb.Get() {
					allTarget = append(allTarget, orgDb.OrgName)
				}
			}
			target = strings.Join(allTarget, ",")
		}
	} else {
		var t TargetStrut
		err := json.Unmarshal([]byte(args), &t)
		if err != nil {
			target = args
		} else {
			target = t.Target
		}
	}
	if len(target) > displayedLength {
		return fmt.Sprintf("%s...", target[:displayedLength])
	}

	return
}

func addGlobalFilterWord(rule string) string {
	// 从custom目录中读取定义的过滤词，每一个关键词一行：
	filterFile := filepath.Join(conf.GetRootPath(), "thirdparty/custom", "fofa_filter_keyword.txt")
	if utils.CheckFileExist(filterFile) == false {
		logging.RuntimeLog.Errorf("fofa filter file not exist:%s", filterFile)
		return rule
	}
	inputFile, err := os.Open(filterFile)
	if err != nil {
		logging.RuntimeLog.Errorf("Could not read fofa filter file: %s\n", err)
		return rule
	}
	defer inputFile.Close()
	scanner := bufio.NewScanner(inputFile)
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if text == "" {
			continue
		}
		rule = rule + "&& body !=\"" + text + "\""
	}
	//globalFilterWords := strings.Split(GlobalFilterWords, "||")
	//if len(globalFilterWords) > 0 {
	//	for _, globalFilterWord := range globalFilterWords {
	//		rule = rule + "&& body !=\"" + globalFilterWord + "\""
	//	}
	//}
	return rule
}

func addFoFaSearchRule(searchkey taskKeySearchParam) []string {
	defaultCheckMod := "title"
	var rules []string

	if searchkey.CheckMod == "" {
		searchkey.CheckMod = defaultCheckMod
	}

	CheckMods := strings.Split(searchkey.CheckMod, "&&")
	for _, checkMod := range CheckMods {
		rule := ""
		if checkMod != "" {
			if checkMod == "self" {
				rule = searchkey.KeyWord
			} else {
				rule = searchkey.CheckMod + "=\"" + searchkey.KeyWord + "\""
			}
		}
		//添加反向匹配词
		if searchkey.ExcludeWords != "" {
			exclude_words := strings.Split(searchkey.ExcludeWords, "||")
			for _, exclude_word := range exclude_words {
				rule += " && body!=\"" + exclude_word + "\""
			}
		}
		//是否大陆地区
		if searchkey.IsCN {
			rule += "&& country=\"CN\" && region!=\"HK\" && region!=\"TW\"  && region!=\"MO\""
		}
		rule = addGlobalFilterWord(rule)

		//判断检索日期
		if searchkey.SearchTime == "" {

		} else if searchkey.SearchTime == time.Now().Format("2006-01-02") {
			break
		} else {
			rule += " && after=\"" + searchkey.SearchTime + "\""
		}
		rules = append(rules, rule)
	}

	return rules
}

func searchKeyword(req XScanRequestParam) (fofaKeyword map[string]int) {
	fofaKeyword = make(map[string]int)
	keyWords := db.KeyWord{}
	//传入org_id
	searchMap := make(map[string]interface{})
	if req.OrgId > 0 {
		searchMap["org_id"] = req.OrgId
	}
	results, _ := keyWords.Gets(searchMap, 0, 99999)
	//fofa检索词拼接
	for _, searchkey := range results {
		taskSearchKey := taskKeySearchParam{}
		taskSearchKey.IsCN = req.IsCn
		taskSearchKey.KeyWord = searchkey.KeyWord
		taskSearchKey.ExcludeWords = searchkey.ExcludeWords
		taskSearchKey.ExcludeWords = searchkey.ExcludeWords
		taskSearchKey.SearchTime = searchkey.SearchTime
		taskSearchKey.CheckMod = searchkey.CheckMod
		rules := addFoFaSearchRule(taskSearchKey)
		for _, rule := range rules {
			fofaKeyword[rule] = searchkey.Count
		}
		kw := db.KeyWord{Id: searchkey.Id}
		updateMap := make(map[string]interface{})
		updateMap["search_time"] = time.Now().Format("2006-01-02")
		kw.Update(updateMap)
	}
	return
}
