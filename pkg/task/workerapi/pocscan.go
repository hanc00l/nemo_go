package workerapi

import (
	"errors"
	"github.com/hanc00l/nemo_go/pkg/task/pocscan"
)

// PocScan 漏洞验证任务
func PocScan(taskId, configJSON string) (result string, err error) {
	isRevoked, err := CheckIsExistOrRevoked(taskId)
	if err != nil {
		return FailedTask(err.Error()), err
	}
	if isRevoked {
		return RevokedTask(""), nil
	}

	config := pocscan.Config{}
	if err = ParseConfig(configJSON, &config); err != nil {
		return FailedTask(err.Error()), err
	}

	var scanResult []pocscan.Result
	if config.CmdBin == "pocsuite" {
		p := pocscan.NewPocsuite(config)
		p.Do()
		scanResult = p.Result
	} else if config.CmdBin == "xray" {
		x := pocscan.NewXray(config)
		if !x.CheckXrayBinFile() {
			return FailedTask("xray binfile not exist or download fail"), errors.New("xray binfile not exist or download fail")
		}
		x.Do()
		scanResult = x.Result
	}
	result = pocscan.SaveResult(scanResult)

	return SucceedTask(result), nil
}
