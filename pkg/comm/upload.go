package comm

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"io"
	"net/http"
)

type Upload struct {
}

type ScreenshotFileInfo struct {
	Domain   string `json:"domain"`
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
	Content  []byte `json:"content"`
}

// PostData  worker向server上传数据，返回状态和信息
func PostData(url string, data []byte) (r *ResponseStatus, err error) {
	postData := EncryptData(data)

	req, err := http.NewRequest("POST", url, bytes.NewReader(postData))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("status code:%d", resp.StatusCode))
	}
	defer resp.Body.Close()
	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if len(responseData) == 0 {
		return nil, errors.New("response empty")
	}
	var rs ResponseStatus
	err = json.Unmarshal(DecryptData(responseData), &rs)
	if err != nil {
		return nil, err
	}

	return &rs, nil
}

// DoUploadScreenshot 上传screenshot
func DoUploadScreenshot(sfi ScreenshotFileInfo) bool {
	url := fmt.Sprintf("http://%s:%d/upload-screenshot", getServerHost(), conf.Nemo.Web.Port)
	sfiData, _ := json.Marshal(sfi)

	r,err := PostData(url,sfiData)
	if err != nil {
		logging.RuntimeLog.Errorf("upload screenshot fail:%v", err)
		return false
	}
	if r.Status != Success {
		logging.RuntimeLog.Errorf("upload screenshot fail,response:%s", r.Msg)
		return false
	}
	return true
}

// DoUploadICPInfo 上传ICP信息
func DoUploadICPInfo(icpInfoBytes []byte) bool {
	url := fmt.Sprintf("http://%s:%d/upload-icpinfo", getServerHost(), conf.Nemo.Web.Port)
	r, err := PostData(url, icpInfoBytes)
	if err != nil {
		logging.RuntimeLog.Errorf("upload icpinfo fail:%v", err)
		return false
	}
	if r.Status == Success {
		return true
	} else {
		logging.RuntimeLog.Errorf("upload icpinfo fail:%v", r.Msg)
		return false
	}
}
