package workerapi

import (
	"context"
	"errors"
	"fmt"
	"github.com/hanc00l/nemo_go/pkg/comm"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/fingerprint"
	"github.com/hanc00l/nemo_go/pkg/task/portscan"
	"strings"
)

// BatchScan 批量扫描任务
// 先对少量端口进行扫描（存活探测），提取扫描结果IP的C段作为，再进行详细扫描
// 与portscan参数和过程相同；存活探测需扫描的port与资产详细扫描的port，用 “|” 进行分隔
func BatchScan(taskId, configJSON string) (result string, err error) {
	var ok bool
	if ok, result, err = CheckTaskStatus(taskId); !ok {
		return result, err
	}
	config := portscan.Config{}
	if err = ParseConfig(configJSON, &config); err != nil {
		return FailedTask(err.Error()), err
	}
	// 提取两个阶段的port
	ports := strings.Split(config.Port, "|")
	if len(ports) != 2 || strings.TrimSpace(ports[0]) == "" || strings.TrimSpace(ports[1]) == "" {
		return FailedTask("ports error"), errors.New("ports error:" + config.Port)
	}
	var resultPortScan portscan.Result
	// 存活探测扫描
	config.Port = ports[0]
	if config.CmdBin == "nmap" {
		nmap := portscan.NewNmap(config)
		nmap.Do()
		resultPortScan = nmap.Result
	} else {
		mascan := portscan.NewMasscan(config)
		mascan.Do()
		resultPortScan = mascan.Result
	}
	// 详细端口扫描
	ipSubnetList := getResultIPSubnetList(&resultPortScan)
	if ipSubnetList != "" {
		config.Port = ports[1]
		config.Target = ipSubnetList
		if config.CmdBin == "nmap" {
			nmap := portscan.NewNmap(config)
			nmap.Do()
			resultPortScan = nmap.Result
		} else {
			mascan := portscan.NewMasscan(config)
			mascan.Do()
			resultPortScan = mascan.Result
		}
		// IP位置
		if config.IsIpLocation {
			doLocation(&resultPortScan)
		}
		// 指纹识别
		fpConfig := fingerprint.Config{OrgId: config.OrgId}
		if config.IsHttpx {
			httpx := fingerprint.NewHttpx(fpConfig)
			httpx.ResultPortScan = resultPortScan
			httpx.Do()
		}
	}
	// 保存结果
	x := comm.NewXClient()
	resultArgs := comm.ScanResultArgs{
		IPConfig: &config,
		IPResult: resultPortScan.IPResult,
	}
	err = x.Call(context.Background(), "SaveScanResult", &resultArgs, &result)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return FailedTask(err.Error()), err
	}
	if config.IsScreenshot {
		ss := fingerprint.NewScreenShot()
		ss.ResultPortScan = resultPortScan
		ss.Do()
		resultScreenshot := ss.LoadResult()
		var result2 string
		err = x.Call(context.Background(), "SaveScreenshotResult", &resultScreenshot, &result2)
		if err != nil {
			logging.RuntimeLog.Error(err)
			return FailedTask(err.Error()), err
		} else {
			result = strings.Join([]string{result, result2}, ",")
		}
	}

	return SucceedTask(result), nil
}

func getResultIPSubnetList(resultPortScan *portscan.Result) string {
	ipSubnets := make(map[string]struct{})
	for ip, _ := range resultPortScan.IPResult {
		ipArray := strings.Split(ip, ".")
		if len(ipArray) != 4 {
			continue
		}
		s := fmt.Sprintf("%s.%s.%s.0/24", ipArray[0], ipArray[1], ipArray[2])
		if _, ok := ipSubnets[s]; !ok {
			ipSubnets[s] = struct{}{}
		}
	}
	var ipSubnetList []string
	for k, _ := range ipSubnets {
		ipSubnetList = append(ipSubnetList, k)
	}
	if len(ipSubnetList) > 0 {
		return strings.Join(ipSubnetList, ",")
	} else {
		return ""
	}
}
