package workerapi

import (
	"context"
	"fmt"
	"github.com/hanc00l/nemo_go/pkg/comm"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/fingerprint"
	"github.com/hanc00l/nemo_go/pkg/task/portscan"
	"github.com/remeh/sizedwaitgroup"
	"strings"
)

const (
	fpNmapThreadNumber = 10
)

// PortScan 端口扫描任务
func PortScan(taskId, configJSON string) (result string, err error) {
	var ok bool
	if ok, result, err = CheckTaskStatus(taskId); !ok {
		return result, err
	}
	config := portscan.Config{}
	if err = ParseConfig(configJSON, &config); err != nil {
		return FailedTask(err.Error()), err
	}
	var resultPortScan portscan.Result
	// 端口扫描
	if config.CmdBin == "masnmap" {
		resultPortScan = doMasscanPlusNmap(config)
	} else if config.CmdBin == "nmap" {
		nmap := portscan.NewNmap(config)
		nmap.Do()
		resultPortScan = nmap.Result
	} else {
		masscan := portscan.NewMasscan(config)
		masscan.Do()
		resultPortScan = masscan.Result
	}
	// IP位置
	if config.IsIpLocation {
		doLocation(&resultPortScan)
	}
	// 指纹识别
	fpConfig := fingerprint.Config{OrgId: config.OrgId}
	if config.IsWhatWeb {
		whatweb := fingerprint.NewWhatweb(fpConfig)
		whatweb.ResultPortScan = resultPortScan
		whatweb.Do()
	}
	if config.IsHttpx {
		httpx := fingerprint.NewHttpx(fpConfig)
		httpx.ResultPortScan = resultPortScan
		httpx.Do()
	}
	if config.IsWappalyzer {
		wappalyzer := fingerprint.NewWappalyzer()
		wappalyzer.ResultPortScan = resultPortScan
		wappalyzer.Do()
	}
	if config.IsFingerprintHub {
		fp := fingerprint.NewFingerprintHub(fpConfig)
		fp.ResultPortScan = resultPortScan
		fp.Do()
	}
	// 保存结果
	resultArgs := comm.ScanResultArgs{
		IPConfig: &config,
		IPResult: resultPortScan.IPResult,
	}
	x := comm.NewXClient()
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

// doMasscanPlusNmap masscan进行端口扫描，nmap -sV进行详细扫描
func doMasscanPlusNmap(config portscan.Config) (resultPortScan portscan.Result) {
	resultPortScan.IPResult = make(map[string]*portscan.IPResult)
	//masscan扫描
	masscan := portscan.NewMasscan(config)
	masscan.Do()
	ipPortMap := getResultIPPortMap(masscan.Result.IPResult)
	//nmap多线程扫描
	swg := sizedwaitgroup.New(fpNmapThreadNumber)
	for ip, port := range ipPortMap {
		nmapConfig := config
		nmapConfig.Target = ip
		nmapConfig.Port = port
		nmapConfig.Tech = "-sV"
		swg.Add()
		go func(c portscan.Config) {
			nmap := portscan.NewNmap(c)
			nmap.Do()
			resultPortScan.Lock()
			for nip, r := range nmap.Result.IPResult {
				resultPortScan.IPResult[nip] = r
			}
			resultPortScan.Unlock()
			swg.Done()
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
