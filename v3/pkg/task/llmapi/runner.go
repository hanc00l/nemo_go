package llmapi

import (
	"context"
	"fmt"
	"github.com/hanc00l/nemo_go/v3/pkg/conf"
	"github.com/hanc00l/nemo_go/v3/pkg/logging"
	"github.com/hanc00l/nemo_go/v3/pkg/task/execute"
	"github.com/sashabaranov/go-openai"
	"strings"
)

type Result []string

type Executor interface {
	Run(target string, api conf.APIToken, config execute.LLMAPIConfig) (result Result)
}

func NewExecutor(executeName string, isProxy bool) (Executor, conf.APIToken) {
	executorMap := map[string]Executor{
		"kimi":     &Kimi{IsProxy: isProxy},
		"deepseek": &DeepSeekOpenAI{IsProxy: isProxy},
		"qwen":     &Qwen{IsProxy: isProxy},
	}
	apiKeyMap := map[string]conf.APIToken{
		"kimi":     conf.GlobalWorkerConfig().LLMAPI.Kimi,
		"deepseek": conf.GlobalWorkerConfig().LLMAPI.DeepSeek,
		"qwen":     conf.GlobalWorkerConfig().LLMAPI.Qwen,
	}
	return executorMap[executeName], apiKeyMap[executeName]
}

func Do(taskInfo execute.ExecutorTaskInfo) (result Result) {
	if len(taskInfo.LLMAPI) <= 0 {
		return
	}
	config, ok := taskInfo.LLMAPI[taskInfo.Executor]
	if !ok {
		logging.RuntimeLog.Errorf("executor config for %s not found", taskInfo.Executor)
		return
	}
	executor, api := NewExecutor(taskInfo.Executor, taskInfo.IsProxy)
	if executor == nil {
		logging.RuntimeLog.Errorf("executor %s not found", taskInfo.Executor)
		return
	}
	for _, line := range strings.Split(taskInfo.Target, ",") {
		exeResult := executor.Run(line, api, config)
		result = append(result, exeResult...)
	}
	// 检查 result 是否有效并返回结果
	vr := ValidResult(result)
	logging.RuntimeLog.Infof("llmapi %s,result:%v,valide result:%v", taskInfo.Executor, result, vr)

	return vr
}

func CallAPI(apiBase, model, apiKey string, target string) (string, error) {
	// 创建自定义配置
	config := openai.DefaultConfig(apiKey)
	config.BaseURL = apiBase
	// 创建客户端
	client := openai.NewClientWithConfig(config)

	// 设置请求消息
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: GetSystemContent(),
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: GetUserPrompt(target),
		},
	}
	// 调用API
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    model,
			Messages: messages,
		},
	)
	if err != nil {
		return "", err
	}
	// 输出结果
	return resp.Choices[0].Message.Content, nil
}

func GetUserPrompt(target string) string {
	prompt := fmt.Sprintf("请收集%s在互联网上公开的域名", target)
	return prompt
}
func GetSystemContent() string {
	content := "你是一个网络信息收集助手，帮助用户收集指定公司或组织在互联网上公开的域名。请确保域名的备案主体是属于查询的公司或组织，返回的结果确保没有重复，并且这些域名是实际存在的，不是AI生成的，请仔细核实。返回的结果中，域名是根域名，不是包含主机的子域名，域名格式为以\"<<\"作为开始符号，以\">>\"作为结束符号，域名用\",\"分隔。"
	return content
}

func ValidResult(result []string) (r Result) {
	resultMap := make(map[string]struct{})

	addToResultMap := func(domain string) {
		if _, ok := resultMap[domain]; !ok {
			resultMap[domain] = struct{}{}
		}
	}
	for _, line := range result {
		domain := cleanDomain(line)
		if len(domain) > 0 {
			addToResultMap(domain)
		}
	}
	// 返回结果
	for domain := range resultMap {
		r = append(r, domain)
	}

	return
}

func cleanDomain(line string) string {
	// 移除首尾空格
	domain := strings.TrimSpace(line)

	// 替换协议和斜杠
	domain = strings.TrimPrefix(domain, "http://")
	domain = strings.TrimPrefix(domain, "https://")
	domain = strings.Trim(domain, "/")

	return domain
}
