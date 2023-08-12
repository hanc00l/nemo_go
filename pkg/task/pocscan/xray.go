package pocscan

import (
	"bufio"
	"bytes"
	"encoding/json"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/custom"
	"github.com/hanc00l/nemo_go/pkg/utils"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Xray struct {
	Config Config
	Result []Result
}

// NewXray 创建xray对象
func NewXray(config Config) *Xray {
	return &Xray{Config: config}
}

// Do 调用xray执行一次webscan
func (x *Xray) Do() {
	resultTempFile := utils.GetTempPathFileName()
	inputTargetFile := utils.GetTempPathFileName()
	defer os.Remove(resultTempFile)
	defer os.Remove(inputTargetFile)

	btc := custom.NewBlackTargetCheck(custom.CheckAll)
	urls := strings.Split(x.Config.Target, ",")
	for idx, url := range urls {
		if btc.CheckBlack(utils.HostStrip(url)) {
			logging.RuntimeLog.Warningf("%s is in blacklist,skip...", url)
			continue
		}
		if strings.HasSuffix(url, ":443") {
			urls[idx] = "https://" + url
		}
	}
	err := os.WriteFile(inputTargetFile, []byte(strings.Join(urls, "\n")), 0666)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		return
	}
	cmdBin := filepath.Join(conf.GetAbsRootPath(), "thirdparty/xray", utils.GetThirdpartyBinNameByPlatform(utils.Xray))
	var cmdArgs []string
	// pocType: default（使用xray内置的poc）、custom（使用自定义的poc）
	pocType := "default"
	pocFile := x.Config.PocFile
	if strings.Contains(x.Config.PocFile, "|") {
		if pocArray := strings.Split(x.Config.PocFile, "|"); len(pocArray) == 2 {
			pocType = pocArray[0]
			pocFile = pocArray[1]
		}
	}
	// check poc file name
	if strings.Contains(pocFile, "..") || strings.Contains(pocFile, "/") || strings.Contains(pocFile, "\\") {
		logging.RuntimeLog.Warningf("invalid poc file:%s", pocFile)
		return
	}
	// format xray cmdline
	cmdArgs = append(
		cmdArgs,
		"--log-level", "error", "webscan", "--json-output", resultTempFile, "--url-file", inputTargetFile,
	)
	if pocType == "default" && pocFile != "" {
		cmdArgs = append(
			cmdArgs, "--plugins", "phantasm", "--poc", pocFile,
		)
	}
	if pocType == "custom" {
		if pocType != "" {
			cmdArgs = append(
				cmdArgs, "--plugins", "phantasm", "--poc",
				filepath.Join(conf.GetAbsRootPath(), conf.GlobalWorkerConfig().Pocscan.Xray.PocPath, x.Config.PocFile),
			)
		} else {
			cmdArgs = append(
				cmdArgs, "--plugins", "phantasm", "--poc",
				filepath.Join(conf.GetAbsRootPath(), conf.GlobalWorkerConfig().Pocscan.Xray.PocPath, "*"),
			)
		}
	}
	cmd := exec.Command(cmdBin, cmdArgs...)
	//Fix:必须指定绝对路径，才能正确读取到配置文件
	cmd.Dir = filepath.Join(conf.GetAbsRootPath(), "thirdparty/xray")
	var stderr bytes.Buffer
	cmd.Stdout = os.Stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		logging.RuntimeLog.Error(err, stderr)
		logging.CLILog.Error(err, stderr)
		return
	}
	x.parseXrayResult(resultTempFile)
}

// parseXrayResult 解析xray的运行结果
func (x *Xray) parseXrayResult(outputTempFile string) {
	content, err := os.ReadFile(outputTempFile)
	if err != nil || len(content) == 0 {
		return
	}

	var xr []xrayJSONResult
	err = json.Unmarshal(content, &xr)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		return
	}
	for _, r := range xr {
		var extraAll []string
		for _, s := range r.Detail.Snapshot {
			extraAll = append(extraAll, strings.Join(s, ""))
		}
		host := utils.HostStrip(r.Target.Url)
		if host == "" || strings.Contains(r.Plugin, "baseline") || strings.Contains(r.Plugin, "dirscan") {
			continue
		}
		x.Result = append(x.Result, Result{
			Target:      host,
			Url:         r.Target.Url,
			PocFile:     r.Plugin,
			Source:      "xray",
			Extra:       strings.Join(extraAll, ""),
			WorkspaceId: x.Config.WorkspaceId,
		})
	}
}

// LoadPocFile 加载poc文件列表
func (x *Xray) LoadPocFile() (pocs []string) {
	files, _ := filepath.Glob(filepath.Join(conf.GetRootPath(), conf.GlobalWorkerConfig().Pocscan.Xray.PocPath, "*.yml"))
	for _, file := range files {
		_, pocFile := filepath.Split(file)
		pocs = append(pocs, pocFile)
	}
	return
}

// LoadDefaultPocFile 加载xray内置的poc文件列表
func (x *Xray) LoadDefaultPocFile() (pocs []string) {
	inputFile, err := os.Open(filepath.Join(conf.GetRootPath(), "thirdparty/xray", "poc.list"))
	if err != nil {
		logging.RuntimeLog.Errorf("could not read poc.list: %s", err)
		return
	}
	defer inputFile.Close()

	scanner := bufio.NewScanner(inputFile)
	for scanner.Scan() {
		text := strings.ToLower(scanner.Text())
		if text == "" || strings.HasPrefix(text, "#") {
			continue
		}
		pocs = append(pocs, text)
	}
	return
}
