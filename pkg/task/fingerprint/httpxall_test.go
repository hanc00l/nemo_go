package fingerprint

import (
	"github.com/hanc00l/nemo_go/pkg/task/domainscan"
	"github.com/hanc00l/nemo_go/pkg/task/portscan"
	"testing"
)

func TestHttpxInvoke_DoHttpxAndFingerPrint1(t *testing.T) {
	v := NewHttpxAll()
	v.ResultPortScan = &portscan.Result{
		IPResult: make(map[string]*portscan.IPResult),
	}
	v.ResultPortScan.SetIP("172.16.222.1")
	v.ResultPortScan.SetPort("172.16.222.1", 8000)
	v.Do()

	for ip, r := range v.ResultPortScan.IPResult {
		t.Log(ip, r)
		for port, p := range r.Ports {
			t.Log(port, p)
		}
	}
}

func TestHttpxInvode_DoHttpxAndFingerPrint2(t *testing.T) {
	v := NewHttpxAll()
	v.ResultDomainScan = &domainscan.Result{
		DomainResult: make(map[string]*domainscan.DomainResult),
	}
	v.IsProxy = true
	v.ResultDomainScan.SetDomain("www.163.com")

	v.Do()

	for domain, r := range v.ResultDomainScan.DomainResult {
		t.Log(domain, r)
	}
}
