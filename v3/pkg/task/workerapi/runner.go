package workerapi

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/hanc00l/nemo_go/v3/pkg/core"
	"github.com/hanc00l/nemo_go/v3/pkg/db"
	"github.com/hanc00l/nemo_go/v3/pkg/logging"
	"github.com/hanc00l/nemo_go/v3/pkg/task/domainscan"
	"github.com/hanc00l/nemo_go/v3/pkg/task/execute"
	"github.com/hanc00l/nemo_go/v3/pkg/task/fingerprint"
	"github.com/hanc00l/nemo_go/v3/pkg/task/llmapi"
	"github.com/hanc00l/nemo_go/v3/pkg/task/onlineapi"
	"github.com/hanc00l/nemo_go/v3/pkg/task/pocscan"
	"github.com/hanc00l/nemo_go/v3/pkg/task/portscan"
	"github.com/hanc00l/nemo_go/v3/pkg/utils"
	"strings"
)

func checkTask(configJSON string) (ok bool, config execute.ExecutorTaskInfo, result string, err error) {
	if err = ParseConfig(configJSON, &config); err != nil {
		logging.RuntimeLog.Error(err)
		result = FailedTask(err.Error())
		return
	}
	if config.Target == "" {
		logging.RuntimeLog.Warningf("任务目标为空，taskId:%s（任务目标可能未设置或者被阻止）", config.TaskId)
		result = FailedTask("任务目标为空（任务目标可能未设置或者被阻止）")
		return
	}
	if ok, result, err = CheckTaskStatus(config.TaskId, config.MainTaskId); !ok {
		return
	}
	ok = true

	return
}

func newNextExecutorTask(config execute.ExecutorTaskInfo, target string, executor string) (err error) {
	executorConfig := config
	executorConfig.Target = target
	executorConfig.Executor = executor
	executorConfig.PreTaskId = config.TaskId
	executorConfig.TaskId = uuid.New().String()

	var result bool
	err = core.CallXClient("NewTask", &executorConfig, &result)

	return
}

func domainDoPortscan(config execute.ExecutorTaskInfo, domainscanConfig execute.DomainscanConfig, result []db.AssetDocument) {
	// 获取IP地址
	ipMap := make(map[string]struct{})
	subnetMap := make(map[string]struct{})
	// 获取所有域名结果
	for _, doc := range result {
		if len(doc.Ip.IpV4) > 0 {
			for _, ip := range doc.Ip.IpV4 {
				ipMap[ip.IPName] = struct{}{}
				ipArray := strings.Split(ip.IPName, ".")
				subnet := fmt.Sprintf("%s.%s.%s.0/24", ipArray[0], ipArray[1], ipArray[2])
				subnetMap[subnet] = struct{}{}
			}
		}
		if len(doc.Ip.IpV6) > 0 {
			// ipv6只进行ip扫描，不做C段扫描
			for _, ip := range doc.Ip.IpV6 {
				ipMap[ip.IPName] = struct{}{}
			}
		}
	}
	if len(ipMap) == 0 {
		logging.RuntimeLog.Warningf("域名任务的扫描结果为空，taskId:%s（下一步端口扫描任务将不会执行）", config.TaskId)
		return
	}
	// 生成独立的新任务：去除原来的domainscan配置，新增portscan配置，其它的fingerprint、pocscan保留不变：
	newConfig := config
	newConfig.DomainScan = nil
	newConfig.PortScan = make(map[string]execute.PortscanConfig)
	newConfig.PortScan[domainscanConfig.ResultPortscanBin] = *domainscanConfig.ResultPortscanConfig
	ipSlice := core.NewTaskSlice()
	ipSlice.IpSliceNumber = config.TargetSliceNum
	ipSlice.TaskMode = config.TargetSliceType
	// C段和IP扫描只需要扫描一次，C段扫描优先级高于IP扫描
	if domainscanConfig.IsIPSubnetPortScan {
		ipSlice.IpTarget = utils.SetToSlice(subnetMap)
	} else if domainscanConfig.IsIPPortScan {
		ipSlice.IpTarget = utils.SetToSlice(ipMap)
	}
	targets, _ := ipSlice.DoIpSlice()
	for _, target := range targets {
		_ = newNextExecutorTask(newConfig, target, domainscanConfig.ResultPortscanBin)
	}
}

func getResultTarget(result []db.AssetDocument, chunkSize int) [][]string {
	var slice []string
	for _, doc := range result {
		slice = append(slice, doc.Authority)
	}
	var chunks [][]string
	for i := 0; i < len(slice); i += chunkSize {
		end := i + chunkSize
		if end > len(slice) {
			end = len(slice)
		}
		chunks = append(chunks, slice[i:end])
	}
	return chunks
}

func doNextFingerAndPocscanExecutorTask(config execute.ExecutorTaskInfo, result []db.AssetDocument, doFingerprint bool) {
	// 如果任务获取到结果，则新建下一步任务
	var TaskSliceSize = 50
	targetSlice := getResultTarget(result, TaskSliceSize)
	for _, target := range targetSlice {
		if doFingerprint && len(config.FingerPrint) > 0 {
			for executor, _ := range config.FingerPrint {
				_ = newNextExecutorTask(config, strings.Join(target, ","), executor)
			}
		}
		// 如果有指纹识别，则在指纹识别结果中新建pocscan的任务
		if !doFingerprint && len(config.PocScan) > 0 {
			for executor, _ := range config.PocScan {
				_ = newNextExecutorTask(config, strings.Join(target, ","), executor)
			}
		}
	}
}

func PortScan(configJSON string) (result string, err error) {
	var ok bool
	var config execute.ExecutorTaskInfo
	ok, config, result, err = checkTask(configJSON)
	if !ok || err != nil {
		return
	}
	// 执行任务
	taskResult := portscan.Do(config)
	resultArgs := core.TaskAssetDocumentResultArgs{
		WorkspaceId: config.WorkspaceId,
		MainTaskId:  config.MainTaskId,
		Result:      taskResult.ParseResult(config),
	}
	// 保存结果
	if err = core.CallXClient("SaveTaskResult", &resultArgs, &result); err != nil {
		logging.RuntimeLog.Error(err)
		return FailedTask(err.Error()), err
	}
	// 如果任务获取到结果，则新建下一步任务
	doNextFingerAndPocscanExecutorTask(config, resultArgs.Result, true)

	return SucceedTask(result), nil
}

func DomainScan(configJSON string) (result string, err error) {
	var ok bool
	var config execute.ExecutorTaskInfo
	ok, config, result, err = checkTask(configJSON)
	if !ok || err != nil {
		return
	}
	// 执行任务
	taskResult := domainscan.Do(config)
	resultArgs := core.TaskAssetDocumentResultArgs{
		WorkspaceId: config.WorkspaceId,
		MainTaskId:  config.MainTaskId,
		Result:      taskResult.ParseResult(config),
	}
	// 保存结果
	if err = core.CallXClient("SaveTaskResult", &resultArgs, &result); err != nil {
		logging.RuntimeLog.Error(err)
		return FailedTask(err.Error()), err
	}
	// 如果任务获取到结果，则新建下一步任务
	doNextFingerAndPocscanExecutorTask(config, resultArgs.Result, true)
	// 如果有端口扫描，则新建portscan的任务
	var domainscanConfig execute.DomainscanConfig
	for _, v := range config.DomainScan {
		domainscanConfig = v
		break
	}
	if domainscanConfig.IsIPSubnetPortScan || domainscanConfig.IsIPPortScan {
		domainDoPortscan(config, domainscanConfig, resultArgs.Result)
	}

	return SucceedTask(result), nil
}

func OnlineAPI(configJSON string) (result string, err error) {
	var ok bool
	var config execute.ExecutorTaskInfo
	ok, config, result, err = checkTask(configJSON)
	if !ok || err != nil {
		return
	}
	// 执行任务
	taskResult := onlineapi.Do(config)
	resultArgs := core.TaskAssetDocumentResultArgs{
		WorkspaceId: config.WorkspaceId,
		MainTaskId:  config.MainTaskId,
		Result:      onlineapi.ParseResult(config, taskResult),
	}
	// 保存结果
	if err = core.CallXClient("SaveTaskResult", &resultArgs, &result); err != nil {
		logging.RuntimeLog.Error(err)
		return FailedTask(err.Error()), err
	}
	// 如果任务获取到结果，则新建下一步任务
	doNextFingerAndPocscanExecutorTask(config, resultArgs.Result, true)

	return SucceedTask(result), nil
}

func QueryData(configJSON string) (result string, err error) {
	var ok bool
	var config execute.ExecutorTaskInfo
	ok, config, result, err = checkTask(configJSON)
	if !ok || err != nil {
		return
	}
	// 执行任务
	// icp、whois
	taskResult := onlineapi.DoQuery(config)
	if len(taskResult) > 0 {
		var docResult []db.QueryDocument
		for _, doc := range taskResult {
			docResult = append(docResult, db.QueryDocument{
				Domain:   doc.Domain,
				Category: doc.Category,
				Content:  doc.Content,
			})
		}
		if err = core.CallXClient("SaveQueryData", &docResult, &result); err != nil {
			logging.RuntimeLog.Error(err)
		}
	}
	// 不需要新建下一步任务
	return SucceedTask(result), nil
}

func Fingerprint(configJSON string) (result string, err error) {
	var ok bool
	var config execute.ExecutorTaskInfo
	ok, config, result, err = checkTask(configJSON)
	if !ok || err != nil {
		return
	}
	// 执行指纹识别
	taskResult := fingerprint.Do(config)
	docResult, screenshotResult := fingerprint.ParseResult(config, &taskResult)
	// 保存指纹结果
	resultArgs := core.TaskAssetDocumentResultArgs{
		WorkspaceId: config.WorkspaceId,
		MainTaskId:  config.MainTaskId,
		Result:      docResult,
	}
	if err = core.CallXClient("SaveTaskResult", &resultArgs, &result); err != nil {
		logging.RuntimeLog.Error(err)
		return FailedTask(err.Error()), err
	}
	// 保存截图结果
	var resultScreenshot string
	if err = core.CallXClient("SaveScreenShotResult", &screenshotResult, &resultScreenshot); err != nil {
		logging.RuntimeLog.Error(err)
		return FailedTask(err.Error()), err
	}
	// 如果任务获取到结果，则新建下一步任务
	if len(config.PocScan) > 0 {
		// 获取pocscan配置，目前只支持nuclei模式
		var pocscanConfig execute.PocscanConfig
		for _, v := range config.PocScan {
			pocscanConfig = v
			break
		}
		// 匹配指纹识别结果，只有在指纹识别模块才能启用
		if pocscanConfig.PocType == "matchFinger" {
			// 根据指纹结果匹配poc文件，并启动nuclei任务
			targetPocMapResult := core.MatchAssetPoc(resultArgs.Result, pocscanConfig)
			if len(targetPocMapResult) > 0 {
				for target, newPocConfig := range targetPocMapResult {
					newConfig := config
					newConfig.PocScan["nuclei"] = newPocConfig
					_ = newNextExecutorTask(newConfig, target, "nuclei")
				}
			}
			// 如果有密码爆破，则匹配service，并启动zombie任务
			if pocscanConfig.IsBrutePassword {
				serviceTargets := core.MatchAssetService(resultArgs.Result)
				if len(serviceTargets) > 0 {
					newConfig := config
					newConfig.PocScan["zombie"] = pocscanConfig
					_ = newNextExecutorTask(newConfig, strings.Join(serviceTargets, ","), "zombie")
				}
			}
		} else {
			// 指定poc文件方式
			// 基于http状态码扫描
			if pocscanConfig.IsScanBaseWebStatus {
				var newResult []db.AssetDocument
				for _, doc := range resultArgs.Result {
					if len(doc.HttpStatus) > 0 {
						newResult = append(newResult, doc)
					}
				}
				if len(newResult) > 0 {
					doNextFingerAndPocscanExecutorTask(config, newResult, false)
				}
			} else {
				doNextFingerAndPocscanExecutorTask(config, resultArgs.Result, false)
			}
		}
	}

	return SucceedTask(fmt.Sprintf("%s %s", result, resultScreenshot)), nil
}

func PocScan(configJSON string) (result string, err error) {
	var ok bool
	var config execute.ExecutorTaskInfo
	ok, config, result, err = checkTask(configJSON)
	if !ok || err != nil {
		return
	}
	// 执行任务
	taskResult := pocscan.Do(config)
	resultArgs := core.VulResultArgs{
		WorkspaceId: config.WorkspaceId,
		MainTaskId:  config.MainTaskId,
		Result:      pocscan.ParseResult(config, &taskResult),
	}
	// 保存结果
	if err = core.CallXClient("SaveVulResult", &resultArgs, &result); err != nil {
		logging.RuntimeLog.Error(err)
		return FailedTask(err.Error()), err
	}

	return SucceedTask(result), nil
}

func LLMScan(configJSON string) (result string, err error) {
	var ok bool
	var config execute.ExecutorTaskInfo
	ok, config, result, err = checkTask(configJSON)
	if !ok || err != nil {
		return
	}
	// 执行任务
	taskResult := llmapi.Do(config)
	// 结果不保存到数据库，只返回任务结果，并启动下一步任务
	result = fmt.Sprintf("域名：%d%s", len(taskResult), taskResult)
	// 如果任务获取到结果，则新建下一步任务(domainscan或者onlineapi)
	if len(taskResult) > 0 {
		var llmConfig execute.LLMAPIConfig
		for _, v := range config.LLMAPI {
			llmConfig = v
			break
		}
		// 自动关联组织
		if llmConfig.AutoAssociateOrg {
			if orgId, err := callAssociateOrg(config.Target, config.WorkspaceId); err == nil && orgId != "" {
				config.OrgId = orgId
			}
		}
		for _, target := range taskResult {
			for executor, _ := range config.DomainScan {
				_ = newNextExecutorTask(config, target, executor)
			}
			for executor, _ := range config.OnlineAPI {
				_ = newNextExecutorTask(config, target, executor)
			}
		}
	}
	return SucceedTask(result), nil
}

func ICPPlusScan(configJSON string) (result string, err error) {
	var ok bool
	var config execute.ExecutorTaskInfo
	ok, config, result, err = checkTask(configJSON)
	if !ok || err != nil {
		return
	}
	// 执行任务
	taskResult := onlineapi.DoICPPlusQuery(config)
	var domainResult []string
	if len(taskResult) > 0 {
		var docResult []db.QueryDocument
		for _, doc := range taskResult {
			docResult = append(docResult, db.QueryDocument{
				Domain:   doc.Domain,
				Category: doc.Category,
				Content:  doc.Content,
			})
			domainResult = append(domainResult, doc.Domain)
		}
		if err = core.CallXClient("SaveQueryData", &docResult, &result); err != nil {
			logging.RuntimeLog.Error(err)
		}
	}
	result = fmt.Sprintf("域名：%d%s", len(domainResult), domainResult)
	// 下一步任务：
	if len(domainResult) > 0 {
		var llmConfig execute.LLMAPIConfig
		for _, v := range config.LLMAPI {
			llmConfig = v
			break
		}
		// 自动关联组织
		if llmConfig.AutoAssociateOrg {
			if orgId, err := callAssociateOrg(config.Target, config.WorkspaceId); err == nil && orgId != "" {
				config.OrgId = orgId
			}
		}
		for _, target := range domainResult {
			for executor, _ := range config.DomainScan {
				_ = newNextExecutorTask(config, target, executor)
			}
			for executor, _ := range config.OnlineAPI {
				_ = newNextExecutorTask(config, target, executor)
			}
		}
	}
	return SucceedTask(result), nil
}

func callAssociateOrg(orgName string, workspaceId string) (orgId string, err error) {
	args := core.AssociateOrgArgs{
		WorkspaceId: workspaceId,
		OrgName:     orgName,
	}
	err = core.CallXClient("AssociateOrg", &args, &orgId)
	if err != nil {
		logging.RuntimeLog.Errorf("关联组织失败：%s", err)
	} else {
		if orgId == "" {
			logging.RuntimeLog.Errorf("自动关联组织失败：%s", orgName)
		}
	}

	return
}
