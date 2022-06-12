package portscan

import (
	"fmt"
	"github.com/hanc00l/nemo_go/pkg/task/custom"
	"github.com/hanc00l/nemo_go/pkg/task/pocscan"
	"github.com/hanc00l/nemo_go/pkg/utils"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

// FScan 导入fscan的扫描结果
type FScan struct {
	Config    Config
	Result    Result
	VulResult []pocscan.Result
}

// NewFScan 创建FScan对象
func NewFScan(config Config) *FScan {
	return &FScan{Config: config}
}

// ParseTxtContentResult 解析fscan扫描的文本结果
func (f *FScan) ParseTxtContentResult(content []byte) {
	s := custom.NewService()
	if f.Result.IPResult == nil {
		f.Result.IPResult = make(map[string]*IPResult)
	}
	lines := strings.Split(string(content), "\n")
	for _, l := range lines {
		line := strings.TrimSpace(strings.Trim(l, "\r"))
		if strings.HasPrefix(line, "(icmp)") || strings.HasPrefix(line, "(ping)") {
			//(ping) Target 192.168.3.183   is alive
			patternHost := "Target (.+) +is alive"
			regHost, _ := regexp.Compile(patternHost)
			hosts := regHost.FindStringSubmatch(line)
			if len(hosts) != 2 || utils.CheckIPV4(hosts[1]) == false {
				continue
			}
			if !f.Result.HasIP(hosts[1]) {
				f.Result.SetIP(hosts[1])
			}
		} else if strings.HasPrefix(line, "[*] WebTitle") {
			//[*] WebTitle:http://192.168.3.242      code:200 len:86     title:None
			patternWebTitle := "\\[\\*\\] WebTitle:(http.?://.+) +code:(\\d{1,3}) +len:\\d+ +title:(.+)"
			regWebTitle, _ := regexp.Compile(patternWebTitle)
			titles := regWebTitle.FindStringSubmatch(line)
			if len(titles) != 4 {
				continue
			}
			url := titles[1]
			statusCode := titles[2]
			title := titles[3]
			if ok, ip, port := f.parseUrlForIPPortResult(url); ok {
				f.Result.IPResult[ip].Ports[port].Status = statusCode
				if title != "None" {
					par := PortAttrResult{
						Source:  "fscan",
						Tag:     "title",
						Content: title,
					}
					f.Result.SetPortAttr(ip, port, par)
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
			if ok, ip, port := f.parseUrlForIPPortResult(url); ok {
				par := PortAttrResult{
					Source:  "fscan",
					Tag:     "fingerprint",
					Content: info,
				}
				f.Result.SetPortAttr(ip, port, par)
			}
		} else if strings.HasPrefix(line, "[+] ") {
			//[+] http://192.168.10.237:8085 poc-yaml-vmware-vcenter-cve-2021-21985-rce
			//[+] SSH:192.168.10.236:22:root root@123
			patternHostPort := "(\\d{1,3}.\\d{1,3}.\\d{1,3}.\\d{1,3}):(\\d{1,5})"
			regHosPort, _ := regexp.Compile(patternHostPort)
			hostPort := regHosPort.FindStringSubmatch(line)
			if len(hostPort) != 3 || utils.CheckIPV4(hostPort[1]) == false {
				continue
			}
			if ok, ip, port := f.parseUrlForIPPortResult(fmt.Sprintf("http://%s:%s", hostPort[1], hostPort[2])); ok {
				f.Result.SetPortAttr(ip, port, PortAttrResult{
					Source:  "fscan",
					Tag:     "banner",
					Content: strings.TrimLeft(line, "[+] "),
				})
				lineSeps := strings.Split(line, " ")
				if len(lineSeps) >= 3 && strings.HasPrefix(lineSeps[2], "poc-yaml") {
					f.VulResult = append(f.VulResult, pocscan.Result{
						Target:  ip,
						Url:     lineSeps[1],
						PocFile: lineSeps[2],
						Source:  "fscan",
						Extra:   line,
					})
				}
			}
		} else {
			//192.168.3.242:80 open
			patternHostPort := "^(.+):(\\d{1,5}) +open"
			regHosPort, _ := regexp.Compile(patternHostPort)
			hostPort := regHosPort.FindStringSubmatch(line)
			if len(hostPort) != 3 || utils.CheckIPV4(hostPort[1]) == false {
				continue
			}
			if ok, ip, port := f.parseUrlForIPPortResult(fmt.Sprintf("http://%s:%s", hostPort[1], hostPort[2])); ok {
				service := s.FindService(port, ip)
				f.Result.SetPortAttr(ip, port, PortAttrResult{
					Source:  "fscan",
					Tag:     "service",
					Content: service,
				})
			}
		}
	}
}

func (f *FScan) parseUrlForIPPortResult(httpUrl string) (ok bool, ip string, port int) {
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
	if err != nil || utils.CheckIPV4(ip) == false {
		return
	}
	if !f.Result.HasIP(ip) {
		f.Result.SetIP(ip)
	}
	if !f.Result.HasPort(ip, port) {
		f.Result.SetPort(ip, port)
	}

	ok = true
	return
}
