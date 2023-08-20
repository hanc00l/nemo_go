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
	"strconv"
	"strings"
	"time"
)

type Hunter struct {
}

// HunterServiceInfo 查询结果的返回数据
type HunterServiceInfo struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Total        int    `json:"total"`
		ConsumeQuota string `json:"consume_quota"`
		RestQuota    string `json:"rest_quota"`
		Arr          []struct {
			IP         string `json:"ip"`
			Port       int    `json:"port"`
			Domain     string `json:"domain"`
			Title      string `json:"web_title"`
			Protocol   string `json:"protocol"`
			Country    string `json:"country"`
			City       string `json:"city"`
			Banner     string `json:"banner"`
			StatusCode int    `json:"status_code"`
		} `json:"arr"`
	} `json:"data"`
}

func (h *Hunter) MakeSearchSyntax(syntax map[SyntaxType]string, condition SyntaxType, checkMod SyntaxType, value string) string {
	return fmt.Sprintf("%s%s\"%s\"", syntax[checkMod], syntax[condition], value)
}

func (h *Hunter) GetSyntaxMap() (syntax map[SyntaxType]string) {
	syntax = make(map[SyntaxType]string)
	syntax[And] = "and"
	syntax[Or] = "or"
	syntax[Equal] = "="
	syntax[Not] = "!="
	syntax[After] = "(NOT SUPPORT YET)"
	syntax[Title] = "web.title"
	syntax[Body] = "web.body"

	return
}

func (h *Hunter) GetQueryString(domain string, config OnlineAPIConfig, filterKeyword map[string]struct{}) (query string) {
	if config.SearchByKeyWord {
		query = config.Target
	} else {
		if utils.CheckIPV4(domain) || utils.CheckIPV4Subnet(domain) {
			query = fmt.Sprintf("ip=\"%s\"", domain)
		} else {
			query = fmt.Sprintf("domain=\"%s\"", domain)
		}
	}
	if words := h.getFilterTitleKeyword(filterKeyword); len(words) > 0 {
		query = fmt.Sprintf("(%s) and (%s)", query, words)
	}
	if config.IsIgnoreOutofChina {
		query = fmt.Sprintf("(%s) and (ip.country=\"CN\" and ip.province!=\"香港\")", query)
	}
	return
}

func (h *Hunter) getFilterTitleKeyword(filterKeyword map[string]struct{}) string {
	var words []string
	for k := range filterKeyword {
		words = append(words, fmt.Sprintf("web.body!=\"%s\"", k))
	}

	return strings.Join(words, " and ")
}

func (h *Hunter) Run(query string, apiKey string, pageIndex int, pageSize int, config OnlineAPIConfig) (pageResult []onlineSearchResult, sizeTotal int, err error) {
	var startTime, endTime string
	if len(config.SearchStartTime) > 0 {
		//指定了查询的开始时间
		et := time.Now()
		endTime = et.Format("2006-01-02")
		startTime = config.SearchStartTime
	} else {
		//查询的起始时间段：最近3个月数据
		et := time.Now()
		endTime = et.Format("2006-01-02")
		st := et.AddDate(0, -3, 0)
		startTime = st.Format("2006-01-02")
	}
	var serviceInfo HunterServiceInfo
	request, err := http.NewRequest(http.MethodGet, "https://hunter.qianxin.com/openApi/search", nil)
	if err != nil {
		return
	}
	params := make(url.Values)
	params.Add("api-key", apiKey)
	params.Add("search", base64.StdEncoding.EncodeToString([]byte(query)))
	params.Add("page", strconv.Itoa(pageIndex))
	params.Add("page_size", strconv.Itoa(pageSize))
	params.Add("is_web", "3") //资产类型，1代表”web资产“，2代表”非web资产“，3代表”全部“
	params.Add("start_time", startTime)
	params.Add("end_time", endTime)
	request.URL.RawQuery = params.Encode()
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return
	}
	content, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return
	}
	err = json.Unmarshal(content, &serviceInfo)
	if err != nil {
		return
	}
	if serviceInfo.Code != 200 {
		err = errors.Newf("Hunter Search Error:%s", serviceInfo.Message)
		return
	}
	sizeTotal = serviceInfo.Data.Total
	for _, data := range serviceInfo.Data.Arr {
		qsr := onlineSearchResult{
			IP:     data.IP,
			Domain: data.Domain,
			Host:   data.Domain,
			Port:   fmt.Sprintf("%d", data.Port),
			Title:  data.Title,
		}
		pageResult = append(pageResult, qsr)
	}

	return
}

func (h *Hunter) ParseContentResult(content []byte) (ipResult portscan.Result, domainResult domainscan.Result) {
	s := custom.NewService()
	ipResult.IPResult = make(map[string]*portscan.IPResult)
	domainResult.DomainResult = make(map[string]*domainscan.DomainResult)

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
		domain := strings.TrimSpace(row[6])
		ip := strings.TrimSpace(row[2])
		port, portErr := strconv.Atoi(row[4])
		title := strings.TrimSpace(row[5])
		service := strings.TrimSpace(row[8])
		banners := strings.Split(strings.TrimSpace(row[11]), ",")

		//域名属性：
		if len(domain) > 0 && utils.CheckIPV4(domain) == false {
			if domainResult.HasDomain(domain) == false {
				domainResult.SetDomain(domain)
			}
			if len(ip) > 0 {
				domainResult.SetDomainAttr(domain, domainscan.DomainAttrResult{
					Source:  "hunter",
					Tag:     "A",
					Content: ip,
				})
			}
			if len(title) > 0 {
				domainResult.SetDomainAttr(domain, domainscan.DomainAttrResult{
					Source:  "hunter",
					Tag:     "title",
					Content: title,
				})
			}
			if len(banners) > 0 {
				for _, banner := range banners {
					if len(banner) > 0 {
						domainResult.SetDomainAttr(domain, domainscan.DomainAttrResult{
							Source:  "hunter",
							Tag:     "banner",
							Content: banner,
						})
					}
				}
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
				Source:  "hunter",
				Tag:     "title",
				Content: title,
			})
		}
		if len(service) <= 0 || service == "unknown" {
			service = s.FindService(port, "")
		}
		if len(service) > 0 {
			ipResult.SetPortAttr(ip, port, portscan.PortAttrResult{
				Source:  "hunter",
				Tag:     "service",
				Content: service,
			})
		}
		if len(banners) > 0 {
			for _, banner := range banners {
				if len(banner) > 0 {
					ipResult.SetPortAttr(ip, port, portscan.PortAttrResult{
						Source:  "hunter",
						Tag:     "banner",
						Content: banner,
					})
				}
			}
		}
	}
	return
}
