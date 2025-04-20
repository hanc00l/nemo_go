package domainscan

import (
	"github.com/hanc00l/nemo_go/v3/pkg/conf"
	"github.com/hanc00l/nemo_go/v3/pkg/core"
	"github.com/hanc00l/nemo_go/v3/pkg/logging"
	"github.com/hanc00l/nemo_go/v3/pkg/resource"
	"github.com/hanc00l/nemo_go/v3/pkg/task/execute"
	"github.com/hanc00l/nemo_go/v3/pkg/utils"
	"path/filepath"
	"strings"
)

type Subfinder struct {
	Config  execute.DomainscanConfig
	IsProxy bool
}

func (s *Subfinder) GetRequiredResources() (re []core.RequiredResource) {
	re = append(re, core.RequiredResource{
		Category: resource.SubfinderCategory,
		Name:     utils.GetThirdpartyBinNameByPlatform(utils.Subfinder),
	})
	re = append(re, core.RequiredResource{
		Category: resource.DictCategory,
		Name:     "resolver.txt",
	})
	//CDN
	re = append(re, core.RequiredResource{
		Category: resource.Geolite2Category,
		Name:     "GeoLite2-ASN.mmdb",
	})
	return
}

func (s *Subfinder) IsExecuteFromCmd() bool {
	return true
}

func (s *Subfinder) GetExecuteCmd() string {
	return filepath.Join(conf.GetRootPath(), "thirdparty/subfinder", utils.GetThirdpartyBinNameByPlatform(utils.Subfinder))
}

func (s *Subfinder) GetExecuteArgs(inputTempFile, outputTempFile string) (cmdArgs []string) {
	cmdArgs = append(cmdArgs,
		"-dL", inputTempFile, "-all", "-o", outputTempFile, "-disable-update-check",
		"-rL", filepath.Join(conf.GetRootPath(), "thirdparty/dict/resolver.txt"),
		"-no-color", "-v",
		"-active", //display active subdomains only
		// 	-config和-pc使用默认的配置，不再指定了
		//	-config string                flag config file (default "$CONFIG/subfinder/config.yaml")
		//  -pc, -provider-config string  provider config file (default "$CONFIG/subfinder/provider-config.yaml")
	)
	if s.IsProxy {
		if proxy := conf.GetProxyConfig(); proxy != "" {
			cmdArgs = append(cmdArgs, "-proxy", proxy)
		} else {
			logging.RuntimeLog.Warning("获取代理配置失败或禁用了代理功能，代理被跳过")
		}
	}
	return cmdArgs

}

func (s *Subfinder) Run(target []string) (result Result) {
	//TODO implement me
	panic("implement me")
}

func (s *Subfinder) ParseContentResult(content []byte) (result Result) {
	var domains []string
	for _, line := range strings.Split(string(content), "\n") {
		domain := strings.TrimSpace(line)
		if domain == "" {
			continue
		}
		domains = append(domains, domain)
	}
	if len(domains) > 0 {
		resolver := Resolve{}
		return resolver.Run(domains)
	}

	return
}
