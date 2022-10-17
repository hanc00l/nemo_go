package pocscan

import (
	"encoding/json"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/utils"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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

	urls := strings.Split(x.Config.Target, ",")
	err := os.WriteFile(inputTargetFile, []byte(strings.Join(urls, "\n")), 0666)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		return
	}
	cmdBin := filepath.Join(conf.GetAbsRootPath(), "thirdparty/xray", "xray_darwin_amd64")
	if runtime.GOOS == "linux" {
		cmdBin = filepath.Join(conf.GetAbsRootPath(), "thirdparty/xray", "xray_linux_amd64")
	} else if runtime.GOOS == "windows" {
		cmdBin = filepath.Join(conf.GetAbsRootPath(), "thirdparty/xray", "xray_windows_amd64.exe")
	}

	var cmdArgs []string
	cmdArgs = append(
		cmdArgs,
		"--log-level", "error", "webscan", "--plugins", "phantasm", "--poc",
		filepath.Join(conf.GetAbsRootPath(), conf.GlobalWorkerConfig().Pocscan.Xray.PocPath, x.Config.PocFile),
		"--json-output", resultTempFile, "--url-file", inputTargetFile,
	)
	cmd := exec.Command(cmdBin, cmdArgs...)
	_, err = cmd.CombinedOutput()
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
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
		if host == "" {
			continue
		}
		x.Result = append(x.Result, Result{
			Target:  host,
			Url:     r.Target.Url,
			PocFile: r.Plugin,
			Source:  "xray",
			Extra:   strings.Join(extraAll, ""),
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
