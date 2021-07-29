package workerapi

import "github.com/hanc00l/nemo_go/pkg/task/onlineapi"

// ICPQuery ICP备案查询任务
func ICPQuery(taskId, configJSON string) (result string, err error) {
	isRevoked, err := CheckIsExistOrRevoked(taskId)
	if err != nil {
		return FailedTask(err.Error()), err
	}
	if isRevoked {
		return RevokedTask(""), nil
	}

	config := onlineapi.ICPQueryConfig{}
	if err = ParseConfig(configJSON, &config); err != nil {
		return FailedTask(err.Error()), err
	}

	icp := onlineapi.NewICPQuery(config)
	icp.Do()
	result = icp.UploadICPInfo()

	return SucceedTask(result), nil
}