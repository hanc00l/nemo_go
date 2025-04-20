package fingerprint

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/hanc00l/nemo_go/v3/pkg/conf"
	"github.com/hanc00l/nemo_go/v3/pkg/core"
	"github.com/hanc00l/nemo_go/v3/pkg/logging"
	"github.com/hanc00l/nemo_go/v3/pkg/resource"
	"github.com/hanc00l/nemo_go/v3/pkg/utils"
	"path/filepath"
)

type Fingerprintx struct {
	OptimizationMode bool
	IsProxy          bool
}

func (f *Fingerprintx) GetRequiredResources() (re []core.RequiredResource) {
	re = append(re, core.RequiredResource{
		Category: resource.FingerprintxCategory,
		Name:     utils.GetThirdpartyBinNameByPlatform(utils.Fingerprintx),
	})
	return
}

func (f *Fingerprintx) IsExecuteFromCmd() bool {
	return true
}

func (f *Fingerprintx) GetExecuteCmd() string {
	return filepath.Join(conf.GetRootPath(), "thirdparty/fingerprintx", utils.GetThirdpartyBinNameByPlatform(utils.Fingerprintx))
}

func (f *Fingerprintx) GetExecuteArgs(inputTempFile, outputTempFile string) (cmdArgs []string) {
	cmdArgs = append(cmdArgs,
		"--json", "-l", inputTempFile, "-o", outputTempFile,
	)
	if f.IsProxy {
		if proxy := conf.GetProxyConfig(); proxy != "" {
			cmdArgs = append(cmdArgs, "-p", proxy)
		} else {
			logging.RuntimeLog.Warning("获取代理配置失败或禁用了代理功能，代理被跳过")
			logging.CLILog.Warning("get proxy config fail or disabled by worker,skip proxy!")
		}
	}
	return
}

func (f *Fingerprintx) Run(target []string) (result Result) {
	//TODO implement me
	panic("implement me")
}

func (f *Fingerprintx) ParseContentResult(content []byte) (result Result) {
	result.FingerResults = make(map[string]interface{})

	lines := bytes.Split(content, []byte{'\n'})
	for _, line := range lines {
		if len(line) > 0 {
			r := f.ParseFingerprintxJson(line)
			if len(r.Host) > 0 {
				result.FingerResults[(fmt.Sprintf("%s:%d", r.Host, r.Port))] = r
			}
		}
	}
	return
}

// ParseFingerprintxJson 解析一条JSON记录
func (f *Fingerprintx) ParseFingerprintxJson(content []byte) (resultJSON FingerprintxResult) {
	err := json.Unmarshal(content, &resultJSON)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return
	}
	// 去除ssh协议中无意义的algo字段部份
	if resultJSON.Protocol == "ssh" {
		var data map[string]json.RawMessage
		if err = json.Unmarshal(resultJSON.Raw, &data); err == nil {
			delete(data, "algo")
			resultJSON.Raw, _ = json.Marshal(data)
		}
	}
	if len(resultJSON.Host) <= 0 {
		resultJSON.Host = resultJSON.IP
	}
	return
}
