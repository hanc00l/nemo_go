package workerapi

import (
	"context"
	"github.com/hanc00l/nemo_go/pkg/comm"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/domainscan"
	"github.com/hanc00l/nemo_go/pkg/task/onlineapi"
	"github.com/hanc00l/nemo_go/pkg/task/portscan"
)

// Quake Quake任务
func Quake(taskId, configJSON string) (result string, err error) {
	var ok bool
	if ok, result, err = CheckTaskStatus(taskId); !ok {
		return result, err
	}

	config := onlineapi.QuakeConfig{}
	if err = ParseConfig(configJSON, &config); err != nil {
		return FailedTask(err.Error()), err
	}

	q := onlineapi.NewQuake(config)
	q.Do()
	if config.IsIPLocation {
		doLocation(&q.IpResult)
	}
	//指纹识别：
	if len(q.IpResult.IPResult) > 0 {
		portscanConfig := portscan.Config{
			IsHttpx:          config.IsHttpx,
			IsWhatWeb:        config.IsWhatWeb,
			IsWappalyzer:     config.IsWappalyzer,
			IsFingerprintHub: config.IsFingerprintHub,
			IsIconHash:       config.IsIconHash,
		}
		DoIPFingerPrint(portscanConfig, &q.IpResult)
		if q.Config.IsScreenshot {
			DoScreenshotAndSave(&q.IpResult, nil)
		}
	}
	if len(q.DomainResult.DomainResult) > 0 {
		domainscanConfig := domainscan.Config{
			IsHttpx:          config.IsHttpx,
			IsWhatWeb:        config.IsWhatWeb,
			IsWappalyzer:     config.IsWappalyzer,
			IsFingerprintHub: config.IsFingerprintHub,
			IsIconHash:       config.IsIconHash,
		}
		DoDomainFingerPrint(domainscanConfig, &q.DomainResult)
		if q.Config.IsScreenshot {
			DoScreenshotAndSave(nil, &q.DomainResult)
		}
	}
	// 保存结果
	x := comm.NewXClient()
	args := comm.ScanResultArgs{
		TaskID:       taskId,
		IPConfig:     &portscan.Config{OrgId: config.OrgId},
		DomainConfig: &domainscan.Config{OrgId: config.OrgId},
		IPResult:     q.IpResult.IPResult,
		DomainResult: q.DomainResult.DomainResult,
	}
	err = x.Call(context.Background(), "SaveScanResult", &args, &result)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return FailedTask(err.Error()), err
	}

	return SucceedTask(result), nil
}


