package portscan

import (
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/custom"
	"github.com/hanc00l/nemo_go/pkg/utils"
	"os"
	"os/exec"
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
	return &Masscan{Config: config}
}

// Do 执行masscan
func (m *Masscan) Do() {
	m.Result.IPResult = make(map[string]*IPResult)
	resultTempFile := utils.GetTempPathFileName()
	defer os.Remove(resultTempFile)

	var cmdArgs []string
	cmdArgs = append(
		cmdArgs,
		"--open", "--rate", strconv.Itoa(m.Config.Rate), "-oL", resultTempFile,
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
	cmdArgs = append(cmdArgs, m.Config.Target)
	cmd := exec.Command(m.Config.CmdBin, cmdArgs...)
	_, err := cmd.CombinedOutput()
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		return
	}
	m.parsResult(resultTempFile)
	filterIPHasTooMuchPort(m.Result)
}

// parsResult 解析结果
func (m *Masscan) parsResult(outputTempFile string) {
	content, err := os.ReadFile(outputTempFile)
	if err != nil {
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
