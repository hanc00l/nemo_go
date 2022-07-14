package runner

import (
	"encoding/json"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/custom"
	"github.com/hanc00l/nemo_go/pkg/task/domainscan"
	"github.com/hanc00l/nemo_go/pkg/task/onlineapi"
	"github.com/hanc00l/nemo_go/pkg/task/pocscan"
	"github.com/hanc00l/nemo_go/pkg/task/portscan"
	"github.com/hanc00l/nemo_go/pkg/task/serverapi"
	"github.com/hanc00l/nemo_go/pkg/utils"
	"strings"
)

// StartPortScanTask 端口扫描任务
func StartPortScanTask(req PortscanRequestParam, cronTaskId string) (taskId string, err error) {
	// 解析参数
	ts := utils.NewTaskSlice()
	ts.TaskMode = req.TaskMode
	ts.IpTarget = req.Target
	ts.Port = req.Port
	tc := conf.GlobalServerConfig().Task
	ts.IpSliceNumber = tc.IpSliceNumber
	ts.PortSliceNumber = tc.PortSliceNumber
	targets, ports := ts.DoIpSlice()
	for _, t := range targets {
		for _, p := range ports {
			// 端口扫描
			if req.IsPortScan {
				if taskId, err = doPortscan(cronTaskId, t, p, req); err != nil {
					return
				}
			}
			// IP归属地：如果有端口执行任务，则IP归属地任务在端口扫描中执行，否则单独执行
			if !req.IsPortScan && req.IsIPLocation {
				if taskId, err = doIPLocation(cronTaskId, t, &req.OrgId); err != nil {
					return
				}
			}
			// FOFA
			if req.IsFofa {
				if taskId, err = doOnlineAPISearch(cronTaskId, "fofa", t, &req.OrgId, req.IsIPLocation, req.IsHttpx, req.IsWappalyzer, req.IsFingerprintHub, req.IsScreenshot, req.IsIconHash); err != nil {
					return
				}
			}
			// Quake
			if req.IsQuake {
				if taskId, err = doOnlineAPISearch(cronTaskId, "quake", t, &req.OrgId, req.IsIPLocation, req.IsHttpx, req.IsWappalyzer, req.IsFingerprintHub, req.IsScreenshot, req.IsIconHash); err != nil {
					return
				}
			}
			// Hunter
			if req.IsHunter {
				if taskId, err = doOnlineAPISearch(cronTaskId, "hunter", t, &req.OrgId, req.IsIPLocation, req.IsHttpx, req.IsWappalyzer, req.IsFingerprintHub, req.IsScreenshot, req.IsIconHash); err != nil {
					return
				}
			}
		}
	}
	return taskId, nil
}

// StartBatchScanTask 探测+扫描任务
func StartBatchScanTask(req PortscanRequestParam, cronTaskId string) (taskId string, err error) {
	ts := utils.NewTaskSlice()
	ts.TaskMode = req.TaskMode
	ts.IpTarget = req.Target
	ts.Port = req.Port
	tc := conf.GlobalServerConfig().Task
	ts.IpSliceNumber = tc.IpSliceNumber
	ts.PortSliceNumber = tc.PortSliceNumber
	targets, ports := ts.DoIpSlice()
	for _, t := range targets {
		for _, p := range ports {
			// 端口扫描
			if taskId, err = doBatchScan(cronTaskId, t, p, req); err != nil {
				return
			}
		}
	}
	return taskId, nil
}

// StartDomainScanTask 域名任务
func StartDomainScanTask(req DomainscanRequestParam, cronTaskId string) (taskId string, err error) {
	allTarget := req.Target
	// 域名的FLD
	if req.IsFldDomain {
		fldList := getDomainFLD(req.Target)
		if len(fldList) > 0 {
			allTarget = req.Target + "," + strings.Join(fldList, ",")
		}
	}
	ts := utils.NewTaskSlice()
	ts.TaskMode = req.TaskMode
	ts.DomainTarget = allTarget
	targets := ts.DoDomainSlice()
	for _, t := range targets {
		// 每个获取子域名的方式采用独立任务，以提高速度
		var taskStarted bool
		if req.IsSubfinder {
			subConfig := req
			subConfig.IsSubdomainBrute = false
			subConfig.IsCrawler = false
			if taskId, err = doDomainscan(cronTaskId, t, subConfig); err != nil {
				return
			}
			taskStarted = true
		}
		if req.IsSubdomainBrute {
			subConfig := req
			subConfig.IsSubfinder = false
			subConfig.IsCrawler = false
			if taskId, err = doDomainscan(cronTaskId, t, subConfig); err != nil {
				return
			}
			taskStarted = true
		}
		if req.IsCrawler {
			subConfig := req
			subConfig.IsSubfinder = false
			subConfig.IsSubdomainBrute = false
			if taskId, err = doDomainscan(cronTaskId, t, subConfig); err != nil {
				return
			}
			taskStarted = true
		}
		// 如果没有子域名任务，则至少启动一个域名解析任务
		if !taskStarted {
			if taskId, err = doDomainscan(cronTaskId, t, req); err != nil {
				return
			}
		}
		if req.IsFofa {
			if taskId, err = doOnlineAPISearch(cronTaskId, "fofa", t, &req.OrgId, true, req.IsHttpx, req.IsWappalyzer, req.IsFingerprintHub, req.IsScreenshot, req.IsIconHash); err != nil {
				return
			}
		}
		if req.IsQuake {
			if taskId, err = doOnlineAPISearch(cronTaskId, "quake", t, &req.OrgId, true, req.IsHttpx, req.IsWappalyzer, req.IsFingerprintHub, req.IsScreenshot, req.IsIconHash); err != nil {
				return
			}
		}
		if req.IsHunter {
			if taskId, err = doOnlineAPISearch(cronTaskId, "hunter", t, &req.OrgId, true, req.IsHttpx, req.IsWappalyzer, req.IsFingerprintHub, req.IsScreenshot, req.IsIconHash); err != nil {
				return
			}
		}
		if req.IsICPQuery {
			if taskId, err = doICPQuery(cronTaskId, t); err != nil {
				return
			}
		}
		if req.IsWhoisQuery {
			if taskId, err = doWhoisQuery(cronTaskId, t); err != nil {
				return
			}
		}
	}
	return taskId, nil
}

// StartPocScanTask pocscan任务
func StartPocScanTask(req PocscanRequestParam, cronTaskId string) (taskId string, err error) {
	var targetList []string
	for _, t := range strings.Split(req.Target, "\n") {
		if tt := strings.TrimSpace(t); tt != "" {
			targetList = append(targetList, tt)
		}
	}
	if req.IsPocsuiteVerify && req.PocsuitePocFile != "" {
		config := pocscan.Config{Target: strings.Join(targetList, ","), PocFile: req.PocsuitePocFile, CmdBin: "pocsuite", LoadOpenedPort: req.LoadOpenedPort}
		configJSON, _ := json.Marshal(config)
		taskId, err = serverapi.NewTask("pocsuite", string(configJSON), cronTaskId)
		if err != nil {
			return
		}
	}
	if req.IsXrayVerify && req.XrayPocFile != "" {
		config := pocscan.Config{Target: strings.Join(targetList, ","), PocFile: req.XrayPocFile, CmdBin: "xray", LoadOpenedPort: req.LoadOpenedPort}
		configJSON, _ := json.Marshal(config)
		taskId, err = serverapi.NewTask("xray", string(configJSON), cronTaskId)
		if err != nil {
			return
		}
	}
	if req.IsNucleiVerify && req.NucleiPocFile != "" {
		config := pocscan.Config{Target: strings.Join(targetList, ","), PocFile: req.NucleiPocFile, CmdBin: "nuclei", LoadOpenedPort: req.LoadOpenedPort}
		configJSON, _ := json.Marshal(config)
		taskId, err = serverapi.NewTask("nuclei", string(configJSON), cronTaskId)
		if err != nil {
			return
		}
	}
	if req.IsDirsearch && req.DirsearchExtName != "" {
		config := pocscan.Config{Target: strings.Join(targetList, ","), PocFile: req.DirsearchExtName, CmdBin: "dirsearch", LoadOpenedPort: req.LoadOpenedPort}
		configJSON, _ := json.Marshal(config)
		taskId, err = serverapi.NewTask("dirsearch", string(configJSON), cronTaskId)
		if err != nil {
			return
		}
	}
	return taskId, nil
}

// doPortscan 端口扫描
func doPortscan(cronTaskId string, target string, port string, req PortscanRequestParam) (taskId string, err error) {
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
		IsWhatWeb:        req.IsWhatweb,
		IsScreenshot:     req.IsScreenshot,
		IsWappalyzer:     req.IsWappalyzer,
		IsFingerprintHub: req.IsFingerprintHub,
		IsIconHash:       req.IsIconHash,
		CmdBin:           req.CmdBin,
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
	taskId, err = serverapi.NewTask("portscan", string(configJSON), cronTaskId)
	if err != nil {
		logging.RuntimeLog.Errorf("start portscan fail:%s", err.Error())
		return "", err
	}
	return taskId, nil
}

// doBatchScan 探测+端口扫描
func doBatchScan(cronTaskId string, target string, port string, req PortscanRequestParam) (taskId string, err error) {
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
		IsWhatWeb:        req.IsWhatweb,
		IsScreenshot:     req.IsScreenshot,
		IsWappalyzer:     req.IsWappalyzer,
		IsFingerprintHub: req.IsFingerprintHub,
		IsIconHash:       req.IsIconHash,
		CmdBin:           "masscan",
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
	taskId, err = serverapi.NewTask("batchscan", string(configJSON), cronTaskId)
	if err != nil {
		logging.RuntimeLog.Errorf("start batchscan fail:%s", err.Error())
		return "", err
	}
	return taskId, nil
}

// doDomainscan 域名任务
func doDomainscan(cronTaskId string, target string, req DomainscanRequestParam) (taskId string, err error) {
	config := domainscan.Config{
		Target:             target,
		OrgId:              &req.OrgId,
		IsSubDomainFinder:  req.IsSubfinder,
		IsSubDomainBrute:   req.IsSubdomainBrute,
		IsCrawler:          req.IsCrawler,
		IsHttpx:            req.IsHttpx,
		IsWhatWeb:          req.IsWhatweb,
		IsIPPortScan:       req.IsIPPortscan,
		IsIPSubnetPortScan: req.IsSubnetPortscan,
		IsScreenshot:       req.IsScreenshot,
		IsWappalyzer:       req.IsWappalyzer,
		IsFingerprintHub:   req.IsFingerprintHub,
		IsIconHash:         req.IsIconHash,
		PortTaskMode:       req.PortTaskMode,
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
	taskId, err = serverapi.NewTask("domainscan", string(configJSON), cronTaskId)
	if err != nil {
		logging.RuntimeLog.Errorf("start domainscan fail:%s", err.Error())
		return "", err
	}
	return taskId, nil
}

// doOnlineAPISearch Fofa,hunter,quaker的查询
func doOnlineAPISearch(cronTaskId string, apiName string, target string, orgId *int, isIplocation, isHttp, isWappalyzer, isFingerprintHub, isScreenshot bool, isIconHash bool) (taskId string, err error) {
	config := onlineapi.OnlineAPIConfig{
		Target:           target,
		OrgId:            orgId,
		IsIPLocation:     isIplocation,
		IsHttpx:          isHttp,
		IsWappalyzer:     isWappalyzer,
		IsFingerprintHub: isFingerprintHub,
		IsScreenshot:     isScreenshot,
		IsIconHash:       isIconHash,
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
	taskId, err = serverapi.NewTask(apiName, string(configJSON), cronTaskId)
	if err != nil {
		logging.RuntimeLog.Errorf("start %s fail:%s", apiName, err.Error())
		return "", err
	}
	return taskId, nil
}

// doICPQuery ICP备案信息查询
func doICPQuery(cronTaskId string, target string) (taskId string, err error) {
	config := onlineapi.ICPQueryConfig{Target: target}
	configJSON, err := json.Marshal(config)
	if err != nil {
		logging.RuntimeLog.Errorf("start icpquery fail:%s", err.Error())
		return "", err
	}
	taskId, err = serverapi.NewTask("icpquery", string(configJSON), cronTaskId)
	if err != nil {
		logging.RuntimeLog.Errorf("start icpquery fail:%s", err.Error())
		return "", err
	}
	return taskId, nil
}

// doWhoisQuery Whois信息查询
func doWhoisQuery(cronTaskId string, target string) (taskId string, err error) {
	config := onlineapi.WhoisQueryConfig{Target: target}
	configJSON, err := json.Marshal(config)
	if err != nil {
		logging.RuntimeLog.Errorf("start whoisquery fail:%s", err.Error())
		return "", err
	}
	taskId, err = serverapi.NewTask("whoisquery", string(configJSON), cronTaskId)
	if err != nil {
		logging.RuntimeLog.Errorf("start whoisquery fail:%s", err.Error())
		return "", err
	}
	return taskId, nil
}

// doIPLocation IP归属地
func doIPLocation(cronTaskId string, target string, orgId *int) (taskId string, err error) {
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
	taskId, err = serverapi.NewTask("iplocation", string(configJSON), cronTaskId)
	if err != nil {
		logging.RuntimeLog.Errorf("start iplocation fail:%s", err.Error())
		return "", err
	}
	return taskId, nil
}

// getDomainFLD 提取域名的FLD
func getDomainFLD(target string) (fldDomain []string) {
	domains := make(map[string]struct{})
	tld := domainscan.NewTldExtract()
	for _, t := range strings.Split(target, "\n") {
		domain := strings.TrimSpace(t)
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
