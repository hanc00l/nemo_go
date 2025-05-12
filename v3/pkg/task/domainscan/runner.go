package domainscan

import (
	"bytes"
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
	GetExecuteArgs(inputTempFile, outputTempFile string) (cmdArgs []string)
	GetRequiredResources() (re []core.RequiredResource)
	Run(target []string) (result Result)
	ParseContentResult(content []byte) (result Result)
}

func NewExecutor(executeName string, config execute.DomainscanConfig, isProxy bool) Executor {
	executorMap := map[string]Executor{
		"resolve":   &Resolve{Config: config},
		"massdns":   &Massdns{Config: config},
		"subfinder": &Subfinder{Config: config, IsProxy: isProxy},
	}

	return executorMap[executeName]
}

func Do(taskInfo execute.ExecutorTaskInfo) (result Result) {
	result = Result{DomainResult: make(map[string]*DomainResult)}

	config, ok := taskInfo.DomainScan[taskInfo.Executor]
	if !ok {
		logging.RuntimeLog.Errorf("子任务的executor配置不存在：%s", taskInfo.Executor)
		return
	}
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
	if executor.IsExecuteFromCmd() {
		runByCmd(executor, strings.Split(taskInfo.Target, ","), &result)
	} else {
		exeResult := executor.Run(strings.Split(taskInfo.Target, ","))
		for domain, domainResult := range exeResult.DomainResult {
			result.DomainResult[domain] = domainResult
		}
	}
	return
}

func runByCmd(executor Executor, target []string, r *Result) {
	inputTargetFile := utils.GetTempPathFileName()
	resultTempFile := utils.GetTempPathFileName()
	defer func() {
		_ = os.Remove(inputTargetFile)
		_ = os.Remove(resultTempFile)
	}()

	err := os.WriteFile(inputTargetFile, []byte(strings.Join(target, "\n")), 0666)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		return
	}
	cmdArgs := executor.GetExecuteArgs(inputTargetFile, resultTempFile)
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
	for domain, domainResult := range result.DomainResult {
		r.DomainResult[domain] = domainResult
	}
}
