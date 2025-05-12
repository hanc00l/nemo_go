package onlineapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hanc00l/nemo_go/v3/pkg/conf"
	"github.com/hanc00l/nemo_go/v3/pkg/core"
	"github.com/hanc00l/nemo_go/v3/pkg/db"
	"github.com/hanc00l/nemo_go/v3/pkg/logging"
	"github.com/hanc00l/nemo_go/v3/pkg/task/execute"
	"math/rand"
	"strings"
	"time"
)

// Executor 网络空间资产搜索引擎interface
type Executor interface {
	GetRequiredResources() (re []core.RequiredResource)
	GetSyntaxMap() (syntax map[SyntaxType]string)
	MakeSearchSyntax(syntax map[SyntaxType]string, condition SyntaxType, checkMod SyntaxType, value string) string
	GetQueryString(domain string, config execute.OnlineAPIConfig, filterKeyword map[string]struct{}) (query string)
	Run(query string, apiKey string, pageIndex int, pageSize int, config execute.OnlineAPIConfig) (pageResult []OnlineSearchResult, sizeTotal int, err error)
}

// ExecutorQueryAndLookup 用于icp、whois等api的interface
type ExecutorQueryAndLookup interface {
	Run(domain string, apiKey string) (result QueryDataResult)
}

type SyntaxType int

type APIKey struct {
	apiName     string
	apiKey      string
	apiKeyInUse string
}

const (
	And SyntaxType = iota
	Or
	Equal
	Not
	After
	Title
	Body
)

func NewExecutor(executeName string, isProxy bool) (Executor, APIKey) {
	executorMap := map[string]Executor{
		"fofa":   &FOFA{IsProxy: isProxy},
		"hunter": &Hunter{IsProxy: isProxy},
		"quake":  &Quake{IsProxy: isProxy},
	}
	apiKeyMap := map[string]APIKey{
		"fofa":   {apiName: "fofa", apiKey: conf.GlobalWorkerConfig().API.Fofa.Key},
		"hunter": {apiName: "hunter", apiKey: conf.GlobalWorkerConfig().API.Hunter.Key},
		"quake":  {apiName: "quake", apiKey: conf.GlobalWorkerConfig().API.Quake.Key},
	}
	return executorMap[executeName], apiKeyMap[executeName]
}

func NewQueryAndLookupExecutor(executeName string, isProxy bool) (ExecutorQueryAndLookup, APIKey) {
	executorMap := map[string]ExecutorQueryAndLookup{
		"whois":   &Whois{IsProxy: isProxy},
		"icp":     &ICPQuery{IsProxy: isProxy},
		"icpPlus": &ICPPlusQuery{IsProxy: isProxy},
	}
	apiKeyMap := map[string]APIKey{
		"whois":   {apiName: "whois", apiKey: ""},
		"icp":     {apiName: "icp", apiKey: conf.GlobalWorkerConfig().API.ICP.Key},
		"icpPlus": {apiName: "icpPlus", apiKey: conf.GlobalWorkerConfig().API.ICPPlus.Key},
	}
	return executorMap[executeName], apiKeyMap[executeName]
}

func Do(taskInfo execute.ExecutorTaskInfo) (result []OnlineSearchResult) {
	// 执行fofa、hunter、quake
	if taskInfo.Executor != "fofa" && taskInfo.Executor != "hunter" && taskInfo.Executor != "quake" {
		return
	}
	if len(taskInfo.OnlineAPI) <= 0 {
		return
	}
	config, ok := taskInfo.OnlineAPI[taskInfo.Executor]
	if !ok {
		logging.RuntimeLog.Errorf("子任务的executor配置不存在：%s", taskInfo.Executor)
		return
	}
	executor, apiKey := NewExecutor(taskInfo.Executor, taskInfo.IsProxy)
	if executor == nil {
		logging.RuntimeLog.Errorf("子任务的executor不存在：%s", taskInfo.Executor)
		return
	}
	re := executor.GetRequiredResources()
	if len(re) > 0 {
		err := core.CheckRequiredResource(re, false)
		if err != nil {
			logging.RuntimeLog.Errorf("任务资源检查和请求失败:%s", err.Error())
			return
		}
	}
	// 请求参数的限制在任务里设置，不再从全局配置里读取
	//config.SearchLimitCount = conf.GlobalWorkerConfig().API.SearchLimitCount
	//if config.SearchPageSize = conf.GlobalWorkerConfig().API.SearchPageSize; config.SearchPageSize <= 0 {
	//	config.SearchPageSize = pageSizeDefault
	//}
	filterKeyword := loadFilterKeyword()
	config.Target = taskInfo.Target
	for _, line := range strings.Split(taskInfo.Target, ",") {
		domain := strings.TrimSpace(line)
		if domain == "" {
			continue
		}
		resultQuery := query(executor, config, domain, filterKeyword, apiKey)
		result = append(result, resultQuery...)
	}

	return
}

func DoQuery(taskInfo execute.ExecutorTaskInfo) (result []QueryDataResult) {
	// 执行whois、icp
	if taskInfo.Executor != "whois" && taskInfo.Executor != "icp" {
		return
	}
	if len(taskInfo.OnlineAPI) <= 0 {
		return
	}
	_, ok := taskInfo.OnlineAPI[taskInfo.Executor]
	if !ok {
		logging.RuntimeLog.Errorf("子任务的executor配置不存在：%s", taskInfo.Executor)
		return
	}
	executor, apiKey := NewQueryAndLookupExecutor(taskInfo.Executor, taskInfo.IsProxy)
	if executor == nil {
		logging.RuntimeLog.Errorf("子任务的executor不存在：%s", taskInfo.Executor)
		return
	}
	for _, target := range strings.Split(taskInfo.Target, ",") {
		// 根据域名查询备案信息或whois
		// 先从数据库中查询数据，如果存在直接返回
		rData := DoLookup(target, taskInfo.Executor)
		if rData.Content != "" {
			result = append(result, rData)
			continue
		}
		// 数据库中不存在，则查询api
		r := executor.Run(target, apiKey.apiKey)
		if r.Content != "" {
			result = append(result, r)
		}
	}

	return
}

func DoICPPlusQuery(taskInfo execute.ExecutorTaskInfo) (result []QueryDataResult) {
	// 但是executor配置还是在onlineapi中；因此这里要注意
	executor, apiKey := NewQueryAndLookupExecutor(taskInfo.Executor, taskInfo.IsProxy)
	if executor == nil {
		logging.RuntimeLog.Errorf("子任务的executor不存在：%s", taskInfo.Executor)
		return
	}
	for _, target := range strings.Split(taskInfo.Target, ",") {
		//根据组织机构查询备案信息
		r := executor.Run(target, apiKey.apiKey)
		if r.Content == "" {
			continue
		}
		// 查询的结果为数组，因此需要反序列化后，重新赋值给QueryDataResult
		var resultData []ICPInfo
		err := json.Unmarshal([]byte(r.Content), &resultData)
		if err != nil {
			logging.RuntimeLog.Error(err)
			continue
		}
		for _, data := range resultData {
			icpInfoData, _ := json.Marshal(data)
			result = append(result, QueryDataResult{
				Domain:   data.Domain,
				Category: db.QueryICP,
				Content:  string(icpInfoData),
			})
		}
	}

	return
}

func DoLookup(domain, category string) (result QueryDataResult) {
	qData := db.QueryDocument{Domain: domain, Category: category}
	rData := db.QueryDocument{}
	err := core.CallXClient("LookupQueryData", &qData, &rData)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return
	}
	result.Domain = rData.Domain
	result.Category = rData.Category
	result.Content = rData.Content

	return
}

// query 查询一个domain
func query(executor Executor, config execute.OnlineAPIConfig, domain string, filterKeyword map[string]struct{}, apiKey APIKey) (result []OnlineSearchResult) {
	queryString := executor.GetQueryString(domain, config, filterKeyword)
	pageResult, sizeTotal, err := retriedQuery(executor, config, queryString, 1, config.SearchPageSize, &apiKey)
	if err != nil {
		logging.RuntimeLog.Error(err)
		logging.CLILog.Error(err)
		return
	}
	if config.SearchLimitCount > 0 && sizeTotal > config.SearchLimitCount {
		msg := fmt.Sprintf("%s 搜索 %s 结果超过限制， total:%d, limited to:%d", apiKey.apiName, domain, sizeTotal, config.SearchLimitCount)
		logging.RuntimeLog.Warning(msg)
		sizeTotal = config.SearchLimitCount
	}
	result = append(result, pageResult...)
	pageTotalNum := sizeTotal / config.SearchPageSize
	if sizeTotal%config.SearchPageSize > 0 {
		pageTotalNum++
	}
	for i := 2; i <= pageTotalNum; i++ {
		pageResult, _, err = retriedQuery(executor, config, queryString, i, config.SearchPageSize, &apiKey)
		if err != nil {
			logging.RuntimeLog.Error(err)
			logging.CLILog.Error(err)
			return
		}
		result = append(result, pageResult...)
		time.Sleep(1 * time.Second)
	}
	return
}

// retriedQuery 执行一次查询，允许重试N次
func retriedQuery(executor Executor, config execute.OnlineAPIConfig, query string, pageIndex int, pageSize int, apiKey *APIKey) (pageResult []OnlineSearchResult, sizeTotal int, err error) {
	RETRIED := 3
	allKeys := strings.Split(apiKey.apiKey, ",")
	if len(allKeys) > 1 {
		RETRIED = 2 * len(allKeys)
	}
	if apiKey.apiKeyInUse == "" {
		if apiKey.apiKeyInUse = getOneAPIKey(apiKey); apiKey.apiKeyInUse == "" {
			return nil, 0, errors.New(fmt.Sprintf("%s没有可用的key", apiKey.apiName))
		}
	}
	var retriedCount int
	for ; retriedCount < RETRIED; retriedCount++ {
		pageResult, sizeTotal, err = executor.Run(query, apiKey.apiKeyInUse, pageIndex, pageSize, config)
		if err == nil {
			return
		}
		msg := fmt.Sprintf("查询失败：%s -> key %s，error:%v", apiKey.apiName, desensitizationAPIKey(*apiKey), err)
		logging.RuntimeLog.Error(msg)
		logging.CLILog.Error(msg)

		if apiKey.apiKeyInUse = getOneAPIKey(apiKey); apiKey.apiKeyInUse == "" {
			return nil, 0, errors.New(fmt.Sprintf("%s没有可用的key", apiKey.apiName))
		}
	}
	if retriedCount >= RETRIED {
		return nil, 0, errors.New(fmt.Sprintf("%s 查询重试次数达到上限", apiKey.apiName))
	}
	return
}

// desensitizationAPIKey 脱敏APIKey，格式为xxxx****xxxx或者xxxx****
func desensitizationAPIKey(key APIKey) string {
	l := len(key.apiKeyInUse)
	if l == 0 {
		return ""
	}
	if l >= 4 {
		keyPre := key.apiKeyInUse[:4]
		if l >= 8 {
			var keyMid, keyEnd string
			if l >= 12 {
				keyMid = "****"
				keyEnd = key.apiKeyInUse[l-4:]
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
func getOneAPIKey(apiKey *APIKey) string {
	allKeys := strings.Split(apiKey.apiKey, ",")

	if len(allKeys) == 0 {
		return ""
	}
	if len(allKeys) == 1 {
		return allKeys[0]
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for {
		n := r.Intn(len(allKeys))
		if allKeys[n] != apiKey.apiKeyInUse {
			return allKeys[n]
		}
	}
}

// loadFilterKeyword 从文件中加载需要过滤的标题关键词
func loadFilterKeyword() (filterKeyword map[string]struct{}) {
	filterKeyword = make(map[string]struct{})
	/*
		inputFile, err := os.Open(filepath.Join(conf.GetRootPath(), "thirdparty/custom/onlineapi_filter_keyword.txt"))
		if err != nil {
			logging.RuntimeLog.Error(err)
			logging.CLILog.Error(err)
			return
		}
		defer inputFile.Close()

		scanner := bufio.NewScanner(inputFile)
		for scanner.Scan() {
			text := strings.ToLower(scanner.Text())
			if text == "" || strings.HasPrefix(text, "#") {
				continue
			}
			if _, ok := filterKeyword[text]; !ok {
				filterKeyword[text] = struct{}{}
			}
		}
	*/
	return
}
