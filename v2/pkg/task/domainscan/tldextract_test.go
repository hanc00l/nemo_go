package domainscan

import (
	"strings"
	"testing"
)

func TestTldExtract_ExtractFLD(t *testing.T) {
	tld := NewTldExtract()
	urls := []string{"www.10086.cn", "test.api.cc.org.cn", "shop.sh.jd.com.cn"}
	for _, u := range urls {
		t.Log(tld.ExtractFLD(u))
	}
}

func TestTldExtract_ExtractFLD2(t *testing.T) {
	var tldDomain []string
	target := "www.10086.cn,test.api.cc.org.cn,shop.sh.jd.com.cn,a.b.c.10086.cn"
	domains := make(map[string]struct{})
	tld := NewTldExtract()
	for _, t := range strings.Split(target, ",") {
		domain := strings.TrimSpace(t)
		fld := tld.ExtractFLD(domain)
		if _, ok := domains[fld]; !ok {
			domains[fld] = struct{}{}
		}
	}
	for k, _ := range domains {
		tldDomain = append(tldDomain, k)
	}
	t.Log(tldDomain)
}
