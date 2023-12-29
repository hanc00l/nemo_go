package workerapi

import (
	"fmt"
	"github.com/hanc00l/nemo_go/pkg/comm"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/portscan"
	"github.com/hanc00l/nemo_go/pkg/utils"
	"github.com/remeh/sizedwaitgroup"
	"strconv"
	"strings"
)

var fpNmapThreadNumber = make(map[string]int)

func init() {
	fpNmapThreadNumber[conf.HighPerformance] = 10
	fpNmapThreadNumber[conf.NormalPerformance] = 5
}

// PortScan 端口扫描任务
func PortScan(taskId, mainTaskId, configJSON string) (result string, err error) {
	var ok bool
	if ok, result, err = CheckTaskStatus(taskId); !ok {
		return result, err
	}
	config := portscan.Config{}
	if err = ParseConfig(configJSON, &config); err != nil {
		logging.RuntimeLog.Error(err)
		return FailedTask(err.Error()), err
	}
	var resultPortScan *portscan.Result
	resultPortScan, result, err = doPortScanAndSave(taskId, mainTaskId, config)
	//指纹识别任务
	_, err = NewFingerprintTask(taskId, mainTaskId, resultPortScan, nil, FingerprintTaskConfig{
		IsHttpx:          config.IsHttpx,
		IsFingerprintHub: config.IsFingerprintHub,
		IsIconHash:       config.IsIconHash,
		IsScreenshot:     config.IsScreenshot,
		IsFingerprintx:   config.IsFingerprintx,
		WorkspaceId:      config.WorkspaceId,
		IsProxy:          config.IsProxy,
	})
	if err != nil {
		return FailedTask(err.Error()), err
	}
	return SucceedTask(result), nil
}

func doPortScanAndSave(taskId string, mainTaskId string, config portscan.Config) (resultPortScan *portscan.Result, result string, err error) {
	//端口扫描：
	if config.IsPortscan {
		if config.CmdBin == "masnmap" {
			resultPortScan = doMasscanPlusNmap(config)
		} else if config.CmdBin == "nmap" {
			nmap := portscan.NewNmap(config)
			nmap.Do()
			resultPortScan = &nmap.Result
		} else if config.CmdBin == "gogo" {
			gogo := portscan.NewGogo(config)
			gogo.Do()
			resultPortScan = &gogo.Result
		} else {
			masscan := portscan.NewMasscan(config)
			masscan.Do()
			resultPortScan = &masscan.Result
		}
	} else {
		resultPortScan.IPResult = make(map[string]*portscan.IPResult)
	}
	// IP位置
	if config.IsIpLocation {
		doLocation(resultPortScan)
	}
	// 保存结果
	resultArgs := comm.ScanResultArgs{
		TaskID:     taskId,
		MainTaskId: mainTaskId,
		IPConfig:   &config,
		IPResult:   resultPortScan.IPResult,
	}
	err = comm.CallXClient("SaveScanResult", &resultArgs, &result)
	if err != nil {
		logging.RuntimeLog.Error(err)
	}
	// 读取目标的数据库中已保存的开放端口
	var resultIPPorts string
	if config.IsLoadOpenedPort {
		args := comm.LoadIPOpenedPortArgs{
			WorkspaceId: config.WorkspaceId,
			Target:      config.Target,
		}
		err = comm.CallXClient("LoadOpenedPort", &args, &resultIPPorts)
		if err == nil && resultIPPorts != "" {
			allTargets := strings.Split(resultIPPorts, ",")
			for _, target := range allTargets {
				// 必须是ip:port格式
				dataArray := strings.Split(target, ":")
				if len(dataArray) != 2 {
					continue
				}
				ip := dataArray[0]
				port, err := strconv.Atoi(dataArray[1])
				if !utils.CheckIP(ip) || err != nil {
					continue
				}
				if !resultPortScan.HasIP(ip) {
					resultPortScan.SetIP(ip)
				}
				if !resultPortScan.HasPort(ip, port) {
					resultPortScan.SetPort(ip, port)
				}
			}
		} else {
			logging.RuntimeLog.Error(err)
		}
	}
	return
}

// doMasscanPlusNmap masscan进行端口扫描，nmap -sV进行详细扫描
func doMasscanPlusNmap(config portscan.Config) (resultPortScan *portscan.Result) {
	resultPortScan.IPResult = make(map[string]*portscan.IPResult)
	//masscan扫描
	masscan := portscan.NewMasscan(config)
	masscan.Do()
	ipPortMap := getResultIPPortMap(masscan.Result.IPResult)
	//nmap多线程扫描
	swg := sizedwaitgroup.New(fpNmapThreadNumber[conf.WorkerPerformanceMode])
	for ip, port := range ipPortMap {
		nmapConfig := config
		nmapConfig.Target = ip
		nmapConfig.Port = port
		nmapConfig.Tech = "-sV"
		swg.Add()
		go func(c portscan.Config) {
			defer swg.Done()
			nmap := portscan.NewNmap(c)
			nmap.Do()
			resultPortScan.Lock()
			for nip, r := range nmap.Result.IPResult {
				resultPortScan.IPResult[nip] = r
			}
			resultPortScan.Unlock()
		}(nmapConfig)
	}
	swg.Wait()

	return
}

// getResultIPPortMap 提取扫描结果的ip和port
func getResultIPPortMap(result map[string]*portscan.IPResult) (ipPortMap map[string]string) {
	ipPortMap = make(map[string]string)
	for ip, r := range result {
		var ports []string
		for p, _ := range r.Ports {
			ports = append(ports, fmt.Sprintf("%d", p))
		}
		if len(ports) > 0 {
			ipPortMap[ip] = strings.Join(ports, ",")
		}
	}
	return
}
