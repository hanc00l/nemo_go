package workerapi

import (
	"encoding/json"
	"fmt"
	"github.com/hanc00l/nemo_go/v2/pkg/comm"
	"github.com/hanc00l/nemo_go/v2/pkg/conf"
	"github.com/hanc00l/nemo_go/v2/pkg/logging"
	"github.com/hanc00l/nemo_go/v2/pkg/task/custom"
	"github.com/hanc00l/nemo_go/v2/pkg/task/domainscan"
	"github.com/hanc00l/nemo_go/v2/pkg/task/portscan"
	"github.com/hanc00l/nemo_go/v2/pkg/utils"
	"net/netip"
	"strings"
)

// DomainScan 域名任务
func DomainScan(taskId, mainTaskId, configJSON string) (result string, err error) {
	var ok bool
	if ok, result, err = CheckTaskStatus(taskId); !ok {
		return result, err
	}

	config := domainscan.Config{}
	if err = ParseConfig(configJSON, &config); err != nil {
		logging.RuntimeLog.Error(err)
		return FailedTask(err.Error()), err
	}
	resultDomainScan := doDomainScan(config)
	// 如果有端口扫描的选项
	if config.IsIPPortScan || config.IsIPSubnetPortScan {
		doPortScanByDomainscan(taskId, mainTaskId, config, resultDomainScan)
	}
	// 保存结果
	resultArgs := comm.ScanResultArgs{
		TaskID:       taskId,
		MainTaskId:   mainTaskId,
		DomainConfig: &config,
		DomainResult: resultDomainScan.DomainResult,
	}
	err = comm.CallXClient("SaveScanResult", &resultArgs, &result)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return FailedTask(err.Error()), err
	}
	_, err = NewFingerprintTask(taskId, mainTaskId, nil, resultDomainScan, FingerprintTaskConfig{
		IsHttpx:          config.IsHttpx,
		IsFingerprintHub: config.IsFingerprintHub,
		IsIconHash:       config.IsIconHash,
		IsScreenshot:     config.IsIconHash,
		IsFingerprintx:   config.IsFingerprintx,
		WorkspaceId:      config.WorkspaceId,
		IsProxy:          config.IsProxy,
	})
	if err != nil {
		logging.RuntimeLog.Error(err)
		return FailedTask(err.Error()), err
	}
	return SucceedTask(result), nil
}

// doDomainScan 域名收集任务
func doDomainScan(config domainscan.Config) (resultDomainScan *domainscan.Result) {
	// 子域名枚举
	if config.IsSubDomainFinder {
		subdomain := domainscan.NewSubFinder(config)
		subdomain.Do()
		resultDomainScan = &subdomain.Result
	}
	// 子域名爆破
	if config.IsSubDomainBrute {
		massdns := domainscan.NewMassdns(config)
		massdns.Do()
		resultDomainScan = &massdns.Result
	}
	//  Crawler
	if config.IsCrawler {
		crawler := domainscan.NewCrawler(config)
		crawler.Do()
		resultDomainScan = &crawler.Result
	}
	// 域名解析
	resolve := domainscan.NewResolve(config)
	if !config.IsSubDomainFinder && !config.IsSubDomainBrute && !config.IsCrawler {
		// 对config中Target进行域名解析
		resolve.Do()
		resultDomainScan = &resolve.Result
	} else {
		// 对域名任务（子域名枚举或爆破）的结果进行域名解析
		resolve.Result.DomainResult = resultDomainScan.DomainResult
		resolve.Do()
	}
	// 去除结果中无域名解析A或CNAME记录的域名
	checkDomainResolveResult(resultDomainScan)
	// 对域名结果中同一个IP对应太多进行过滤
	domainscan.FilterDomainResult(resultDomainScan)

	return resultDomainScan
}

// doPortScanByDomainscan 对IP进行端口扫描
func doPortScanByDomainscan(taskId, mainTaskId string, config domainscan.Config, resultDomainScan *domainscan.Result) {
	ipResult, ipSubnetResult := getResultIPList(resultDomainScan)
	if len(ipResult) == 0 {
		return
	}
	portsConfig := conf.GlobalWorkerConfig().Portscan
	ts := utils.NewTaskSlice()
	ts.TaskMode = config.PortTaskMode
	ts.Port = portsConfig.Port
	ts.IpTarget = ipResult
	if config.IsIPSubnetPortScan {
		ts.IpTarget = ipSubnetResult
	}
	// worker任务执行，只能使用默认配置，无法读取server.yml中的配置
	ts.IpSliceNumber = utils.DefaultIpSliceNumber
	ts.PortSliceNumber = utils.DefaultPortSliceNumber
	targets, ports := ts.DoIpSlice()
	for _, t := range targets {
		for _, p := range ports {
			configPortScan := portscan.Config{
				OrgId:            config.OrgId,
				Target:           t,
				Port:             p,
				Rate:             portsConfig.Rate,
				CmdBin:           portsConfig.Cmdbin,
				IsPing:           portsConfig.IsPing,
				Tech:             portsConfig.Tech,
				IsIpLocation:     true,
				IsHttpx:          config.IsHttpx,
				IsScreenshot:     config.IsScreenshot,
				IsFingerprintHub: config.IsFingerprintHub,
				IsIconHash:       config.IsIconHash,
				IsFingerprintx:   config.IsFingerprintx,
				IsLoadOpenedPort: false, //只扫描当前结果
				IsPortscan:       true,
				WorkspaceId:      config.WorkspaceId,
				IsProxy:          config.IsProxy,
			}
			configPortScanJSON, _ := json.Marshal(configPortScan)
			// 创建端口扫描任务
			newTaskArgs := comm.NewTaskArgs{
				TaskName:      "portscan",
				LastRunTaskId: taskId,
				MainTaskID:    mainTaskId,
				ConfigJSON:    string(configPortScanJSON),
			}
			var result string
			err := comm.CallXClient("NewTask", &newTaskArgs, &result)
			if err != nil {
				logging.RuntimeLog.Error("Start Portscan task fail:", err)
				logging.CLILog.Error("Start Portscan task fail:", err)
			} else {
				logging.CLILog.Info("Start Portscan task...")
			}
		}
	}
}

// getResultIPList 提取域名收集结果的IP
func getResultIPList(resultDomainScan *domainscan.Result) (ipResult, ipSubnetResult []string) {
	ips := make(map[string]struct{})
	ipSubnets := make(map[string]struct{})
	cdnCheck := custom.NewCDNCheck()
	for domain, da := range resultDomainScan.DomainResult {
		if isCdn, _, _ := cdnCheck.CheckCName(domain); isCdn {
			continue
		}
		for _, dar := range da.DomainAttrs {
			if dar.Tag == "A" {
				ipArray := strings.Split(dar.Content, ".")
				if len(ipArray) != 4 {
					continue
				}
				if cdnCheck.CheckIP(dar.Content) || cdnCheck.CheckASN(dar.Content) {
					continue
				}
				if _, ok := ips[dar.Content]; !ok {
					ips[dar.Content] = struct{}{}
				}
				s := fmt.Sprintf("%s.%s.%s.0/24", ipArray[0], ipArray[1], ipArray[2])
				if _, ok := ipSubnets[s]; !ok {
					ipSubnets[s] = struct{}{}
				}
			} else if dar.Tag == "AAAA" {
				_, err := netip.ParseAddr(dar.Tag)
				if err != nil || !utils.CheckIPV6(dar.Content) {
					continue
				}
				if _, ok := ips[dar.Content]; !ok {
					ips[dar.Content] = struct{}{}
				}
				s := utils.GetIPV6SubnetC(dar.Content)
				if _, ok := ipSubnets[s]; !ok {
					ipSubnets[s] = struct{}{}
				}
			}
		}
	}
	for k, _ := range ips {
		ipResult = append(ipResult, k)
	}
	for k, _ := range ipSubnets {
		ipSubnetResult = append(ipSubnetResult, k)
	}
	return
}

// checkDomainResolveResult 检查域名结果，去除没有解析记录的无效域名
func checkDomainResolveResult(resultDomainScan *domainscan.Result) {
	var removedDomain []string
	for domain, domainResult := range resultDomainScan.DomainResult {
		if isHaveResolveRecord(&domainResult.DomainAttrs) == false {
			removedDomain = append(removedDomain, domain)
		}
	}
	for _, domain := range removedDomain {
		delete(resultDomainScan.DomainResult, domain)
	}
}

func isHaveResolveRecord(domainAttrs *[]domainscan.DomainAttrResult) bool {
	if len(*domainAttrs) == 0 {
		return false
	}
	for _, dar := range *domainAttrs {
		if dar.Tag == "A" || dar.Tag == "AAAA" || dar.Tag == "CNAME" {
			return true
		}
	}
	return false
}
