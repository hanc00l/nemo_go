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
	if config.IsHttpx {
		httpx := fingerprint.NewHttpx()
		httpx.ResultPortScan = *resultPortScan
		httpx.Do()
	}
	if config.IsFingerprintHub {
		fp := fingerprint.NewFingerprintHub()
		fp.ResultPortScan = *resultPortScan
		fp.Do()
	}
	if config.IsIconHash {
		DoIconHashAndSave(resultPortScan, nil)
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
	if config.IsFingerprintHub {
		fp := fingerprint.NewFingerprintHub()
		fp.ResultDomainScan = *resultDomainScan
		fp.Do()
	}
	if config.IsIconHash {
		DoIconHashAndSave(nil, resultDomainScan)
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

// DoIconHashAndSave 获取icon，并将icon image保存到服务端
func DoIconHashAndSave(resultIPScan *portscan.Result, resultDomainScan *domainscan.Result) string {
	hash := fingerprint.NewIconHash()
	if resultIPScan != nil {
		hash.ResultPortScan = *resultIPScan
	}
	if resultDomainScan != nil {
		hash.ResultDomainScan = *resultDomainScan
	}
	hash.Do()

	if len(hash.IconHashInfoResult.Result) <= 0 {
		return ""
	}
	x := comm.NewXClient()
	var result2 string
	err := x.Call(context.Background(), "SaveIconImageResult", &hash.IconHashInfoResult.Result, &result2)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return err.Error()
	}
	return result2
}
