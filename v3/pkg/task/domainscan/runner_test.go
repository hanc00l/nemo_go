package domainscan

import (
	"github.com/hanc00l/nemo_go/v3/pkg/task/execute"
	"testing"
)

func TestResult_ParseResult(t *testing.T) {
	executorConfig := execute.ExecutorConfig{
		DomainScan: map[string]execute.DomainscanConfig{
			"massdns":   {},
			"subfinder": {},
		},
	}
	taskInfo := execute.ExecutorTaskInfo{
		MainTaskInfo: execute.MainTaskInfo{
			WorkspaceId:    "test",
			Target:         "",
			OrgId:          "test1",
			MainTaskId:     "domaintest",
			ExecutorConfig: executorConfig,
		},
		Executor: "subfinder",
	}

	taskInfo.MainTaskInfo.ExecutorConfig.DomainScan = executorConfig.DomainScan

	result := Do(taskInfo)
	docs := result.ParseResult(taskInfo)
	for _, doc := range docs {
		t.Log(doc)
	}
}
