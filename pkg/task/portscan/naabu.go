package portscan

import (
	"bufio"
	"bytes"
	"github.com/hanc00l/nemo_go/pkg/task/custom"
	"github.com/hanc00l/nemo_go/pkg/utils"
	"strconv"
	"strings"
)

// Naabu 导入naabu的扫描结果
type Naabu struct {
	Config Config
	Result Result
}

// NewNaabu 创建Naabu对象
func NewNaabu(config Config) *Naabu {
	return &Naabu{Config: config}
}

// ParseTxtContentResult 解析naabu扫描的文本结果
func (n *Naabu) ParseTxtContentResult(content []byte) {
	s := custom.NewService()
	if n.Result.IPResult == nil {
		n.Result.IPResult = make(map[string]*IPResult)
	}
	scanner := bufio.NewScanner(bytes.NewReader(content))
	for scanner.Scan() {
		data := scanner.Text()
		if data == "" {
			continue
		}
		dataArray := strings.Split(data, ":")
		if len(dataArray) != 2 {
			continue
		}
		ip := dataArray[0]
		port, err := strconv.Atoi(dataArray[1])
		if utils.CheckIPV4(ip) == false || err != nil {
			continue
		}
		if !n.Result.HasIP(ip) {
			n.Result.SetIP(ip)
		}
		if !n.Result.HasPort(ip, port) {
			n.Result.SetPort(ip, port)
		}
		service := s.FindService(port, "")
		n.Result.SetPortAttr(ip, port, PortAttrResult{
			Source:  "naabu",
			Tag:     "service",
			Content: service,
		})
	}
}
