package pocscan

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/custom"
	"github.com/hanc00l/nemo_go/pkg/utils"
	"github.com/tidwall/pretty"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
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

	blackDomain := custom.NewBlackDomain()
	blackIP := custom.NewBlackIP()
	urls := strings.Split(n.Config.Target, ",")
	var urlsFormatted []string
	//由于nuclei要求url要http或https开始，非http/https协议不进行漏洞检测，节约扫描时间
	for _, u := range urls {
		if utils.CheckDomain(utils.HostStrip(u)) && blackDomain.CheckBlack(utils.HostStrip(u)) {
			continue
		}
		if utils.CheckIPV4(utils.HostStrip(u)) && blackIP.CheckBlack(utils.HostStrip(u)) {
			continue
		}
		if strings.HasPrefix(u, "http") == false {
			protocol := getProtocol(u, 4)
			if protocol == "" || protocol == "tcp" {
				continue
			}
			urlsFormatted = append(urlsFormatted, fmt.Sprintf("%s://%s", protocol, u))
		} else {
			urlsFormatted = append(urlsFormatted, u)
		}
	}
	//没有需要检测的端口,直接返回
	if len(urlsFormatted) == 0 {
		return
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
		"-c", fmt.Sprintf("%d", nucleiConcurrencyThreadNumber[conf.WorkerPerformanceMode]),
		"-bs", fmt.Sprintf("%d", nucleiConcurrencyThreadNumber[conf.WorkerPerformanceMode]),
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
		Target:      host,
		Url:         xr.Host,
		PocFile:     xr.TemplateID,
		Source:      "nuclei",
		Extra:       string(pretty.Pretty(content)),
		WorkspaceId: n.Config.WorkspaceId,
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

// 获取协议，返回空“”（端口未开放），http/https,tcp
func getProtocol(host string, Timeout int64) string {
	conn, err := tls.DialWithDialer(&net.Dialer{Timeout: time.Duration(Timeout) * time.Second}, "tcp", host, &tls.Config{InsecureSkipVerify: true})
	defer func() {
		if conn != nil {
			conn.Close()
		}
	}()
	//端口未开放
	if err != nil && strings.Contains(err.Error(), "No connection could be made") {
		return ""
	}
	protocol := "http"
	if err == nil || strings.Contains(err.Error(), "handshake failure") {
		protocol = "https"
	}
	req, _ := http.NewRequest("GET", protocol+"://"+host, nil)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	Client := &http.Client{
		Transport: tr,
		Timeout:   time.Duration(Timeout) * time.Second,
	}
	_, err = Client.Do(req)
	if err != nil {
		protocol = "tcp"
	}
	return protocol
}
