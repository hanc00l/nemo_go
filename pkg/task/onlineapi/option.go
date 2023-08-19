package onlineapi

import (
	"github.com/hanc00l/nemo_go/pkg/task/domainscan"
	"github.com/hanc00l/nemo_go/pkg/task/portscan"
)

// Engine 网络空间资产搜索引擎interface
type Engine interface {
	GetSyntaxMap() (syntax map[SyntaxType]string)
	MakeSearchSyntax(syntax map[SyntaxType]string, condition SyntaxType, checkMod SyntaxType, value string) string
	GetQueryString(domain string, config OnlineAPIConfig, filterKeyword map[string]struct{}) (query string)
	Run(domain string, apiKey string, pageIndex int, pageSize int, config OnlineAPIConfig) (pageResult []onlineSearchResult, sizeTotal int, err error)
	ParseContentResult(content []byte) (ipResult portscan.Result, domainResult domainscan.Result)
}

type SyntaxType int

const (
	And SyntaxType = iota
	Or
	Equal
	Not
	After
	Title
	Body
)
