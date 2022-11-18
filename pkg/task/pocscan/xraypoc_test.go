package pocscan

import (
	"fmt"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/utils"
	xv2 "github.com/hanc00l/nemo_go/pkg/xraypocv2"
	"github.com/remeh/sizedwaitgroup"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

var loadMutex sync.Mutex

func Test1(t *testing.T) {
	config := XrayPocConfig{
		IPPort: make(map[string][]int),
		Domain: make(map[string]struct{}),
	}
	config.IPPort["127.0.0.1"] = append(config.IPPort["127.0.0.1"], 8080)
	config.IPPort["127.0.0.1"] = append(config.IPPort["127.0.0.1"], 8161)
	config.Domain["localhost:8848"] = struct{}{}
	p := NewXrayPoc(config)
	p.Do()
	t.Log(p.VulResult)
}

func Test2(t *testing.T) {
	config := XrayPocConfig{
		IPPort: make(map[string][]int),
		Domain: make(map[string]struct{}),
	}
	config.Domain["localhost:8080"] = struct{}{}
	p := NewXrayPoc(config)
	p.Do()
	t.Log(p.VulResult)
}

func Test3(t *testing.T) {
	swg := sizedwaitgroup.New(4)
	swg.Add()
	go testOne(&swg, "127.0.0.1:8080", t)
	swg.Add()
	go testOne(&swg, "127.0.0.1:8848", t)
	swg.Add()
	go testOne(&swg, "127.0.0.1:8000", t)
	swg.Add()
	go testOne(&swg, "127.0.0.1:8161", t)

	swg.Wait()
}

func Test4(t *testing.T) {
	swg := sizedwaitgroup.New(4)

	swg.Add()
	go testOneOne(&swg)
	swg.Add()
	go testOneOne2(&swg)
	swg.Add()
	go testOneOne(&swg)
	swg.Add()
	go testOneOne2(&swg)

	swg.Wait()
}

func testOneOne2(swg *sizedwaitgroup.SizedWaitGroup) {
	defer swg.Done()
	config := XrayPocConfig{
		IPPort: make(map[string][]int),
		Domain: make(map[string]struct{}),
	}
	config.IPPort["127.0.0.1"] = append(config.IPPort["127.0.0.1"], 8080)
	config.IPPort["127.0.0.1"] = append(config.IPPort["127.0.0.1"], 8000)
	config.Domain["localhost:8848"] = struct{}{}
	p := NewXrayPoc(config)
	p.Do()
	fmt.Println(p.VulResult)
}

func testOneOne(swg *sizedwaitgroup.SizedWaitGroup) {
	defer swg.Done()
	config := XrayPocConfig{
		IPPort: make(map[string][]int),
		Domain: make(map[string]struct{}),
	}
	config.IPPort["127.0.0.1"] = append(config.IPPort["127.0.0.1"], 8080)
	config.IPPort["127.0.0.1"] = append(config.IPPort["127.0.0.1"], 8161)
	config.Domain["localhost:8848"] = struct{}{}
	p := NewXrayPoc(config)
	p.Do()
	fmt.Println(p.VulResult)
}

func testOne(swg *sizedwaitgroup.SizedWaitGroup, url string, t *testing.T) {
	loadMutex.Lock()
	var pocAll [][]byte
	files, _ := filepath.Glob(filepath.Join(conf.GetRootPath(), conf.GlobalWorkerConfig().Pocscan.Xray.PocPath, "*.yml"))
	for _, file := range files {
		pocContent, err := os.ReadFile(file)
		if err != nil {
			logging.CLILog.Error(err)
			continue
		}
		pocAll = append(pocAll, pocContent)
	}
	loadMutex.Unlock()
	fmt.Println(url)
	protocol := utils.GetProtocol(url, 5)
	x := xv2.InitXrayV2Poc("", "", "")
	aa := x.RunXrayMultiPocByQuery(fmt.Sprintf("%s://%s", protocol, url), pocAll, []xv2.Content{})
	t.Log(aa)
	swg.Done()
}
