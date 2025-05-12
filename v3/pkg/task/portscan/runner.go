package portscan

import (
	"bytes"
	"github.com/hanc00l/nemo_go/v3/pkg/conf"
	"github.com/hanc00l/nemo_go/v3/pkg/core"
	"github.com/hanc00l/nemo_go/v3/pkg/logging"
	"github.com/hanc00l/nemo_go/v3/pkg/task/execute"
	"github.com/hanc00l/nemo_go/v3/pkg/utils"
	"os"
	"os/exec"
	"strings"
)

type Executor interface {
	IsExecuteFromCmd() bool
	GetExecuteCmd() string
	GetRequiredResources() (re []core.RequiredResource)
	GetExecuteArgs(inputTempFile, outputTempFile string, ipv6 bool) (cmdArgs []string)
	Run(target []string, ipv6 bool)
	ParseContentResult(content []byte) (result Result)
}

func NewExecutor(executeName string, config execute.PortscanConfig, isProxy bool) Executor {
	executorMap := map[string]Executor{
		"nmap":    &Nmap{Config: config},
		"masscan": &Masscan{Config: config},
		"gogo":    &Gogo{Config: config, IsProxy: isProxy},
	}

	return executorMap[executeName]
}

func Do(taskInfo execute.ExecutorTaskInfo) (result Result) {
	result.IPResult = make(map[string]*IPResult)

	if len(taskInfo.PortScan) <= 0 {
		return
	}
	config, ok := taskInfo.PortScan[taskInfo.Executor]
	if !ok {
		logging.RuntimeLog.Errorf("子任务的executor配置不存在：%s", taskInfo.Executor)
		return
	}
	config.ExcludeTarget = taskInfo.ExcludeTarget
	executor := NewExecutor(taskInfo.Executor, config, taskInfo.IsProxy)
	if executor == nil {
		logging.RuntimeLog.Errorf("子任务的executor不存在：%s", taskInfo.Executor)
		return
	}
	re := executor.GetRequiredResources()
	if len(re) > 0 {
		err := core.CheckRequiredResource(re, false)
		if err != nil {
			logging.RuntimeLog.Errorf("任务资源检查和请求失败:%s", err.Error())
			return
		}
	}
	var targetIpV4, targetIpV6 []string
	for _, target := range strings.Split(taskInfo.Target, ",") {
		t := strings.TrimSpace(target)
		if utils.CheckIPV4(t) || utils.CheckIPV4Subnet(t) {
			targetIpV4 = append(targetIpV4, t)
		} else if utils.CheckIPV6(t) || utils.CheckIPV6Subnet(t) {
			targetIpV6 = append(targetIpV6, t)
		}
	}
	if executor.IsExecuteFromCmd() {
		if len(targetIpV4) > 0 {
			runByCmd(executor, targetIpV4, false, &result)
		}
		if len(targetIpV6) > 0 {
			if conf.Ipv6Support {
				runByCmd(executor, targetIpV6, true, &result)
			} else {
				logging.RuntimeLog.Warning("ipv6扫描选项不支持，跳过对ipv6地址的扫描:", targetIpV6)
			}
		}
	} else {
		// 目前portscan任务全部为cmd执行方式
	}
	return
}

func runByCmd(executor Executor, targetIP []string, ipv6 bool, r *Result) {
	inputTargetFile := utils.GetTempPathFileName()
	resultTempFile := utils.GetTempPathFileName()
	defer func() {
		_ = os.Remove(inputTargetFile)
		_ = os.Remove(resultTempFile)
	}()

	err := os.WriteFile(inputTargetFile, []byte(strings.Join(targetIP, "\n")), 0666)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		return
	}
	cmdArgs := executor.GetExecuteArgs(inputTargetFile, resultTempFile, ipv6)
	cmd := exec.Command(executor.GetExecuteCmd(), cmdArgs...)
	var stderr bytes.Buffer
	cmd.Stdout = os.Stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		logging.RuntimeLog.Error(err, stderr.String())
		logging.CLILog.Error(err, stderr.String())
		return
	}
	parseResult(executor, resultTempFile, r)
}

func parseResult(executor Executor, outputTempFile string, r *Result) {
	content, err := os.ReadFile(outputTempFile)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return
	}
	result := executor.ParseContentResult(content)
	for ip, ipr := range result.IPResult {
		r.IPResult[ip] = ipr
	}
}
