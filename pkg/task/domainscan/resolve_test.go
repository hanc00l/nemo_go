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
