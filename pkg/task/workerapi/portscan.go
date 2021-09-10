package workerapi

import (
	"context"
	"github.com/hanc00l/nemo_go/pkg/comm"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/fingerprint"
	"github.com/hanc00l/nemo_go/pkg/task/portscan"
	"strings"
)

// PortScan 端口扫描任务
func PortScan(taskId, configJSON string) (result string, err error) {
	var ok bool
	if ok, result, err = CheckTaskStatus(taskId); !ok {
		return result, err
	}
	config := portscan.Config{}
	if err = ParseConfig(configJSON, &config); err != nil {
		return FailedTask(err.Error()), err
	}
	var resultPortScan portscan.Result
	// 端口扫描
	if config.CmdBin == "nmap" {
		nmap := portscan.NewNmap(config)
		nmap.Do()
		resultPortScan = nmap.Result
	} else {
		mascan := portscan.NewMasscan(config)
		mascan.Do()
		resultPortScan = mascan.Result
	}
	// IP位置
	if config.IsIpLocation {
		doLocation(&resultPortScan)
	}
	// 指纹识别
	fpConfig := fingerprint.Config{OrgId: config.OrgId}
	if config.IsWhatWeb {
		whatweb := fingerprint.NewWhatweb(fpConfig)
		whatweb.ResultPortScan = resultPortScan
		whatweb.Do()
	}
	if config.IsHttpx {
		httpx := fingerprint.NewHttpx(fpConfig)
		httpx.ResultPortScan = resultPortScan
		httpx.Do()
	}
	if config.IsWappalyzer {
		wappalyzer := fingerprint.NewWappalyzer()
		wappalyzer.ResultPortScan = resultPortScan
		wappalyzer.Do()
	}
	// 保存结果
	x := comm.NewXClient()

	resultArgs := comm.ScanResultArgs{
		IPConfig: &config,
		IPResult: resultPortScan.IPResult,
	}
	err = x.Call(context.Background(), "SaveScanResult", &resultArgs, &result)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return FailedTask(err.Error()), err
	}
	if config.IsScreenshot {
		ss := fingerprint.NewScreenShot()
		ss.ResultPortScan = resultPortScan
		ss.Do()
		resultScreenshot := ss.LoadResult()
		var result2 string
		err = x.Call(context.Background(), "SaveScreenshotResult", &resultScreenshot, &result2)
		if err != nil {
			logging.RuntimeLog.Error(err)
			return FailedTask(err.Error()), err
		} else {
			result = strings.Join([]string{result, result2}, ",")
		}
	}

	return SucceedTask(result), nil
}
