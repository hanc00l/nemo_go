package workerapi

import (
	"context"
	"github.com/hanc00l/nemo_go/pkg/comm"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/domainscan"
	"github.com/hanc00l/nemo_go/pkg/task/fingerprint"
	"github.com/hanc00l/nemo_go/pkg/task/portscan"
)

// DoIPFingerPrint 对 IP结果进行指纹识别
func DoIPFingerPrint(config portscan.Config, resultPortScan *portscan.Result) {
	if config.IsWhatWeb {
		whatweb := fingerprint.NewWhatweb()
		whatweb.ResultPortScan = *resultPortScan
		whatweb.Do()
	}
	if config.IsHttpx {
		httpx := fingerprint.NewHttpx()
		httpx.ResultPortScan = *resultPortScan
		httpx.Do()
	}
	if config.IsWappalyzer {
		wappalyzer := fingerprint.NewWappalyzer()
		wappalyzer.ResultPortScan = *resultPortScan
		wappalyzer.Do()
	}
	if config.IsFingerprintHub {
		fp := fingerprint.NewFingerprintHub()
		fp.ResultPortScan = *resultPortScan
		fp.Do()
	}
}

// DoDomainFingerPrint 对域名结果进行指纹识别
func DoDomainFingerPrint(config domainscan.Config, resultDomainScan *domainscan.Result) {
	// 指纹识别
	if config.IsHttpx {
		httpx := fingerprint.NewHttpx()
		httpx.ResultDomainScan = *resultDomainScan
		httpx.Do()
	}
	if config.IsWhatWeb {
		whatweb := fingerprint.NewWhatweb()
		whatweb.ResultDomainScan = *resultDomainScan
		whatweb.Do()
	}
	if config.IsWappalyzer {
		wappalyzer := fingerprint.NewWappalyzer()
		wappalyzer.ResultDomainScan = *resultDomainScan
		wappalyzer.Do()
	}
	if config.IsFingerprintHub {
		fp := fingerprint.NewFingerprintHub()
		fp.ResultDomainScan = *resultDomainScan
		fp.Do()
	}
}

// DoScreenshotAndSave 执行Screenshot并保存
func DoScreenshotAndSave(resultIPScan *portscan.Result, resultDomainScan *domainscan.Result) string {
	ss := fingerprint.NewScreenShot()
	if resultIPScan != nil {
		ss.ResultPortScan = *resultIPScan
	}
	if resultDomainScan != nil {
		ss.ResultDomainScan = *resultDomainScan
	}
	ss.Do()

	resultScreenshot := ss.LoadResult()
	x := comm.NewXClient()
	var result2 string
	err := x.Call(context.Background(), "SaveScreenshotResult", &resultScreenshot, &result2)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return err.Error()
	}
	return result2
}
