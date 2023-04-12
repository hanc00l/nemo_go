package onlineapi

import (
	"bytes"
	"encoding/csv"
	"github.com/hanc00l/nemo_go/pkg/task/custom"
	"github.com/hanc00l/nemo_go/pkg/task/domainscan"
	"github.com/hanc00l/nemo_go/pkg/task/portscan"
	"github.com/hanc00l/nemo_go/pkg/utils"
	"io"
	"strconv"
	"strings"
)

type ZeroZone struct {
	Config OnlineAPIConfig
	//DomainResult 整理后的域名结果
	DomainResult domainscan.Result
	//IpResult 整理后的IP结果
	IpResult portscan.Result
}

func NewZeroZone(config OnlineAPIConfig) *ZeroZone {
	return &ZeroZone{Config: config}
}

// ParseCSVContentResult 解析零零信安中导出的CSV文本结果
func (z *ZeroZone) ParseCSVContentResult(content []byte) {
	s := custom.NewService()
	if z.IpResult.IPResult == nil {
		z.IpResult.IPResult = make(map[string]*portscan.IPResult)
	}
	if z.DomainResult.DomainResult == nil {
		z.DomainResult.DomainResult = make(map[string]*domainscan.DomainResult)
	}
	blackDomain := domainscan.NewBlankDomain()
	r := csv.NewReader(bytes.NewReader(content))
	for index := 0; ; index++ {
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		//忽略第一行的标题行
		if err != nil || index == 0 {
			continue
		}
		domain := utils.HostStrip(strings.TrimSpace(row[1]))
		ip := strings.TrimSpace(row[4])
		port, portErr := strconv.Atoi(row[2])
		title := strings.TrimSpace(row[5])
		service := strings.TrimSpace(row[7])
		//域名属性：
		if len(domain) > 0 && utils.CheckIPV4(domain) == false {
			if blackDomain.CheckBlank(domain) {
				continue
			}
			if z.DomainResult.HasDomain(domain) == false {
				z.DomainResult.SetDomain(domain)
			}
			if len(ip) > 0 {
				z.DomainResult.SetDomainAttr(domain, domainscan.DomainAttrResult{
					Source:  "0zone",
					Tag:     "A",
					Content: ip,
				})
			}
			if len(title) > 0 {
				z.DomainResult.SetDomainAttr(domain, domainscan.DomainAttrResult{
					Source:  "0zone",
					Tag:     "title",
					Content: title,
				})
			}
		}
		//IP属性（由于不是主动扫描，忽略导入StatusCode）
		if len(ip) == 0 || utils.CheckIPV4(ip) == false || portErr != nil {
			continue
		}
		if z.IpResult.HasIP(ip) == false {
			z.IpResult.SetIP(ip)
		}
		if z.IpResult.HasPort(ip, port) == false {
			z.IpResult.SetPort(ip, port)
		}
		if len(title) > 0 {
			z.IpResult.SetPortAttr(ip, port, portscan.PortAttrResult{
				Source:  "0zone",
				Tag:     "title",
				Content: title,
			})
		}
		if len(service) <= 0 || service == "unknown" {
			service = s.FindService(port, "")
		}
		if len(service) > 0 {
			z.IpResult.SetPortAttr(ip, port, portscan.PortAttrResult{
				Source:  "0zone",
				Tag:     "service",
				Content: service,
			})
		}
	}
}
