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

func (n *Nuclei) loadPocFilesFromBasePath(pocBase string) (pocFiles []string) {
	basePath := filepath.Base(pocBase)
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
				pocFiles = append(pocFiles, fmt.Sprintf("%s/%s", basePath, pocFile))
			}
			return nil
		})
	if err != nil {
		logging.RuntimeLog.Error(err)
	}
	return
}

func (n *Nuclei) LoadPocFiles() (pocFiles []string) {
	sourcePaths := map[string]string{
		"some_nuclei_templates": "some_nuclei_templates",
		"nuclei-template":       "nuclei-templates",
	}
	for _, p := range sourcePaths {
		pocBase := filepath.Join(conf.GetAbsRootPath(), "thirdparty/nuclei", p)
		pocFiles = append(pocFiles, n.loadPocFilesFromBasePath(pocBase)...)
	}
	// 排序
	sort.Strings(pocFiles)

	return
}

func (n *Nuclei) IsExecuteFromCmd() bool {
	return true
}

func (n *Nuclei) GetDir() string {
	return filepath.Join(conf.GetAbsRootPath(), "thirdparty/nuclei")
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
		"-t", n.Config.PocFile,
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
		Category: resource.PocFileCategory,
		Name:     "nuclei-templates",
	})
	re = append(re, core.RequiredResource{
		Category: resource.PocFileCategory,
		Name:     "some_nuclei_templates",
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
		if len(line) == 0 {
			continue
		}
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
			Name:      xr.Info.Name,
			Severity:  xr.Info.Severity,
			Extra:     string(pretty.Pretty(line)),
		})
	}
	return
}
