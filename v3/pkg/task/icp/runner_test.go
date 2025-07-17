package icp

import (
	"github.com/hanc00l/nemo_go/v3/pkg/task/execute"
	"testing"
)

func getOnlineAPITaskInfo() execute.ExecutorTaskInfo {
	executorConfig := execute.ExecutorConfig{
		ICP: map[string]execute.ICPConfig{
			"icp": {
				APIName: []string{"chinaz", "beianx"},
			},
			"icpPlus2": {
				APIName: []string{"chinaz", "beianx"},
			},
		},
	}
	taskInfo := execute.ExecutorTaskInfo{
		MainTaskInfo: execute.MainTaskInfo{
			WorkspaceId:    "test",
			OrgId:          "test1",
			MainTaskId:     "onlineapi_test",
			ExecutorConfig: executorConfig,
		},
	}
	taskInfo.MainTaskInfo.ExecutorConfig = executorConfig

	return taskInfo
}

func TestDo(t *testing.T) {
	taskInfo := getOnlineAPITaskInfo()
	taskInfo.Executor = "icpPlus2"
	taskInfo.Target = "chinaz.com"
	result := Do(taskInfo, false)
	for _, item := range result {
		t.Log(item)
	}
}
