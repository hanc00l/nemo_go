package portscan

import (
	"bufio"
	"bytes"
	"github.com/hanc00l/nemo_go/pkg/task/custom"
	"github.com/hanc00l/nemo_go/pkg/utils"
	"regexp"
	"strconv"
	"strings"
)

type TXPortMap struct {
	Config Config
	Result Result
}

func NewTXPortMap(config Config) *TXPortMap {
	return &TXPortMap{
		Config: config,
	}
}

// ParseTxtContentResult 解析TXPortMap扫描的文本结果
func (tx *TXPortMap) ParseTxtContentResult(content []byte) {
	s := custom.NewService()
	if tx.Result.IPResult == nil {
		tx.Result.IPResult = make(map[string]*IPResult)
	}
	scanner := bufio.NewScanner(bytes.NewReader(content))
	for scanner.Scan() {
		data := scanner.Text()
		if data == "" {
			continue
		}
		allDataArray := strings.Split(data, " ")
		if len(allDataArray) <= 0 {
			continue
		}
		//提取ip和端口
		dataArray := strings.Split(allDataArray[0], ":")
		if len(dataArray) != 2 {
			continue
		}
		ip := dataArray[0]
		port, err := strconv.Atoi(dataArray[1])
		if utils.CheckIPV4(ip) == false || err != nil {
			continue
		}
		if !tx.Result.HasIP(ip) {
			tx.Result.SetIP(ip)
		}
		if !tx.Result.HasPort(ip, port) {
			tx.Result.SetPort(ip, port)
		}
		service := s.FindService(port, "")
		tx.Result.SetPortAttr(ip, port, PortAttrResult{
			Source:  "txportmap",
			Tag:     "service",
			Content: service,
		})
		//将其它的信息加入到bannner中：
		allDataArray = allDataArray[1:]
		if len(allDataArray) > 0 {
			banner := strings.TrimSpace(strings.Join(allDataArray, " "))
			if len(banner) > 0 {
				tx.Result.SetPortAttr(ip, port, PortAttrResult{
					Source:  "txportmap",
					Tag:     "banner",
					Content: banner,
				})
				//正则检查是否有[200],[403]，增加端口的状态码
				patternStatus := `\[(\d{3})\]`
				r := regexp.MustCompile(patternStatus)
				statusCodeResult := r.FindStringSubmatch(banner)
				if len(statusCodeResult) >= 2 {
					tx.Result.IPResult[ip].Ports[port].Status = statusCodeResult[1]
				}
			}
		}
	}
}
