package pocscan

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/utils"
	"github.com/tidwall/pretty"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

type Nuclei struct {
	Config Config
	Result []Result
}

func NewNuclei(config Config) *Nuclei {
	return &Nuclei{Config: config}
}

func (n *Nuclei) Do() {
	resultTempFile := utils.GetTempPathFileName()
	inputTargetFile := utils.GetTempPathFileName()
	defer os.Remove(resultTempFile)
	defer os.Remove(inputTargetFile)

	urls := strings.Split(n.Config.Target, ",")
	var urlsFormatted []string
	//由于nuclei要求url要http或https开始：
	for _, u := range urls {
		if strings.HasPrefix(u, "http") == false {
			urlsFormatted = append(urlsFormatted, fmt.Sprintf("http://%s", u))
			urlsFormatted = append(urlsFormatted, fmt.Sprintf("https://%s", u))
		} else {
			urlsFormatted = append(urlsFormatted, u)
		}
	}
	err := os.WriteFile(inputTargetFile, []byte(strings.Join(urlsFormatted, "\n")), 0666)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		return
	}
	cmdBin := filepath.Join(conf.GetAbsRootPath(), "thirdparty/nuclei", utils.GetThirdpartyBinNameByPlatform(utils.Nuclei))
	var cmdArgs []string
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
		"--timeout", "5", "-no-color",
		"-c", fmt.Sprintf("%d", nucleiConcurrencyThreadNumber),
		"-bs", fmt.Sprintf("%d", nucleiConcurrencyThreadNumber),
		"-rl", fmt.Sprintf("%d", nucleiConcurrencyThreadNumber[conf.WorkerPerformanceMode]*6),
		"-t", filepath.Join(conf.GetAbsRootPath(), conf.GlobalWorkerConfig().Pocscan.Nuclei.PocPath, n.Config.PocFile),
		"-json", "-o", resultTempFile, "-l", inputTargetFile,
	)
	cmd := exec.Command(cmdBin, cmdArgs...)
	_, err = cmd.CombinedOutput()
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		return
	}
	n.parseNucleiResult(resultTempFile)
}

// parseNucleiContentResult 解析nuclei的运行结果
func (n *Nuclei) parseNucleiContentResult(content []byte) {
	var xr nucleiJSONResult
	err := json.Unmarshal(content, &xr)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		return
	}
	host := utils.HostStrip(xr.Host)
	if host == "" {
		return
	}
	n.Result = append(n.Result, Result{
		Target:  host,
		Url:     xr.Host,
		PocFile: xr.TemplateID,
		Source:  "nuclei",
		Extra:   string(pretty.Pretty(content)),
	})
}

// parseNucleiResult 解析nuclei的运行结果
func (n *Nuclei) parseNucleiResult(outputTempFile string) {
	inputFile, err := os.Open(outputTempFile)
	if err != nil {
		logging.RuntimeLog.Errorf("Could not read nuclei result: %s\n", err)
		return
	}
	defer inputFile.Close()
	scanner := bufio.NewScanner(inputFile)
	for scanner.Scan() {
		n.parseNucleiContentResult(scanner.Bytes())
	}
}

// LoadPocFile 加载poc文件列表
func (n *Nuclei) LoadPocFile() (pocs []string) {
	pocBase := filepath.Join(conf.GetRootPath(), conf.GlobalWorkerConfig().Pocscan.Nuclei.PocPath)
	//统一路径为“/”
	if runtime.GOOS == "windows" {
		pocBase = strings.ReplaceAll(pocBase, "\\", "/")
	}
	err := filepath.Walk(pocBase,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
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
				pocs = append(pocs, pocFile)
			}
			return nil
		})
	if err != nil {
		logging.RuntimeLog.Error(err)
	}
	sort.Strings(pocs)
	return
}
