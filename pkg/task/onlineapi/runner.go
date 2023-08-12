package onlineapi

import (
	"errors"
	"fmt"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/custom"
	"github.com/hanc00l/nemo_go/pkg/task/domainscan"
	"github.com/hanc00l/nemo_go/pkg/task/portscan"
	"math/rand"
	"strings"
	"time"
)

type OnlineSearch struct {
	apiName      string
	apiKey       string
	apiKeyInUse  string
	searchEngine Engine
	//Config 配置参数：查询的目标、关联的组织
	Config OnlineAPIConfig
	//Result quake api查询后的结果
	Result []onlineSearchResult
	//DomainResult 整理后的域名结果
	DomainResult domainscan.Result
	//IpResult 整理后的IP结果
	IpResult portscan.Result
}

func NewOnlineAPISearch(config OnlineAPIConfig, apiName string) *OnlineSearch {
	s := &OnlineSearch{Config: config, apiName: apiName}
	switch apiName {
	case "fofa":
		s.searchEngine = new(FOFA)
		s.apiKey = conf.GlobalWorkerConfig().API.Fofa.Key
	case "hunter":
		s.searchEngine = new(Hunter)
		s.apiKey = conf.GlobalWorkerConfig().API.Hunter.Key
	case "quake":
		s.searchEngine = new(Quake)
		s.apiKey = conf.GlobalWorkerConfig().API.Quake.Key
	case "0zone":
		s.searchEngine = new(ZeroZone)
	}
	s.Config.SearchLimitCount = conf.GlobalWorkerConfig().API.SearchLimitCount
	if s.Config.SearchPageSize = conf.GlobalWorkerConfig().API.SearchPageSize; s.Config.SearchPageSize <= 0 {
		s.Config.SearchPageSize = pageSizeDefault
	}

	return s
}

// Do 执行查询
func (s *OnlineSearch) Do() {
	if s.searchEngine == nil {
		logging.RuntimeLog.Errorf("invalid api:%s,exit search", s.apiName)
		logging.CLILog.Errorf("invalid api:%s,exit search", s.apiName)
		return
	}
	if len(s.apiKey) == 0 {
		logging.RuntimeLog.Warningf("no %s api key,exit search", s.apiName)
		logging.CLILog.Warningf("no %s api key,exit search", s.apiName)
		return
	}
	btc := custom.NewBlackTargetCheck(custom.CheckAll)
	for _, line := range strings.Split(s.Config.Target, ",") {
		domain := strings.TrimSpace(line)
		if domain == "" {
			continue
		}
		if btc.CheckBlack(domain) {
			logging.RuntimeLog.Warningf("%s is in blacklist,skip...", domain)
			continue
		}
		s.Query(domain)
	}
	s.processResult()
}

// Query 查询一个domain
func (s *OnlineSearch) Query(domain string) {
	query := s.searchEngine.GetQueryString(domain, s.Config)
	pageResult, sizeTotal, err := s.retriedQuery(query, 1, s.Config.SearchPageSize)
	if err != nil {
		logging.RuntimeLog.Error(err)
		logging.CLILog.Error(err)
		return
	}
	if s.Config.SearchLimitCount > 0 && sizeTotal > s.Config.SearchLimitCount {
		msg := fmt.Sprintf("%s search %s result total:%d, limited to:%d", s.apiName, domain, sizeTotal, s.Config.SearchLimitCount)
		logging.RuntimeLog.Warning(msg)
		logging.CLILog.Warning(msg)
		sizeTotal = s.Config.SearchLimitCount
	}
	s.Result = append(s.Result, pageResult...)
	pageTotalNum := sizeTotal / s.Config.SearchPageSize
	if sizeTotal%s.Config.SearchPageSize > 0 {
		pageTotalNum++
	}
	for i := 2; i <= pageTotalNum; i++ {
		pageResult, _, err = s.retriedQuery(query, i, s.Config.SearchPageSize)
		if err != nil {
			logging.RuntimeLog.Error(err)
			logging.CLILog.Error(err)
			return
		}
		s.Result = append(s.Result, pageResult...)
		time.Sleep(1 * time.Second)
	}
}

// ParseContentResult 从文件内容中导入结果
func (s *OnlineSearch) ParseContentResult(content []byte) {
	s.IpResult, s.DomainResult = s.searchEngine.ParseContentResult(content)
}

// retriedQuery 执行一次查询，允许重试N次
func (s *OnlineSearch) retriedQuery(query string, pageIndex int, pageSize int) (pageResult []onlineSearchResult, sizeTotal int, err error) {
	RETRIED := 3
	allKeys := strings.Split(s.apiKey, ",")
	if len(allKeys) > 1 {
		RETRIED = 2 * len(allKeys)
	}
	if s.apiKeyInUse == "" {
		if s.apiKeyInUse = s.getOneAPIKey(); s.apiKeyInUse == "" {
			return nil, 0, errors.New(fmt.Sprintf("%s no  key to available", s.apiName))
		}
	}
	var retriedCount int
	for ; retriedCount < RETRIED; retriedCount++ {
		pageResult, sizeTotal, err = s.searchEngine.Run(query, s.apiKeyInUse, pageIndex, pageSize, s.Config)
		if err == nil {
			return
		}
		msg := fmt.Sprintf("api %s with key %s has error:%v", s.apiName, s.desensitizationAPIKey(), err)
		logging.RuntimeLog.Error(msg)
		logging.CLILog.Error(msg)

		if s.apiKeyInUse = s.getOneAPIKey(); s.apiKeyInUse == "" {
			return nil, 0, errors.New(fmt.Sprintf("%s no  key to available", s.apiName))
		}
	}
	if retriedCount >= RETRIED {
		return nil, 0, errors.New(fmt.Sprintf("%s search retried failed to over max", s.apiName))
	}
	return
}

// desensitizationAPIKey 脱敏APIKey，格式为xxxx****xxxx或者xxxx****
func (s *OnlineSearch) desensitizationAPIKey() string {
	l := len(s.apiKeyInUse)
	if l == 0 {
		return ""
	}
	if l >= 4 {
		keyPre := s.apiKeyInUse[:4]
		if l >= 8 {
			var keyMid, keyEnd string
			if l >= 12 {
				keyMid = "****"
				keyEnd = s.apiKeyInUse[l-4:]
			} else {
				keyMid = ""
				keyEnd = "****"
			}
			return fmt.Sprintf("%s%s%s", keyPre, keyMid, keyEnd)
		}
		return fmt.Sprintf("%s****", keyPre)
	}
	return "****"
}

// getOneAPIKey 选择一个查询的key
func (s *OnlineSearch) getOneAPIKey() string {
	allKeys := strings.Split(s.apiKey, ",")

	if len(allKeys) == 0 {
		return ""
	}
	if len(allKeys) == 1 {
		return allKeys[0]
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for {
		n := r.Intn(len(allKeys))
		if allKeys[n] != s.apiKeyInUse {
			return allKeys[n]
		}
	}
}

// processResult 转换查询的IP和域名结果保存
func (s *OnlineSearch) processResult() {
	s.IpResult = portscan.Result{IPResult: make(map[string]*portscan.IPResult)}
	s.DomainResult = domainscan.Result{DomainResult: make(map[string]*domainscan.DomainResult)}

	btc := custom.NewBlackTargetCheck(custom.CheckAll)
	for _, fsr := range s.Result {
		parseIpPort(s.IpResult, fsr, s.apiName, btc)
		parseDomainIP(s.DomainResult, fsr, s.apiName, btc)
	}

	domainscan.FilterDomainHasTooMuchIP(&s.DomainResult)
}
