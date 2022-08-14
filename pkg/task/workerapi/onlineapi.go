package workerapi

import (
	"context"
	"github.com/hanc00l/nemo_go/pkg/comm"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/custom"
	"github.com/hanc00l/nemo_go/pkg/task/domainscan"
	"github.com/hanc00l/nemo_go/pkg/task/onlineapi"
	"github.com/hanc00l/nemo_go/pkg/task/portscan"
	"github.com/hanc00l/nemo_go/pkg/utils"
)

// Fofa Fofa任务
func Fofa(taskId, configJSON string) (result string, err error) {
	var ok bool
	if ok, result, err = CheckTaskStatus(taskId); !ok {
		return result, err
	}

	config := onlineapi.OnlineAPIConfig{}
	if err = ParseConfig(configJSON, &config); err != nil {
		return FailedTask(err.Error()), err
	}

	fofa := onlineapi.NewFofa(config)
	fofa.Do()
	checkIgnoreResult(&fofa.IpResult, &fofa.DomainResult, config)
	if config.IsIPLocation {
		doLocation(&fofa.IpResult)
	}
	result, err = doFingerAndSave(taskId, &fofa.IpResult, &fofa.DomainResult, config)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return FailedTask(err.Error()), err
	}

	return SucceedTask(result), nil
}

// Quake Quake任务
func Quake(taskId, configJSON string) (result string, err error) {
	var ok bool
	if ok, result, err = CheckTaskStatus(taskId); !ok {
		return result, err
	}

	config := onlineapi.OnlineAPIConfig{}
	if err = ParseConfig(configJSON, &config); err != nil {
		return FailedTask(err.Error()), err
	}

	quake := onlineapi.NewQuake(config)
	quake.Do()
	checkIgnoreResult(&quake.IpResult, &quake.DomainResult, config)
	if config.IsIPLocation {
		doLocation(&quake.IpResult)
	}
	result, err = doFingerAndSave(taskId, &quake.IpResult, &quake.DomainResult, config)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return FailedTask(err.Error()), err
	}

	return SucceedTask(result), nil
}

// Hunter Hunter
func Hunter(taskId, configJSON string) (result string, err error) {
	var ok bool
	if ok, result, err = CheckTaskStatus(taskId); !ok {
		return result, err
	}

	config := onlineapi.OnlineAPIConfig{}
	if err = ParseConfig(configJSON, &config); err != nil {
		return FailedTask(err.Error()), err
	}

	hunter := onlineapi.NewHunter(config)
	hunter.Do()
	checkIgnoreResult(&hunter.IpResult, &hunter.DomainResult, config)
	if config.IsIPLocation {
		doLocation(&hunter.IpResult)
	}
	result, err = doFingerAndSave(taskId, &hunter.IpResult, &hunter.DomainResult, config)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return FailedTask(err.Error()), err
	}

	return SucceedTask(result), nil
}

// doFingerAndSave 主动探测、指纹识别及保存结果
func doFingerAndSave(taskId string, portScanResult *portscan.Result, domainScanResult *domainscan.Result, config onlineapi.OnlineAPIConfig) (result string, err error) {
	//指纹识别：
	if len(portScanResult.IPResult) > 0 {
		portscanConfig := portscan.Config{
			IsHttpx:          config.IsHttpx,
			IsWhatWeb:        config.IsWhatWeb,
			IsFingerprintHub: config.IsFingerprintHub,
			IsIconHash:       config.IsIconHash,
		}
		DoIPFingerPrint(portscanConfig, portScanResult)
		if config.IsScreenshot {
			DoScreenshotAndSave(portScanResult, nil)
		}
	}
	if len(domainScanResult.DomainResult) > 0 {
		domainscanConfig := domainscan.Config{
			IsHttpx:          config.IsHttpx,
			IsWhatWeb:        config.IsWhatWeb,
			IsFingerprintHub: config.IsFingerprintHub,
			IsIconHash:       config.IsIconHash,
		}
		DoDomainFingerPrint(domainscanConfig, domainScanResult)
		if config.IsScreenshot {
			DoScreenshotAndSave(nil, domainScanResult)
		}
	}
	// 保存结果
	x := comm.NewXClient()
	args := comm.ScanResultArgs{
		TaskID:       taskId,
		IPConfig:     &portscan.Config{OrgId: config.OrgId},
		DomainConfig: &domainscan.Config{OrgId: config.OrgId},
		IPResult:     portScanResult.IPResult,
		DomainResult: domainScanResult.DomainResult,
	}
	err = x.Call(context.Background(), "SaveScanResult", &args, &result)

	return result, err
}

// ICPQuery ICP备案查询任务
func ICPQuery(taskId, configJSON string) (result string, err error) {
	var ok bool
	if ok, result, err = CheckTaskStatus(taskId); !ok {
		return result, err
	}

	config := onlineapi.ICPQueryConfig{}
	if err = ParseConfig(configJSON, &config); err != nil {
		return FailedTask(err.Error()), err
	}

	icp := onlineapi.NewICPQuery(config)
	icp.Do()
	// 保存结果
	x := comm.NewXClient()

	err = x.Call(context.Background(), "SaveICPResult", &icp.QueriedICPInfo, &result)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return FailedTask(err.Error()), err
	}

	return SucceedTask(result), nil
}

// WhoisQuery Whois查询任务
func WhoisQuery(taskId, configJSON string) (result string, err error) {
	var ok bool
	if ok, result, err = CheckTaskStatus(taskId); !ok {
		return result, err
	}

	config := onlineapi.WhoisQueryConfig{}
	if err = ParseConfig(configJSON, &config); err != nil {
		return FailedTask(err.Error()), err
	}

	whois := onlineapi.NewWhois(config)
	whois.Do()
	// 保存结果
	x := comm.NewXClient()

	err = x.Call(context.Background(), "SaveWhoisResult", &whois.QueriedWhoisInfo, &result)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return FailedTask(err.Error()), err
	}

	return SucceedTask(result), nil
}

// checkIgnoreResult 检查资产查询API中的IP资产，非中国IP或CDN，则不保存该结果
func checkIgnoreResult(portScanResult *portscan.Result, domainScanResult *domainscan.Result, config onlineapi.OnlineAPIConfig) {
	iplocation := custom.NewIPLocation()
	cdnCheck := custom.NewCDNCheck()
	if len(portScanResult.IPResult) > 0 && (config.IsIgnoreOutofChina || config.IsIgnoreCDN) {
		for ip := range portScanResult.IPResult {
			ipl := iplocation.FindPublicIP(ip)
			if config.IsIgnoreOutofChina && utils.CheckIPLocationInChinaMainLand(ipl) == false {
				delete(portScanResult.IPResult, ip)
				continue
			}
			if config.IsIgnoreCDN && (cdnCheck.CheckIP(ip) || cdnCheck.CheckASN(ip)) {
				delete(portScanResult.IPResult, ip)
			}
		}
	}
	/*
		if len(domainScanResult.DomainResult) > 0 && config.IsIgnoreCDN {
			for domain := range domainScanResult.DomainResult {
				iscdn, _, _ := cdnCheck.CheckCName(domain)
				if iscdn {
					delete(domainScanResult.DomainResult, domain)
				}
			}
		}*/
}
