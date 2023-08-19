package onlineapi

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/domainscan"
	"github.com/hanc00l/nemo_go/pkg/task/portscan"
	"github.com/hanc00l/nemo_go/pkg/utils"
	"io"
	"net/http"
	"strings"
	"time"
)

type Quake struct {
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

func (q *Quake) MakeSearchSyntax(syntax map[SyntaxType]string, condition SyntaxType, checkMod SyntaxType, value string) string {
	if condition == Not {
		// NOT title:"百度"
		return fmt.Sprintf("%s %s:\"%s\"", syntax[condition], syntax[checkMod], value)
	}
	// body:"百度"
	return fmt.Sprintf("%s%s\"%s\"", syntax[checkMod], syntax[condition], value)
}

func (q *Quake) GetSyntaxMap() (syntax map[SyntaxType]string) {
	syntax = make(map[SyntaxType]string)
	syntax[And] = "AND"
	syntax[Or] = "OR"
	syntax[Equal] = ":"
	syntax[Not] = "NOT"
	syntax[After] = "(NOT SUPPORT YET)"
	syntax[Title] = "title"
	syntax[Body] = "body"

	return
}

func (q *Quake) GetQueryString(domain string, config OnlineAPIConfig, filterKeyword map[string]struct{}) (query string) {
	if utils.CheckIPV4(domain) || utils.CheckIPV4Subnet(domain) {
		query = fmt.Sprintf("ip:\"%s\"", domain)
	} else {
		query = fmt.Sprintf("domain:\"%s\"", domain)
	}
	if words := q.getFilterTitleKeyword(filterKeyword); len(words) > 0 {
		query = fmt.Sprintf("(%s) AND (%s)", query, filterKeyword)
	}
	if config.IsIgnoreOutofChina {
		query = fmt.Sprintf("(%s) AND country:\"CN\" AND NOT province:\"Hongkong\"", query)
	}
	return
}

func (q *Quake) getFilterTitleKeyword(filterKeyword map[string]struct{}) string {
	var words []string
	for k := range filterKeyword {
		words = append(words, fmt.Sprintf("NOT body:\"%s\"", k))
	}

	return strings.Join(words, " AND ")
}

func (q *Quake) Run(query string, apiKey string, pageIndex int, pageSize int, config OnlineAPIConfig) (pageResult []onlineSearchResult, sizeTotal int, err error) {
	client := &http.Client{
		Timeout: time.Duration(30) * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}}
	data := quakePostData{
		Query:       query,
		Latest:      true,
		IgnoreCache: true,
		ShortCuts:   []string{},
		Start:       (pageIndex - 1) * pageSize,
		Size:        pageSize,
		Include:     []string{"ip", "port", "hostname", "transport", "service.name", "service.http.host", "service.http.title"},
	}
	jsonData, _ := json.Marshal(data)
	var request *http.Request
	request, err = http.NewRequest("POST", "https://quake.360.cn/api/v3/search/quake_service", bytes.NewBuffer(jsonData))
	if err != nil {
		return
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-QuakeToken", apiKey)
	request.Header.Add("User-Agent", userAgent)
	var response *http.Response
	response, err = client.Do(request)
	if err != nil {
		return
	}
	if response.Body != nil {
		defer response.Body.Close()
		var body []byte
		body, err = io.ReadAll(response.Body)
		if err != nil {
			return
		}
		if strings.Contains(string(body), "/quake/login") {
			return nil, 0, errors.New("quake token invalid")
		}
		pageResult, _, sizeTotal = q.parseQuakeSearchResult(body)
	}

	return
}

func (q *Quake) parseQuakeSearchResult(queryResult []byte) (result []onlineSearchResult, finish bool, sizeTotal int) {
	var serviceInfo QuakeServiceInfo
	err := json.Unmarshal(queryResult, &serviceInfo)
	if err != nil {
		//json数据反序列化失败
		//如果是json: cannot unmarshal object into Go struct field QuakeServiceInfo.data of type []struct { Time time.Time "json:\"time\""; Transport string "json:\"transport\""; Service struct { HTTP struct
		//则基本上是API的key失效，或积分不足无法读取
		logging.RuntimeLog.Error(err)
		logging.CLILog.Error(err)
		return
	}
	if strings.HasPrefix(serviceInfo.Message, "Successful") == false {
		logging.CLILog.Errorf("Quake Search Error:%s", serviceInfo.Message)
		return
	}
	for _, data := range serviceInfo.Data {
		qsr := onlineSearchResult{
			Host:   data.Service.HTTP.Host,
			IP:     data.IP,
			Port:   fmt.Sprintf("%d", data.Port),
			Title:  data.Service.HTTP.Title,
			Server: data.Service.HTTP.Server,
		}
		result = append(result, qsr)
	}
	sizeTotal = serviceInfo.Meta.Pagination.Total
	// 如果是API有效、正确获取到数据，count为0，表示已是最后一页了
	if serviceInfo.Meta.Pagination.Count == 0 || sizeTotal == 0 {
		finish = true
	}

	return
}

func (q *Quake) ParseContentResult(content []byte) (ipResult portscan.Result, domainResult domainscan.Result) {
	logging.RuntimeLog.Error("not apply")
	logging.CLILog.Error("not apply")

	return
}
