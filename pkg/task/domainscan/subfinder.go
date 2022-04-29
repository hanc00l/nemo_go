package domainscan

import (
	"bufio"
	"bytes"
	"context"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/utils"
	"github.com/projectdiscovery/subfinder/v2/pkg/passive"
	"github.com/projectdiscovery/subfinder/v2/pkg/resolve"
	"github.com/projectdiscovery/subfinder/v2/pkg/runner"
	"github.com/remeh/sizedwaitgroup"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type SubFinder struct {
	Config Config
	Result Result
}

var defaultResolvers []string

// NewSubFinder 创建subfinder
func NewSubFinder(config Config) *SubFinder {
	loadDefaultResolver()
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
	//subfinder的options
	options := &runner.Options{
		Verbose:            true,
		Threads:            10,                              // Thread controls the number of threads to use for active enumerations
		Timeout:            30,                              // Timeout is the seconds to wait for sources to respond
		MaxEnumerationTime: 10,                              // MaxEnumerationTime is the maximum amount of time in mins to wait for enumeration
		Resolvers:          defaultResolvers,                // Use the default list of resolvers by marshaling it to the config
		Sources:            passive.DefaultSources,          // Use the default list of passive sources
		AllSources:         passive.DefaultAllSources,       // Use the default list of all passive sources
		Recursive:          passive.DefaultRecursiveSources, // Use the default list of recursive sources
		Providers:          &runner.Providers{},             // Use empty api keys for all providers
	}
	if len(defaultResolvers)==0{
		options.Resolvers = resolve.DefaultResolvers
	}
	//读取provider配置文件，参见https://github.com/projectdiscovery/subfinder#post-installation-instructions
	if len(conf.GlobalWorkerConfig().Domainscan.ProviderConfig) > 0 {
		providerFile := filepath.Join(conf.GetRootPath(), "thirdparty/dict", conf.GlobalWorkerConfig().Domainscan.ProviderConfig)
		if utils.CheckFileExist(providerFile) {
			err := options.Providers.UnmarshalFrom(providerFile)
			if err != nil {
				logging.RuntimeLog.Error(err)
			}
		}else{
			logging.RuntimeLog.Errorf("provider-config file not exist:%s",providerFile)
		}
	}
	//执行subfinder
	runnerInstance, err := runner.NewRunner(options)
	buf := bytes.Buffer{}
	err = runnerInstance.EnumerateSingleDomain(context.Background(), domain, []io.Writer{&buf})
	if err != nil {
		logging.RuntimeLog.Error(err)
		return
	}
	//读取结果
	data, err := io.ReadAll(&buf)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return
	}
	s.parseResultContent(data)
}

// parseResult 解析子域名枚举结果文件
func (s *SubFinder) parseResult(outputTempFile string) {
	content, err := os.ReadFile(outputTempFile)
	if err != nil {
		return
	}

	s.parseResultContent(content)
}

// parseResult 解析子域名枚举结果
func (s *SubFinder) parseResultContent(content []byte) {

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
// loadDefaultResolver  读取Resolvers
func loadDefaultResolver(){
	inputFile, err := os.Open(filepath.Join(conf.GetRootPath(), "thirdparty/dict", conf.GlobalWorkerConfig().Domainscan.Resolver))
	if err != nil {
		logging.RuntimeLog.Errorf("Could not read default resolver: %s\n", err)
		return
	}
	inputFile.Close()
	scanner := bufio.NewScanner(inputFile)
	for scanner.Scan() {
		text := strings.ToLower(scanner.Text())
		if text == "" {
			continue
		}
		defaultResolvers = append(defaultResolvers, text)
	}
}