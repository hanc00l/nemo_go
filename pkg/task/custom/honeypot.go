package custom

import (
	"fmt"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type HoneyPot struct {
	honeypotMap map[string]*HoneyDef
}

type HoneyDef struct {
	SystemName string
	PortDef    map[int]string
}

// NewHoneyPot 创建HoneyPot对象
func NewHoneyPot() *HoneyPot {
	hp := &HoneyPot{}
	hp.loadHoneyPot()
	return hp
}

// loadHoneyPot 加载honeypot配置
func (hp *HoneyPot) loadHoneyPot() {
	hp.honeypotMap = make(map[string]*HoneyDef)
	content, err := os.ReadFile(filepath.Join(conf.GetRootPath(), "thirdparty/custom/honeypot.txt"))
	if err != nil {
		logging.RuntimeLog.Error(err)
		logging.CLILog.Error(err)
	} else {
		for _, line := range strings.Split(string(content), "\n") {
			txt := strings.TrimSpace(line)
			if txt == "" || strings.HasPrefix(txt, "#") {
				continue
			}
			arrays := strings.Split(txt, " ")
			if len(arrays) < 3 {
				continue
			}
			domain := strings.TrimSpace(arrays[0])
			system := strings.TrimSpace(arrays[2])
			if _, ok := hp.honeypotMap[domain]; !ok {
				hp.honeypotMap[domain] = &HoneyDef{SystemName: system, PortDef: make(map[int]string)}
			}
			if strings.TrimSpace(arrays[1]) == "-" {
				continue
			}
			ports := strings.Split(arrays[1], ",")
			for _, p := range ports {
				port, err := strconv.Atoi(p)
				if err != nil {
					logging.RuntimeLog.Error(err)
					logging.CLILog.Error(err)
					continue
				}
				hp.honeypotMap[domain].PortDef[port] = system
			}
		}
	}
}

// CheckHoneyPot 给定domain/ip和端口，匹配是否是预定义的honeypot
func (hp *HoneyPot) CheckHoneyPot(domain, ports string) (isChecked bool, systemList []string) {
	if _, ok := hp.honeypotMap[domain]; !ok {
		return false, nil
	}
	if len(hp.honeypotMap[domain].PortDef) == 0 || ports == "" {
		return true, []string{fmt.Sprintf("%s/%s", domain, hp.honeypotMap[domain].SystemName)}
	}
	for _, p := range strings.Split(ports, ",") {
		port, err := strconv.Atoi(p)
		if err != nil {
			logging.RuntimeLog.Error(err)
			logging.CLILog.Error(err)
			continue
		}
		if _, ok := hp.honeypotMap[domain].PortDef[port]; !ok {
			continue
		}
		systemList = append(systemList, fmt.Sprintf("%s:%d/%s", domain, port, hp.honeypotMap[domain].PortDef[port]))
	}
	if len(systemList) == 0 {
		return false, nil
	} else {
		return true, systemList
	}
}
