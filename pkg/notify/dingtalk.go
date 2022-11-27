package notify

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// https://open.dingtalk.com/document/group/custom-robot-access

type DingTalk struct {
}

type DingTalkResponseInfo struct {
	Code    int    `json:"errcode"`
	Message string `json:"errmsg"`
}

func (d *DingTalk) Send(token string, message string) (err error) {
	url := fmt.Sprintf("https://oapi.dingtalk.com/robot/send?access_token=%s", token)
	//-d '{"msgtype": "text","text": {"content":"Nemo任务通知：\n我就是我, 是不一样的烟火"}}'
	text := make(map[string]string)
	text["content"] = fmt.Sprintf("Nemo任务通知：\n%s", message)
	data := make(map[string]interface{})
	data["text"] = text
	data["msgtype"] = "text"
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
	var msgData DingTalkResponseInfo
	if err = json.Unmarshal(responseData, &msgData); err != nil {
		return
	}
	//{"errcode":0,"errmsg":"ok"}
	//{"errcode":310000,"errmsg":"description:关键词不匹配;solution:请联系群管理员查看此机器人的关键词，并在发送的信息中包含此关键词;"}
	if msgData.Code != 0 {
		err = errors.New(msgData.Message)
	}
	return
}
