package pocscan

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/hanc00l/nemo_go/v3/pkg/conf"
	"github.com/hanc00l/nemo_go/v3/pkg/core"
	"github.com/hanc00l/nemo_go/v3/pkg/db"
	"github.com/hanc00l/nemo_go/v3/pkg/logging"
	"github.com/hanc00l/nemo_go/v3/pkg/resource"
	"github.com/hanc00l/nemo_go/v3/pkg/task/execute"
	"github.com/hanc00l/nemo_go/v3/pkg/utils"
	"github.com/tidwall/pretty"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

type Nuclei struct {
	Config  execute.PocscanConfig
	IsProxy bool
}

func (n *Nuclei) LoadPocFiles() (pocFiles []string) {
	pocBase := filepath.Join(conf.GetRootPath(), "thirdparty/nuclei/nuclei-templates")
	//统一路径为“/”
	if runtime.GOOS == "windows" {
		pocBase = strings.ReplaceAll(pocBase, "\\", "/")
	}
	err := filepath.Walk(pocBase,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				logging.RuntimeLog.Error(err)
				return err
			}
			//统一路径为“/”
			if runtime.GOOS == "windows" {
				path = strings.ReplaceAll(path, "\\", "/")
			}
			if path == pocBase {
				return nil
			}
			// 替换去除路径
			pocFile := strings.Replace(path, fmt.Sprintf("%s/", pocBase), "", 1)
			// 忽略.开头的隐藏目录
			if strings.HasPrefix(pocFile, ".") || strings.HasPrefix(info.Name(), ".") {
				return nil
			}
			if info.IsDir() || strings.HasSuffix(path, ".yaml") {
				pocFiles = append(pocFiles, pocFile)
			}
			return nil
		})
	if err != nil {
		logging.RuntimeLog.Error(err)
	}
	sort.Strings(pocFiles)
	return

}

func (n *Nuclei) IsExecuteFromCmd() bool {
	return true
}

func (n *Nuclei) GetExecuteCmd() string {
	return filepath.Join(conf.GetAbsRootPath(), "thirdparty/nuclei", utils.GetThirdpartyBinNameByPlatform(utils.Nuclei))

}

func (n *Nuclei) GetExecuteArgs(inputTempFile, outputTempFile string) (cmdArgs []string) {
	/*RATE-LIMIT:
	  -rl, -rate-limit int            maximum number of requests to send per second (default 150)
	  -rlm, -rate-limit-minute int    maximum number of requests to send per minute
	  -bs, -bulk-size int             maximum number of hosts to be analyzed in parallel per template (default 25)
	  -c, -concurrency int            maximum number of templates to be executed in parallel (default 25)
	  -hbs, -headless-bulk-size int   maximum number of headless hosts to be analyzed in parallel per template (default 10)
	  -hc, -headless-concurrency int  maximum number of headless templates to be executed in parallel (default 10)
	*/
	cmdArgs = append(
		cmdArgs,
		"--timeout", "5", "-no-color", "-disable-update-check", "-silent",
		"-t", filepath.Join(conf.GetAbsRootPath(), "thirdparty/nuclei/nuclei-templates", n.Config.PocFile),
		"-j", "-o", outputTempFile, "-l", inputTempFile,
	)
	if n.IsProxy {
		if proxy := conf.GetProxyConfig(); proxy != "" {
			cmdArgs = append(cmdArgs, "-p", proxy)
		} else {
			logging.RuntimeLog.Warning("获取代理配置失败或禁用了代理功能，代理被跳过")
			logging.CLILog.Warning("get proxy config fail or disabled by worker,skip proxy!")
		}
	}
	return
}

func (n *Nuclei) GetRequiredResources() (re []core.RequiredResource) {
	re = append(re, core.RequiredResource{
		Category: resource.NucleiCategory,
		Name:     utils.GetThirdpartyBinNameByPlatform(utils.Nuclei),
	})
	re = append(re, core.RequiredResource{
		Category: resource.NucleiCategory,
		Name:     "nuclei-templates",
	})
	return
}

func (n *Nuclei) Run(target []string) (result Result) {
	//TODO implement me
	panic("implement me")
}

func (n *Nuclei) ParseContentResult(content []byte) (result Result) {
	lines := bytes.Split(content, []byte{'\n'})
	for _, line := range lines {
		var xr NucleiJSONResult
		err := json.Unmarshal(line, &xr)
		if err != nil {
			logging.RuntimeLog.Error(err.Error())
			continue
		}
		host := utils.ParseHost(xr.Host)
		if host == "" {
			continue
		}
		result.VulResult = append(result.VulResult, db.VulDocument{
			Authority: xr.Host,
			Host:      host,
			Url:       xr.URL,
			PocFile:   xr.TemplateID,
			Source:    "nuclei",
			Extra:     string(pretty.Pretty(line)),
		})
	}
	return
}
