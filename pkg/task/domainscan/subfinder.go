package domainscan

import (
	"bufio"
	"context"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/utils"
	"github.com/projectdiscovery/subfinder/v2/pkg/passive"
	"github.com/projectdiscovery/subfinder/v2/pkg/runner"
	"github.com/remeh/sizedwaitgroup"
	"os"
	"path/filepath"
	"strings"
)

type SubFinder struct {
	Config Config
	Result Result
}

// NewSubFinder 创建subfinder
func NewSubFinder(config Config) *SubFinder {
	return &SubFinder{Config: config}
}

// Do 执行子域名枚举
func (s *SubFinder) Do() {
	s.Result.DomainResult = make(map[string]*DomainResult)
	swg := sizedwaitgroup.New(subfinderThreadNumber)

	for _, line := range strings.Split(s.Config.Target, ",") {
		domain := strings.TrimSpace(line)
		if domain == "" || utils.CheckIPV4(domain) || utils.CheckIPV4Subnet(domain) {
			continue
		}
		swg.Add()
		go func(d string) {
			s.RunSubFinder(d)
			swg.Done()
		}(domain)
	}
	swg.Wait()
}

// RunSubFinder 执行subfinder
func (s *SubFinder) RunSubFinder(domain string) {
	resultTempFile := utils.GetTempPathFileName()
	defer os.Remove(resultTempFile)

	var defaultResolvers []string
	inputFile, err := os.Open(filepath.Join(conf.GetRootPath(), "thirdparty/dict", conf.GlobalWorkerConfig().Domainscan.Resolver))
	if err != nil {
		logging.RuntimeLog.Errorf("Could not read bruteforce wordlist: %s\n", err)
		return
	}
	scanner := bufio.NewScanner(inputFile)
	for scanner.Scan() {
		text := strings.ToLower(scanner.Text())
		if text == "" {
			continue
		}
		defaultResolvers = append(defaultResolvers, text)
	}
	inputFile.Close()

	options := &runner.Options{
		Verbose:            true,
		Threads:            10,
		Timeout:            30,
		MaxEnumerationTime: 10,
		Domain:             domain,
		Output:             os.Stdout,
		OutputFile:         resultTempFile,
		YAMLConfig: runner.ConfigFile{
			Resolvers:  defaultResolvers,
			Sources:    passive.DefaultSources,
			AllSources: passive.DefaultAllSources,
			Recursive:  passive.DefaultRecursiveSources,
		},
	}
	newRunner, err := runner.NewRunner(options)
	if err != nil {
		logging.RuntimeLog.Error(err)
	}
	err = newRunner.RunEnumeration(context.Background())
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
	}
	s.parseResult(resultTempFile)
}

// parseResult 解析子域名枚举结果文件
func (s *SubFinder) parseResult(outputTempFile string) {
	content, err := os.ReadFile(outputTempFile)
	if err != nil {
		return
	}

	for _, line := range strings.Split(string(content), "\n") {
		domain := strings.TrimSpace(line)
		if domain == "" {
			continue
		}
		if !s.Result.HasDomain(domain) {
			s.Result.SetDomain(domain)
		}
	}
}
