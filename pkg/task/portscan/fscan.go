package portscan

import (
	"fmt"
	"github.com/hanc00l/nemo_go/pkg/task/custom"
	"github.com/hanc00l/nemo_go/pkg/utils"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

// FScan 导入fscan的扫描结果
type FScan struct {
	//VulResult []pocscan.Result
}

// ParseContentResult 解析fscan扫描的文本结果
func (f *FScan) ParseContentResult(content []byte) (result Result) {
	result.IPResult = make(map[string]*IPResult)

	s := custom.NewService()
	lines := strings.Split(string(content), "\n")
	for _, l := range lines {
		line := strings.TrimSpace(strings.Trim(l, "\r"))
		if strings.HasPrefix(line, "(icmp)") || strings.HasPrefix(line, "(ping)") {
			//(ping) Target 192.168.3.183   is alive
			patternHost := "Target (.+) +is alive"
			regHost, _ := regexp.Compile(patternHost)
			hosts := regHost.FindStringSubmatch(line)
			if len(hosts) != 2 || !(utils.CheckIPV4(hosts[1]) || utils.CheckIPV6(hosts[1])) {
				continue
			}
			if !result.HasIP(hosts[1]) {
				result.SetIP(hosts[1])
			}
		} else if strings.HasPrefix(line, "[*] WebTitle") {
			//[*] WebTitle: https://10.192.117.161:15093 code:200 len:3773   title:一张表
			//[*] WebTitle:http://192.168.3.242      code:200 len:86     title:None
			patternWebTitle := "\\[\\*\\] WebTitle: *(http.?://.+) +code:(\\d{1,3}) +len:\\d+ +title:(.+)"
			regWebTitle, _ := regexp.Compile(patternWebTitle)
			titles := regWebTitle.FindStringSubmatch(line)
			if len(titles) != 4 {
				continue
			}
			url := titles[1]
			statusCode := titles[2]
			title := titles[3]
			if ok, ip, port := f.parseUrlForIPPortResult(url, &result); ok {
				result.IPResult[ip].Ports[port].Status = statusCode
				if title != "None" {
					par := PortAttrResult{
						Source:  "fscan",
						Tag:     "title",
						Content: title,
					}
					result.SetPortAttr(ip, port, par)
				}
			}
		} else if strings.HasPrefix(line, "[+] InfoScan") {
			//[+] InfoScan:http://10.58.26.183:7080/login [SpringBoot]
			patternInfo := "\\[\\+\\] InfoScan:(http.?://.+) +\\[(.+)\\]"
			regInfo, _ := regexp.Compile(patternInfo)
			titles := regInfo.FindStringSubmatch(line)
			if len(titles) != 3 {
				continue
			}
			url := titles[1]
			info := titles[2]
			if ok, ip, port := f.parseUrlForIPPortResult(url, &result); ok {
				par := PortAttrResult{
					Source:  "fscan",
					Tag:     "fingerprint",
					Content: info,
				}
				result.SetPortAttr(ip, port, par)
			}
			/*} else if strings.HasPrefix(line, "[+] ") {
			//[+] mysql:10.192.117.150:3306:root Aa123456
			//[+] https://10.192.117.122:10008 poc-yaml-go-pprof-leak
			//[+] http://192.168.10.237:8085 poc-yaml-vmware-vcenter-cve-2021-21985-rce
			//[+] SSH:192.168.10.236:22:root root@123
			patternHostPort := "(\\d{1,3}.\\d{1,3}.\\d{1,3}.\\d{1,3}):(\\d{1,5})"
			regHosPort, _ := regexp.Compile(patternHostPort)
			hostPort := regHosPort.FindStringSubmatch(line)
			if len(hostPort) != 3 || utils.CheckIPV4(hostPort[1]) == false {
				continue
			}
			if ok, ip, port := f.parseUrlForIPPortResult(fmt.Sprintf("http://%s:%s", hostPort[1], hostPort[2])); ok {
				result.SetPortAttr(ip, port, PortAttrResult{
					Source:  "fscan",
					Tag:     "banner",
					Content: strings.TrimLeft(line, "[+] "),
				})
				lineSeps := strings.Split(line, " ")
				if len(lineSeps) >= 3 && strings.HasPrefix(lineSeps[2], "poc-yaml") {
					f.VulResult = append(f.VulResult, pocscan.Result{
						Target:      ip,
						Url:         lineSeps[1],
						PocFile:     lineSeps[2],
						Source:      "fscan",
						Extra:       line,
						WorkspaceId: f.Config.WorkspaceId,
					})
				}
			}*/
		} else {
			//192.168.3.242:80 open
			patternHostPort := "^(.+):(\\d{1,5}) +open"
			regHosPort, _ := regexp.Compile(patternHostPort)
			hostPort := regHosPort.FindStringSubmatch(line)
			if len(hostPort) != 3 || !(utils.CheckIPV4(hostPort[1]) || utils.CheckIPV6(hostPort[1])) {
				continue
			}
			if ok, ip, port := f.parseUrlForIPPortResult(fmt.Sprintf("http://%s:%s", hostPort[1], hostPort[2]), &result); ok {
				service := s.FindService(port, ip)
				result.SetPortAttr(ip, port, PortAttrResult{
					Source:  "fscan",
					Tag:     "service",
					Content: service,
				})
			}
		}
	}
	return
}

func (f *FScan) parseUrlForIPPortResult(httpUrl string, result *Result) (ok bool, ip string, port int) {
	u, err := url.Parse(strings.TrimSpace(httpUrl))
	if err != nil {
		return
	}
	ip = u.Hostname()
	portValue := u.Port()
	if portValue == "" {
		if u.Scheme == "http" {
			portValue = "80"
		} else {
			portValue = "443"
		}
	}
	port, err = strconv.Atoi(portValue)
	if err != nil || !(utils.CheckIPV4(ip) || utils.CheckIPV6(ip)) {
		return
	}
	if !result.HasIP(ip) {
		result.SetIP(ip)
	}
	if !result.HasPort(ip, port) {
		result.SetPort(ip, port)
	}

	ok = true
	return
}
