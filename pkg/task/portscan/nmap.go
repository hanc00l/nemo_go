package portscan

import (
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
	resultTempFile := utils.GetTempPathFileName()
	defer os.Remove(resultTempFile)

	var cmdArgs []string
	cmdArgs = append(
		cmdArgs,
		nmap.Config.Tech, "-T4", "--open", "-n", "--randomize-hosts",
		"--min-rate", strconv.Itoa(nmap.Config.Rate), "-oG", resultTempFile,
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
	cmdArgs = append(cmdArgs, nmap.Config.Target)
	cmd := exec.Command(nmap.Config.CmdBin, cmdArgs...)
	_, err := cmd.CombinedOutput()
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		return
	}
	nmap.parseResult(resultTempFile)
	FilterIPHasTooMuchPort(&nmap.Result, false)
}

// parseResult 解析nmap结果
func (nmap *Nmap) parseResult(outputTempFile string) {
	content, err := os.ReadFile(outputTempFile)
	if err != nil {
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

// ParseXMLResult 解析nmap的XML文件
func (nmap *Nmap) ParseXMLResult(outputTempFile string) {
	content, err := os.ReadFile(outputTempFile)
	if err != nil {
		return
	}
	nmap.ParseXMLContentResult(content)
}

// ParseXMLContentResult 解析nmap的XML文件
func (nmap *Nmap) ParseXMLContentResult(content []byte) {
	nmapRunner, err := gonmap.Parse(content)
	if err != nil {
		return
	}
	if nmap.Result.IPResult == nil {
		nmap.Result.IPResult = make(map[string]*IPResult)
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
		if !nmap.Result.HasIP(ip) {
			nmap.Result.SetIP(ip)
		}
		for _, port := range host.Ports {
			if port.State.State == "open" && port.Protocol == "tcp" {
				if !nmap.Result.HasPort(ip, port.PortId) {
					nmap.Result.SetPort(ip, port.PortId)
				}
				service := port.Service.Name
				if service == "" {
					service = s.FindService(port.PortId, ip)
				}
				nmap.Result.SetPortAttr(ip, port.PortId, PortAttrResult{
					Source:  "portscan",
					Tag:     "service",
					Content: service,
				})
				banner := strings.Join([]string{port.Service.Product, port.Service.Version}, " ")
				if strings.TrimSpace(banner) != "" {
					nmap.Result.SetPortAttr(ip, port.PortId, PortAttrResult{
						Source:  "portscan",
						Tag:     "banner",
						Content: banner,
					})
				}
			}
		}
	}
}
