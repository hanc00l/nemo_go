package onlineapi

import (
	"encoding/json"
	"github.com/hanc00l/nemo_go/v3/pkg/logging"
	"github.com/likexian/whois"
	whoisparser "github.com/likexian/whois-parser"
	"golang.org/x/net/proxy"
)

type Whois struct {
	IsProxy bool
}

func (w *Whois) Run(domain string, apiKey string) (result QueryDataResult) {
	c := whois.NewClient()
	c.SetDialer(proxy.FromEnvironment())
	text, err := c.Whois(domain, "")
	if err != nil {
		return
	}
	whoisInfo, err := whoisparser.Parse(text)
	if err != nil {
		return
	}
	//去除不常用的记录
	whoisInfo.Billing = nil
	whoisInfo.Administrative = nil
	whoisInfo.Technical = nil

	data, err := json.Marshal(whoisInfo)
	if err != nil {
		logging.RuntimeLog.Error("whoisparser json marshal fail", err)
		return
	}

	result.Domain = domain
	result.Category = "whois"
	result.Content = string(data)

	return
}
