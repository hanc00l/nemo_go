package fingerprint

import (
	"context"
	"fmt"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/hanc00l/nemo_go/pkg/comm"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/domainscan"
	"github.com/hanc00l/nemo_go/pkg/task/portscan"
	"github.com/hanc00l/nemo_go/pkg/utils"
	"github.com/remeh/sizedwaitgroup"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	MaxWidth    = 1980
	MinHeight   = 1080
	SavedWidth  = 1024
	SavedHeight = 0 //根据比例自动缩放

)

type ScreenShot struct {
	ResultPortScan   portscan.Result
	ResultDomainScan domainscan.Result
	ResultScreenShot ScreenshotResult
}

// NewScreenShot 创建ScreenShot对象
func NewScreenShot() *ScreenShot {
	s := &ScreenShot{
		ResultScreenShot: ScreenshotResult{Result: make(map[string][]ScreenshotInfo)}}

	return s
}

// Do 执行任务
func (s *ScreenShot) Do() {
	swg := sizedwaitgroup.New(fpScreenshotThreadNum)

	if s.ResultPortScan.IPResult != nil {
		bport := make(map[int]struct{})
		for _, p := range IgnorePort {
			bport[p] = struct{}{}
		}
		for ipName, ipResult := range s.ResultPortScan.IPResult {
			for portNumber, _ := range ipResult.Ports {
				if _, ok := bport[portNumber]; ok {
					continue
				}
				for _, protocol := range []string{"http", "https"} {
					swg.Add()
					go s.doScreenshotAndResize(&swg, ipName, portNumber, protocol)
				}
			}
		}
	}
	if s.ResultDomainScan.DomainResult != nil {
		for domain, _ := range s.ResultDomainScan.DomainResult {
			for _, protocol := range []string{"http", "https"} {
				portNumber := 80
				if protocol == "https" {
					portNumber = 443
				}
				swg.Add()
				go s.doScreenshotAndResize(&swg, domain, portNumber, protocol)
			}
		}
	}
	swg.Wait()
}

// UploadResult 将screenshot的结果上传到server
func (s *ScreenShot) UploadResult() string {
	var count int
	for domain, r := range s.ResultScreenShot.Result {
		for _, si := range r {
			sfi := comm.ScreenshotFileInfo{
				Domain:   domain,
				Port:     si.Port,
				Protocol: si.Protocol,
			}
			var err error
			sfi.Content, err = os.ReadFile(si.FilePathName)
			if err != nil {
				logging.RuntimeLog.Error(err)
				continue
			}
			if comm.DoUploadScreenshot(sfi) {
				count++
				os.Remove(si.FilePathName)
			}
		}
	}

	return fmt.Sprintf("screenshots:%d", count)
}

// LoadScreenshotFile 获取screenshot文件
func (s *ScreenShot) LoadScreenshotFile(domain string) (r []string) {
	if !utils.CheckDomain(domain) && !utils.CheckIPV4(domain) {
		return
	}
	files, _ := filepath.Glob(filepath.Join(conf.Nemo.Web.ScreenshotPath, domain, "*.png"))
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

	u := fmt.Sprintf("%s://%s:%d", protocol, domain, port)
	file1 := utils.GetTempPNGPathFileName()
	defer os.Remove(file1)
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
func (s *ScreenShot) Delete(domain string) bool {
	if !utils.CheckDomain(domain) && !utils.CheckIPV4(domain) {
		logging.RuntimeLog.Errorf("invalid domain:%s", domain)
		return false
	}
	domainPath := filepath.Join(conf.Nemo.Web.ScreenshotPath, domain)
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
		chromedp.Flag("enable-features","NetworkService"),
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
	if err := ioutil.WriteFile(path, buf, 0644); err != nil {
		logging.RuntimeLog.Error(err)
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
				return err
			}
			return nil
		}),
	}
}
