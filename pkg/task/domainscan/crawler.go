package domainscan

import (
	"encoding/json"
	"fmt"
	"github.com/Qianlitp/crawlergo/pkg"
	"github.com/Qianlitp/crawlergo/pkg/config"
	"github.com/Qianlitp/crawlergo/pkg/logger"
	model2 "github.com/Qianlitp/crawlergo/pkg/model"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/custom"
	"github.com/hanc00l/nemo_go/pkg/utils"
	"github.com/remeh/sizedwaitgroup"
	"github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
)

type Crawler struct {
	Config Config
	Result Result
}

type UrlResponse struct {
	URL              string
	Method           string
	Headers          map[string]interface{}
	PostData         string
	StatusCode       int
	ContentLength    int64
	RedirectLocation string
}

// NewCrawler 创建Crawler对象
func NewCrawler(config Config) *Crawler {
	return &Crawler{Config: config}
}

// Do 执行爬虫获取子域名
func (c *Crawler) Do() {
	c.Result.DomainResult = make(map[string]*DomainResult)
	swg := sizedwaitgroup.New(crawlerThreadNumber[conf.WorkerPerformanceMode])
	blackDomain := custom.NewBlackTargetCheck(custom.CheckDomain)

	for _, line := range strings.Split(c.Config.Target, ",") {
		domain := strings.TrimSpace(line)
		if domain == "" || utils.CheckIPOrSubnet(domain) {
			continue
		}
		if blackDomain.CheckBlack(domain) {
			logging.RuntimeLog.Warningf("%s is in blacklist,skip...", domain)
			continue
		}
		protocol := utils.GetProtocol(domain, 5)
		swg.Add()
		go func(d string) {
			defer swg.Done()
			c.RunCrawler(d)
		}(fmt.Sprintf("%s://%s", protocol, domain))
	}
	swg.Wait()
}

// RunCrawler 爬取一个网站
func (c *Crawler) RunCrawler(domainUrl string) {
	taskConfig := initTaskConfig()
	option := getOption(&taskConfig)
	if taskConfig.ChromiumPath == "" {
		logging.RuntimeLog.Error("no chrome or chromium-browser found in default path")
		logging.CLILog.Error("no chrome or chromium-browser found in default path")
		return
	}
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	// 设置Crawler的日志输出级别
	logger.Logger.SetLevel(logrus.InfoLevel)
	logger.Logger.SetFormatter(logging.GetCustomLoggerFormatter())
	// 格式化target
	var targets []*model2.Request
	url, err := model2.GetUrl(domainUrl)
	if err != nil {
		logging.RuntimeLog.Error(err)
		logging.CLILog.Error(err)
		return
	}
	req := model2.GetRequest(config.GET, url, option)
	targets = append(targets, &req)
	// 开始爬虫任务
	task, err := pkg.NewCrawlerTask(targets, taskConfig)
	if err != nil {
		logging.RuntimeLog.Error(fmt.Sprintf("create crawler task failed:%s.", domainUrl))
		logging.CLILog.Error(err)
		return
	}
	go handleExit(task, signalChan)
	logging.CLILog.Info(fmt.Sprintf("start crawling %s...", domainUrl))
	task.Run()
	result := task.Result
	logging.CLILog.Info(fmt.Sprintf("task finished, %d results, %d requests, %d subdomains, %d domains found.",
		len(result.ReqList), len(result.AllReqList), len(result.SubDomainList), len(result.AllDomainList)))
	// 结果解析
	c.parseResult(result.SubDomainList)
}

// parseResult 解析子域名枚举结果文件
func (c *Crawler) parseResult(result []string) {
	blackDomain := custom.NewBlackTargetCheck(custom.CheckDomain)
	for _, line := range result {
		domain := strings.TrimSpace(line)
		if domain == "" {
			continue
		}
		if blackDomain.CheckBlack(domain) {
			logging.RuntimeLog.Warningf("%s is in blacklist,skip...", domain)
			continue
		}
		logging.CLILog.Info(domain)
		if !c.Result.HasDomain(domain) {
			c.Result.SetDomain(domain)
		}
	}
}

func getOption(taskConfig *pkg.TaskConfig) model2.Options {
	var option model2.Options
	if taskConfig.ExtraHeadersString != "" {
		err := json.Unmarshal([]byte(taskConfig.ExtraHeadersString), &taskConfig.ExtraHeaders)
		if err != nil {
			logging.RuntimeLog.Error("custom headers can't be Unmarshal.")
		}
		option.Headers = taskConfig.ExtraHeaders
	}
	return option
}

// initTaskConfig 初始化爬虫参数
func initTaskConfig() pkg.TaskConfig {
	/*	type TaskConfig struct {
		MaxCrawlCount           int    // 最大爬取的数量
		FilterMode              string // simple、smart、strict
		ExtraHeaders            map[string]interface{}
		ExtraHeadersString      string
		AllDomainReturn         bool // 全部域名收集
		SubDomainReturn         bool // 子域名收集
		NoHeadless              bool // headless模式
		DomContentLoadedTimeout time.Duration
		TabRunTimeout           time.Duration     // 单个标签页超时
		PathByFuzz              bool              // 通过字典进行Path Fuzz
		FuzzDictPath            string            //Fuzz目录字典
		PathFromRobots          bool              // 解析Robots文件找出路径
		MaxTabsCount            int               // 允许开启的最大标签页数量 即同时爬取的数量
		ChromiumPath            string            // Chromium的程序路径  `/home/zhusiyu1/chrome-linux/chrome`
		EventTriggerMode        string            // 事件触发的调用方式： 异步 或 顺序
		EventTriggerInterval    time.Duration     // 事件触发的间隔
		BeforeExitDelay         time.Duration     // 退出前的等待时间，等待DOM渲染，等待XHR发出捕获
		EncodeURLWithCharset    bool              // 使用检测到的字符集自动编码URL
		IgnoreKeywords          []string          // 忽略的关键字，匹配上之后将不再扫描且不发送请求
		Proxy                   string            // 请求代理
		CustomFormValues        map[string]string // 自定义表单填充参数
		CustomFormKeywordValues map[string]string // 自定义表单关键词填充内容
	}*/
	taskConfig := pkg.TaskConfig{}
	taskConfig.ChromiumPath = findExecPath()
	taskConfig.ExtraHeadersString = fmt.Sprintf(`{"User-Agent": "%s"}`, config.DefaultUA)
	taskConfig.MaxTabsCount = config.MaxCrawlCount
	taskConfig.FilterMode = config.SmartFilterMode
	taskConfig.MaxTabsCount = config.MaxTabsCount
	taskConfig.PathByFuzz = false
	taskConfig.TabRunTimeout = config.TabRunTimeout
	taskConfig.DomContentLoadedTimeout = config.DomContentLoadedTimeout
	taskConfig.EventTriggerMode = config.EventTriggerAsync
	taskConfig.EventTriggerInterval = config.EventTriggerInterval
	taskConfig.BeforeExitDelay = config.BeforeExitDelay
	taskConfig.NoHeadless = false
	taskConfig.EncodeURLWithCharset = false
	taskConfig.PathFromRobots = false
	taskConfig.IgnoreKeywords = config.DefaultIgnoreKeywords

	return taskConfig
}

// handleExit 退出清理
func handleExit(t *pkg.CrawlerTask, signalChan chan os.Signal) {
	select {
	case <-signalChan:
		//fmt.Println("exit ...")
		t.Pool.Tune(1)
		t.Pool.Release()
		t.Browser.Close()
	}
}

// findExecPath tries to find the Chrome browser somewhere in the current
// system. It finds in different locations on different OS systems.
// It could perform a rather aggressive search. That may make it a bit slow,
// but it will only be run when creating a new ExecAllocator.
// forked from https://github.com/chromedp/chromedp/blob/master/allocate.go
func findExecPath() string {
	var locations []string
	switch runtime.GOOS {
	case "darwin":
		locations = []string{
			// Mac
			"/Applications/Chromium.app/Contents/MacOS/Chromium",
			"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
		}
	case "windows":
		locations = []string{
			// Windows
			"chrome",
			"chrome.exe", // in case PATHEXT is misconfigured
			`C:\Program Files (x86)\Google\Chrome\Application\chrome.exe`,
			`C:\Program Files\Google\Chrome\Application\chrome.exe`,
			filepath.Join(os.Getenv("USERPROFILE"), `AppData\Local\Google\Chrome\Application\chrome.exe`),
			filepath.Join(os.Getenv("USERPROFILE"), `AppData\Local\Chromium\Application\chrome.exe`),
		}
	default:
		locations = []string{
			// Unix-like
			"/usr/lib/chromium-browser/chromium-browser",             //ubuntu 18.04LTS
			"/snap/chromium/current/usr/lib/chromium-browser/chrome", //ubuntu 20.04LTS
			"/opt/google/chrome",                                     //kali 2021.3
			"headless_shell",
			"headless-shell",
			"chromium",
			"chromium-browser",
			"google-chrome",
			"google-chrome-stable",
			"google-chrome-beta",
			"google-chrome-unstable",
			"/usr/bin/google-chrome",
			"/usr/local/bin/chrome",
			"/snap/bin/chromium",
			"chrome",
		}
	}

	for _, path := range locations {
		found, err := exec.LookPath(path)
		if err == nil {
			return found
		}
	}
	// Fall back to something simple and sensible, to give a useful error
	// message.
	return ""
}
