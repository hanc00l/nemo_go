package workerapi

import (
	"github.com/hanc00l/nemo_go/pkg/task/custom"
	"github.com/hanc00l/nemo_go/pkg/task/portscan"
	"github.com/hanc00l/nemo_go/pkg/utils"
	"strings"
)

// IPLocation IP归属任务
func IPLocation(taskId, configJSON string) (result string, err error) {
	isRevoked, err := CheckIsExistOrRevoked(taskId)
	if err != nil {
		return FailedTask(err.Error()), err
	}
	if isRevoked {
		return RevokedTask(""), nil
	}

	config := custom.Config{}
	if err = ParseConfig(configJSON, &config); err != nil {
		return FailedTask(err.Error()), err
	}

	resultPortScan := portscan.Result{IPResult: make(map[string]*portscan.IPResult)}
	ips := strings.Split(config.Target, ",")
	for _, ip := range ips {
		lists := utils.ParseIP(ip)
		for _, oneIp := range lists {
			if !resultPortScan.HasIP(oneIp) {
				resultPortScan.SetIP(oneIp)
			}
		}
	}
	doLocation(&resultPortScan)
	result = resultPortScan.SaveResult(portscan.Config{OrgId: config.OrgId})

	return SucceedTask(result), nil
}

// doLocation 执行IP位置查询
func doLocation(portScanResult *portscan.Result) {
	ipl := custom.NewIPLocation()
	for ip, _ := range portScanResult.IPResult {
		location := ipl.FindCustomIP(ip)
		if location == "" {
			location = ipl.FindPublicIP(ip)
		}
		if location != "" {
			portScanResult.IPResult[ip].Location = location
		}
	}
}
