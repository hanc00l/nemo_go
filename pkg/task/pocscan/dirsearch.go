package pocscan

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"github.com/evilsocket/brutemachine"
	"github.com/evilsocket/dirsearch"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/custom"
	"github.com/hanc00l/nemo_go/pkg/utils"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
)

type Dirsearch struct {
	Config Config
	Result []Result

	// resultUrl 暂存dirsearch的结果
	resultUrl []string
	// resultMutex 协程互斥锁
	resultMutex sync.RWMutex
	// blacklist
	http403BlackList []string
	http400BlackList []string
}

// DirsearcchResult represents the resul of each HEAD request to a given URL.
type DirsearcchResult struct {
	url      string
	status   int
	location string
	err      error
}

var (
	m         *brutemachine.Machine
	errorsInt = uint64(0)
	base, ext string
	wordlist  = filepath.Join(conf.GetRootPath(), "thirdparty/dict/dicc.txt")
	consumers = 8
	only200   = false
	maxerrors = 20
)

// NewDirsearch 创建Dirsearch对象
func NewDirsearch(config Config) *Dirsearch {
	d := &Dirsearch{Config: config}
	d.http403BlackList = d.readBlankList("403_blacklist.txt")
	d.http400BlackList = d.readBlankList("400_blacklist.txt")

	return d
}

// readBlankList 从文件中读取列表
func (d *Dirsearch) readBlankList(file string) (blackList []string) {
	inputFile, err := os.Open(filepath.Join(conf.GetRootPath(), "thirdparty/dict", file))
	if err != nil {
		logging.CLILog.Error(err)
		return
	}
	scanner := bufio.NewScanner(inputFile)
	for scanner.Scan() {
		text := strings.ToLower(scanner.Text())
		if text == "" {
			continue
		}
		blackList = append(blackList, text)
	}
	inputFile.Close()
	return
}

// Do 执行Dirsearch
func (d *Dirsearch) Do() {
	btc := custom.NewBlackTargetCheck(custom.CheckAll)
	for _, line := range strings.Split(d.Config.Target, ",") {
		target := strings.TrimSpace(line)
		if target == "" {
			continue
		}
		if btc.CheckBlack(utils.HostStrip(target)) {
			logging.RuntimeLog.Warningf("%s is in blacklist,skip...", target)
			continue
		}
		for _, protocol := range []string{"http", "https"} {
			//清空上一次暂存的结果
			d.resultUrl = []string{}
			url := fmt.Sprintf("%s://%s/", protocol, target) //注意最后要加/
			d.RunDirsearch(url)
			//保存结果
			if len(d.resultUrl) > 0 {
				d.Result = append(d.Result, Result{
					Target:      target,
					Url:         url,
					PocFile:     d.Config.PocFile,
					Source:      "dirsearch",
					Extra:       strings.Join(d.resultUrl, "\n"),
					WorkspaceId: d.Config.WorkspaceId,
				})
			}
		}
	}
}

// RunDirsearch 对一个url执行Dirsearch
func (d *Dirsearch) RunDirsearch(url string) {
	if d.initDirsearch(url) == false {
		return
	}
	m = brutemachine.New(consumers, wordlist, d.DoRequest, d.OnResult)
	logging.CLILog.Infof("start dirsearch:%s", url)
	if err := m.Start(); err != nil {
		logging.CLILog.Error(err)
	}
	m.Wait()
	printStats(url)
}

// DoRequest 执行一次请求
func (d *Dirsearch) DoRequest(page string) interface{} {
	url := strings.Replace(fmt.Sprintf("%s%s", base, page), "%EXT%", ext, -1)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	// Do not verify certificates, do not follow redirects.
	client := &http.Client{
		Transport: tr,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}}

	// Create HEAD request with random user agent.
	req, _ := http.NewRequest("HEAD", url, nil)
	req.Header.Set("User-Agent", dirsearch.GetRandomUserAgent())

	if resp, err := client.Do(req); err == nil {
		resp.Body.Close()
		if resp.StatusCode == 200 || !only200 {
			return DirsearcchResult{url, resp.StatusCode, resp.Header.Get("Location"), nil}
		}
	} else {
		atomic.AddUint64(&errorsInt, 1)
	}
	client.CloseIdleConnections()

	return nil
}

// OnResult 请求结果处理
func (d *Dirsearch) OnResult(res interface{}) {
	result, ok := res.(DirsearcchResult)
	if !ok {
		logging.CLILog.Error("Error while converting result.")
		return
	}

	switch {
	// error not due to 404 response
	case result.err != nil && result.status != 404:
		logging.CLILog.Infof("[???] %s : %v", result.url, result.err)
	// 2xx
	case result.status >= 200 && result.status < 300:
		logging.CLILog.Infof("[%d] %s", result.status, result.url)
		d.resultMutex.Lock()
		d.resultUrl = append(d.resultUrl, fmt.Sprintf("[%d] %s", result.status, result.url))
		d.resultMutex.Unlock()
	// 3xx
	case !only200 && result.status >= 300 && result.status < 400:
		logging.CLILog.Infof("[%d] %s -> %s", result.status, result.url, result.location)
		d.resultMutex.Lock()
		d.resultUrl = append(d.resultUrl, fmt.Sprintf("[%d] %s -> %s", result.status, result.url, result.location))
		d.resultMutex.Unlock()
	// 4xx
	case result.status >= 400 && result.status < 500 && result.status != 404:
		// 401、403
		if result.status > 400 && checkBlackList(result.url, d.http403BlackList) == false && checkBlackList(result.url, d.http400BlackList) == false {
			logging.CLILog.Infof("[%d] %s", result.status, result.url)
			d.resultMutex.Lock()
			d.resultUrl = append(d.resultUrl, fmt.Sprintf("[%d] %s", result.status, result.url))
			d.resultMutex.Unlock()
		}
	// 5xx
	case result.status >= 500 && result.status < 600:
		logging.CLILog.Infof("[%d] %s", result.status, result.url)
	}
}

// Do some initialization.
// NOTE: We can't call this in the 'init' function otherwise
// flags are gonna be mandatory for unit test modules.
func (d *Dirsearch) initDirsearch(url string) bool {
	base = url
	ext = d.Config.PocFile
	if d.checkHttp(url) == false {
		logging.CLILog.Infof("check %s http fail,skip... ", url)
		return false
	}
	return true
}

// checkHttp 检查指定url的http请求是否正常
func (d *Dirsearch) checkHttp(url string) bool {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	// Do not verify certificates, do not follow redirects.
	client := &http.Client{
		Transport: tr,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}}
	// Create HEAD request with random user agent.
	req, _ := http.NewRequest("HEAD", url, nil)
	req.Header.Set("User-Agent", dirsearch.GetRandomUserAgent())

	if _, err := client.Do(req); err == nil {
		return true
	} else {
		return false
	}
}

// printStats Print some stats
func printStats(url string) {
	m.UpdateStats()
	logging.CLILog.Infof("%s -> Requests:%d, Errors:%d, Results:%d, Time:%fs,Req/s: %f",
		url, m.Stats.Execs, errorsInt, m.Stats.Results, m.Stats.Total.Seconds(), m.Stats.Eps)
}

// checkBlackList 检查是否是黑名单
func checkBlackList(location string, blacklist []string) bool {
	u, err := url.Parse(location)
	if err != nil {
		return false
	}
	for _, s := range blacklist {
		if strings.Index(u.Path, s) >= 0 {
			return true
		}
	}
	return false
}
