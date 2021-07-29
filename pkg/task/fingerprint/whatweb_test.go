package fingerprint

import (
	"github.com/hanc00l/nemo_go/pkg/task/domainscan"
	"github.com/hanc00l/nemo_go/pkg/task/portscan"
	"regexp"
	"testing"
)

func TestWhatweb_Run(t *testing.T) {
	subdomain := domainscan.SubFinder{
		Config: domainscan.Config{Target: "800best.com"},
	}
	subdomain.Do()
	t.Log(subdomain.Result)

	whatweb := Whatweb{Config: Config{}}
	whatweb.ResultDomainScan = subdomain.Result
	whatweb.Do()
	t.Log(whatweb.ResultDomainScan)

	whatweb.ResultDomainScan.SaveResult(subdomain.Config)
}

func TestWhatweb_Run2(t *testing.T) {
	nmap := &portscan.Nmap{Config: portscan.Config{
		Target:       "47.98.181.116",
		Port:         "80,443",
		OrgId:        nil,
		Rate:         1000,
		Tech:         "-sS",
		CmdBin:       "nmap",
	}}
	nmap.Do()
	t.Log(nmap.Result)

	whatweb := NewWhatweb(Config{})
	whatweb.ResultPortScan = nmap.Result
	whatweb.Do()
	t.Log(whatweb.ResultPortScan)

	nmap.Result.SaveResult(nmap.Config)
}

func Test1(t *testing.T) {
	//s := `http://47.98.181.116:80 [200 OK] Bootstrap, Country[CANADA][CA], HTML5, Domain[47.98.181.116], JQuery, Script[JavaScript], Title[邻里驿站官网——你身边的服务驿站], X-Powered-By[ASP.NET], X-UA-Compatible[IE=edge]`
	s := `http://s9.800best.com [303 See Other] Access-Control-Allow-Methods[GET, PUT, POST, DELETE, PATCH, OPTIONS], Country[CHINA][CN], Domain[202.107.193.31], RedirectLocation[https://account.800best.com/uc/login?service=https://s9.800best.com/ssoCallback&appid=uc16cb111780834e75ba1cfbfab568d78e], UncommonHeaders[access-control-allow-origin,access-control-allow-credentials,access-control-allow-methods,access-control-allow-headers]
https://account.800best.com/uc/login?service=https://s9.800best.com/ssoCallback&appid=uc16cb111780834e75ba1cfbfab568d78e [302 Found] Access-Control-Allow-Methods[GET, PUT, POST, DELETE, PATCH, OPTIONS], Content-Language[en-US], Cookies[SESSION], Country[CHINA][CN], HttpOnly[SESSION], Domain[202.107.193.24], RedirectLocation[https://account.800best.com?redirectKey=5e999abf-1734-43e4-8489-d96586cfdcaf&mode=0&authType=1&appid=uc16cb111780834e75ba1cfbfab568d78e&sysCode=S9], UncommonHeaders[x-content-type-options,access-control-allow-origin,access-control-allow-credentials,access-control-allow-methods,access-control-allow-headers], X-Frame-Options[DENY], X-XSS-Protection[1; mode=block]
https://account.800best.com?redirectKey=5e999abf-1734-43e4-8489-d96586cfdcaf&mode=0&authType=1&appid=uc16cb111780834e75ba1cfbfab568d78e&sysCode=S9 [200 OK] Access-Control-Allow-Methods[GET, PUT, POST, DELETE, PATCH, OPTIONS], Country[CHINA][CN], HTML5, Domain[202.107.193.24], Script[text/javascript], Title[百世账号中心], UncommonHeaders[access-control-allow-origin,access-control-allow-credentials,access-control-allow-methods,access-control-allow-headers], X-Powered-By[Express]`
	statusRegx := regexp.MustCompile(`\[(\d{3})\s.*?\]`)
	m := statusRegx.FindAllStringSubmatch(s, -1)
	t.Log(len(m))
	t.Log(m)
	t.Log(m[len(m)-1][1])
}
