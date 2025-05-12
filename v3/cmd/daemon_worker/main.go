package main

import (
	"github.com/hanc00l/nemo_go/v3/pkg/conf"
	"github.com/hanc00l/nemo_go/v3/pkg/core"
	"github.com/hanc00l/nemo_go/v3/pkg/logging"
	"github.com/hanc00l/nemo_go/v3/pkg/resource"
	"github.com/hanc00l/nemo_go/v3/pkg/utils"
	"path/filepath"
)

func prepareWorkerRunEnv() bool {
	// 准备日志目录
	logPath, err := filepath.Abs(filepath.Join(conf.GetRootPath(), "log"))
	if err != nil {
		logging.RuntimeLog.Error(err)
		return false
	}
	if !utils.MakePath(logPath) {
		logging.RuntimeLog.Error("create log path fail")
		return false
	}
	// 准备worker可执行文件
	workerBin := utils.GetThirdpartyBinNameByPlatform(utils.Worker)
	workerPathName, err := filepath.Abs(filepath.Join(conf.GetRootPath(), workerBin))
	if err != nil {
		logging.RuntimeLog.Error(err)
		return false
	}
	if utils.CheckFileExist(workerPathName) {
		return true
	}
	var re []core.RequiredResource
	re = append(re, core.RequiredResource{
		Category: resource.WorkerCategory,
		Name:     workerBin,
	})
	err = core.CheckRequiredResource(re, false)
	if err != nil {
		logging.RuntimeLog.Errorf("获取资源失败:%s", err.Error())
		return false
	}
	return true
}

func main() {
	core.DaemonRunOption = core.PrepareWorkerOptions()
	if core.DaemonRunOption == nil {
		return
	}
	if !prepareWorkerRunEnv() {
		return
	}
	go core.SetupCloseHandler()
	core.StartWorkerDaemon()
}
