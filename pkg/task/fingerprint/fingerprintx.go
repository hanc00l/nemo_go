package fingerprint

import (
	"bytes"
	"encoding/json"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/custom"
	"github.com/hanc00l/nemo_go/pkg/task/portscan"
	"github.com/hanc00l/nemo_go/pkg/utils"
	"github.com/remeh/sizedwaitgroup"
	"os"
	"os/exec"
	"path/filepath"
)

type Fingerprintx struct {
	ResultPortScan   *portscan.Result
	OptimizationMode bool
	IsProxy          bool
}

type FingerprintxService struct {
	Host      string          `json:"host,omitempty"`
	IP        string          `json:"ip"`
	Port      int             `json:"port"`
	Protocol  string          `json:"protocol"`
	TLS       bool            `json:"tls"`
	Transport string          `json:"transport"`
	Version   string          `json:"version,omitempty"`
	Raw       json.RawMessage `json:"metadata"`
}

func NewFingerprintx() *Fingerprintx {
	//默认是启用OptimizationMode
	f := &Fingerprintx{OptimizationMode: true}

	return f
}

func (f *Fingerprintx) Do() {
	swg := sizedwaitgroup.New(fpFingerprintxThreadNumber[conf.WorkerPerformanceMode])
	btc := custom.NewBlackTargetCheck(custom.CheckAll)
	if f.ResultPortScan != nil && f.ResultPortScan.IPResult != nil {
		for ipName, ipResult := range f.ResultPortScan.IPResult {
			if btc.CheckBlack(ipName) {
				logging.RuntimeLog.Warningf("%s is in blacklist,skip...", ipName)
				continue
			}
			for portNumber, _ := range ipResult.Ports {
				// 优化模式启用时，如果httpx已有扫描结果了，则不进行指纹获取；注意这里与screenshot等http指纹是相反的
				if f.OptimizationMode {
					if CheckForHttpxFingerResult(ipName, "", portNumber, f.ResultPortScan, nil) {
						continue
					}
				}
				url := utils.FormatHostUrl("", ipName, portNumber) //fmt.Sprintf("%v:%v", ipName, portNumber)
				swg.Add()
				go func(ip string, port int, u string) {
					defer swg.Done()
					fingerPrintResult := f.RunFingerprintx(u)
					if len(fingerPrintResult) > 0 {
						for _, fpa := range fingerPrintResult {
							par := portscan.PortAttrResult{
								Source:  "fingerprintx",
								Tag:     fpa.Tag,
								Content: fpa.Content,
							}
							f.ResultPortScan.SetPortAttr(ip, port, par)
						}
					}
				}(ipName, portNumber, url)
			}
		}
	}
	swg.Wait()
}

func (f *Fingerprintx) RunFingerprintx(domain string) (result []FingerAttrResult) {
	resultTempFile := utils.GetTempPathFileName()
	defer os.Remove(resultTempFile)

	var cmdArgs []string
	cmdArgs = append(cmdArgs,
		"--json", "-t", domain, "-o", resultTempFile,
	)
	if f.IsProxy {
		if proxy := conf.GetProxyConfig(); proxy != "" {
			cmdArgs = append(cmdArgs, "-p", proxy)
		} else {
			logging.RuntimeLog.Warning("get proxy config fail or disabled by worker,skip proxy!")
			logging.CLILog.Warning("get proxy config fail or disabled by worker,skip proxy!")
		}
	}
	binPath := filepath.Join(conf.GetRootPath(), "thirdparty/fingerprintx", utils.GetThirdpartyBinNameByPlatform(utils.Fingerprintx))
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
	result = f.parseResult(resultTempFile)
	return
}

func (f *Fingerprintx) parseResult(outputTempFile string) (result []FingerAttrResult) {
	content, err := os.ReadFile(outputTempFile)
	if err != nil || len(content) == 0 {
		if err != nil {
			logging.RuntimeLog.Error(err)
			logging.CLILog.Error(err)
		}
		return
	}
	var service FingerprintxService
	err = json.Unmarshal(content, &service)
	if err != nil {
		logging.RuntimeLog.Error(err)
		logging.CLILog.Error(err)
		return
	}
	farBanner := FingerAttrResult{
		Tag:     "fingerprintx",
		Content: string(content),
	}
	result = append(result, farBanner)
	if service.Protocol != "" {
		farProtocol := FingerAttrResult{
			Tag:     "service",
			Content: service.Protocol,
		}
		result = append(result, farProtocol)
	}
	if service.Version != "" {
		farVersion := FingerAttrResult{
			Tag:     "version",
			Content: service.Version,
		}
		result = append(result, farVersion)
	}
	return
}
