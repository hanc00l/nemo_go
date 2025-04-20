package domainscan

import (
	"github.com/hanc00l/nemo_go/v3/pkg/conf"
	"github.com/hanc00l/nemo_go/v3/pkg/core"
	"github.com/hanc00l/nemo_go/v3/pkg/logging"
	"github.com/hanc00l/nemo_go/v3/pkg/resource"
	"github.com/hanc00l/nemo_go/v3/pkg/task/execute"
	"github.com/hanc00l/nemo_go/v3/pkg/utils"
	mdrunner "github.com/projectdiscovery/shuffledns/pkg/runner"

	"github.com/remeh/sizedwaitgroup"
	"os"
	"path/filepath"
	"strings"
)

type Massdns struct {
	Config execute.DomainscanConfig
}

func (m *Massdns) GetRequiredResources() (re []core.RequiredResource) {
	re = append(re, core.RequiredResource{
		Category: resource.MassdnsCategory,
		Name:     utils.GetThirdpartyBinNameByPlatform(utils.MassDns),
	})
	re = append(re, core.RequiredResource{
		Category: resource.DictCategory,
		Name:     "resolver.txt",
	})
	re = append(re, core.RequiredResource{
		Category: resource.DictCategory,
		Name:     "subnames.txt",
	})
	re = append(re, core.RequiredResource{
		Category: resource.DictCategory,
		Name:     "subnames_medium.txt",
	})
	//CDN
	re = append(re, core.RequiredResource{
		Category: resource.Geolite2Category,
		Name:     "GeoLite2-ASN.mmdb",
	})
	return
}

func (m *Massdns) IsExecuteFromCmd() bool {
	return false
}

func (m *Massdns) GetExecuteCmd() string {
	return ""
}

func (m *Massdns) GetExecuteArgs(inputTempFile, outputTempFile string) (cmdArgs []string) {
	return
}

func (m *Massdns) Run(target []string) (result Result) {
	result.DomainResult = make(map[string]*DomainResult)

	swg := sizedwaitgroup.New(massdnsThreadNumber[conf.WorkerPerformanceMode])
	for _, line := range target {
		domain := strings.TrimSpace(line)
		if domain == "" || utils.CheckIPOrSubnet(domain) {
			continue
		}
		if utils.CheckIPOrSubnet(domain) {
			logging.RuntimeLog.Warningf("目标是IP地址: %s，skip...", domain)
			continue
		}
		swg.Add()
		go func(d string, r *Result) {
			defer swg.Done()
			m.RunMassdns(domain, r)
		}(domain, &result)
	}
	swg.Wait()

	return
}

func (m *Massdns) ParseContentResult(content []byte) (result Result) {
	//TODO implement me
	panic("implement me")
}

// parseResult 解析子域名枚举结果文件
func (m *Massdns) parseResult(outputTempFile string, result *Result) {
	content, err := os.ReadFile(outputTempFile)
	if err != nil {
		return
	}
	resolver := Resolve{}
	for _, line := range strings.Split(string(content), "\n") {
		domain := strings.TrimSpace(line)
		if domain == "" {
			continue
		}
		if !result.HasDomain(domain) {
			result.SetDomain(domain)
		}
		resolver.RunResolve(domain, result)
	}
}

// RunMassdns runs the massdns tool on the list of inputs
func (m *Massdns) RunMassdns(domain string, result *Result) {
	tempOutputFile := utils.GetTempPathFileName()
	defer func(name string) {
		_ = os.Remove(name)
	}(tempOutputFile)

	tempDir, err := os.MkdirTemp("", utils.GetRandomString2(8))
	if err != nil {
		logging.RuntimeLog.Error(err)
		return
	}
	defer func(path string) {
		_ = os.RemoveAll(path)
	}(tempDir)

	wordFile := "subnames.txt"
	if m.Config.WordlistFile == "medium" {
		wordFile = "subnames_medium.txt"
	}
	options := &mdrunner.Options{
		Directory:          tempDir,
		Domains:            []string{domain},
		SubdomainsList:     "",
		ResolversFile:      filepath.Join(conf.GetRootPath(), "thirdparty/dict/resolver.txt"),
		Wordlist:           filepath.Join(conf.GetRootPath(), "thirdparty/dict", wordFile),
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
	}
	massdnsRunner, err := mdrunner.New(options)
	if err != nil {
		logging.RuntimeLog.Errorf("创建massdns失败: %s", err)
		return
	}
	massdnsRunner.RunEnumeration()
	m.parseResult(tempOutputFile, result)
}
