package portscan

import (
	"bytes"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/custom"
	"os"
	"os/exec"
	"regexp"
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
	inputTargetFile := utils.GetTempPathFileName()
	resultTempFile := utils.GetTempPathFileName()
	defer os.Remove(inputTargetFile)
	defer os.Remove(resultTempFile)

	btc := custom.NewBlackTargetCheck(custom.CheckIP)
	var targets []string
	for _, target := range strings.Split(nmap.Config.Target, ",") {
		t := strings.TrimSpace(target)
		if btc.CheckBlack(t) {
			logging.RuntimeLog.Warningf("%s is in blacklist,skip...", t)
			continue
		}
		targets = append(targets, t)
	}
	err := os.WriteFile(inputTargetFile, []byte(strings.Join(targets, "\n")), 0666)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		return
	}

	var cmdArgs []string
	cmdArgs = append(
		cmdArgs,
		nmap.Config.Tech, "-T4", "--open", "-n", "--randomize-hosts",
		"--min-rate", strconv.Itoa(nmap.Config.Rate), "-oG", resultTempFile, "-iL", inputTargetFile,
	)
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
	FilterIPHasTooMuchPort(&nmap.Result, false)
}

// parseResult 解析nmap结果
func (nmap *Nmap) parseResult(outputTempFile string) {
	content, err := os.ReadFile(outputTempFile)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return
	}

	hostAndPortsReg := regexp.MustCompile("^Host:(.+)Ports:(.+)")
	ipReg := regexp.MustCompile("\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}")

	s := custom.Service{}
	for _, line := range strings.Split(string(content), "\n") {
		txt := strings.TrimSpace(line)
		if txt == "" || strings.HasPrefix(txt, "#") {
			continue
		}
		//check exist ip and ports
		hostAndPorts := hostAndPortsReg.FindAllStringSubmatch(txt, -1)
		if len(hostAndPorts) < 1 || len(hostAndPorts[0]) != 3 {
			continue
		}
		//ip
		ip := ipReg.FindString(hostAndPorts[0][1])
		if ip == "" {
			continue
		}
		if !nmap.Result.HasIP(ip) {
			nmap.Result.SetIP(ip)
		}
		//ports
		for _, p := range strings.Split(strings.TrimSpace(hostAndPorts[0][2]), ",") {
			portInfo := strings.Split(strings.TrimSpace(p), "/")
			portNumber, err := strconv.Atoi(portInfo[0])
			if err != nil {
				logging.RuntimeLog.Error(err)
				continue
			}
			if !nmap.Result.HasPort(ip, portNumber) {
				nmap.Result.SetPort(ip, portNumber)
			}
			service := portInfo[4]
			if service == custom.UnknownService || service == "" {
				service = s.FindService(portNumber, ip)
			}
			nmap.Result.SetPortAttr(ip, portNumber, PortAttrResult{
				Source:  "portscan",
				Tag:     "service",
				Content: service,
			})
			if portInfo[6] != "" {
				nmap.Result.SetPortAttr(ip, portNumber, PortAttrResult{
					Source:  "portscan",
					Tag:     "banner",
					Content: portInfo[6],
				})
			}
		}
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
			if addr.AddrType == "ipv4" {
				ip = addr.Addr
				break
			}
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
