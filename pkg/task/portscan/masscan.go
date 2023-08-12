package portscan

import (
	"bytes"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/custom"
	"github.com/hanc00l/nemo_go/pkg/utils"
	gonmap "github.com/lair-framework/go-nmap"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

type Masscan struct {
	Config Config
	Result Result
}

// NewMasscan 创建masscan对象
func NewMasscan(config Config) *Masscan {
	config.CmdBin = "masscan"
	if runtime.GOOS == "windows" {
		config.CmdBin = "masscan.exe"
	}
	return &Masscan{Config: config}
}

// Do 执行masscan
func (m *Masscan) Do() {
	m.Result.IPResult = make(map[string]*IPResult)
	inputTargetFile := utils.GetTempPathFileName()
	resultTempFile := utils.GetTempPathFileName()
	defer os.Remove(inputTargetFile)
	defer os.Remove(resultTempFile)

	btc := custom.NewBlackTargetCheck(custom.CheckIP)
	var targets []string
	for _, target := range strings.Split(m.Config.Target, ",") {
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
		"--open", "--rate", strconv.Itoa(m.Config.Rate), "-iL", inputTargetFile, "-oL", resultTempFile,
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
	cmd := exec.Command(m.Config.CmdBin, cmdArgs...)
	var stderr bytes.Buffer
	cmd.Stdout = os.Stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		logging.RuntimeLog.Error(err, stderr)
		logging.CLILog.Error(err, stderr)
		return
	}
	m.parsResult(resultTempFile)
	FilterIPHasTooMuchPort(&m.Result, false)
}

// parsResult 解析结果
func (m *Masscan) parsResult(outputTempFile string) {
	content, err := os.ReadFile(outputTempFile)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return
	}

	s := custom.NewService()
	for _, line := range strings.Split(string(content), "\n") {
		txt := strings.TrimSpace(line)
		if txt == "" || strings.HasPrefix(txt, "#") {
			continue
		}
		data := strings.Split(txt, " ")
		if data[0] == "open" && data[1] == "tcp" {
			ip := strings.TrimSpace(data[3])
			portNumber, err := strconv.Atoi(data[2])
			if err != nil {
				logging.RuntimeLog.Error(err)
				continue
			}
			if !m.Result.HasIP(ip) {
				m.Result.SetIP(ip)
			}
			if !m.Result.HasPort(ip, portNumber) {
				m.Result.SetPort(ip, portNumber)
			}
			service := s.FindService(portNumber, ip)
			m.Result.SetPortAttr(ip, portNumber, PortAttrResult{
				Source:  "portscan",
				Tag:     "service",
				Content: service,
			})
		}
	}
}

// ParseXMLResult 解析XML格式的masscan扫描结果
func (m *Masscan) ParseXMLResult(outputTempFile string) {
	content, err := os.ReadFile(outputTempFile)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return
	}
	m.ParseXMLContentResult(content)
}

// ParseXMLContentResult 解析XML格式的masscan扫描结果
func (m *Masscan) ParseXMLContentResult(content []byte) {
	s := custom.NewService()
	// masscan的XML结果兼容Nmap，但是没有service信息
	nmapRunner, err := gonmap.Parse(content)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return
	}
	if m.Result.IPResult == nil {
		m.Result.IPResult = make(map[string]*IPResult)
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
		if !m.Result.HasIP(ip) {
			m.Result.SetIP(ip)
		}
		for _, port := range host.Ports {
			if port.State.State == "open" && port.Protocol == "tcp" {
				if !m.Result.HasPort(ip, port.PortId) {
					m.Result.SetPort(ip, port.PortId)
				}
				service := s.FindService(port.PortId, ip)
				m.Result.SetPortAttr(ip, port.PortId, PortAttrResult{
					Source:  "portscan",
					Tag:     "service",
					Content: service,
				})
			}
		}
	}
}
