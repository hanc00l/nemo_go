package domainscan

import (
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/utils"
	"github.com/projectdiscovery/shuffledns/pkg/runner"
	"github.com/remeh/sizedwaitgroup"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type Massdns struct {
	Config Config
	Result Result
}

// NewMassdns 创建Massdns对象
func NewMassdns(config Config) *Massdns {
	return &Massdns{Config: config}
}

// Do 执行Massdns任务
func (m *Massdns) Do() {
	m.Result.DomainResult = make(map[string]*DomainResult)
	swg := sizedwaitgroup.New(massdnsThreadNumber[conf.WorkerPerformanceMode])

	for _, line := range strings.Split(m.Config.Target, ",") {
		domain := strings.TrimSpace(line)
		if domain == "" || utils.CheckIPV4(domain) || utils.CheckIPV4Subnet(domain) {
			continue
		}
		swg.Add()
		go func(d string) {
			defer swg.Done()
			m.RunMassdns(domain)
		}(domain)
	}
	swg.Wait()
}

// parseResult 解析子域名枚举结果文件
func (m *Massdns) parseResult(outputTempFile string) {
	content, err := os.ReadFile(outputTempFile)
	if err != nil {
		return
	}

	for _, line := range strings.Split(string(content), "\n") {
		domain := strings.TrimSpace(line)
		if domain == "" {
			continue
		}
		if !m.Result.HasDomain(domain) {
			m.Result.SetDomain(domain)
		}
	}
}

// RunMassdns runs the massdns tool on the list of inputs
func (m *Massdns) RunMassdns(domain string) {
	//massdns不支持windows平台
	if runtime.GOOS == "windows" {
		logging.RuntimeLog.Error("Widnows don't support to run massdns!")
		return
	}
	tempOutputFile := utils.GetTempPathFileName()
	defer os.Remove(tempOutputFile)

	tempDir, err := os.MkdirTemp(filepath.Join(conf.GetRootPath(), "thirdparty/massdns/temp"), utils.GetRandomString2(8))
	if err != nil {
		return
	}
	defer os.RemoveAll(tempDir)

	options := &runner.Options{
		Directory:          tempDir,
		Domain:             domain,
		SubdomainsList:     "",
		ResolversFile:      filepath.Join(conf.GetRootPath(), "thirdparty/dict", conf.GlobalWorkerConfig().Domainscan.Resolver),
		Wordlist:           filepath.Join(conf.GetRootPath(), "thirdparty/dict", conf.GlobalWorkerConfig().Domainscan.Wordlist),
		MassdnsPath:        filepath.Join(conf.GetRootPath(), "thirdparty/massdns", utils.GetThirdpartyBinNameByPlatform(utils.MassDns)),
		Output:             tempOutputFile,
		Json:               false,
		Silent:             false,
		Version:            false,
		Retries:            5,
		Verbose:            true,
		NoColor:            true,
		Threads:            massdnsRunnerThreads[conf.WorkerPerformanceMode],
		MassdnsRaw:         "",
		WildcardThreads:    25,
		StrictWildcard:     true,
		WildcardOutputFile: "",
		Stdin:              false,
	}
	massdnsRunner, err := runner.New(options)
	if err != nil {
		logging.RuntimeLog.Errorf("Could not create runner: %s\n", err)
	}

	massdnsRunner.RunEnumeration()
	m.parseResult(tempOutputFile)
}
