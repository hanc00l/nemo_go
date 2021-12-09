package workerapi

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hanc00l/nemo_go/pkg/comm"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/domainscan"
	"github.com/hanc00l/nemo_go/pkg/task/portscan"
	"github.com/hanc00l/nemo_go/pkg/task/serverapi"
	"github.com/hanc00l/nemo_go/pkg/utils"
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
	DoDomainFingerPrint(config, &resultDomainScan)
	// 保存结果
	x := comm.NewXClient()
	resultArgs := comm.ScanResultArgs{
		TaskID:       taskId,
		DomainConfig: &config,
		DomainResult: resultDomainScan.DomainResult,
	}
	err = x.Call(context.Background(), "SaveScanResult", &resultArgs, &result)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return FailedTask(err.Error()), err
	}
	// Screenshot
	if config.IsScreenshot {
		result2 := DoScreenshotAndSave(nil, &resultDomainScan)
		result = strings.Join([]string{result, result2}, ",")
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

// doPortScan 对IP进行端口扫描
func doPortScan(config domainscan.Config, resultDomainScan *domainscan.Result) {
	ipResult, ipSubnetResult := getResultIPList(resultDomainScan)
	if ipResult == "" {
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
				IsWhatWeb:        config.IsWhatWeb,
				IsScreenshot:     config.IsScreenshot,
				IsFingerprintHub: config.IsFingerprintHub,
			}
			configPortScanJSON, _ := json.Marshal(configPortScan)
			serverapi.NewTask("portscan", string(configPortScanJSON))
		}
	}
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

	return strings.Join(ipList, "\n"), strings.Join(ipSubnetList, "\n")
}
