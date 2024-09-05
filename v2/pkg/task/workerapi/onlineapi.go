package workerapi

import (
	"bufio"
	"github.com/hanc00l/nemo_go/v2/pkg/comm"
	"github.com/hanc00l/nemo_go/v2/pkg/conf"
	"github.com/hanc00l/nemo_go/v2/pkg/logging"
	"github.com/hanc00l/nemo_go/v2/pkg/task/custom"
	"github.com/hanc00l/nemo_go/v2/pkg/task/domainscan"
	"github.com/hanc00l/nemo_go/v2/pkg/task/onlineapi"
	"github.com/hanc00l/nemo_go/v2/pkg/task/portscan"
	"github.com/hanc00l/nemo_go/v2/pkg/utils"
	"os"
	"path/filepath"
	"strings"
)

// Fofa Fofa任务
func Fofa(taskId, mainTaskId, configJSON string) (result string, err error) {
	return doOnlineAPI(taskId, mainTaskId, configJSON, "fofa")
}

// Quake Quake任务
func Quake(taskId, mainTaskId, configJSON string) (result string, err error) {
	return doOnlineAPI(taskId, mainTaskId, configJSON, "quake")
}

// Hunter Hunter
func Hunter(taskId, mainTaskId, configJSON string) (result string, err error) {
	return doOnlineAPI(taskId, mainTaskId, configJSON, "hunter")
}

// doOnlineAPI 执行fofa、hunter及quake的资产搜索任务
func doOnlineAPI(taskId string, mainTaskId string, configJSON string, apiName string) (result string, err error) {
	// 检查任务状态
	var ok bool
	if ok, result, err = CheckTaskStatus(taskId); !ok {
		return result, err
	}
	// 解析任务参数
	config := onlineapi.OnlineAPIConfig{}
	if err = ParseConfig(configJSON, &config); err != nil {
		logging.RuntimeLog.Error(err)
		return FailedTask(err.Error()), err
	}
	//执行任务
	var ipResult *portscan.Result
	var domainResult *domainscan.Result
	ipResult, domainResult, result, err = doOnlineAPIAndSave(taskId, mainTaskId, apiName, config)
	//fingerprint
	_, err = NewFingerprintTask(taskId, mainTaskId, ipResult, domainResult, FingerprintTaskConfig{
		IsHttpx:          config.IsHttpx,
		IsFingerprintHub: config.IsFingerprintHub,
		IsIconHash:       config.IsIconHash,
		IsScreenshot:     config.IsScreenshot,
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

// doOnlineAPIAndSave 执行fofa、hunter及quake的资产搜索，并保存结果
func doOnlineAPIAndSave(taskId string, mainTaskId string, apiName string, config onlineapi.OnlineAPIConfig) (ipResult *portscan.Result, domainResult *domainscan.Result, result string, err error) {
	s := onlineapi.NewOnlineAPISearch(config, apiName)
	s.Do()
	ipResult = &s.IpResult
	domainResult = &s.DomainResult
	checkIgnoreResult(ipResult, domainResult, config)

	if config.IsIPLocation {
		doLocation(ipResult)
	}
	// 保存结果
	args := comm.ScanResultArgs{
		TaskID:       taskId,
		MainTaskId:   mainTaskId,
		IPConfig:     &portscan.Config{OrgId: config.OrgId, WorkspaceId: config.WorkspaceId},
		DomainConfig: &domainscan.Config{OrgId: config.OrgId, WorkspaceId: config.WorkspaceId},
		IPResult:     ipResult.IPResult,
		DomainResult: domainResult.DomainResult,
	}
	err = comm.CallXClient("SaveScanResult", &args, &result)
	if err != nil {
		logging.RuntimeLog.Error(err)
	}
	return
}

// ICPQuery ICP备案查询任务
func ICPQuery(taskId, mainTaskId, configJSON string) (result string, err error) {
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
	err = comm.CallXClient("SaveICPResult", &icp.QueriedICPInfo, &result)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return FailedTask(err.Error()), err
	}

	return SucceedTask(result), nil
}

// WhoisQuery Whois查询任务
func WhoisQuery(taskId, mainTaskId, configJSON string) (result string, err error) {
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
	err = comm.CallXClient("SaveWhoisResult", &whois.QueriedWhoisInfo, &result)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return FailedTask(err.Error()), err
	}

	return SucceedTask(result), nil
}

// checkIgnoreResult 检查资产查询API中的IP资产，非中国IP或CDN，则不保存该结果
func checkIgnoreResult(portScanResult *portscan.Result, domainScanResult *domainscan.Result, config onlineapi.OnlineAPIConfig) {
	iplocation := custom.NewIPv4Location()
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
	// 关键词过滤
	globalFilterWords := addGlobalFilterWord()
	if len(portScanResult.IPResult) > 0 {
		for ip := range portScanResult.IPResult {
			ipInfo := portScanResult.IPResult[ip]
			needDelete := false
			if len(ipInfo.Ports) > 0 {
				for port := range ipInfo.Ports {
					portInfo := ipInfo.Ports[port]
					for _, attr := range portInfo.PortAttrs {
						if (attr.Source == "fofa" || attr.Source == "hunter" || attr.Source == "quake") && attr.Tag == "title" {
							if len(attr.Content) > 100 {
								needDelete = true
								break
							}
							if len(globalFilterWords) > 0 {
								for _, globalFilterWord := range globalFilterWords {
									if strings.Contains(attr.Content, globalFilterWord) {
										needDelete = true
										break
									}
								}
							}
						}
						if needDelete {
							break
						}
					}
					if needDelete {
						break
					}
				}
				if needDelete {
					delete(portScanResult.IPResult, ip)
					continue
				}
			}
		}
	}
}

func addGlobalFilterWord() (globalLocalFilterWords []string) {
	// 从custom目录中读取定义的过滤词，每一个关键词一行：
	filterFile := filepath.Join(conf.GetRootPath(), "thirdparty/custom", "onlineapi_filter_keyword_local.txt")
	if utils.CheckFileExist(filterFile) == false {
		logging.RuntimeLog.Warningf("fofa filter file not exist:%s", filterFile)
		return
	}
	inputFile, err := os.Open(filterFile)
	if err != nil {
		logging.RuntimeLog.Warningf("Could not read fofa filter file: %s", err)
		return
	}
	defer inputFile.Close()
	scanner := bufio.NewScanner(inputFile)
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if text == "" || strings.HasPrefix(text, "#") {
			continue
		}
		globalLocalFilterWords = append(globalLocalFilterWords, text)
	}
	return
}
