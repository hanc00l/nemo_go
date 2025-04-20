package portscan

import (
	"github.com/hanc00l/nemo_go/v3/pkg/core"
	"github.com/hanc00l/nemo_go/v3/pkg/db"
	"github.com/hanc00l/nemo_go/v3/pkg/logging"
	"github.com/hanc00l/nemo_go/v3/pkg/resource"
	"github.com/hanc00l/nemo_go/v3/pkg/task/custom"
	"github.com/hanc00l/nemo_go/v3/pkg/task/execute"
	"github.com/hanc00l/nemo_go/v3/pkg/utils"
	gonmap "github.com/lair-framework/go-nmap"
	"runtime"
	"strconv"
	"strings"
)

type Masscan struct {
	Config execute.PortscanConfig
}

func (m Masscan) GetRequiredResources() (re []core.RequiredResource) {
	// masscan 调用custom解析service信息，需要nmap-services
	re = append(re, core.RequiredResource{
		Category: resource.NmapCategory,
		Name:     "nmap-services",
	})
	return
}

func (m Masscan) IsExecuteFromCmd() bool {
	return true
}

func (m Masscan) GetExecuteCmd() string {
	if runtime.GOOS == "windows" {
		return "masscan.exe"
	}
	return "masscan"
}

func (m Masscan) GetExecuteArgs(inputTempFile, outputTempFile string, ipv6 bool) (cmdArgs []string) {
	cmdArgs = append(
		cmdArgs,
		"--open", "--rate", strconv.Itoa(m.Config.Rate), "-iL", inputTempFile, "-oX", outputTempFile,
	)
	if strings.HasPrefix(m.Config.Port, "--top-ports") {
		cmdArgs = append(cmdArgs, "-p")
		switch strings.Split(m.Config.Port, " ")[1] {
		case "1000":
			cmdArgs = append(cmdArgs, utils.TopPorts1000)
		case "100":
			cmdArgs = append(cmdArgs, utils.TopPorts100)
		case "10":
			cmdArgs = append(cmdArgs, utils.TopPorts10)
		default:
			cmdArgs = append(cmdArgs, utils.TopPorts100)
		}
	} else {
		cmdArgs = append(cmdArgs, "-p", m.Config.Port)
	}
	if m.Config.ExcludeTarget != "" {
		cmdArgs = append(cmdArgs, "--exclude", m.Config.ExcludeTarget)
	}
	return
}

func (m Masscan) Run(target []string, ipv6 bool) {
	//TODO implement me
}

func (m Masscan) ParseContentResult(content []byte) (result Result) {
	result.IPResult = make(map[string]*IPResult)
	if len(content) == 0 {
		return
	}
	s := custom.NewService()
	// masscan的XML结果兼容Nmap，但是没有service信息
	nmapRunner, err := gonmap.Parse(content)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return
	}
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
				service := s.FindService(port.PortId)
				result.SetPortAttr(ip, port.PortId, PortAttrResult{
					Source:  "portscan",
					Tag:     db.FingerService,
					Content: service,
				})
			}
		}
	}
	return
}
