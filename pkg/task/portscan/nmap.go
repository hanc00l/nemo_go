package portscan

import (
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/custom"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/hanc00l/nemo_go/pkg/utils"
)

type Nmap struct {
	Config Config
	Result Result
}

//NewNmap 创建nmap对象
func NewNmap(config Config) *Nmap {
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
			if service == custom.UnknownService {
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
