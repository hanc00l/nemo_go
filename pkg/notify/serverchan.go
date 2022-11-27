package notify

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// https://sct.ftqq.com/

type ServerChan struct {
}

type ServerChanResponseInfo struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Info    string `json:"info"`
}

func (s *ServerChan) Send(token string, message string) (err error) {
	u := fmt.Sprintf("https://sctapi.ftqq.com/%s.send", token)
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
	}
	return
}
