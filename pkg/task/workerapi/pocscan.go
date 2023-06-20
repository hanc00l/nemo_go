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
	} else if config.CmdBin == "goby" {
		g := pocscan.NewGoby(config)
		g.Do()
		scanResult = g.Result
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
