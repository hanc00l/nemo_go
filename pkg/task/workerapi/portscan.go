package workerapi

import (
	"github.com/hanc00l/nemo_go/pkg/task/fingerprint"
	"github.com/hanc00l/nemo_go/pkg/task/portscan"
	"strings"
)

// PortScan 端口扫描任务
func PortScan(taskId, configJSON string) (result string, err error) {
	isRevoked, err := CheckIsExistOrRevoked(taskId)
	if err != nil {
		return FailedTask(err.Error()), err
	}
	if isRevoked {
		return RevokedTask(""), nil
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
	// 保存结果
	result = resultPortScan.SaveResult(config)

	if config.IsScreenshot {
		ss := fingerprint.NewScreenShot()
		ss.ResultPortScan = resultPortScan
		ss.Do()
		resultScreenshot := ss.UploadResult()
		result = strings.Join([]string{result, resultScreenshot}, ",")
	}

	return SucceedTask(result), nil
}
