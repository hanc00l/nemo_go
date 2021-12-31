package domainscan

import (
	"encoding/json"
	"fmt"
	"github.com/hanc00l/crawlergo/pkg"
	"github.com/hanc00l/crawlergo/pkg/config"
	"github.com/hanc00l/crawlergo/pkg/logger"
	model2 "github.com/hanc00l/crawlergo/pkg/model"
	"github.com/hanc00l/nemo_go/pkg/logging"
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

// NewCrawler 创建Crawler对象
func NewCrawler(config Config) *Crawler {
	return &Crawler{Config: config}
}

// Do 执行爬虫获取子域名
func (c *Crawler) Do() {
	c.Result.DomainResult = make(map[string]*DomainResult)
	swg := sizedwaitgroup.New(crawlerThreadNumber)
	for _, line := range strings.Split(c.Config.Target, ",") {
		domain := strings.TrimSpace(line)
		if domain == "" || utils.CheckIPV4(domain) || utils.CheckIPV4Subnet(domain) {
			continue
		}
		swg.Add()
		go func(d string) {
			c.RunCrawler(d)
			swg.Done()
		}(fmt.Sprintf("%s://%s", "http", domain))
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
		logging.CLILog.Error(err)
		return
	}
	req := model2.GetRequest(config.GET, url, option)
	targets = append(targets, &req)
	// 开始爬虫任务
	task, err := pkg.NewCrawlerTask(targets, taskConfig)
	if err != nil {
		logging.RuntimeLog.Error(fmt.Sprintf("create crawler task failed:%s.", domainUrl))
		return
	}
	go handleExit(task, signalChan)
	logging.CLILog.Info(fmt.Sprintf("Start crawling %s...", domainUrl))
	task.Run()
	result := task.Result
	logging.CLILog.Info(fmt.Sprintf("Task finished, %d results, %d requests, %d subdomains, %d domains found.",
		len(result.ReqList), len(result.AllReqList), len(result.SubDomainList), len(result.AllDomainList)))
	// 结果解析
	c.parseResult(result.SubDomainList)
}

// parseResult 解析子域名枚举结果文件
func (c *Crawler) parseResult(result []string) {
	for _, line := range result {
		domain := strings.TrimSpace(line)
		if domain == "" {
			continue
		}
		logging.CLILog.Println(domain)
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
	taskConfig := pkg.TaskConfig{}
	taskConfig.ChromiumPath = findExecPath()
	taskConfig.ExtraHeadersString = fmt.Sprintf(`{"User-Agent": "%s"}`, config.DefaultUA)
	taskConfig.MaxTabsCount = config.MaxCrawlCount
	taskConfig.FilterMode = "smart"
	taskConfig.IncognitoContext = true
	taskConfig.MaxTabsCount = 8
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

// getDefaultChromePath 获取默认安装的chrome路径
func getDefaultChromePath() (path string) {
	chromeDarwin := []string{"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"}
	chromeLinux := []string{
		"/usr/lib/chromium-browser/chromium-browser",             //ubuntu 18.04LTS
		"/snap/chromium/current/usr/lib/chromium-browser/chrome", //ubuntu 20.04LTS
		"/opt/google/chrome",                                     //kali 2021.3
	}
	if runtime.GOOS == "darwin" {
		for _, p := range chromeDarwin {
			if utils.CheckFileExist(p) {
				return p
			}
		}
	} else if runtime.GOOS == "linux" {
		for _, p := range chromeLinux {
			if utils.CheckFileExist(p) {
				return p
			}
		}
	}
	return
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
