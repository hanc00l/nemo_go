package fingerprint

import (
	"context"
	"fmt"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/hanc00l/nemo_go/v2/pkg/conf"
	"github.com/hanc00l/nemo_go/v2/pkg/logging"
	"github.com/hanc00l/nemo_go/v2/pkg/task/custom"
	"github.com/hanc00l/nemo_go/v2/pkg/task/domainscan"
	"github.com/hanc00l/nemo_go/v2/pkg/task/portscan"
	"github.com/hanc00l/nemo_go/v2/pkg/utils"
	"github.com/remeh/sizedwaitgroup"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	MaxWidth       = 1280 //1980
	MinHeight      = 800  //1080
	SavedWidth     = 1280 //1024
	SavedHeight    = 800  //768
	thumbnailWidth = 120
)

// 不直接调用ScreenShot
// 目前采用httpx调用chrome-headless的方式获取screenshot，并做相应的缩放与裁剪后获得screenshot

type ScreenShot struct {
	ResultPortScan   *portscan.Result
	ResultDomainScan *domainscan.Result
	ResultScreenShot ScreenshotResult
	DomainTargetPort map[string]map[int]struct{}
	OptimizationMode bool
}

type ScreenshotFileInfo struct {
	Domain   string `json:"domain"`
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
	Content  []byte `json:"content"`
}

// NewScreenShot 创建ScreenShot对象
func NewScreenShot() *ScreenShot {
	s := &ScreenShot{
		ResultScreenShot: ScreenshotResult{Result: make(map[string][]ScreenshotInfo)}}

	return s
}

// Do 执行任务
func (s *ScreenShot) Do() {
	swg := sizedwaitgroup.New(fpScreenshotThreadNum[conf.WorkerPerformanceMode])

	btc := custom.NewBlackTargetCheck(custom.CheckAll)
	if s.ResultPortScan != nil && s.ResultPortScan.IPResult != nil {
		for ipName, ipResult := range s.ResultPortScan.IPResult {
			if btc.CheckBlack(ipName) {
				logging.RuntimeLog.Warningf("%s is in blacklist,skip...", ipName)
				continue
			}
			for portNumber := range ipResult.Ports {
				if _, ok := blankPort[portNumber]; ok {
					continue
				}
				if s.OptimizationMode {
					if !CheckForHttpxFingerResult(ipName, "", portNumber, s.ResultPortScan, s.ResultDomainScan) {
						continue
					}
				}
				protocol := utils.GetProtocol(utils.FormatHostUrl("", ipName, portNumber), 5)
				swg.Add()
				go s.doScreenshotAndResize(&swg, ipName, portNumber, protocol)

			}
		}
	}
	if s.ResultDomainScan != nil && s.ResultDomainScan.DomainResult != nil {
		if s.DomainTargetPort == nil {
			s.DomainTargetPort = make(map[string]map[int]struct{})
		}
		for domain := range s.ResultDomainScan.DomainResult {
			if btc.CheckBlack(domain) {
				logging.RuntimeLog.Warningf("%s is in blacklist,skip...", domain)
				continue
			}
			//如果无域名对应的端口，默认80和443
			if _, ok := s.DomainTargetPort[domain]; !ok || len(s.DomainTargetPort[domain]) == 0 {
				s.DomainTargetPort[domain] = make(map[int]struct{})
				s.DomainTargetPort[domain][80] = struct{}{}
				s.DomainTargetPort[domain][443] = struct{}{}
			}
			for port := range s.DomainTargetPort[domain] {
				if _, ok := blankPort[port]; ok {
					continue
				}
				if s.OptimizationMode {
					if !CheckForHttpxFingerResult("", domain, port, s.ResultPortScan, s.ResultDomainScan) {
						continue
					}
				}
				protocol := utils.GetProtocol(utils.FormatHostUrl("", domain, port), 5)
				swg.Add()
				go s.doScreenshotAndResize(&swg, domain, port, protocol)
			}
		}
	}
	swg.Wait()
}

// LoadResult 获取screenshot的结果文件
func (s *ScreenShot) LoadResult() (result []ScreenshotFileInfo) {
	for domain, r := range s.ResultScreenShot.Result {
		for _, si := range r {
			sfi := ScreenshotFileInfo{
				Domain:   domain,
				Port:     si.Port,
				Protocol: si.Protocol,
			}
			var err error
			sfi.Content, err = os.ReadFile(si.FilePathName)
			if err != nil {
				logging.RuntimeLog.Error(err)
				logging.CLILog.Error(err)
				continue
			}
			result = append(result, sfi)
			os.Remove(si.FilePathName)
		}
	}
	return
}

// SaveFile 保存screenshot文件到本地
func (s *ScreenShot) SaveFile(localSavePath string, result []ScreenshotFileInfo) (count int) {
	for _, sfi := range result {
		// check
		if sfi.Port == 0 || sfi.Domain == "" || sfi.Protocol == "" || len(sfi.Content) == 0 {
			logging.RuntimeLog.Error("empty upload attribute")
			continue
		}
		if !utils.CheckIP(sfi.Domain) && !utils.CheckDomain(sfi.Domain) {
			logging.RuntimeLog.Errorf("invalid domain:%s", sfi.Domain)
			continue
		}
		if strings.Contains(sfi.Domain, "..") || strings.Contains(sfi.Domain, "/") {
			logging.RuntimeLog.Errorf("invalid domain:%s", sfi.Domain)
			continue
		}
		domainPath := filepath.Join(localSavePath, sfi.Domain)
		if !utils.MakePath(localSavePath) || !utils.MakePath(domainPath) {
			logging.RuntimeLog.Errorf("check upload path fail:%s", domainPath)
			continue
		}
		//保存文件
		fileName := filepath.Join(domainPath, fmt.Sprintf("%d_%s.png", sfi.Port, sfi.Protocol))
		err := os.WriteFile(fileName, sfi.Content, 0666)
		if err != nil {
			logging.RuntimeLog.Errorf("write file %s fail:%v", fileName, err)
			continue
		}
		//生成缩略图
		fileNameThumbnail := filepath.Join(domainPath, fmt.Sprintf("%d_%s_thumbnail.png", sfi.Port, sfi.Protocol))
		if utils.ReSizePicture(fileName, fileNameThumbnail, thumbnailWidth, 0) {
			count++
		} else {
			logging.RuntimeLog.Error("generate thumbnail picature fail")
		}
	}
	return
}

// LoadScreenshotFile 获取screenshot文件
func (s *ScreenShot) LoadScreenshotFile(workspaceGUID, domain string) (r []string) {
	if !utils.CheckDomain(domain) && !utils.CheckIP(domain) {
		return
	}
	files, _ := filepath.Glob(filepath.Join(conf.GlobalServerConfig().Web.WebFiles, workspaceGUID, "screenshot", domain, "*.png"))
	for _, file := range files {
		_, f := filepath.Split(file)
		if !strings.HasSuffix(f, "_thumbnail.png") {
			r = append(r, f)
		}
	}
	return
}

// doScreenshotAndResize 屏幕截图并进行缩放
func (s *ScreenShot) doScreenshotAndResize(swg *sizedwaitgroup.SizedWaitGroup, domain string, port int, protocol string) {
	defer swg.Done()

	u := utils.FormatHostUrl(protocol, domain, port)
	file1 := utils.GetTempPNGPathFileName()
	defer os.Remove(file1)
	logging.CLILog.Infof("headless-chrome screenshot -> %s", u)
	if DoFullScreenshot(u, file1) {
		fileResized := utils.GetTempPNGPathFileName()
		if utils.ReSizePicture(file1, fileResized, SavedWidth, SavedHeight) {
			si := ScreenshotInfo{
				Port:         port,
				Protocol:     protocol,
				FilePathName: fileResized,
			}
			s.ResultScreenShot.SetScreenshotInfo(domain, si)
		}
	}
}

// Delete 删除指定domain、IP下保存的screenshot文件
func (s *ScreenShot) Delete(workspaceGUID, domain string) bool {
	if !utils.CheckDomain(domain) && !utils.CheckIP(domain) {
		logging.RuntimeLog.Errorf("invalid domain:%s", domain)
		return false
	}
	domainPath := filepath.Join(conf.GlobalServerConfig().Web.WebFiles, workspaceGUID, "screenshot", domain)
	if err := os.RemoveAll(domainPath); err != nil {
		return false
	}
	return true
}

// DoFullScreenshot 调用chromedp执行截图
func DoFullScreenshot(url, path string) bool {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("ignore-certificate-errors", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-crash-reporter", true),
		chromedp.Flag("disable-notifications", true),
		chromedp.Flag("hide-scrollbars", true),
		chromedp.Flag("mute-audio", true),
		chromedp.Flag("incognito", true),
		chromedp.Flag("enable-features", "NetworkService"),
		chromedp.UserAgent(`Mozilla/5.0 (Windows NT 6.3; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/73.0.3683.103 Safari/537.36`),
		chromedp.WindowSize(MaxWidth, MinHeight),
	)
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()
	// 创建chrome实例
	ctx, cancel := chromedp.NewContext(
		allocCtx,
		chromedp.WithLogf(log.Printf),
	)
	defer cancel()
	// 创建超时时间
	ctx, cancel = context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	// 缓存对象
	var buf []byte
	// 运行截屏
	if err := chromedp.Run(ctx, fullScreenshot(url, 100, &buf)); err != nil {
		return false
	}
	// 保存文件
	if err := os.WriteFile(path, buf, 0644); err != nil {
		logging.RuntimeLog.Error(err)
		logging.CLILog.Error(err)
		return false
	}

	return true
}

// fullScreenshot 全屏截图
func fullScreenshot(url string, quality int64, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(url),
		////延时：等待有些页面有js自动跳转，待js跳转后再执行截图操作
		chromedp.Sleep(5 * time.Second),
		chromedp.ActionFunc(func(ctx context.Context) (err error) {
			*res, err = page.CaptureScreenshot().WithQuality(quality).WithClip(&page.Viewport{
				X:      0,
				Y:      0,
				Width:  MaxWidth,
				Height: MinHeight,
				Scale:  1,
			}).Do(ctx)
			if err != nil {
				logging.CLILog.Error(err)
				return err
			}
			return nil
		}),
	}
}
