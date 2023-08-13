package pocscan

import (
	"fmt"
	"github.com/hanc00l/nemo_go/pkg/utils"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

type FScan struct {
}

func (f *FScan) ParseContentResult(content []byte) (result []Result) {
	lines := strings.Split(string(content), "\n")
	for _, l := range lines {
		line := strings.TrimSpace(strings.Trim(l, "\r"))
		if strings.HasPrefix(line, "[+] InfoScan") {
			//[+] InfoScan:http://10.58.26.183:7080/login [SpringBoot]
			continue
		} else if strings.HasPrefix(line, "[+] ") {
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
			if ok, ip, _ := f.parseUrlForIPPortResult(fmt.Sprintf("http://%s:%s", hostPort[1], hostPort[2])); ok {
				lineSeps := strings.Split(line, " ")
				if len(lineSeps) >= 3 && strings.HasPrefix(lineSeps[2], "poc-yaml") {
					result = append(result, Result{
						Target:  ip,
						Url:     lineSeps[1],
						PocFile: lineSeps[2],
						Source:  "fscan",
						Extra:   line,
					})
				}
			}
		}
	}
	return
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
	ok = true
	return
}
