package onlineapi

import (
	"encoding/json"
	"github.com/hanc00l/nemo_go/v3/pkg/db"
	"github.com/hanc00l/nemo_go/v3/pkg/logging"
	"github.com/hanc00l/nemo_go/v3/pkg/utils"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type ICPQuery struct {
	IsProxy bool `query:"isProxy"`
}

type ICPPlusQuery struct {
	IsProxy bool `query:"isProxy"`
}

func (i *ICPPlusQuery) Run(companyName string, apiKey string) (result QueryDataResult) {
	request, err := http.NewRequest(http.MethodGet, "https://openapi.chinaz.net/v1/1001/getdamainplus", nil)
	if err != nil {
		return
	}
	params := make(url.Values)
	params.Add("page", "1")
	params.Add("ChinazVer", "1.0")
	params.Add("APIKey", selectOneAPIKey(apiKey))
	params.Add("companyname", companyName)
	request.URL.RawQuery = params.Encode()
	resp, err := utils.GetProxyHttpClient(i.IsProxy).Do(request)
	if err != nil {
		return
	}
	responseData, err := io.ReadAll(resp.Body)
	resp.Body.Close()

	var r icpPlusQueryResult
	err = json.Unmarshal(responseData, &r)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return
	}
	if r.StateCode == 1 && len(r.Result) > 0 {
		// 这里需要将UnitName字段替换成CompanyName字段，因为ICPPlus的接口返回的字段名是UnitName，而ICP的接口返回的字段名是CompanyName
		for k := range r.Result {
			r.Result[k].CompanyName = r.Result[k].UnitName
		}
		result.Domain = companyName
		result.Category = db.QueryICPPlus
		content, _ := json.Marshal(r.Result)
		result.Content = string(content)
	} else {
		logging.RuntimeLog.Warningf("根据组织名称查询备案信息失败, domain:%s,result:%v", companyName, r)
	}
	return
}

func (i *ICPQuery) Run(domain, apiKey string) (result QueryDataResult) {
	request, err := http.NewRequest(http.MethodGet, "https://apidatav2.chinaz.com/single/newicp", nil)
	if err != nil {
		return
	}
	params := make(url.Values)
	params.Add("ChinazVer", "1.0")
	params.Add("key", selectOneAPIKey(apiKey))
	params.Add("domain", domain)
	request.URL.RawQuery = params.Encode()
	resp, err := utils.GetProxyHttpClient(i.IsProxy).Do(request)
	if err != nil {
		return
	}
	responseData, err := io.ReadAll(resp.Body)
	resp.Body.Close()

	var r icpQueryResult
	err = json.Unmarshal(responseData, &r)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return
	}
	if r.StateCode == 1 && len(r.Result.Domain) > 0 {
		result.Domain = domain
		result.Category = db.QueryICP
		content, _ := json.Marshal(r.Result)
		result.Content = string(content)

	} else {
		logging.RuntimeLog.Warningf("icp query domain:%s,result:%v", domain, r)
	}
	return
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
