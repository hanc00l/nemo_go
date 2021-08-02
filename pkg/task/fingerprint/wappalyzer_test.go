package fingerprint

import (
	"github.com/hanc00l/nemo_go/pkg/task/domainscan"
	"testing"
)

func TestWappalyzer_RunWappalyzer(t *testing.T) {
	w := NewWappalyzer()
	domainResult := domainscan.Result{
		DomainResult: make(map[string]*domainscan.DomainResult),
	}
	domainResult.SetDomain("www.freebuf.com")
	domainResult.SetDomain("www.baidu.com")
	domainResult.SetDomain("www.jianshu.com")
	domainResult.SetDomain("www.800best.com")
	w.ResultDomainScan = domainResult
	w.Do()
	for _,dr :=range domainResult.DomainResult{
		t.Log(dr.DomainAttrs)
	}
}
