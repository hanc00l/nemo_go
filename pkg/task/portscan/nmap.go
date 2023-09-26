package portscan

import (
	"bytes"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/custom"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/hanc00l/nemo_go/pkg/utils"
	gonmap "github.com/lair-framework/go-nmap"
)

type Nmap struct {
	Config Config
	Result Result
}

// NewNmap 创建nmap对象
func NewNmap(config Config) *Nmap {
	config.CmdBin = "nmap"
	if runtime.GOOS == "windows" {
		config.CmdBin = "nmap.exe"
	}
	return &Nmap{Config: config}
}

// Do 执行nmap
func (nmap *Nmap) Do() {
	nmap.Result.IPResult = make(map[string]*IPResult)

	var targetIpV4, targetIpV6 []string
	btc := custom.NewBlackTargetCheck(custom.CheckIP)
	for _, target := range strings.Split(nmap.Config.Target, ",") {
		t := strings.TrimSpace(target)
		if btc.CheckBlack(t) {
			logging.RuntimeLog.Warningf("%s is in blacklist,skip...", t)
			continue
		}
		if utils.CheckIPV4(t) || utils.CheckIPV4Subnet(t) {
			targetIpV4 = append(targetIpV4, t)
		} else if utils.CheckIPV6(t) || utils.CheckIPV6Subnet(t) {
			targetIpV6 = append(targetIpV6, t)
		}
	}
	if len(targetIpV4) > 0 {
		nmap.RunNmap(targetIpV4, false)
	}
	if len(targetIpV6) > 0 {
		nmap.RunNmap(targetIpV6, true)
	}
	FilterIPHasTooMuchPort(&nmap.Result, false)
}

// RunNmap 调用并执行一次nmap
func (nmap *Nmap) RunNmap(targets []string, ipv6 bool) {
	inputTargetFile := utils.GetTempPathFileName()
	resultTempFile := utils.GetTempPathFileName()
	defer os.Remove(inputTargetFile)
	defer os.Remove(resultTempFile)

	err := os.WriteFile(inputTargetFile, []byte(strings.Join(targets, "\n")), 0666)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		return
	}

	var cmdArgs []string
	cmdArgs = append(
		cmdArgs,
		nmap.Config.Tech, "-T4", "--open", "-n", "--randomize-hosts",
		"--min-rate", strconv.Itoa(nmap.Config.Rate), "-oX", resultTempFile, "-iL", inputTargetFile,
	)
	if ipv6 {
		cmdArgs = append(cmdArgs, "-6")
	}
	if !nmap.Config.IsPing {
		cmdArgs = append(cmdArgs, "-Pn")
	}
	if strings.HasPrefix(nmap.Config.Port, "--top-ports") {
		cmdArgs = append(cmdArgs, "--top-ports")
		cmdArgs = append(cmdArgs, strings.Split(nmap.Config.Port, " ")[1])
	} else {
		cmdArgs = append(cmdArgs, "-p", nmap.Config.Port)
	}
	if nmap.Config.ExcludeTarget != "" {
		cmdArgs = append(cmdArgs, "--exclude", nmap.Config.ExcludeTarget)
	}
	cmd := exec.Command(nmap.Config.CmdBin, cmdArgs...)
	var stderr bytes.Buffer
	cmd.Stdout = os.Stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		logging.RuntimeLog.Error(err, stderr)
		logging.CLILog.Error(err, stderr)
		return
	}
	nmap.parseResult(resultTempFile)
}

// parseResult 解析nmap结果
func (nmap *Nmap) parseResult(outputTempFile string) {
	content, err := os.ReadFile(outputTempFile)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return
	}
	result := nmap.ParseContentResult(content)
	for ip, ipr := range result.IPResult {
		nmap.Result.IPResult[ip] = ipr
	}
}

// ParseContentResult 解析nmap的XML文件
func (nmap *Nmap) ParseContentResult(content []byte) (result Result) {
	result.IPResult = make(map[string]*IPResult)
	nmapRunner, err := gonmap.Parse(content)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return
	}
	s := custom.NewService()
	for _, host := range nmapRunner.Hosts {
		if len(host.Ports) == 0 || len(host.Addresses) == 0 {
			continue
		}
		var ip string
		for _, addr := range host.Addresses {
			ip = addr.Addr
			break
		}
		if ip == "" {
			continue
		}
		if !result.HasIP(ip) {
			result.SetIP(ip)
		}
		for _, port := range host.Ports {
			if port.State.State == "open" && port.Protocol == "tcp" {
				if !result.HasPort(ip, port.PortId) {
					result.SetPort(ip, port.PortId)
				}
				service := port.Service.Name
				if service == "" {
					service = s.FindService(port.PortId, ip)
				}
				result.SetPortAttr(ip, port.PortId, PortAttrResult{
					Source:  "portscan",
					Tag:     "service",
					Content: service,
				})
				banner := strings.Join([]string{port.Service.Product, port.Service.Version}, " ")
				if strings.TrimSpace(banner) != "" {
					result.SetPortAttr(ip, port.PortId, PortAttrResult{
						Source:  "portscan",
						Tag:     "banner",
						Content: banner,
					})
				}
			}
		}
	}
	return
}
