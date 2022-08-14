package onlineapi

import (
	"encoding/base64"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/domainscan"
	"github.com/hanc00l/nemo_go/pkg/task/portscan"
	"github.com/hanc00l/nemo_go/pkg/utils"
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
	// Create a Resty Client
	var query string
	if utils.CheckIPV4(domain) || utils.CheckIPV4Subnet(domain) {
		query = fmt.Sprintf("ip=\"%s\"", domain)
	} else {
		query = fmt.Sprintf("domain=\"%s\" or cert.subject=\"%s\"", domain, domain)
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
	client := resty.New()
	//查询的起始时间段：最近3个月数据
	endTime := time.Now()
	startTime := endTime.AddDate(0, -3, 0)
	for i := 0; i < RETRIED; i++ {
		var serviceInfo HunterServiceInfo
		_, err := client.R().
			SetQueryParams(map[string]string{
				"api-key":    conf.GlobalWorkerConfig().API.Hunter.Key,
				"search":     base64.URLEncoding.EncodeToString([]byte(query)),
				"page":       fmt.Sprintf("%d", page),
				"page_size":  fmt.Sprintf("%d", pageSize),
				"is_web":     "3", //资产类型，1代表”web资产“，2代表”非web资产“，3代表”全部“
				"start_time": startTime.Format("2006-01-02 15:04:05"),
				"end_time":   endTime.Format("2006-01-02 15:04:05"),
			}).
			SetHeader("User-Agent", userAgent).
			SetResult(&serviceInfo).
			Get("https://hunter.qianxin.com/openApi/search")
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
