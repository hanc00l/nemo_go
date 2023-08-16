package onlineapi

import (
	"bytes"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/hanc00l/nemo_go/pkg/task/custom"
	"github.com/hanc00l/nemo_go/pkg/task/domainscan"
	"github.com/hanc00l/nemo_go/pkg/task/portscan"
	"github.com/hanc00l/nemo_go/pkg/utils"
	"gopkg.in/errgo.v2/fmt/errors"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

type FOFA struct {
}

func (f *FOFA) GetQueryString(domain string, config OnlineAPIConfig) (query string) {
	if config.SearchByKeyWord {
		query = config.Target
	} else {
		if utils.CheckIPV4(domain) || utils.CheckIPV4Subnet(domain) {
			query = fmt.Sprintf("ip=\"%s\"", domain)
		} else {
			// cert.subject相比更精准，但信息量更少；cert="xxx.com"干扰太多，暂时不用（没想法好优的方案）
			// 在域名前加.减少模糊匹配带来的部份干扰
			//domainCert := domain
			//if strings.HasPrefix(domain, ".") == false {
			//	domainCert = "." + domain
			//}
			//query = fmt.Sprintf("domain=\"%s\" || cert=\"%s\" || cert.subject=\"%s\"", domain, domainCert, domainCert)
			query = fmt.Sprintf("domain=\"%s\"", domain)
		}
	}
	if config.IsIgnoreOutofChina {
		query = fmt.Sprintf("(%s) && country=\"CN\" && region!=\"HK\" && region!=\"TW\"  && region!=\"MO\"", query)
	}

	return
}

func (f *FOFA) Run(query string, apiKey string, pageIndex int, pageSize int, config OnlineAPIConfig) (pageResult []onlineSearchResult, sizeTotal int, err error) {
	fields := "domain,host,ip,port,title,country,city,server,banner"
	arr := strings.Split(apiKey, ":")
	if len(arr) != 2 {
		err = errors.Newf("invalid fofa key:%s", apiKey)
		return
	}
	request, err := http.NewRequest(http.MethodGet, "https://fofa.info/api/v1/search/all", nil)
	if err != nil {
		return
	}
	params := make(url.Values)
	params.Add("email", arr[0])
	params.Add("key", arr[1])
	params.Add("qbase64", base64.StdEncoding.EncodeToString([]byte(query)))
	params.Add("fields", fields)
	params.Add("page", strconv.Itoa(pageIndex))
	params.Add("size", strconv.Itoa(pageSize))
	request.URL.RawQuery = params.Encode()
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return
	}
	content, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	pageResult, sizeTotal, err = f.parseFofaSearchResult(content)

	return
}

func (f *FOFA) parseFofaSearchResult(queryResult []byte) (result []onlineSearchResult, sizeTotal int, err error) {
	r := fofaQueryResult{}
	err = json.Unmarshal(queryResult, &r)
	if err != nil {
		return
	}
	if r.IsError {
		err = errors.New(r.ErrorMessage)
		return
	}
	sizeTotal = r.Size
	for _, line := range r.Results {
		fsr := onlineSearchResult{
			Domain: line[0], Host: line[1], IP: line[2], Port: line[3], Title: line[4],
			Country: line[5], City: line[6], Server: line[7], Banner: line[8],
		}
		lowerBanner := strings.ToLower(fsr.Banner)
		//过滤部份无意义的banner
		if strings.HasPrefix(fsr.Banner, "HTTP/") || strings.HasPrefix(lowerBanner, "<html>") || strings.HasPrefix(lowerBanner, "<!doctype html") || strings.HasPrefix(lowerBanner, "<?xml version") {
			fsr.Banner = ""
		}
		//过滤所有的\x00 \x15等不可见字符
		re := regexp.MustCompile("\\\\x[0-9a-f]{2}")
		fsr.Banner = re.ReplaceAllString(fsr.Banner, "")
		result = append(result, fsr)
	}

	return
}

func (f *FOFA) ParseContentResult(content []byte) (ipResult portscan.Result, domainResult domainscan.Result) {
	ipResult.IPResult = make(map[string]*portscan.IPResult)
	domainResult.DomainResult = make(map[string]*domainscan.DomainResult)
	s := custom.NewService()
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
		domain := utils.HostStrip(strings.TrimSpace(row[0]))
		ip := strings.TrimSpace(row[2])
		port, portErr := strconv.Atoi(row[3])
		title := strings.TrimSpace(row[4])
		service := strings.TrimSpace(row[6])
		//域名属性：
		if len(domain) > 0 && utils.CheckIPV4(domain) == false {
			if domainResult.HasDomain(domain) == false {
				domainResult.SetDomain(domain)
			}
			if len(ip) > 0 {
				domainResult.SetDomainAttr(domain, domainscan.DomainAttrResult{
					Source:  "fofa",
					Tag:     "A",
					Content: ip,
				})
			}
			if len(title) > 0 {
				domainResult.SetDomainAttr(domain, domainscan.DomainAttrResult{
					Source:  "fofa",
					Tag:     "title",
					Content: title,
				})
			}
		}
		//IP属性（由于不是主动扫描，忽略导入StatusCode）
		if len(ip) == 0 || utils.CheckIPV4(ip) == false || portErr != nil {
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
				Source:  "fofa",
				Tag:     "title",
				Content: title,
			})
		}
		if len(service) <= 0 || service == "unknown" {
			service = s.FindService(port, "")
		}
		if len(service) > 0 {
			ipResult.SetPortAttr(ip, port, portscan.PortAttrResult{
				Source:  "fofa",
				Tag:     "service",
				Content: service,
			})
		}
	}
	return
}
