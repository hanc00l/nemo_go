package llmapi

import (
	"github.com/hanc00l/nemo_go/v3/pkg/task/execute"
	"testing"
)

func TestKimiRun(t *testing.T) {
	executorConfig := execute.ExecutorConfig{
		LLMAPI: map[string]execute.LLMAPIConfig{
			"kimi":     {},
			"deepseek": {},
		}}
	taskInfo := execute.ExecutorTaskInfo{
		MainTaskInfo: execute.MainTaskInfo{
			WorkspaceId:    "test",
			Target:         "",
			OrgId:          "test1",
			MainTaskId:     "llmtest",
			ExecutorConfig: executorConfig,
		},
		Executor: "kimi",
	}

	taskInfo.MainTaskInfo.ExecutorConfig = executorConfig

	result := Do(taskInfo)
	for _, r := range result {
		t.Log(r)
	}
}

func TestRun(t *testing.T) {
	executorConfig := execute.ExecutorConfig{
		LLMAPI: map[string]execute.LLMAPIConfig{
			"kimi":     {},
			"deepseek": {},
			"qwen":     {},
		}}
	taskInfo := execute.ExecutorTaskInfo{
		MainTaskInfo: execute.MainTaskInfo{
			WorkspaceId:    "test",
			Target:         "百度在线网络技术（北京）有限公司",
			OrgId:          "test1",
			MainTaskId:     "llmtest",
			ExecutorConfig: executorConfig,
		},
		//Executor: "qwen",
		//Executor: "deepseek",
		//Executor: "kimi",
	}

	for _, api := range []string{"kimi", "qwen", "deepseek"} {
		taskInfo.Executor = api
		taskInfo.MainTaskInfo.ExecutorConfig = executorConfig
		result := Do(taskInfo)
		t.Log("---", api, "---")
		for _, r := range result {
			t.Log(r)
		}
	}
}
