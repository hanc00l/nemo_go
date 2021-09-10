package workerapi

import (
	"context"
	"github.com/hanc00l/nemo_go/pkg/comm"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/onlineapi"
)

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
