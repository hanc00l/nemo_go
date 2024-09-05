package domainscan

import (
	"bytes"
	"github.com/hanc00l/nemo_go/v2/pkg/conf"
	"github.com/hanc00l/nemo_go/v2/pkg/logging"
	"github.com/hanc00l/nemo_go/v2/pkg/task/custom"
	"github.com/hanc00l/nemo_go/v2/pkg/utils"
	"github.com/remeh/sizedwaitgroup"
	"os"
	"os/exec"
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
	swg := sizedwaitgroup.New(subfinderThreadNumber[conf.WorkerPerformanceMode])
	blackDomain := custom.NewBlackTargetCheck(custom.CheckDomain)

	for _, line := range strings.Split(s.Config.Target, ",") {
		domain := strings.TrimSpace(line)
		if domain == "" || utils.CheckIPOrSubnet(domain) {
			continue
		}
		if blackDomain.CheckBlack(domain) {
			logging.RuntimeLog.Warningf("%s is in blacklist,skip...", domain)
			continue
		}
		swg.Add()
		go func(d string) {
			defer swg.Done()
			s.RunSubFinder(d)
		}(domain)
	}
	swg.Wait()
}

// RunSubFinder 执行subfinder
func (s *SubFinder) RunSubFinder(domain string) {
	resultTempFile := utils.GetTempPathFileName()
	defer os.Remove(resultTempFile)

	var cmdArgs []string
	cmdArgs = append(cmdArgs,
		"-d", domain, "-all", "-o", resultTempFile, "-disable-update-check",
		"-rlist", filepath.Join(conf.GetRootPath(), "thirdparty/dict", conf.GlobalWorkerConfig().Domainscan.Resolver),
		"-provider-config", filepath.Join(conf.GetRootPath(), "thirdparty/dict", conf.GlobalWorkerConfig().Domainscan.ProviderConfig),
		"-no-color", "-v",
		"-active", //RemoveWildcard
	)
	if s.Config.IsProxy {
		if proxy := conf.GetProxyConfig(); proxy != "" {
			cmdArgs = append(cmdArgs, "-proxy", proxy)
		} else {
			logging.RuntimeLog.Warning("get proxy config fail or disabled by worker,skip proxy!")
			logging.CLILog.Warning("get proxy config fail or disabled by worker,skip proxy!")
		}
	}
	binPath := filepath.Join(conf.GetRootPath(), "thirdparty/subfinder", utils.GetThirdpartyBinNameByPlatform(utils.Subfinder))
	cmd := exec.Command(binPath, cmdArgs...)
	var stderr bytes.Buffer
	cmd.Stdout = os.Stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		logging.RuntimeLog.Error(err, stderr)
		logging.CLILog.Error(err, stderr)
		return
	}
	//读取结果
	data, err := os.ReadFile(resultTempFile)
	if err != nil {
		logging.RuntimeLog.Error(err)
		logging.CLILog.Error(err)
		return
	}
	s.parseResultContent(data)
}

// parseResult 解析子域名枚举结果文件
func (s *SubFinder) parseResult(outputTempFile string) {
	content, err := os.ReadFile(outputTempFile)
	if err != nil {
		logging.RuntimeLog.Error(err)
		logging.CLILog.Error(err)
		return
	}

	s.parseResultContent(content)
}

// parseResult 解析子域名枚举结果
func (s *SubFinder) parseResultContent(content []byte) {
	blackDomain := custom.NewBlackTargetCheck(custom.CheckDomain)
	for _, line := range strings.Split(string(content), "\n") {
		domain := strings.TrimSpace(line)
		if domain == "" {
			continue
		}
		if blackDomain.CheckBlack(domain) {
			logging.RuntimeLog.Warningf("%s is in blacklist,skip...", domain)
			continue
		}
		if !s.Result.HasDomain(domain) {
			s.Result.SetDomain(domain)
		}
	}
}
