package workerapi

import (
	"github.com/hanc00l/nemo_go/pkg/task/onlineapi"
)

// Fofa Fofa任务
func Fofa(taskId, configJSON string) (result string, err error) {
	isRevoked, err := CheckIsExistOrRevoked(taskId)
	if err != nil {
		return FailedTask(err.Error()), err
	}
	if isRevoked {
		return RevokedTask(""), nil
	}

	config := onlineapi.FofaConfig{}
	if err = ParseConfig(configJSON, &config); err != nil {
		return FailedTask(err.Error()), err
	}

	fofa := onlineapi.NewFofa(config)
	fofa.Do()
	if config.IsIPLocation {
		doLocation(&fofa.IpResult)
	}
	result = fofa.SaveResult()

	return SucceedTask(result), nil
}
