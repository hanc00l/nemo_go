package workerapi

import (
	"context"
	"encoding/json"
	"github.com/hanc00l/nemo_go/pkg/comm"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/domainscan"
	"github.com/hanc00l/nemo_go/pkg/task/fingerprint"
	"github.com/hanc00l/nemo_go/pkg/task/portscan"
	"strings"
)

const (
	// IPNumberPerFingerprintTask 拆分指纹识别子任务的粒度
	IPNumberPerFingerprintTask     = 10
	DomainNumberPerFingerprintTask = 20
)

type FingerprintTaskConfig struct {
	IsHttpx          bool
	IsFingerprintHub bool
	IsIconHash       bool
	IsScreenshot     bool
	IPTargetMap      map[string][]int
	DomainTargetMap  map[string]struct{}
}

// Fingerprint 指纹识别任务，将所有指纹识别任务整合后，可由worker将端口扫描与域名收集后的结果进行二次拆分为多个任务，提高指纹识别效率
func Fingerprint(taskId, configJSON string) (result string, err error) {
	var ok bool
	if ok, result, err = CheckTaskStatus(taskId); !ok {
		return result, err
	}
	config := FingerprintTaskConfig{}
	if err = ParseConfig(configJSON, &config); err != nil {
		return FailedTask(err.Error()), err
	}
	//
	var resultScreenshot string
	resultPortScan := portscan.Result{IPResult: make(map[string]*portscan.IPResult)}
	resultDomainScan := domainscan.Result{DomainResult: make(map[string]*domainscan.DomainResult)}
	//
	resultArgs := comm.ScanResultArgs{
		TaskID: taskId,
	}
	if len(config.IPTargetMap) > 0 {
		for ip, ports := range config.IPTargetMap {
			resultPortScan.SetIP(ip)
			for _, port := range ports {
				resultPortScan.SetPort(ip, port)
			}
		}
		portscanConfig := portscan.Config{
			IsHttpx:          config.IsHttpx,
			IsFingerprintHub: config.IsFingerprintHub,
			IsIconHash:       config.IsIconHash,
		}
		doIPFingerPrint(portscanConfig, &resultPortScan)
		resultArgs.IPConfig = &portscanConfig
		resultArgs.IPResult = resultPortScan.IPResult
	}
	if len(config.DomainTargetMap) > 0 {
		for domain := range config.DomainTargetMap {
			resultDomainScan.SetDomain(domain)
		}
		domainscanConfig := domainscan.Config{
			IsHttpx:          config.IsHttpx,
			IsFingerprintHub: config.IsFingerprintHub,
			IsIconHash:       config.IsIconHash,
		}
		doDomainFingerPrint(domainscanConfig, &resultDomainScan)
		resultArgs.DomainConfig = &domainscanConfig
		resultArgs.DomainResult = resultDomainScan.DomainResult
	}
	// 保存结果
	x := comm.NewXClient()
	err = x.Call(context.Background(), "SaveScanResult", &resultArgs, &result)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return FailedTask(err.Error()), err
	}
	// screenshot任务
	if config.IsScreenshot {
		resultScreenshot = doScreenshotAndSave(&resultPortScan, &resultDomainScan)
	}
	//返回全部结果
	result = strings.Join([]string{result, resultScreenshot}, ",")

	return
}

// doIPFingerPrint 对 IP结果进行指纹识别
func doIPFingerPrint(config portscan.Config, resultPortScan *portscan.Result) {
	if config.IsHttpx {
		httpx := fingerprint.NewHttpx()
		httpx.ResultPortScan = *resultPortScan
		httpx.Do()
	}
	if config.IsFingerprintHub {
		fp := fingerprint.NewFingerprintHub()
		fp.ResultPortScan = *resultPortScan
		fp.Do()
	}
	if config.IsIconHash {
		doIconHashAndSave(resultPortScan, nil)
	}
}

// doDomainFingerPrint 对域名结果进行指纹识别
func doDomainFingerPrint(config domainscan.Config, resultDomainScan *domainscan.Result) {
	// 指纹识别
	if config.IsHttpx {
		httpx := fingerprint.NewHttpx()
		httpx.ResultDomainScan = *resultDomainScan
		httpx.Do()
	}
	if config.IsFingerprintHub {
		fp := fingerprint.NewFingerprintHub()
		fp.ResultDomainScan = *resultDomainScan
		fp.Do()
	}
	if config.IsIconHash {
		doIconHashAndSave(nil, resultDomainScan)
	}
}

// doScreenshotAndSave 执行Screenshot并保存
func doScreenshotAndSave(resultIPScan *portscan.Result, resultDomainScan *domainscan.Result) string {
	ss := fingerprint.NewScreenShot()
	if resultIPScan != nil {
		ss.ResultPortScan = *resultIPScan
	}
	if resultDomainScan != nil {
		ss.ResultDomainScan = *resultDomainScan
	}
	ss.Do()

	resultScreenshot := ss.LoadResult()
	x := comm.NewXClient()
	var result2 string
	err := x.Call(context.Background(), "SaveScreenshotResult", &resultScreenshot, &result2)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return err.Error()
	}
	return result2
}

// doIconHashAndSave 获取icon，并将icon image保存到服务端
func doIconHashAndSave(resultIPScan *portscan.Result, resultDomainScan *domainscan.Result) string {
	hash := fingerprint.NewIconHash()
	if resultIPScan != nil {
		hash.ResultPortScan = *resultIPScan
	}
	if resultDomainScan != nil {
		hash.ResultDomainScan = *resultDomainScan
	}
	hash.Do()

	if len(hash.IconHashInfoResult.Result) <= 0 {
		return ""
	}
	x := comm.NewXClient()
	var result2 string
	err := x.Call(context.Background(), "SaveIconImageResult", &hash.IconHashInfoResult.Result, &result2)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return err.Error()
	}
	return result2
}

// NewFingerprintTask 根据端口及域名扫描结果，根据设置的拆分规模，生成指纹识别子任务
func NewFingerprintTask(portScanResult *portscan.Result, domainScanResult *domainscan.Result, config FingerprintTaskConfig) (result string, err error) {
	if config.IsHttpx == false && config.IsFingerprintHub == false && config.IsIconHash == false && config.IsScreenshot == false {
		return
	}
	//指纹识别：
	if portScanResult != nil && len(portScanResult.IPResult) > 0 {
		index := 0
		mapIpPort := make(map[string][]int)
		for ip, ipr := range portScanResult.IPResult {
			mapIpPort[ip] = make([]int, 0)
			for port := range ipr.Ports {
				mapIpPort[ip] = append(mapIpPort[ip], port)
			}
			index++
			if index%IPNumberPerFingerprintTask == 0 || index == len(portScanResult.IPResult) {
				newConfig := config
				newConfig.IPTargetMap = mapIpPort
				result, err = sendFingerprintTask(newConfig)
				if err != nil {
					return
				}
				mapIpPort = make(map[string][]int)
			}
		}
	}
	if domainScanResult != nil && len(domainScanResult.DomainResult) > 0 {
		index := 0
		mapDomain := make(map[string]struct{})
		for domain := range domainScanResult.DomainResult {
			mapDomain[domain] = struct{}{}
			index++
			if index%DomainNumberPerFingerprintTask == 0 || index == len(domainScanResult.DomainResult) {
				newConfig := config
				newConfig.DomainTargetMap = mapDomain
				result, err = sendFingerprintTask(newConfig)
				if err != nil {
					return
				}
				mapDomain = make(map[string]struct{})
			}
		}
	}

	return
}

// sendFingerprintTask 调用api发送任务
func sendFingerprintTask(config FingerprintTaskConfig) (result string, err error) {
	fpConfigMarshal, _ := json.Marshal(config)
	newTaskArgs := comm.NewTaskArgs{
		TaskName:   "fingerprint",
		ConfigJSON: string(fpConfigMarshal),
	}
	x := comm.NewXClient()
	err = x.Call(context.Background(), "NewTask", &newTaskArgs, &result)
	if err != nil {
		logging.RuntimeLog.Error("Start fingerprint task fail:", err)
		logging.CLILog.Error("Start fingerprint task fail:", err)
	}
	return
}
