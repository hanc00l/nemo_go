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
	"strings"
)

type Pocsuite struct {
	Config Config
	Result []Result
}

// NewPocsuite 创建pocsuite对像
func NewPocsuite(config Config) *Pocsuite {
	return &Pocsuite{Config: config}
}

// Do 调用poc.py运行pocsuite3
func (p *Pocsuite) Do() {
	resultTempFile := utils.GetTempPathFileName()
	inputTargetFile := utils.GetTempPathFileName()
	defer os.Remove(resultTempFile)
	defer os.Remove(inputTargetFile)

	urls := strings.Split(p.Config.Target, ",")
	err := os.WriteFile(inputTargetFile, []byte(strings.Join(urls, "\n")), 0666)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		return
	}
	pocPathPfile := filepath.Join(conf.GetRootPath(), conf.GlobalWorkerConfig().Pocscan.Pocsuite.PocPath, p.Config.PocFile)
	cmdBin := filepath.Join(conf.GetRootPath(), "thirdparty/pocsuite/poc.py")
	var cmdArgs []string
	cmdArgs = append(
		cmdArgs,
		"-r", pocPathPfile, "-f", inputTargetFile, "-o", resultTempFile,
		"--threads", fmt.Sprintf("%d", conf.GlobalWorkerConfig().Pocscan.Pocsuite.Threads),
	)
	cmd := exec.Command(cmdBin, cmdArgs...)
	_, err = cmd.CombinedOutput()
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		return
	}
	p.parsePocsuiteResult(resultTempFile)
}

// parsePocsuiteResult 解析pocsuite3的运行结果
func (p *Pocsuite) parsePocsuiteResult(outputTempFile string) {
	content, err := os.ReadFile(outputTempFile)
	if err != nil || string(content) == "" {
		return
	}

	err = json.Unmarshal(content, &p.Result)
	if err == nil {
		for i, _ := range p.Result {
			_, fileName := filepath.Split(p.Result[i].PocFile)
			p.Result[i].PocFile = fileName
		}
	}
}

// LoadPocFile 加载poc文件列表
func (p *Pocsuite) LoadPocFile() (pocs []string) {
	files, _ := filepath.Glob(filepath.Join(conf.GetRootPath(), conf.GlobalWorkerConfig().Pocscan.Pocsuite.PocPath, "*.py"))
	for _, file := range files {
		_, pocFile := filepath.Split(file)
		pocs = append(pocs, pocFile)
	}
	return
}
