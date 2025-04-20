package llmapi

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hanc00l/nemo_go/v3/pkg/conf"
	"github.com/hanc00l/nemo_go/v3/pkg/logging"
	"github.com/hanc00l/nemo_go/v3/pkg/task/execute"
	"github.com/northes/go-moonshot"
	"go.uber.org/zap"
	"regexp"
)

type Kimi struct {
	IsProxy bool
}

func (k *Kimi) GetPrompt(target string) string {
	prompt := fmt.Sprintf("请搜集%s在互联网上使用的域名，生成的结果确保没有重复，并且确保这些域名是实际存在的，请仔细核实。域名结果是根域名，不是包含主机的子域名，格式为以\"<<\"作为开始符号，以\">>\"作为结束符号，域名用\",\"分隔。", target)
	//fmt.Println(prompt)
	return prompt
}
func (k *Kimi) Run(target string, apiToken conf.APIToken, config execute.LLMAPIConfig) (result Result) {
	response, err := k.Do(target, apiToken)
	if err != nil {
		return
	}
	for _, choice := range response.Choices {
		//fmt.Println(choice.Message.Content)
		result = append(result, k.ParContentResult(choice.Message.Content)...)
	}
	return result
}

func (k *Kimi) Do(target string, apiToken conf.APIToken) (resp *moonshot.ChatCompletionsResponse, err error) {
	cli, err := moonshot.NewClient(apiToken.Token)
	if err != nil {
		logging.RuntimeLog.Error("Failed to create moonshot client", zap.Error(err))
		return
	}
	ctx := context.Background()
	builder := moonshot.NewChatCompletionsBuilder()
	builder.SetModel(moonshot.ModelMoonshotV1Auto)
	builder.SetTemperature(0.3)
	builder.AddUserContent(k.GetPrompt(target))
	//builder.AddSystemContent(GetSystemContent())
	builder.SetTool(&moonshot.ChatCompletionsTool{
		Type: moonshot.ChatCompletionsToolTypeBuiltinFunction,
		Function: &moonshot.ChatCompletionsToolFunction{
			Name: moonshot.BuiltinFunctionWebSearch,
		},
	})

	resp, err = cli.Chat().Completions(ctx, builder.ToRequest())
	if err != nil {
		logging.RuntimeLog.Error("Failed to get completions", zap.Error(err))
		return
	}
	if len(resp.Choices) != 0 {
		choice := resp.Choices[0]
		if choice.FinishReason == moonshot.FinishReasonToolCalls {
			for _, tool := range choice.Message.ToolCalls {
				if tool.Function.Name == moonshot.BuiltinFunctionWebSearch {
					// web search
					arguments := new(moonshot.ChatCompletionsToolBuiltinFunctionWebSearchArguments)
					if err = json.Unmarshal([]byte(tool.Function.Arguments), arguments); err != nil {
						continue
					}
					// do something...
					//fmt.Println(arguments)
					builder.AddMessageFromChoices(resp.Choices)
					builder.AddToolContent(tool.Function.Arguments, tool.Function.Name, tool.ID)
				}
			}
		}
	}

	resp, err = cli.Chat().Completions(ctx, builder.ToRequest())
	if err != nil {
		logging.RuntimeLog.Error("Failed to get completions", zap.Error(err))
		return
	}

	return
}

func (k *Kimi) ParContentResult(content string) (result []string) {
	re := regexp.MustCompile(`<<(.*)>>`)
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		data := matches[1]
		result = regexp.MustCompile(`,`).Split(data, -1)
	}
	return
}
