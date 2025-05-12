package fingerprint

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

func NewExecutor(executeName string, config execute.FingerprintConfig, isProxy bool) Executor {
	executorMap := map[string]Executor{}
	switch executeName {
	case "httpx":
		executorMap[executeName] = &Httpx{Config: config, IsProxy: isProxy}
	case "fingerprintx":
		executorMap[executeName] = &Fingerprintx{IsProxy: isProxy}
	}

	return executorMap[executeName]
}

func Do(taskInfo execute.ExecutorTaskInfo) (result Result) {
	result = Result{FingerResults: make(map[string]interface{})}

	if config, ok := taskInfo.FingerPrint["fingerprint"]; ok {
		if config.IsHttpx {
			exeResult := do1("httpx", config, taskInfo.Target, taskInfo.IsProxy)
			for domain, fingerResult := range exeResult.FingerResults {
				result.FingerResults[domain] = fingerResult
			}
		}
		if config.IsFingerprintx {
			var targetsNew []string
			for _, target := range strings.Split(taskInfo.Target, ",") {
				if _, exist := result.FingerResults[target]; !exist {
					targetsNew = append(targetsNew, target)
				}
			}
			if len(targetsNew) > 0 {
				exeResult := do1("fingerprintx", config, strings.Join(targetsNew, ","), taskInfo.IsProxy)
				for domain, fingerResult := range exeResult.FingerResults {
					result.FingerResults[domain] = fingerResult
				}
			}
		}
	}
	return
}

func do1(executorName string, executorConfig execute.FingerprintConfig, target string, isProxy bool) (result Result) {
	result.FingerResults = make(map[string]interface{})

	executor := NewExecutor(executorName, executorConfig, isProxy)
	if executor == nil {
		logging.RuntimeLog.Errorf("子任务的executor不存在：%s", executorName)
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
	var exeResult Result
	if executor.IsExecuteFromCmd() {
		exeResult = runByCmd(executor, strings.Split(target, ","))
	} else {
		exeResult = executor.Run(strings.Split(target, ","))
	}
	for domain, fingerResult := range exeResult.FingerResults {
		result.FingerResults[domain] = fingerResult
	}
	return
}

func runByCmd(executor Executor, target []string) (result Result) {
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
	content, err := os.ReadFile(resultTempFile)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return
	}
	result = executor.ParseContentResult(content)
	return
}
