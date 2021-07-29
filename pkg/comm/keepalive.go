package comm

import (
	"encoding/json"
	"fmt"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/asynctask"
	"github.com/hanc00l/nemo_go/pkg/utils"
	"os"
	"path/filepath"
	"time"
)

type KeepAliveInfo struct {
	WorkerStatus asynctask.WorkerStatus `json:"worker_status"`
	CustomFiles  map[string]string      `json:"custom_files"`
}

type ResponseStatus struct {
	Status string `json:"status"`
	Msg    string `json:"msg"`
}

const (
	Success = "success"
	Fail    = "fail"
)

// asynFiles 需要同步的自定义配置文件
var asynFiles = []string{
	"thirdparty/custom/honeypot.txt",
	"thirdparty/custom/iplocation-custom.txt",
	"thirdparty/custom/iplocation-custom-B.txt",
	"thirdparty/custom/iplocation-custom-C.txt",
	"thirdparty/custom/services-custom.txt",
	"thirdparty/icp/icp.cache",
}

// getServerHost 获取上传的server地址
func getServerHost() string {
	if conf.RunMode == conf.Debug {
		return "127.0.0.1"
	}
	if conf.Nemo.Web.Host == "" || conf.Nemo.Web.Host == "0.0.0.0" {
		return "127.0.0.1"
	}
	return conf.Nemo.Web.Host
}

// EncryptData 加密数据
func EncryptData(data []byte) []byte {
	if conf.RunMode == conf.Release {
		if len(conf.Nemo.Web.EncryptKey) >= 16 {
			return utils.AesEncryptCBC(data, []byte(conf.Nemo.Web.EncryptKey[:16]))
		}
	}
	return data
}

// DecryptData 解密数据
func DecryptData(data []byte) []byte {
	if conf.RunMode == conf.Release {
		if len(conf.Nemo.Web.EncryptKey) >= 16 {
			return utils.AesDecryptCBC(data, []byte(conf.Nemo.Web.EncryptKey[:16]))
		}
	}
	return data
}

// DoKeepAlive worker请求keepAlive
func DoKeepAlive(ws asynctask.WorkerStatus) bool {
	url := fmt.Sprintf("http://%s:%d/worker-alive", getServerHost(), conf.Nemo.Web.Port)
	kaiData := NewKeepAliveRequestInfo(ws)

	r, err := PostData(url, kaiData)
	if err != nil {
		logging.RuntimeLog.Errorf("keep alive fail:%v", err)
		return false
	}
	if r.Status != Success {
		logging.RuntimeLog.Errorf("keep alive fail:%s", r.Msg)
		return false
	}
	// 自定义配置文件的同步
	if r.Msg != "" {
		customs := make(map[string]string)
		err = json.Unmarshal([]byte(r.Msg), &customs)
		if err == nil {
			SyncCustomFiles(customs)
		}
	}
	return true
}

// NewKeepAliveRequestInfo worker请求的keepAlive数据
func NewKeepAliveRequestInfo(ws asynctask.WorkerStatus) []byte {
	kai := KeepAliveInfo{
		WorkerStatus: ws,
		CustomFiles:  make(map[string]string),
	}
	kai.WorkerStatus.UpdateTime = time.Now()

	for _, file := range asynFiles {
		var txt string
		content, err := os.ReadFile(filepath.Join(conf.GetRootPath(), file))
		if err == nil {
			txt = string(content)
		}
		kai.CustomFiles[file] = utils.MD5(txt)
	}
	jsonData, _ := json.Marshal(kai)

	return jsonData
}

// NewKeepAliveResponseInfo server响应的keepAlive数据
func NewKeepAliveResponseInfo(req map[string]string) []byte {
	syncCustomFiles := make(map[string]string)
	for _, file := range asynFiles {
		if _, ok := req[file]; !ok || req[file] == "" {
			continue
		}
		content, err := os.ReadFile(filepath.Join(conf.GetRootPath(), file))
		if err != nil || len(content) == 0 {
			logging.RuntimeLog.Errorf("load custom file %s fail", file)
			continue
		}
		fileHash := utils.MD5(string(content))
		if fileHash == req[file] {
			continue
		}
		syncCustomFiles[file] = string(content)
	}
	jsonData, _ := json.Marshal(syncCustomFiles)

	return jsonData
}

// SyncCustomFiles 同步自定义配置文件
func SyncCustomFiles(cf map[string]string) {
	for _, file := range asynFiles {
		if _, ok := cf[file]; !ok || cf[file] == "" {
			continue
		}
		err := os.WriteFile(filepath.Join(conf.GetRootPath(), file), []byte(cf[file]), 0666)
		if err != nil {
			logging.RuntimeLog.Error(err)
		}
	}
}
