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
	x := comm.NewXClient()
	//读取资产开放端口
	if config.LoadOpenedPort {
		err = x.Call(context.Background(), "LoadOpenedPort", &config.Target, &result)
		if err == nil {
			config.Target = result
		} else {
			logging.RuntimeLog.Error(err)
		}
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
	} else if config.CmdBin == "dirsearch" {
		d := pocscan.NewDirsearch(config)
		d.Do()
		scanResult = d.Result
	} else if config.CmdBin == "nuclei" {
		n := pocscan.NewNuclei(config)
		n.Do()
		scanResult = n.Result
	}
	// 保存结果
	resultArgs := comm.ScanResultArgs{
		TaskID:              taskId,
		VulnerabilityResult: scanResult,
	}
	err = x.Call(context.Background(), "SaveVulnerabilityResult", &resultArgs, &result)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return FailedTask(err.Error()), err
	}

	return SucceedTask(result), nil
}
