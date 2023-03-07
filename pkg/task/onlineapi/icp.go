package onlineapi

import (
	"encoding/json"
	"fmt"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/domainscan"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type ICPQuery struct {
	Config           ICPQueryConfig
	ICPMap           map[string]*ICPInfo
	QueriedICPInfo   map[string]*ICPInfo
	cachedICpInfoNum int
}

// NewICPQuery 创建ICP备案查询对象
func NewICPQuery(config ICPQueryConfig) *ICPQuery {
	icp := ICPQuery{
		Config:         config,
		ICPMap:         make(map[string]*ICPInfo),
		QueriedICPInfo: make(map[string]*ICPInfo)}
	icp.loadICPCache()

	return &icp
}

// Do 执行任务
func (i *ICPQuery) Do() {
	tld := domainscan.NewTldExtract()
	for _, domain := range strings.Split(i.Config.Target, ",") {
		fldDomain := tld.ExtractFLD(domain)
		if fldDomain == "" {
			continue
		}
		icpInfo := i.LookupICP(fldDomain)
		if icpInfo != nil {
			i.QueriedICPInfo[fldDomain] = icpInfo
			i.cachedICpInfoNum++
			continue
		}
		icpInfo = i.RunICPQuery(fldDomain)
		if icpInfo != nil {
			i.QueriedICPInfo[fldDomain] = icpInfo
			i.ICPMap[fldDomain] = icpInfo
			i.cachedICpInfoNum++
		}
	}
	//保存到本地的ICP缓存中
	i.SaveLocalICPInfo()
}

// LookupICP 从本地缓存的文件中，查询一个domain的ICP备案
func (i *ICPQuery) LookupICP(domain string) *ICPInfo {
	tld := domainscan.NewTldExtract()
	fldDomain := tld.ExtractFLD(domain)
	if fldDomain == "" {
		return nil
	}

	_, ok := i.ICPMap[fldDomain]
	if ok {
		return i.ICPMap[fldDomain]
	}
	return nil
}

// SaveLocalICPInfo 保存icp信息
func (i *ICPQuery) SaveLocalICPInfo() bool {
	data, _ := json.Marshal(i.ICPMap)
	err := os.WriteFile(filepath.Join(conf.GetRootPath(), "thirdparty/icp/icp.cache"), data, 0666)
	if err != nil {
		logging.RuntimeLog.Errorf("save icp cache fail:%v", err)
		return false
	}
	return true
}

// RunICPQuery 通过API在线查询一个域名的ICP备案信息
func (i *ICPQuery) RunICPQuery(domain string) *ICPInfo {
	if conf.GlobalWorkerConfig().API.ICP.Key == "" {
		logging.RuntimeLog.Error("query key is empty")
		return nil
	}
	url := fmt.Sprintf("https://apidatav2.chinaz.com/single/icp?key=%s&domain=%s", conf.GlobalWorkerConfig().API.ICP.Key, domain)
	resp, err := http.Get(url)
	if resp.StatusCode != 200 {
		logging.RuntimeLog.Errorf("get api status code:%v", resp.Status)
		return nil
	}
	defer resp.Body.Close()
	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return nil
	}
	var r icpQueryResult
	err = json.Unmarshal(responseData, &r)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return nil
	}
	if r.StateCode == 1 {
		r.Result.Domain = domain
		return &r.Result
	} else {
		logging.RuntimeLog.Error(r)
	}
	return nil
}

// loadICPCache 从本地缓存中加载ICP备案信息
func (i *ICPQuery) loadICPCache() {
	content, err := os.ReadFile(filepath.Join(conf.GetRootPath(), "thirdparty/icp/icp.cache"))
	if err != nil {
		logging.RuntimeLog.Errorf("Could not open icp cahe file : %v", err)
		return
	}
	err = json.Unmarshal(content, &i.ICPMap)
	if err != nil {
		logging.RuntimeLog.Errorf("read icp cache fail:%v", err)
	}
}
