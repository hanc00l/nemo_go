package workerapi

import (
	"context"
	"errors"
	"github.com/hanc00l/nemo_go/pkg/comm"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/pocscan"
)

// PocScan 漏洞验证任务
func PocScan(taskId, configJSON string) (result string, err error) {
	var ok bool
	if ok, result, err = CheckTaskStatus(taskId); !ok {
		return result, err
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
	// 保存结果
	x := comm.NewXClient()

	err = x.Call(context.Background(), "SaveVulnerabilityResult", &scanResult, &result)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return FailedTask(err.Error()), err
	}

	return SucceedTask(result), nil
}
