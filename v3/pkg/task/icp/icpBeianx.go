package icp

import (
	"encoding/json"
	"fmt"
	"github.com/hanc00l/nemo_go/v3/pkg/logging"
	"github.com/hanc00l/nemo_go/v3/pkg/utils"
	"net/http"
	"net/url"
	"strings"
)

// ICPQueryBeianx 域名ICP备案查询组织机构
type ICPQueryBeianx struct {
	IsProxy         bool `query:"isProxy"`
	IsDisableLookup bool
}

// ICPPlusQueryBeianx 根据组织机构名称查询ICP备案信息
type ICPPlusQueryBeianx struct {
	IsProxy         bool `query:"isProxy"`
	IsDisableLookup bool
}

// ICPPlus2QueryBeianx　域名查询备案名称-根据组织机构查询ICP备案域名
type ICPPlus2QueryBeianx struct {
	IsProxy         bool `query:"isProxy"`
	IsDisableLookup bool
}

type ICPResultBeianx struct {
	Domain   string `json:"domain"`
	Unit     string `json:"unit"`
	Kind     string `json:"kind"`
	ICP      string `json:"icp"`
	MainICP  string `json:"main_icp"`
	PassTime string `json:"pass_time"` // 使用 time.Time 类型
}

type ICPResponseBeianx struct {
	Data    []ICPResultBeianx `json:"data"`
	Elapsed int               `json:"elapsed"`
	Code    int               `json:"code"`
	Msg     string            `json:"msg"`
	Desc    string            `json:"desc"`
}

// 公共查询方法
func queryAPI(keyword, apiKey string, isProxy bool) (*ICPResponseBeianx, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("empty API key")
	}

	params := url.Values{}
	params.Add("keyword", keyword)
	params.Add("api_key", apiKey)

	req, err := http.NewRequest("POST", "https://open.beianx.cn/api/query_icp_v5", strings.NewReader(params.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := utils.GetProxyHttpClient(isProxy)
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	var response ICPResponseBeianx
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	if response.Code != 200 {
		return nil, fmt.Errorf("API error: %s (code: %d)", response.Msg, response.Code)
	}

	return &response, nil
}

// 转换结果格式的公共方法
func convertResults(response *ICPResponseBeianx) []DataResult {
	var results []DataResult
	for _, item := range response.Data {
		results = append(results, DataResult{
			Domain:         item.Domain,
			UnitName:       item.Unit,
			CompanyType:    item.Kind,
			SiteLicense:    item.ICP,
			ServiceLicence: item.MainICP,
			VerifyTime:     item.PassTime,
			Source:         "beianx",
		})
	}
	return results
}

func (i *ICPQueryBeianx) Run(domain, apiKey string) []DataResult {
	//先查询是否已有查询结果
	if !i.IsDisableLookup {
		rData := DoLookup(domain)
		if rData.Domain == domain {
			return []DataResult{rData}
		}
	}
	response, err := queryAPI(domain, apiKey, i.IsProxy)
	if err != nil {
		logging.RuntimeLog.Warningf("ICP query failed for domain:%s, error:%v", domain, err)
		return nil
	}
	return convertResults(response)
}

func (i *ICPPlusQueryBeianx) Run(companyName, apiKey string) []DataResult {
	response, err := queryAPI(companyName, apiKey, i.IsProxy)
	if err != nil {
		logging.RuntimeLog.Warningf("ICP query failed for company:%s, error:%v", companyName, err)
		return nil
	}
	return convertResults(response)
}

func (i *ICPPlus2QueryBeianx) Run(domain, apiKey string) []DataResult {
	if apiKey == "" {
		logging.RuntimeLog.Warningf("empty API key for domain: %s", domain)
		return nil
	}

	// 第一步：查询域名的ICP备案信息获取组织名称
	icpResults := (&ICPQueryBeianx{IsProxy: i.IsProxy, IsDisableLookup: i.IsDisableLookup}).Run(domain, apiKey)
	if len(icpResults) == 0 {
		logging.RuntimeLog.Warningf("no ICP info found for domain: %s", domain)
		return nil
	}

	// 第二步：查询该组织的所有ICP备案信息
	return (&ICPPlusQueryBeianx{IsProxy: i.IsProxy, IsDisableLookup: i.IsDisableLookup}).Run(icpResults[0].UnitName, apiKey)
}
