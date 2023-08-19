package onlineapi

import (
	"bytes"
	"encoding/csv"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/custom"
	"github.com/hanc00l/nemo_go/pkg/task/domainscan"
	"github.com/hanc00l/nemo_go/pkg/task/portscan"
	"github.com/hanc00l/nemo_go/pkg/utils"
	"io"
	"strconv"
	"strings"
)

type ZeroZone struct {
}

func (*ZeroZone) GetQueryString(domain string, config OnlineAPIConfig, filterKeyword map[string]struct{}) (query string) {
	return
}

func (*ZeroZone) Run(domain string, apiKey string, pageIndex int, pageSize int, config OnlineAPIConfig) (pageResult []onlineSearchResult, sizeTotal int, err error) {
	return
}

func (*ZeroZone) MakeSearchSyntax(syntax map[SyntaxType]string, condition SyntaxType, checkMod SyntaxType, value string) string {
	return ""
}
func (*ZeroZone) GetSyntaxMap() (syntax map[SyntaxType]string) {
	return
}
func (z *ZeroZone) ParseContentResult(content []byte) (ipResult portscan.Result, domainResult domainscan.Result) {
	ipResult.IPResult = make(map[string]*portscan.IPResult)
	domainResult.DomainResult = make(map[string]*domainscan.DomainResult)
	s := custom.NewService()
	btc := custom.NewBlackTargetCheck(custom.CheckAll)
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
			if btc.CheckBlack(domain) {
				logging.RuntimeLog.Warningf("%s is in blacklist,skip...", domain)
				continue
			}
			if domainResult.HasDomain(domain) == false {
				domainResult.SetDomain(domain)
			}
			if len(ip) > 0 {
				domainResult.SetDomainAttr(domain, domainscan.DomainAttrResult{
					Source:  "0zone",
					Tag:     "A",
					Content: ip,
				})
			}
			if len(title) > 0 {
				domainResult.SetDomainAttr(domain, domainscan.DomainAttrResult{
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
		if btc.CheckBlack(ip) {
			logging.RuntimeLog.Warningf("%s is in blacklist,skip...", domain)
			continue
		}
		if ipResult.HasIP(ip) == false {
			ipResult.SetIP(ip)
		}
		if ipResult.HasPort(ip, port) == false {
			ipResult.SetPort(ip, port)
		}
		if len(title) > 0 {
			ipResult.SetPortAttr(ip, port, portscan.PortAttrResult{
				Source:  "0zone",
				Tag:     "title",
				Content: title,
			})
		}
		if len(service) <= 0 || service == "unknown" {
			service = s.FindService(port, "")
		}
		if len(service) > 0 {
			ipResult.SetPortAttr(ip, port, portscan.PortAttrResult{
				Source:  "0zone",
				Tag:     "service",
				Content: service,
			})
		}
	}
	return
}
