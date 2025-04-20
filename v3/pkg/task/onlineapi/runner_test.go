package onlineapi

import (
	"github.com/hanc00l/nemo_go/v3/pkg/task/execute"
	"testing"
)

func getOnlineAPITaskInfo() execute.ExecutorTaskInfo {
	executorConfig := execute.ExecutorConfig{
		OnlineAPI: map[string]execute.OnlineAPIConfig{
			"fofa":    {},
			"hunter":  {},
			"quake":   {},
			"whois":   {},
			"icp":     {},
			"icpPlus": {},
		},
	}
	taskInfo := execute.ExecutorTaskInfo{
		MainTaskInfo: execute.MainTaskInfo{
			WorkspaceId:    "test",
			Target:         "",
			OrgId:          "test1",
			MainTaskId:     "onlineapi_test",
			ExecutorConfig: executorConfig,
		},
	}
	taskInfo.MainTaskInfo.ExecutorConfig = executorConfig

	return taskInfo
}

func TestResult_ParseResult(t *testing.T) {
	taskInfo := getOnlineAPITaskInfo()
	taskInfo.Executor = "fofa"
	result := Do(taskInfo)
	docs := ParseResult(taskInfo, result)
	for _, doc := range docs {
		t.Log(doc)
	}
}

func TestWhois_Run(t *testing.T) {
	taskInfo := getOnlineAPITaskInfo()
	taskInfo.Executor = "whois"
	result := DoQuery(taskInfo)
	t.Log(result)
}

func TestICP_Run(t *testing.T) {
	taskInfo := getOnlineAPITaskInfo()
	taskInfo.Executor = "icp"
	result := DoQuery(taskInfo)
	t.Log(result)
}
