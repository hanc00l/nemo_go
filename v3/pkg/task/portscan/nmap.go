package portscan

import (
	"github.com/hanc00l/nemo_go/v3/pkg/core"
	"github.com/hanc00l/nemo_go/v3/pkg/db"
	"github.com/hanc00l/nemo_go/v3/pkg/logging"
	"github.com/hanc00l/nemo_go/v3/pkg/resource"
	"github.com/hanc00l/nemo_go/v3/pkg/task/custom"
	"github.com/hanc00l/nemo_go/v3/pkg/task/execute"
	gonmap "github.com/lair-framework/go-nmap"
	"runtime"
	"strconv"
	"strings"
)

type Nmap struct {
	Config execute.PortscanConfig
}

func (nmap *Nmap) GetRequiredResources() (re []core.RequiredResource) {
	re = append(re, core.RequiredResource{
		Category: resource.NmapCategory,
		Name:     "nmap-services",
	})
	return
}

func (nmap *Nmap) IsExecuteFromCmd() bool {
	return true
}

func (nmap *Nmap) GetExecuteCmd() string {
	if runtime.GOOS == "windows" {
		return "nmap.exe"
	}
	return "nmap"
}

func (nmap *Nmap) GetExecuteArgs(inputTempFile, outputTempFile string, ipv6 bool) (cmdArgs []string) {
	cmdArgs = append(
		cmdArgs,
		nmap.Config.Tech, "-T4", "--open", "-n", "--randomize-hosts",
		"--min-rate", strconv.Itoa(nmap.Config.Rate), "-oX", outputTempFile, "-iL", inputTempFile,
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
	return cmdArgs
}

func (nmap *Nmap) Run(targets []string, ipv6 bool) {
	// 	TODO implement me
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
					service = s.FindService(port.PortId)
				}
				result.SetPortAttr(ip, port.PortId, PortAttrResult{
					Source:  "portscan",
					Tag:     db.FingerService,
					Content: service,
				})
				banner := strings.Join([]string{port.Service.Product, port.Service.Version}, " ")
				if strings.TrimSpace(banner) != "" {
					result.SetPortAttr(ip, port.PortId, PortAttrResult{
						Source:  "portscan",
						Tag:     db.FingerBanner,
						Content: banner,
					})
				}
			}
		}
	}
	return
}
