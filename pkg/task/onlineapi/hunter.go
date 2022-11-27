package onlineapi

import (
	"bytes"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/custom"
	"github.com/hanc00l/nemo_go/pkg/task/domainscan"
	"github.com/hanc00l/nemo_go/pkg/task/portscan"
	"github.com/hanc00l/nemo_go/pkg/utils"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Hunter struct {
	//Config 配置参数：查询的目标、关联的组织
	Config OnlineAPIConfig
	//Result quake api查询后的结果
	Result []onlineSearchResult
	//DomainResult 整理后的域名结果
	DomainResult domainscan.Result
	//IpResult 整理后的IP结果
	IpResult portscan.Result
}

// HunterServiceInfo 查询结果的返回数据
type HunterServiceInfo struct {
	Code    int    `json:"code"`
	Message string `json:"msg"`
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

// NewHunter 创建Hunter对像
func NewHunter(config OnlineAPIConfig) *Hunter {
	return &Hunter{Config: config}
}

// Do 执行查询
func (h *Hunter) Do() {
	if conf.GlobalWorkerConfig().API.Hunter.Key == "" {
		logging.RuntimeLog.Error("no hunter api key,exit hunter search")
		return
	}
	for _, line := range strings.Split(h.Config.Target, ",") {
		domain := strings.TrimSpace(line)
		if domain == "" {
			continue
		}
		h.RunHunter(domain)
	}
}

// RunHunter 调用API查询一个IP或域名
func (h *Hunter) RunHunter(domain string) {
	var query string
	if utils.CheckIPV4(domain) || utils.CheckIPV4Subnet(domain) {
		query = fmt.Sprintf("ip=\"%s\"", domain)
	} else {
		query = fmt.Sprintf("domain=\"%s\" or cert.subject=\"%s\"", domain, domain)
	}
	if h.Config.IsIgnoreOutofChina {
		query = fmt.Sprintf("(%s) and ip.country=\"CN\" and ip.province!=\"香港\"", query)
	}
	// 查询第1页，并获取总共记录数量
	pageResult, sizeTotal := h.retriedPageSearch(query, 1)
	h.Result = append(h.Result, pageResult...)
	// 计算需要查询的页数
	pageTotalNum := sizeTotal / pageSize
	if sizeTotal%pageSize > 0 {
		pageTotalNum++
	}
	for i := 2; i <= pageTotalNum; i++ {
		time.Sleep(1 * time.Second)
		pageResult, _ = h.retriedPageSearch(query, i)
		h.Result = append(h.Result, pageResult...)
	}
	h.parseResult()
}

func (h *Hunter) retriedPageSearch(query string, page int) (pageResult []onlineSearchResult, sizeTotal int) {
	RETRIED := 3
	//查询的起始时间段：最近3个月数据
	endTime := time.Now()
	startTime := endTime.AddDate(0, -3, 0)
	for i := 0; i < RETRIED; i++ {
		var serviceInfo HunterServiceInfo
		request, err := http.NewRequest(http.MethodGet, "https://hunter.qianxin.com/openApi/search", nil)
		if err != nil {
			logging.CLILog.Printf("Hunter Search Error:%s", err.Error())
		}
		params := make(url.Values)
		params.Add("api-key", conf.GlobalWorkerConfig().API.Hunter.Key)
		params.Add("search", base64.URLEncoding.EncodeToString([]byte(query)))
		params.Add("page", strconv.Itoa(page))
		params.Add("page_size", strconv.Itoa(pageSize))
		params.Add("is_web", "3") //资产类型，1代表”web资产“，2代表”非web资产“，3代表”全部“
		params.Add("start_time", startTime.Format("2006-01-02 15:04:05"))
		params.Add("end_time", endTime.Format("2006-01-02 15:04:05"))
		request.URL.RawQuery = params.Encode()
		resp, err := http.DefaultClient.Do(request)
		if err != nil {
			logging.CLILog.Printf("Hunter Search Error:%s", err.Error())
			continue
		}
		content, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			logging.CLILog.Printf("Hunter Search Error:%s", err.Error())
			continue
		}
		err = json.Unmarshal(content, &serviceInfo)
		if err != nil {
			logging.CLILog.Printf("Hunter Search Error:%s", err.Error())
			continue
		}
		if serviceInfo.Code != 200 {
			logging.CLILog.Printf("Hunter Search Error:%s", serviceInfo.Message)
			continue
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
		if len(pageResult) > 0 {
			break
		}
	}
	return
}

// ParseCSVContentResult 解析零零信安中导出的CSV文本结果
func (h *Hunter) ParseCSVContentResult(content []byte) {
	s := custom.NewService()
	if h.IpResult.IPResult == nil {
		h.IpResult.IPResult = make(map[string]*portscan.IPResult)
	}
	if h.DomainResult.DomainResult == nil {
		h.DomainResult.DomainResult = make(map[string]*domainscan.DomainResult)
	}
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
			if h.DomainResult.HasDomain(domain) == false {
				h.DomainResult.SetDomain(domain)
			}
			if len(ip) > 0 {
				h.DomainResult.SetDomainAttr(domain, domainscan.DomainAttrResult{
					Source:  "hunter",
					Tag:     "A",
					Content: ip,
				})
			}
			if len(title) > 0 {
				h.DomainResult.SetDomainAttr(domain, domainscan.DomainAttrResult{
					Source:  "hunter",
					Tag:     "title",
					Content: title,
				})
			}
			if len(banners) > 0 {
				for _, banner := range banners {
					if len(banner) > 0 {
						h.DomainResult.SetDomainAttr(domain, domainscan.DomainAttrResult{
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
		if h.IpResult.HasIP(ip) == false {
			h.IpResult.SetIP(ip)
		}
		if h.IpResult.HasPort(ip, port) == false {
			h.IpResult.SetPort(ip, port)
		}
		if len(title) > 0 {
			h.IpResult.SetPortAttr(ip, port, portscan.PortAttrResult{
				Source:  "hunter",
				Tag:     "title",
				Content: title,
			})
		}
		if len(service) <= 0 || service == "unknown" {
			service = s.FindService(port, "")
		}
		if len(service) > 0 {
			h.IpResult.SetPortAttr(ip, port, portscan.PortAttrResult{
				Source:  "hunter",
				Tag:     "service",
				Content: service,
			})
		}
		if len(banners) > 0 {
			for _, banner := range banners {
				if len(banner) > 0 {
					h.IpResult.SetPortAttr(ip, port, portscan.PortAttrResult{
						Source:  "hunter",
						Tag:     "banner",
						Content: banner,
					})
				}
			}
		}
	}
}
