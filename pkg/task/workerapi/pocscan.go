package workerapi

import (
	"github.com/hanc00l/nemo_go/pkg/comm"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/pocscan"
)

// PocScan 漏洞验证任务
func PocScan(taskId, mainTaskId, configJSON string) (result string, err error) {
	var ok bool
	if ok, result, err = CheckTaskStatus(taskId); !ok {
		return result, err
	}
	config := pocscan.Config{}
	if err = ParseConfig(configJSON, &config); err != nil {
		return FailedTask(err.Error()), err
	}
	//读取资产开放端口
	var resultIPPorts string
	if config.IsLoadOpenedPort {
		args := comm.LoadIPOpenedPortArgs{
			WorkspaceId: config.WorkspaceId,
			Target:      config.Target,
		}
		err = comm.CallXClient("LoadOpenedPort", &args, &resultIPPorts)
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
		MainTaskId:          mainTaskId,
		VulnerabilityResult: scanResult,
	}
	err = comm.CallXClient("SaveVulnerabilityResult", &resultArgs, &result)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return FailedTask(err.Error()), err
	}

	return SucceedTask(result), nil
}

// XrayPocScan 调用本地的xraypoc，批量验证漏洞任务
func XrayPocScan(taskId, mainTaskId, configJSON string) (result string, err error) {
	var ok bool
	if ok, result, err = CheckTaskStatus(taskId); !ok {
		return result, err
	}
	config := pocscan.XrayPocConfig{}
	if err = ParseConfig(configJSON, &config); err != nil {
		return FailedTask(err.Error()), err
	}

	result, err = doXrayPocScanAndSave(taskId, mainTaskId, config)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return FailedTask(err.Error()), err
	}

	return SucceedTask(result), nil
}

// doXrayPocScanAndSave xraypoc扫描
func doXrayPocScanAndSave(taskId string, mainTaskId string, config pocscan.XrayPocConfig) (result string, err error) {
	var scanResult []pocscan.Result
	p := pocscan.NewXrayPoc(config)
	p.Do()
	scanResult = p.VulResult
	// 保存结果
	resultArgs := comm.ScanResultArgs{
		TaskID:              taskId,
		MainTaskId:          mainTaskId,
		VulnerabilityResult: scanResult,
	}
	err = comm.CallXClient("SaveVulnerabilityResult", &resultArgs, &result)
	return
}
