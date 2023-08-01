package onlineapi

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/custom"
	"github.com/hanc00l/nemo_go/pkg/task/domainscan"
	"github.com/hanc00l/nemo_go/pkg/task/portscan"
	"github.com/hanc00l/nemo_go/pkg/utils"
	"io"
	"net/http"
	"strings"
	"time"
)

type Quake struct {
	//Config 配置参数：查询的目标、关联的组织
	Config OnlineAPIConfig
	//Result quake api查询后的结果
	Result []onlineSearchResult
	//DomainResult 整理后的域名结果
	DomainResult domainscan.Result
	//IpResult 整理后的IP结果
	IpResult portscan.Result
}

type quakePostData struct {
	Query       string   `json:"query"`
	Start       int      `json:"start"`
	Size        int      `json:"size"`
	Latest      bool     `json:"latest"`
	IgnoreCache bool     `json:"ignore_cache"`
	ShortCuts   []string `json:"shortcuts"`
	Include     []string `json:"include"`
}

// QuakeServiceInfo Quake查询返回数据 from https://github.com/YetClass/QuakeAPI
type QuakeServiceInfo struct {
	//Code:在查询成功时为0（整形），在失败时为p00XX为字符串，因此无法保证正确unmarshal，因此取消code，用Message来判断
	//Code    int    `json:"code"`
	Message string `json:"message"`
	Data    []struct {
		Time      time.Time `json:"time"`
		Transport string    `json:"transport"`
		Service   struct {
			HTTP struct {
				HTMLHash string `json:"html_hash"`
				Favicon  struct {
					Hash     string `json:"hash"`
					Location string `json:"location"`
					Data     string `json:"data"`
				} `json:"favicon"`
				Robots          string `json:"robots"`
				SitemapHash     string `json:"sitemap_hash"`
				Server          string `json:"server"`
				Body            string `json:"body"`
				XPoweredBy      string `json:"x_powered_by"`
				MetaKeywords    string `json:"meta_keywords"`
				RobotsHash      string `json:"robots_hash"`
				Sitemap         string `json:"sitemap"`
				Path            string `json:"path"`
				Title           string `json:"title"`
				Host            string `json:"host"`
				SecurityText    string `json:"security_text"`
				StatusCode      int    `json:"status_code"`
				ResponseHeaders string `json:"response_headers"`
			} `json:"http"`
			Version  string `json:"version"`
			Name     string `json:"name"`
			Product  string `json:"product"`
			Banner   string `json:"banner"`
			Response string `json:"response"`
		} `json:"service"`
		Images     []interface{} `json:"images"`
		OsName     string        `json:"os_name"`
		Components []interface{} `json:"components"`
		Location   struct {
			DistrictCn  string    `json:"district_cn"`
			ProvinceCn  string    `json:"province_cn"`
			Gps         []float64 `json:"gps"`
			ProvinceEn  string    `json:"province_en"`
			CityEn      string    `json:"city_en"`
			CountryCode string    `json:"country_code"`
			CountryEn   string    `json:"country_en"`
			Radius      float64   `json:"radius"`
			DistrictEn  string    `json:"district_en"`
			Isp         string    `json:"isp"`
			StreetEn    string    `json:"street_en"`
			Owner       string    `json:"owner"`
			CityCn      string    `json:"city_cn"`
			CountryCn   string    `json:"country_cn"`
			StreetCn    string    `json:"street_cn"`
		} `json:"location"`
		Asn       int    `json:"asn"`
		Hostname  string `json:"hostname"`
		Org       string `json:"org"`
		OsVersion string `json:"os_version"`
		IsIpv6    bool   `json:"is_ipv6"`
		IP        string `json:"ip"`
		Port      int    `json:"port"`
	} `json:"data"`
	Meta struct {
		//Modified struct
		Pagination struct {
			Count        int    `json:"count"`
			Total        int    `json:"total"`
			PageIndex    int    `json:"page_index"`
			PageSize     int    `json:"page_size"`
			PaginationID string `json:"pagination_id"`
		} `json:"pagination"`
	} `json:"meta"`
}

// NewQuake 创建Quake对象
func NewQuake(config OnlineAPIConfig) *Quake {
	return &Quake{Config: config}
}

// Do 执行Quake查询
func (q *Quake) Do() {
	if conf.GlobalWorkerConfig().API.Quake.Key == "" {
		logging.RuntimeLog.Warning("no quake api key,exit quake search")
		logging.CLILog.Warning("no quake api key,exit quake search")
		return
	}
	q.Config.SearchLimitCount = conf.GlobalWorkerConfig().API.SearchLimitCount
	if q.Config.SearchPageSize = conf.GlobalWorkerConfig().API.SearchPageSize; q.Config.SearchPageSize <= 0 {
		q.Config.SearchPageSize = pageSizeDefault
	}
	blackDomain := custom.NewBlackDomain()
	blackIP := custom.NewBlackIP()
	for _, line := range strings.Split(q.Config.Target, ",") {
		domain := strings.TrimSpace(line)
		if domain == "" {
			continue
		}
		if utils.CheckIPV4(domain) && blackIP.CheckBlack(domain) {
			continue
		}
		if utils.CheckDomain(domain) && blackDomain.CheckBlack(domain) {
			continue
		}
		q.RunQuake(domain)
	}
}

// RunQuake 调用API接口查询一个ip或域名
func (q *Quake) RunQuake(domain string) {
	var query string
	if utils.CheckIPV4(domain) || utils.CheckIPV4Subnet(domain) {
		query = fmt.Sprintf("ip:\"%s\"", domain)
	} else {
		domainCert := domain
		if strings.HasPrefix(domain, ".") == false {
			domainCert = "." + domain
		}
		query = fmt.Sprintf("domain:\"%s\" OR cert:\"%s\"", domain, domainCert)
	}
	if q.Config.IsIgnoreOutofChina {
		query = fmt.Sprintf("(%s) AND country:\"CN\" AND NOT province:\"Hongkong\"", query)
	}
	//proxy, _ := url.Parse("http://127.0.0.1:8080")
	client := &http.Client{
		Timeout: time.Duration(30) * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			//Proxy:           http.ProxyURL(proxy),
		}}
	// 分页查询
	// 查询第1页（quake页数第0开始），并获取总共记录数量
	pageResult, finish, sizeTotal := q.retriedPageSearch(client, query, 0)
	if q.Config.SearchLimitCount > 0 && sizeTotal > q.Config.SearchLimitCount {
		msg := fmt.Sprintf("search %s result total:%d, limited to:%d", domain, sizeTotal, q.Config.SearchLimitCount)
		logging.RuntimeLog.Warning(msg)
		logging.CLILog.Warning(msg)
		sizeTotal = q.Config.SearchLimitCount
	}
	if len(pageResult) > 0 {
		q.Result = append(q.Result, pageResult...)
	}
	pageTotalNum := sizeTotal / q.Config.SearchPageSize
	if sizeTotal%q.Config.SearchPageSize > 0 {
		pageTotalNum++
	}
	for i := 1; i < pageTotalNum; i++ {
		pageResult, finish, _ = q.retriedPageSearch(client, query, i)
		if len(pageResult) > 0 {
			q.Result = append(q.Result, pageResult...)
		}
		if finish {
			break
		}
	}
	q.parseResult()
}

// retriedPageSearch 分页、N次重试API查询并返回结果
func (q *Quake) retriedPageSearch(client *http.Client, query string, page int) (result []onlineSearchResult, finish bool, sizeTotal int) {
	RetryNumber := 3
	for i := 0; i < RetryNumber; i++ {
		data := quakePostData{
			Query:       query,
			Latest:      true,
			IgnoreCache: true,
			ShortCuts:   []string{},
			Start:       page * q.Config.SearchPageSize,
			Size:        q.Config.SearchPageSize,
			/**
			include 和 exclude参数，可传参字段从获取可筛选服务字段接口获取
			注册用户：
				服务数据：ip，port，hostname，transport，asn，org，service.name，location.country_cn，location.province_cn，location.city_cn、service.http.host
				主机数据：ip、asn，org，location.country_cn，location.province_cn，location.city_cn
			会员用户：
				服务数据：ip，port，hostname，transport，asn，org，service.name，location.country_cn，location.province_cn，location.city_cn，service.http.host，time，service.http.title，service.response，service.cert，components.product_catalog，components.product_type，components.product_level，components.product_vendor，location.country_en，location.province_en，location.city_en，location.district_en，location.district_cn，location.isp，service.http.body
				主机数据：ip、asn，org，location.country_cn，location.province_cn，location.city_cn，hostname，time，location.country_en，location.province_en，location.city_en，location.street_en，location.street_cn，location.owner，location.gps
			*/
			Include: []string{"ip", "port", "hostname", "transport", "service.name", "service.http.host", "service.http.title"},
			//"components.product_catalog", "components.product_type", "components.product_level", "components.product_vendor",
			//"location.country_cn", "location.province_cn", "location.city_cn"
		}
		jsonData, _ := json.Marshal(data)
		searchResult, err := q.sendRequest(client, jsonData)
		if err != nil {
			logging.RuntimeLog.Error(err)
			logging.CLILog.Error(err)
			continue
		}
		//fmt.Println(string(searchResult))
		result, finish, sizeTotal = q.parseQuakeSearchResult(searchResult)
		if finish || len(result) > 0 {
			break
		}
	}
	return
}

// sendRequest 发送查询的HTTP请求
func (q *Quake) sendRequest(client *http.Client, dataBytes []byte) ([]byte, error) {
	request, err := http.NewRequest("POST", "https://quake.360.cn/api/v3/search/quake_service", bytes.NewBuffer(dataBytes))
	defer request.Body.Close()
	if err != nil {
		logging.RuntimeLog.Error(err)
		return nil, err
	}
	defer request.Body.Close()
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-QuakeToken", conf.GlobalWorkerConfig().API.Quake.Key)
	request.Header.Add("User-Agent", userAgent)
	response, err := client.Do(request)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return nil, err
	}
	if response.Body != nil {
		defer response.Body.Close()
		body, err := io.ReadAll(response.Body)
		if err != nil {
			logging.RuntimeLog.Error(err)
			return nil, err
		}
		if strings.Contains(string(body), "/quake/login") {
			return nil, errors.New("quake token invalid")
		}
		return body, nil
	}
	return nil, nil
}
