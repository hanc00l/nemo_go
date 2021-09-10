package workerapi

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hanc00l/nemo_go/pkg/comm"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/domainscan"
	"github.com/hanc00l/nemo_go/pkg/task/fingerprint"
	"github.com/hanc00l/nemo_go/pkg/task/portscan"
	"github.com/hanc00l/nemo_go/pkg/task/serverapi"
	"strings"
)

// DomainScan 域名任务
func DomainScan(taskId, configJSON string) (result string, err error) {
	var ok bool
	if ok, result, err = CheckTaskStatus(taskId); !ok {
		return result, err
	}

	config := domainscan.Config{}
	if err = ParseConfig(configJSON, &config); err != nil {
		return FailedTask(err.Error()), err
	}
	resultDomainScan := doDomainScan(config)
	// 如果有端口扫描的选项
	if config.IsIPPortScan || config.IsIPSubnetPortScan {
		doPortScan(config, &resultDomainScan)
	}
	// 指纹识别
	doFingerPrint(config, &resultDomainScan)
	// 保存结果
	x := comm.NewXClient()

	resultArgs := comm.ScanResultArgs{
		DomainConfig: &config,
		DomainResult: resultDomainScan.DomainResult,
	}
	err = x.Call(context.Background(), "SaveScanResult", &resultArgs, &result)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return FailedTask(err.Error()), err
	}
	if config.IsScreenshot {
		ss := fingerprint.NewScreenShot()
		ss.ResultDomainScan = resultDomainScan
		ss.Do()
		resultScreenshot := ss.LoadResult()
		var result2 string
		err = x.Call(context.Background(), "SaveScreenshotResult", &resultScreenshot, &result2)
		if err != nil {
			logging.RuntimeLog.Error(err)
			return FailedTask(err.Error()), err
		} else {
			result = strings.Join([]string{result, result2}, ",")
		}
	}

	return SucceedTask(result), nil
}

// doDomainScan 域名收集任务
func doDomainScan(config domainscan.Config) (resultDomainScan domainscan.Result) {
	// 子域名枚举
	if config.IsSubDomainFinder {
		subdomain := domainscan.NewSubFinder(config)
		subdomain.Do()
		resultDomainScan = subdomain.Result
	}
	// 子域名爆破
	if config.IsSubDomainBrute {
		massdns := domainscan.NewMassdns(config)
		massdns.Do()
		resultDomainScan = massdns.Result
	}
	//  JSFinder
	if config.IsJSFinder {
		// TODO
	}
	// 域名解析
	resolve := domainscan.NewResolve(config)
	if !config.IsSubDomainFinder && !config.IsSubDomainBrute && !config.IsJSFinder {
		// 对config中Target进行域名解析
		resolve.Do()
		resultDomainScan = resolve.Result
	} else {
		// 对域名任务（子域名枚举或爆破）的结果进行域名解析
		resolve.Result.DomainResult = resultDomainScan.DomainResult
		resolve.Do()
	}

	return resultDomainScan
}

// doFingerPrint 对域名结果进行标题指纹识别
func doFingerPrint(config domainscan.Config, resultDomainScan *domainscan.Result) {
	// 指纹识别
	fpConfig := fingerprint.Config{OrgId: config.OrgId}
	if config.IsHttpx {
		httpx := fingerprint.NewHttpx(fpConfig)
		httpx.ResultDomainScan = *resultDomainScan
		httpx.Do()
	}
	if config.IsWhatWeb {
		whatweb := fingerprint.NewWhatweb(fpConfig)
		whatweb.ResultDomainScan = *resultDomainScan
		whatweb.Do()
	}
	if config.IsWappalyzer {
		wappalyzer := fingerprint.NewWappalyzer()
		wappalyzer.ResultDomainScan = *resultDomainScan
		wappalyzer.Do()
	}
}

// doPortScan 对IP进行端口扫描
func doPortScan(config domainscan.Config, resultDomainScan *domainscan.Result) {
	ipResult, ipSubnetResult := getResultIPList(resultDomainScan)
	if ipResult == "" {
		return
	}
	configPortScan := portscan.Config{
		OrgId:        config.OrgId,
		Target:       ipResult,
		Port:         conf.GlobalWorkerConfig().Portscan.Port,
		Rate:         conf.GlobalWorkerConfig().Portscan.Rate,
		CmdBin:       conf.GlobalWorkerConfig().Portscan.Cmdbin,
		IsPing:       conf.GlobalWorkerConfig().Portscan.IsPing,
		Tech:         conf.GlobalWorkerConfig().Portscan.Tech,
		IsIpLocation: true,
		IsHttpx:      config.IsHttpx,
		IsWhatWeb:    config.IsWhatWeb,
		IsScreenshot: config.IsScreenshot,
	}
	if config.IsIPSubnetPortScan {
		configPortScan.Target = ipSubnetResult
	}
	configPortScanJSON, _ := json.Marshal(configPortScan)
	serverapi.NewTask("portscan", string(configPortScanJSON))
}

// getResultIPList 提取域名收集结果的IP
func getResultIPList(resultDomainScan *domainscan.Result) (ipResult, ipSubnetResult string) {
	ips := make(map[string]struct{})
	ipSubnets := make(map[string]struct{})
	for _, da := range resultDomainScan.DomainResult {
		for _, dar := range da.DomainAttrs {
			if dar.Tag == "A" {
				ipArray := strings.Split(dar.Content, ".")
				if len(ipArray) != 4 {
					continue
				}
				if _, ok := ips[dar.Content]; !ok {
					ips[dar.Content] = struct{}{}
				}
				s := fmt.Sprintf("%s.%s.%s.0/24", ipArray[0], ipArray[1], ipArray[2])
				if _, ok := ipSubnets[s]; !ok {
					ipSubnets[s] = struct{}{}
				}
			}
		}
	}
	var ipList, ipSubnetList []string
	for k, _ := range ips {
		ipList = append(ipList, k)
	}
	for k, _ := range ipSubnets {
		ipSubnetList = append(ipSubnetList, k)
	}

	return strings.Join(ipList, ","), strings.Join(ipSubnetList, ",")
}
