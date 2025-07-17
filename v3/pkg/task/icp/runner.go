package icp

import (
	"fmt"
	"github.com/hanc00l/nemo_go/v3/pkg/conf"
	"github.com/hanc00l/nemo_go/v3/pkg/core"
	"github.com/hanc00l/nemo_go/v3/pkg/db"
	"github.com/hanc00l/nemo_go/v3/pkg/logging"
	"github.com/hanc00l/nemo_go/v3/pkg/task/execute"
	"strings"
)

// Executor 用于icp、whois等api的interface
type Executor interface {
	Run(domain string, apiKey string) (result []DataResult)
}

type APIKey struct {
	apiName     string
	apiKey      string
	apiKeyInUse string
}

// DataResult ICP备案查询的通用结果结构体
type DataResult struct {
	Domain         string `json:"Domain"`
	UnitName       string `json:"UnitName"`
	CompanyType    string `json:"CompanyType"`
	SiteLicense    string `json:"SiteLicense"`
	ServiceLicence string `json:"ServiceLicence"`
	VerifyTime     string `json:"VerifyTime"`
	Source         string `json:"Source"`
	IsLookupData   bool   `json:"IsLookupData"` // 是否为已有的查询结果
}

func NewICPQueryChinaz(executeName string, isProxy bool, isDisableLookup bool) (Executor, APIKey) {
	executorMap := map[string]Executor{
		"icp":      &ICPQueryChinaz{IsProxy: isProxy, IsDisableLookup: isDisableLookup},
		"icpPlus":  &ICPPlusQueryChinaz{IsProxy: isProxy, IsDisableLookup: isDisableLookup},
		"icpPlus2": &ICPPlus2QueryChinaz{IsProxy: isProxy, IsDisableLookup: isDisableLookup},
	}
	apiKeyMap := map[string]APIKey{
		"icp":      {apiName: "icp", apiKey: conf.GlobalWorkerConfig().API.ICPChinaz.Key},
		"icpPlus":  {apiName: "icpPlus", apiKey: conf.GlobalWorkerConfig().API.ICPPlusChinaz.Key},
		"icpPlus2": {apiName: "icpPlus2", apiKey: fmt.Sprintf("%s|%s", conf.GlobalWorkerConfig().API.ICPChinaz.Key, conf.GlobalWorkerConfig().API.ICPPlusChinaz.Key)},
	}
	return executorMap[executeName], apiKeyMap[executeName]
}

func NewICPQueryBeianx(executeName string, isProxy bool, isDisableLookup bool) (Executor, APIKey) {
	executorMap := map[string]Executor{
		"icp":      &ICPQueryBeianx{IsProxy: isProxy, IsDisableLookup: isDisableLookup},
		"icpPlus":  &ICPPlusQueryBeianx{IsProxy: isProxy, IsDisableLookup: isDisableLookup},
		"icpPlus2": &ICPPlus2QueryBeianx{IsProxy: isProxy, IsDisableLookup: isDisableLookup},
	}
	apiKeyMap := map[string]APIKey{
		"icp":      {apiName: "icp", apiKey: conf.GlobalWorkerConfig().API.ICPBeianx.Key},
		"icpPlus":  {apiName: "icpPlus", apiKey: conf.GlobalWorkerConfig().API.ICPBeianx.Key},
		"icpPlus2": {apiName: "icpPlus2", apiKey: conf.GlobalWorkerConfig().API.ICPBeianx.Key},
	}
	return executorMap[executeName], apiKeyMap[executeName]
}

func Do(taskInfo execute.ExecutorTaskInfo, isDisableLookup bool) (result []DataResult) {
	config, ok := taskInfo.ICP[taskInfo.Executor]
	if !ok {
		logging.RuntimeLog.Errorf("子任务的executor配置不存在：%s", taskInfo.Executor)
		return
	}

	for _, apiName := range config.APIName {
		var executor Executor
		var apiKey APIKey
		if apiName == "chinaz" {
			executor, apiKey = NewICPQueryChinaz(taskInfo.Executor, taskInfo.IsProxy, isDisableLookup)
		} else if apiName == "beianx" {
			executor, apiKey = NewICPQueryBeianx(taskInfo.Executor, taskInfo.IsProxy, isDisableLookup)
		} else {
			logging.RuntimeLog.Errorf("子任务的apiName配置不存在：%s", apiName)
			return
		}
		if executor == nil {
			logging.RuntimeLog.Errorf("子任务的executor不存在：%s", taskInfo.Executor)
			continue
		}
		if len(apiKey.apiKey) == 0 {
			logging.RuntimeLog.Errorf("子任务的api key不存在：%s", apiKey.apiName)
			continue
		}

		for _, domain := range strings.Split(taskInfo.Target, ",") {
			resultQuery := executor.Run(domain, apiKey.apiKey)
			result = append(result, resultQuery...)
		}
	}
	return
}

func DoLookup(domain string) (result DataResult) {
	rData := db.ICPDocument{}
	err := core.CallXClient("LookupICPData", &domain, &rData)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return
	}
	if len(rData.Domain) == 0 || rData.Domain != domain {
		return
	}
	result = DataResult{
		Domain:         rData.Domain,
		UnitName:       rData.UnitName,
		CompanyType:    rData.CompanyType,
		SiteLicense:    rData.SiteLicense,
		ServiceLicence: rData.ServiceLicence,
		VerifyTime:     rData.VerifyTime,
		Source:         rData.Source,
		IsLookupData:   true,
	}

	return
}
