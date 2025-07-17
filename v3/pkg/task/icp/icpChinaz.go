package icp

import (
	"encoding/json"
	"fmt"
	"github.com/hanc00l/nemo_go/v3/pkg/logging"
	"github.com/hanc00l/nemo_go/v3/pkg/utils"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// ICPQueryChinaz 域名ICP备案查询组织机构
type ICPQueryChinaz struct {
	IsDisableLookup bool
	IsProxy         bool
}

// ICPPlusQueryChinaz 根据组织机构名称查询ICP备案信息
type ICPPlusQueryChinaz struct {
	IsDisableLookup bool
	IsProxy         bool
}

// ICPPlus2QueryChinaz　域名查询备案名称-根据组织机构查询ICP备案域名
type ICPPlus2QueryChinaz struct {
	IsDisableLookup bool
	IsProxy         bool
}

type ICPResultChinaz struct {
	CompanyName    string `json:"CompanyName"`
	CompanyType    string `json:"CompanyType"`
	SiteLicense    string `json:"SiteLicense"`
	VerifyTime     string `json:"VerifyTime"`
	Domain         string `json:"Domain"`
	ServiceLicence string `json:"ServiceLicence"`
}

type ICPPlusResultChinaz struct {
	UnitName       string `json:"UnitName"`
	CompanyType    string `json:"CompanyType"`
	SiteLicense    string `json:"SiteLicense"`
	VerifyTime     string `json:"VerifyTime"`
	Domain         string `json:"Domain"`
	ServiceLicence string `json:"ServiceLicence"`
}

type ICPResponseChinaz struct {
	StateCode int             `json:"StateCode"`
	Reason    string          `json:"Reason"`
	Result    ICPResultChinaz `json:"Result"`
}

type ICPPlusResponseChinaz struct {
	Page       int                   `json:"Page"`
	PageSize   int                   `json:"PageSize"`
	Reason     string                `json:"Reason"`
	Result     []ICPPlusResultChinaz `json:"Result"`
	StateCode  int                   `json:"StateCode"`
	TotalCount int                   `json:"TotalCount"`
	TotalPage  int                   `json:"TotalPage"`
}

func (i *ICPPlus2QueryChinaz) Run(domain string, apiKey string) (result []DataResult) {
	if apiKey == "" {
		logging.RuntimeLog.Warningf("查询域名ICP备案失败，空的API: %s", domain)
		return nil
	}
	allKeys := strings.Split(apiKey, "|")
	if len(allKeys) != 2 {
		logging.RuntimeLog.Warningf("查询域名ICP备案失败，API无效: %s", apiKey)
	}
	// 第一步：查询域名的ICP备案信息获取组织名称
	icpQuery := ICPQueryChinaz{IsProxy: i.IsProxy, IsDisableLookup: i.IsDisableLookup}
	icpResults := icpQuery.Run(domain, allKeys[0])

	if len(icpResults) == 0 {
		//logging.RuntimeLog.Warningf("No ICPChinaz information found for domain: %s", domain)
		return nil
	}

	// 获取备案组织名称
	companyName := icpResults[0].UnitName
	if companyName == "" {
		//logging.RuntimeLog.Warningf("No company name found in ICPChinaz results for domain: %s", domain)
		return nil
	}

	// 第二步：查询该组织的所有ICP备案信息
	icpPlusQuery := ICPPlusQueryChinaz{IsProxy: i.IsProxy, IsDisableLookup: i.IsDisableLookup}
	return icpPlusQuery.Run(companyName, allKeys[1])
}

func (i *ICPPlusQueryChinaz) Run(companyName string, apiKey string) (result []DataResult) {
	// 内部函数：执行单页查询并返回响应结构
	queryPage := func(page int) (*ICPPlusResponseChinaz, error) {
		request, err := http.NewRequest(http.MethodGet, "https://openapi.chinaz.net/v1/1001/getdamainplus", nil)
		if err != nil {
			return nil, err
		}

		params := make(url.Values)
		params.Add("ChinazVer", "1.0")
		params.Add("APIKey", selectOneAPIKey(apiKey))
		params.Add("companyname", companyName)
		params.Add("page", strconv.Itoa(page))
		request.URL.RawQuery = params.Encode()

		resp, err := utils.GetProxyHttpClient(i.IsProxy).Do(request)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		var response ICPPlusResponseChinaz
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			return nil, err
		}

		if response.StateCode != 1 {
			return nil, fmt.Errorf("API error: %s (StateCode: %d)", response.Reason, response.StateCode)
		}

		return &response, nil
	}

	// 获取第一页数据
	firstPage, err := queryPage(1)
	if err != nil {
		logging.RuntimeLog.Errorf("ICPChinaz query failed for company:%s, error:%v", companyName, err)
		return
	}

	// 处理所有结果（包括第一页）
	processResults := func(response *ICPPlusResponseChinaz) {
		for _, item := range response.Result {
			result = append(result, DataResult{
				Domain:         item.Domain,
				UnitName:       item.UnitName,
				CompanyType:    item.CompanyType,
				SiteLicense:    item.SiteLicense,
				ServiceLicence: item.ServiceLicence,
				VerifyTime:     item.VerifyTime,
				Source:         "chinaz",
			})
		}
	}

	processResults(firstPage)

	// 获取剩余页数据（如果有）
	for page := 2; page <= firstPage.TotalPage; page++ {
		response, err := queryPage(page)
		if err != nil {
			logging.RuntimeLog.Errorf("Failed to fetch page %d for %s: %v", page, companyName, err)
			continue
		}
		processResults(response)
	}

	return result
}

func (i *ICPQueryChinaz) Run(domain, apiKey string) (result []DataResult) {
	//先查询是否已有查询结果
	if !i.IsDisableLookup {
		rData := DoLookup(domain)
		if rData.Domain == domain {
			result = append(result, rData)
			return result
		}
	}
	request, err := http.NewRequest(http.MethodGet, "https://openapi.chinaz.net/v1/1001/newicp", nil)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return
	}

	params := make(url.Values)
	params.Add("ChinazVer", "1.0")
	params.Add("APIKey", selectOneAPIKey(apiKey))
	params.Add("domain", domain)
	request.URL.RawQuery = params.Encode()

	resp, err := utils.GetProxyHttpClient(i.IsProxy).Do(request)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return
	}
	defer resp.Body.Close()

	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return
	}

	var r ICPResponseChinaz
	err = json.Unmarshal(responseData, &r)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return
	}

	if r.StateCode == 1 && len(r.Result.Domain) > 0 {
		result = append(result, DataResult{
			Domain:         r.Result.Domain,
			UnitName:       r.Result.CompanyName,
			CompanyType:    r.Result.CompanyType,
			SiteLicense:    r.Result.SiteLicense,
			ServiceLicence: r.Result.ServiceLicence,
			VerifyTime:     r.Result.VerifyTime,
			Source:         "chinaz",
		})
	} else {
		logging.RuntimeLog.Warningf("icp query domain:%s, result:%v", domain, r)
	}

	return result
}

func selectOneAPIKey(apiKey string) string {
	keys := strings.Split(apiKey, ",")
	if len(keys) == 0 {
		return ""
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	n := r.Intn(len(keys))
	return keys[n]
}
