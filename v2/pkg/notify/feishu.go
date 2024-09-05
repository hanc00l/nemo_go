package notify

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// https://open.feishu.cn/document/ukTMukTMukTM/ucTM5YjL3ETO24yNxkjN

type Feishu struct {
}

type FeishuResponseInfo struct {
	Code    int    `json:"StatusCode"`
	Message string `json:"StatusMessage"`
}

func (f *Feishu) Send(token string, message string) (err error) {
	url := fmt.Sprintf("https://open.feishu.cn/open-apis/bot/v2/hook/%s", token)
	//-d '{"msg_type":"text","content":{"text":"request example"}}' \
	content := make(map[string]string)
	content["text"] = fmt.Sprintf("Nemo任务通知：\n%s", message)
	data := make(map[string]interface{})
	data["content"] = content
	data["msg_type"] = "text"
	b, _ := json.Marshal(data)
	var resp *http.Response
	if resp, err = http.Post(url, "application/json", bytes.NewBuffer(b)); err != nil {
		return
	}
	defer func(Body io.ReadCloser) {
		err = Body.Close()
	}(resp.Body)
	var responseData []byte
	if responseData, err = io.ReadAll(resp.Body); err != nil {
		return
	}
	var msgData FeishuResponseInfo
	if err = json.Unmarshal(responseData, &msgData); err != nil {
		return
	}
	// "Extra": null, "StatusCode": 0, "StatusMessage": "success"
	// "code": 9499, "msg": "Bad Request", "data": {}
	if msgData.Code != 0 && msgData.Message != "success" {
		err = errors.New(string(responseData))
	}
	return
}
