package workerapi

import (
	"context"
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
	var resultIPPorts string
	if config.IsLoadOpenedPort {
		err = x.Call(context.Background(), "LoadOpenedPort", &config.Target, &resultIPPorts)
		if err == nil {
			config.Target = resultIPPorts
		} else {
			logging.RuntimeLog.Error(err)
		}
	}
	var scanResult []pocscan.Result
	if config.CmdBin == "xray" {
		x := pocscan.NewXray(config)
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

// XrayPocScan 调用本地的xraypoc，批量验证漏洞任务
func XrayPocScan(taskId, configJSON string) (result string, err error) {
	var ok bool
	if ok, result, err = CheckTaskStatus(taskId); !ok {
		return result, err
	}
	config := pocscan.XrayPocConfig{}
	if err = ParseConfig(configJSON, &config); err != nil {
		return FailedTask(err.Error()), err
	}

	result, err = doXrayPocScanAndSave(taskId, config)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return FailedTask(err.Error()), err
	}

	return SucceedTask(result), nil
}

func doXrayPocScanAndSave(taskId string, config pocscan.XrayPocConfig) (result string, err error) {
	var scanResult []pocscan.Result
	p := pocscan.NewXrayPoc(config)
	p.Do()
	scanResult = p.VulResult
	// 保存结果
	resultArgs := comm.ScanResultArgs{
		TaskID:              taskId,
		VulnerabilityResult: scanResult,
	}
	x := comm.NewXClient()
	err = x.Call(context.Background(), "SaveVulnerabilityResult", &resultArgs, &result)
	return
}
