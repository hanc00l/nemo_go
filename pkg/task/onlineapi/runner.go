package onlineapi

import (
	"errors"
	"fmt"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/custom"
	"github.com/hanc00l/nemo_go/pkg/task/domainscan"
	"github.com/hanc00l/nemo_go/pkg/task/portscan"
	"github.com/hanc00l/nemo_go/pkg/utils"
	"math/rand"
	"strings"
	"time"
)

type OnlineSearch struct {
	apiName      string
	apiKey       conf.APIKey
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
		s.searchEngine = new(FOFANew)
		s.apiKey = conf.GlobalWorkerConfig().API.Fofa
	case "hunter":
		s.searchEngine = new(Hunter)
		s.apiKey = conf.GlobalWorkerConfig().API.Hunter
	case "quake":
		s.searchEngine = new(Quake)
		s.apiKey = conf.GlobalWorkerConfig().API.Quake
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
	if len(s.apiKey.Key) == 0 {
		logging.RuntimeLog.Warningf("no %s api key,exit search", s.apiName)
		logging.CLILog.Warningf("no %s api key,exit search", s.apiName)
		return
	}
	blackDomain := custom.NewBlackDomain()
	blackIP := custom.NewBlackIP()
	for _, line := range strings.Split(s.Config.Target, ",") {
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
	if s.Config.SearchLimitCount > 0 && sizeTotal > s.Config.SearchPageSize {
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
	if len(s.apiKey.Key) > 1 {
		RETRIED = 2 * len(s.apiKey.Key)
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
		logging.RuntimeLog.Error(err)
		logging.CLILog.Error(err)
		if s.apiKeyInUse = s.getOneAPIKey(); s.apiKeyInUse == "" {
			return nil, 0, errors.New(fmt.Sprintf("%s no  key to available", s.apiName))
		}
	}
	if retriedCount >= RETRIED {
		return nil, 0, errors.New(fmt.Sprintf("%s search retried failed to over max", s.apiName))
	}
	return
}

// getOneAPIKey 选择一个查询的key
func (s *OnlineSearch) getOneAPIKey() string {
	if len(s.apiKey.Key) == 0 {
		return ""
	}
	if len(s.apiKey.Key) == 1 {
		return s.apiKey.Key[0]
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for {
		n := r.Intn(len(s.apiKey.Key))
		if s.apiKey.Key[n] != s.apiKeyInUse {
			return s.apiKey.Key[n]
		}
	}
}

// processResult 转换查询的IP和域名结果保存
func (s *OnlineSearch) processResult() {
	s.IpResult = portscan.Result{IPResult: make(map[string]*portscan.IPResult)}
	s.DomainResult = domainscan.Result{DomainResult: make(map[string]*domainscan.DomainResult)}

	blackDomain := custom.NewBlackDomain()
	blackIP := custom.NewBlackIP()
	for _, fsr := range s.Result {
		parseIpPort(s.IpResult, fsr, s.apiName, blackIP)
		parseDomainIP(s.DomainResult, fsr, s.apiName, blackDomain)
	}

	checkDomainResult(s.DomainResult.DomainResult)
}
