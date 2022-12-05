package workerapi

import (
	"bufio"
	"github.com/hanc00l/nemo_go/pkg/comm"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/custom"
	"github.com/hanc00l/nemo_go/pkg/task/domainscan"
	"github.com/hanc00l/nemo_go/pkg/task/onlineapi"
	"github.com/hanc00l/nemo_go/pkg/task/portscan"
	"github.com/hanc00l/nemo_go/pkg/utils"
	"os"
	"path/filepath"
	"strings"
)

// Fofa Fofa任务
func Fofa(taskId, mainTaskId, configJSON string) (result string, err error) {
	return doFofaOnlineAPI(taskId, mainTaskId, configJSON, "fofa")
}

// Quake Quake任务
func Quake(taskId, mainTaskId, configJSON string) (result string, err error) {
	return doFofaOnlineAPI(taskId, mainTaskId, configJSON, "quake")
}

// Hunter Hunter
func Hunter(taskId, mainTaskId, configJSON string) (result string, err error) {
	return doFofaOnlineAPI(taskId, mainTaskId, configJSON, "hunter")
}

// doFofaOnlineAPI 执行fofa、hunter及quake的资产搜索任务
func doFofaOnlineAPI(taskId string, mainTaskId string, configJSON string, apiName string) (result string, err error) {
	// 检查任务状态
	var ok bool
	if ok, result, err = CheckTaskStatus(taskId); !ok {
		return result, err
	}
	// 解析任务参数
	config := onlineapi.OnlineAPIConfig{}
	if err = ParseConfig(configJSON, &config); err != nil {
		return FailedTask(err.Error()), err
	}
	//执行任务
	var ipResult portscan.Result
	var domainResult domainscan.Result
	ipResult, domainResult, result, err = doFofaAndSave(taskId, mainTaskId, apiName, config)
	//fingerprint
	_, err = NewFingerprintTask(taskId, mainTaskId, &ipResult, &domainResult, FingerprintTaskConfig{
		IsHttpx:          config.IsHttpx,
		IsFingerprintHub: config.IsFingerprintHub,
		IsIconHash:       config.IsIconHash,
		IsScreenshot:     config.IsScreenshot,
	})
	if err != nil {
		logging.RuntimeLog.Error(err)
		return FailedTask(err.Error()), err
	}

	return SucceedTask(result), nil
}

// doFofaAndSave 执行fofa、hunter及quake的资产搜索，并保存结果
func doFofaAndSave(taskId string, mainTaskId string, apiName string, config onlineapi.OnlineAPIConfig) (ipResult portscan.Result, domainResult domainscan.Result, result string, err error) {
	if apiName == "fofa" {
		fofa := onlineapi.NewFofa(config)
		fofa.Do()
		ipResult = fofa.IpResult
		domainResult = fofa.DomainResult
	} else if apiName == "quake" {
		quake := onlineapi.NewQuake(config)
		quake.Do()
		ipResult = quake.IpResult
		domainResult = quake.DomainResult
	} else if apiName == "hunter" {
		hunter := onlineapi.NewHunter(config)
		hunter.Do()
		ipResult = hunter.IpResult
		domainResult = hunter.DomainResult
	}
	portscan.FilterIPHasTooMuchPort(&ipResult, true)
	checkIgnoreResult(&ipResult, &domainResult, config)

	if config.IsIPLocation {
		doLocation(&ipResult)
	}
	// 保存结果
	args := comm.ScanResultArgs{
		TaskID:       taskId,
		MainTaskId:   mainTaskId,
		IPConfig:     &portscan.Config{OrgId: config.OrgId},
		DomainConfig: &domainscan.Config{OrgId: config.OrgId},
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
						if attr.Source == "fofa" && attr.Tag == "title" {
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
	filterFile := filepath.Join(conf.GetRootPath(), "thirdparty/custom", "fofa_filter_keyword_local.txt")
	if utils.CheckFileExist(filterFile) == false {
		logging.RuntimeLog.Errorf("fofa filter file not exist:%s", filterFile)
		return
	}
	inputFile, err := os.Open(filterFile)
	if err != nil {
		logging.RuntimeLog.Errorf("Could not read fofa filter file: %s\n", err)
		return
	}
	defer inputFile.Close()
	scanner := bufio.NewScanner(inputFile)
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if text == "" {
			continue
		}
		globalLocalFilterWords = append(globalLocalFilterWords, text)
	}
	return
}
