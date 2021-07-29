package pocscan

import (
	"encoding/json"
	"fmt"
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
	xray := &Xray{Config: config}

	return xray
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
	cmdBin := filepath.Join(conf.GetRootPath(), "thirdparty/xray", "xray_darwin_amd64")
	if runtime.GOOS == "linux" {
		cmdBin = filepath.Join(conf.GetRootPath(), "thirdparty/xray", "xray_linux_amd64")
	}
	var cmdArgs []string
	cmdArgs = append(
		cmdArgs,
		"--log-level", "error", "webscan", "--plugins", "phantasm", "--poc",
		filepath.Join(conf.GetRootPath(), conf.Nemo.Pocscan.Xray.PocPath, x.Config.PocFile),
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
	files, _ := filepath.Glob(filepath.Join(conf.GetRootPath(), conf.Nemo.Pocscan.Xray.PocPath, "*.yml"))
	for _, file := range files {
		_, pocFile := filepath.Split(file)
		pocs = append(pocs, pocFile)
	}
	return
}

// CheckXrayBinFile 检查xray可执行文件是否存在，如果不存在，尝试从网上下载
// 由于xray持续更新中，下载的版本定义在config.yaml中
func (x *Xray) CheckXrayBinFile() bool {
	cmdBin := filepath.Join(conf.GetRootPath(), "thirdparty/xray", "xray_darwin_amd64")
	if runtime.GOOS == "linux" {
		cmdBin = filepath.Join(conf.GetRootPath(), "thirdparty/xray", "xray_linux_amd64")
	}
	_, err := os.Stat(cmdBin)
	if err == nil {
		return true
	}
	//download file
	logging.RuntimeLog.Info("xray binfile not exist,try to download...")
	tempDownloadPathFile := utils.GetTempPathFileName()
	defer os.Remove(tempDownloadPathFile)

	downloadUrl := fmt.Sprintf("https://github.com/chaitin/xray/releases/download/%s/xray_darwin_amd64.zip", conf.Nemo.Pocscan.Xray.LatestVersion)
	if runtime.GOOS == "linux" {
		downloadUrl = fmt.Sprintf("https://github.com/chaitin/xray/releases/download/%s/xray_linux_amd64.zip", conf.Nemo.Pocscan.Xray.LatestVersion)
	}
	isDownloadSuccess, err := utils.DownloadFile(downloadUrl, tempDownloadPathFile)
	if !isDownloadSuccess {
		logging.RuntimeLog.Errorf("download xray binfile fail: %s!", err)
		return false
	}
	//unzip download file
	logging.RuntimeLog.Info("xray download finish,try to unzip...")
	err = utils.Unzip(tempDownloadPathFile, filepath.Join(conf.GetRootPath(), "thirdparty/xray"))
	if err != nil {
		logging.RuntimeLog.Errorf("unzip xray zip binfile fail: %s!", err)
		return false
	}
	_, err = os.Stat(cmdBin)
	if err == nil {
		logging.RuntimeLog.Info("xray download success")
		return true
	}
	logging.RuntimeLog.Info("xray download and check fail")
	return false
}
