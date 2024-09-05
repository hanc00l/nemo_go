package domainscan

import "testing"

func TestResolveDomain(t *testing.T) {
	config := Config{
		Target: "www.sina.com.cn,www.163.com,www.china.gov.cn",
	}
	r := NewResolve(config)
	r.Do()
	for k, v := range r.Result.DomainResult {
		t.Log(k, v)
	}
}

func TestResolveDomainIPV6(t *testing.T) {
	datas := []string{"www.sina.com.cn", "www.163.com", "www.sgcc.com.cn"}
	for _, d := range datas {
		cname, host := ResolveDomain(d)
		t.Log(d, "->", cname)
		for _, h := range host {
			t.Log(h)
		}
	}
}
