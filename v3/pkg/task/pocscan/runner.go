package pocscan

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
	GetDir() string
	GetExecuteArgs(inputTempFile, outputTempFile string) (cmdArgs []string)
	GetRequiredResources() (re []core.RequiredResource)
	Run(target []string) (result Result)
	ParseContentResult(content []byte) (result Result)
	LoadPocFiles() (pocFiles []string)
}

func NewExecutor(executeName string, config execute.PocscanConfig, isProxy bool) Executor {
	executorMap := map[string]Executor{
		"nuclei": &Nuclei{Config: config, IsProxy: isProxy},
	}

	return executorMap[executeName]
}

func Do(taskInfo execute.ExecutorTaskInfo) (result Result) {
	config, ok := taskInfo.PocScan[taskInfo.Executor]
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
		for domain, vul := range exeResult.VulResult {
			result.VulResult[domain] = vul
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
	// 由于poc file是相对路径，所以需要设置工作目录
	workDir := executor.GetDir()
	if len(workDir) > 0 {
		cmd.Dir = workDir
	}
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
	result := executor.ParseContentResult(content)
	for _, vul := range result.VulResult {
		r.VulResult = append(r.VulResult, vul)
	}
}
