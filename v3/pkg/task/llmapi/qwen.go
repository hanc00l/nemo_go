package llmapi

import (
	"github.com/hanc00l/nemo_go/v3/pkg/conf"
	"github.com/hanc00l/nemo_go/v3/pkg/logging"
	"github.com/hanc00l/nemo_go/v3/pkg/task/execute"
	"regexp"
	"strings"
)

type Qwen struct {
	IsProxy bool
}

func (q *Qwen) Run(target string, api conf.APIToken, config execute.LLMAPIConfig) (result Result) {
	content, err := CallAPI(api.API, api.Model, api.Token, GetSystemContent(), GetUserPrompt(target))
	if err != nil {
		logging.RuntimeLog.Error("调用API失败：", err)
		return
	}
	return q.ParContentResult(content)
}

func (q *Qwen) ParContentResult(content string) (result Result) {
	re := regexp.MustCompile(`<<(.*)>>`)
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		data := matches[1]
		for _, item := range regexp.MustCompile(`,`).Split(data, -1) {
			rr := strings.ReplaceAll(item, "<", "")
			rr = strings.ReplaceAll(rr, ">", "")
			result = append(result, rr)
		}
	}
	return
}
