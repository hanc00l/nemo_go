package core

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hanc00l/nemo_go/v3/pkg/db"
	"github.com/hanc00l/nemo_go/v3/pkg/logging"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

type NotifyExecutor interface {
	Notify(message string) (err error)
}

type NotifyData struct {
	TaskName string
	Target   string
	Runtime  string
	Result   string
}

type FeishuResponseInfo struct {
	Code    int    `json:"StatusCode"`
	Message string `json:"StatusMessage"`
}

type FeishuNotify struct {
	Token string
}

type DingTalkNotify struct {
	Token string
}

type DingTalkResponseInfo struct {
	Code    int    `json:"errcode"`
	Message string `json:"errmsg"`
}

type ServerChan struct {
	Token string
}

type ServerChanResponseInfo struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Info    string `json:"info"`
}

func NewNotifyExecutor(category string, token string) NotifyExecutor {
	executorMap := map[string]NotifyExecutor{
		"feishu":     &FeishuNotify{Token: token},
		"dingtalk":   &DingTalkNotify{Token: token},
		"serverchan": &ServerChan{Token: token},
	}

	return executorMap[category]
}

func Notify(docId []string, data NotifyData) (err error) {
	mongoClient, err := db.GetClient()
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		return
	}
	defer db.CloseClient(mongoClient)

	notify := db.NewNotify(mongoClient)
	wg := new(sync.WaitGroup)
	for _, id := range docId {
		doc, err := notify.Get(id)
		if err != nil {
			logging.RuntimeLog.Errorf("获取通知对象失败,Id: %s, err: %s", id, err.Error())
			continue
		}
		if doc.Status == "disable" {
			logging.RuntimeLog.Warnf("通知对象已禁用: %s，跳过...", doc.Name)
			continue
		}
		// 格式化通知消息
		message := formatNotifyData(doc.Template, data)
		wg.Add(1)
		// 发送通知
		go func(category string, token, message string) {
			defer wg.Done()
			executor := NewNotifyExecutor(category, token)
			if executor == nil {
				logging.RuntimeLog.Errorf("获取通知执行器失败: %s", category)
				return
			}
			if err := executor.Notify(message); err != nil {
				logging.RuntimeLog.Error(err.Error())
				return
			}
			//logging.RuntimeLog.Infof("通知发送成功, category: %s, message: %s", category, message)
		}(doc.Category, doc.Token, message)
	}
	wg.Wait()
	return
}

func (f *FeishuNotify) Notify(message string) (err error) {
	url := fmt.Sprintf("https://open.feishu.cn/open-apis/bot/v2/hook/%s", f.Token)
	//-d '{"msg_type":"text","content":{"text":"request example"}}' \
	content := make(map[string]string)
	content["text"] = message
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

func (d *DingTalkNotify) Notify(message string) (err error) {
	url := fmt.Sprintf("https://oapi.dingtalk.com/robot/send?access_token=%s", d.Token)
	//-d '{"msgtype": "text","text": {"content":"Nemo任务通知：\n我就是我, 是不一样的烟火"}}'
	text := make(map[string]string)
	text["content"] = message
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

func (s *ServerChan) Notify(message string) (err error) {
	u := fmt.Sprintf("https://sctapi.ftqq.com/%s.send", s.Token)
	data := fmt.Sprintf("title=Nemo任务通知&&desp=%s", url.QueryEscape(message))
	var resp *http.Response
	if resp, err = http.Post(u, "application/x-www-form-urlencoded", strings.NewReader(data)); err != nil {
		return
	}
	defer func(Body io.ReadCloser) {
		err = Body.Close()
	}(resp.Body)
	var responseData []byte
	if responseData, err = io.ReadAll(resp.Body); err != nil {
		return
	}
	var msgData ServerChanResponseInfo
	if err = json.Unmarshal(responseData, &msgData); err != nil {
		return
	}
	//success:
	//{"code":0,"message":"","data":{"pushid":"97909234","readkey":"SCTC9MOq0zG6NJY","error":"SUCCESS","errno":0}}%
	//fail:
	//{"message":"[AUTH]\u9519\u8bef\u7684Key","code":40001,"info":"\u9519\u8bef\u7684Key","args":[null],"scode":461}%
	if msgData.Code != 0 {
		err = errors.New(msgData.Message)
		return
	}
	return
}

func formatNotifyData(template string, data NotifyData) (message string) {
	const (
		TaskName = "$taskName$"
		Target   = "$target$"
		Runtime  = "$runtime$"
		Result   = "$result$"
	)
	data.TaskName = strings.Replace(data.TaskName, "\n", ",", -1)

	message = strings.Replace(template, TaskName, data.TaskName, -1)
	message = strings.Replace(message, Target, data.Target, -1)
	message = strings.Replace(message, Runtime, data.Runtime, -1)
	message = strings.Replace(message, Result, data.Result, -1)

	return
}
