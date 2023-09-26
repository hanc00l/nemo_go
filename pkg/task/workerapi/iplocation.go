package workerapi

import (
	"github.com/hanc00l/nemo_go/pkg/comm"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/custom"
	"github.com/hanc00l/nemo_go/pkg/task/portscan"
	"github.com/hanc00l/nemo_go/pkg/utils"
	"strings"
)

// IPLocation IP归属任务
func IPLocation(taskId, mainTaskId, configJSON string) (result string, err error) {
	var ok bool
	if ok, result, err = CheckTaskStatus(taskId); !ok {
		return result, err
	}

	config := custom.Config{}
	if err = ParseConfig(configJSON, &config); err != nil {
		logging.RuntimeLog.Error(err)
		return FailedTask(err.Error()), err
	}

	resultPortScan := portscan.Result{IPResult: make(map[string]*portscan.IPResult)}
	ips := strings.Split(config.Target, ",")
	for _, ip := range ips {
		if utils.CheckIPV4(ip) || utils.CheckIPV4Subnet(ip) {
			lists := utils.ParseIP(ip)
			for _, oneIp := range lists {
				if !resultPortScan.HasIP(oneIp) {
					resultPortScan.SetIP(oneIp)
				}
			}
		} else if utils.CheckIPV6(ip) {
			if !resultPortScan.HasIP(ip) {
				resultPortScan.SetIP(ip)
			}
		}
	}
	doLocation(&resultPortScan)
	// 保存结果
	resultArgs := comm.ScanResultArgs{
		TaskID:     taskId,
		MainTaskId: mainTaskId,
		IPConfig:   &portscan.Config{OrgId: config.OrgId},
		IPResult:   resultPortScan.IPResult,
	}
	err = comm.CallXClient("SaveScanResult", &resultArgs, &result)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return FailedTask(err.Error()), err
	}
	return SucceedTask(result), nil
}

// doLocation 执行IP位置查询
func doLocation(portScanResult *portscan.Result) {
	ipv4 := custom.NewIPv4Location()
	ipv6, _ := custom.NewIPv6Location()
	for ip, _ := range portScanResult.IPResult {
		var location string
		if utils.CheckIPV4(ip) {
			location = ipv4.FindCustomIP(ip)
			if location == "" {
				location = ipv4.FindPublicIP(ip)
			}

		} else if utils.CheckIPV6(ip) {
			location = ipv6.Find(ip)
		}
		if location != "" {
			portScanResult.IPResult[ip].Location = location
		}
	}
}
