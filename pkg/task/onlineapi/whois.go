package onlineapi

import (
	"encoding/json"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/domainscan"
	"github.com/likexian/whois"
	whoisparser "github.com/likexian/whois-parser"
	"golang.org/x/net/proxy"
	"os"
	"path/filepath"
	"strings"
)

type Whois struct {
	Config             WhoisQueryConfig
	WhoisMap           map[string]*whoisparser.WhoisInfo
	QueriedWhoisInfo   map[string]*whoisparser.WhoisInfo
	cachedWhoisInfoNum int
}

// NewWhois 创建Whois对象
func NewWhois(config WhoisQueryConfig) *Whois {
	w := &Whois{
		Config:           config,
		WhoisMap:         make(map[string]*whoisparser.WhoisInfo),
		QueriedWhoisInfo: make(map[string]*whoisparser.WhoisInfo),
	}
	w.loadWhoisCache()
	return w
}

// Do 执行Whois查询
func (w *Whois) Do() {
	tld := domainscan.NewTldExtract()
	for _, domain := range strings.Split(w.Config.Target, ",") {
		fldDomain := tld.ExtractFLD(domain)
		if fldDomain == "" {
			continue
		}
		//从缓存中读取whois信息
		whoisInfo := w.LookupWhois(fldDomain)
		if whoisInfo != nil {
			w.QueriedWhoisInfo[fldDomain] = whoisInfo
			w.cachedWhoisInfoNum++
			continue
		}
		whoisInfo = w.runWhoisQuery(fldDomain)
		if whoisInfo != nil {
			w.QueriedWhoisInfo[fldDomain] = whoisInfo
			w.WhoisMap[fldDomain] = whoisInfo
			w.cachedWhoisInfoNum++
		}
	}
	w.SaveLocalWhoisInfo()
}

// runWhoisQuery 执行查询一个whois
func (w *Whois) runWhoisQuery(domain string) (whoisInfoResult *whoisparser.WhoisInfo) {
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

	return &whoisInfo
}

// loadWhoisCache 从缓存中读取whois信息
func (w *Whois) loadWhoisCache() {
	content, err := os.ReadFile(filepath.Join(conf.GetRootPath(), "thirdparty/whois/whois.cache"))
	if err != nil {
		logging.RuntimeLog.Errorf("Could not open whois cahe file : %v", err)
		return
	}
	err = json.Unmarshal(content, &w.WhoisMap)
	if err != nil {
		logging.RuntimeLog.Errorf("read whois cache fail:%v", err)
	}
}

// SaveLocalWhoisInfo 将whois信息保存为本地文件
func (w *Whois) SaveLocalWhoisInfo() bool {
	data, _ := json.Marshal(w.WhoisMap)
	err := os.WriteFile(filepath.Join(conf.GetRootPath(), "thirdparty/whois/whois.cache"), data, 0666)
	if err != nil {
		logging.RuntimeLog.Errorf("save whois cache fail:%v", err)
		return false
	}
	return true
}

// LookupWhois 查询一个whois信息
func (w *Whois) LookupWhois(domain string) *whoisparser.WhoisInfo {
	tld := domainscan.NewTldExtract()
	fldDomain := tld.ExtractFLD(domain)
	if fldDomain == "" {
		return nil
	}

	_, ok := w.WhoisMap[fldDomain]
	if ok {
		return w.WhoisMap[fldDomain]
	}
	return nil
}
