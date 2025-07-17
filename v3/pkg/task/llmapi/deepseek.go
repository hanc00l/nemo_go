package llmapi

import (
	"bytes"
	"encoding/json"
	"github.com/hanc00l/nemo_go/v3/pkg/conf"
	"github.com/hanc00l/nemo_go/v3/pkg/logging"
	"github.com/hanc00l/nemo_go/v3/pkg/task/execute"
	"github.com/hanc00l/nemo_go/v3/pkg/utils"
	"io"
	"net/http"
	"regexp"
	"strings"
)

type DeepSeek struct {
	IsProxy bool
}

type DeepSeekOpenAI struct {
	IsProxy bool
}

func (d *DeepSeekOpenAI) Run(target string, api conf.APIToken, config execute.LLMAPIConfig) (result Result) {
	content, err := CallAPI(api.API, api.Model, api.Token, GetSystemContent(), GetUserPrompt(target))
	if err != nil {
		logging.RuntimeLog.Error("调用API失败：", err)
		return
	}
	return d.ParContentResult(content)
}

func (d *DeepSeekOpenAI) ParContentResult(content string) (result []string) {
	content = strings.ReplaceAll(content, "\n", "")
	//fmt.Println(content)
	re := regexp.MustCompile(`<<(.*)>>`)
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		data := matches[1]
		result = regexp.MustCompile(`,`).Split(data, -1)
	}
	return
}

// Message 定义了消息结构体
type Message struct {
	Content string `json:"content"`
	Role    string `json:"role"`
}

// RequestData 定义了请求数据结构体
type RequestData struct {
	Messages         []Message `json:"messages"`
	Model            string    `json:"model"`
	FrequencyPenalty float64   `json:"frequency_penalty"`
	MaxTokens        int       `json:"max_tokens"`
	PresencePenalty  float64   `json:"presence_penalty"`
	ResponseFormat   struct {
		Type string `json:"type"`
	} `json:"response_format"`
	Stop          interface{} `json:"stop"`
	Stream        bool        `json:"stream"`
	StreamOptions interface{} `json:"stream_options"`
	Temperature   float64     `json:"temperature"`
	TopP          float64     `json:"top_p"`
	Tools         interface{} `json:"tools"`
	ToolChoice    string      `json:"tool_choice"`
	Logprobs      bool        `json:"logprobs"`
	TopLogprobs   interface{} `json:"top_logprobs"`
}

// Response 定义了响应结构体
type Response struct {
	ID                string   `json:"id"`
	Object            string   `json:"object"`
	Created           int64    `json:"created"`
	Model             string   `json:"model"`
	Choices           []Choice `json:"choices"`
	Usage             Usage    `json:"usage"`
	SystemFingerprint string   `json:"system_fingerprint"`
}

// Choice 定义了选择结构体
type Choice struct {
	Index        int         `json:"index"`
	Message      Message     `json:"message"`
	Logprobs     interface{} `json:"logprobs"` // 使用interface{}来处理可能为null的情况
	FinishReason string      `json:"finish_reason"`
}

// Usage 定义了用量结构体
type Usage struct {
	PromptTokens          int `json:"prompt_tokens"`
	CompletionTokens      int `json:"completion_tokens"`
	TotalTokens           int `json:"total_tokens"`
	PromptCacheHitTokens  int `json:"prompt_cache_hit_tokens"`
	PromptCacheMissTokens int `json:"prompt_cache_miss_tokens"`
}

func (d *DeepSeek) Run(target string, apiToken conf.APIToken, config execute.LLMAPIConfig) (result Result) {
	response, err := d.Do(target, apiToken)
	if err != nil {
		return
	}
	for _, choice := range response.Choices {
		result = append(result, d.ParContentResult(choice.Message.Content)...)
	}
	return
}

func (d *DeepSeek) Do(target string, apiToken conf.APIToken) (response Response, err error) {
	reqData := RequestData{
		Messages: []Message{
			{
				Content: GetSystemContent(),
				Role:    "system",
			},
			{
				Content: GetUserPrompt(target),
				Role:    "user",
			}},
		Model: "deepseek-chat",
		//MaxTokens: 4096,
		ResponseFormat: struct {
			Type string `json:"type"`
		}{
			Type: "text",
		},
		Temperature: 1,
		ToolChoice:  "none",
		TopP:        1,
	}
	payloadBytes, _ := json.Marshal(reqData)
	payload := bytes.NewReader(payloadBytes)
	req, err := http.NewRequest(http.MethodPost, apiToken.API, payload)
	if err != nil {
		return
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+apiToken.Token)
	res, err := utils.GetProxyHttpClient(d.IsProxy).Do(req)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return
	}
	if res.StatusCode != 200 {
		logging.RuntimeLog.Errorf("%s request failed, status code: %d", req.URL.String(), res.StatusCode)
		return
	}
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		logging.RuntimeLog.Error(err)
		return
	}
	err = json.Unmarshal(body, &response)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return
	}
	return
}

func (d *DeepSeek) ParContentResult(content string) (result []string) {
	re := regexp.MustCompile(`<<(.*)>>`)
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		data := matches[1]
		result = regexp.MustCompile(`,`).Split(data, -1)
	}
	return
}
