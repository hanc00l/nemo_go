package fingerprint

import (
	"github.com/hanc00l/nemo_go/pkg/task/domainscan"
	"github.com/hanc00l/nemo_go/pkg/task/portscan"
	"testing"
)

func TestIconHash_RunFetchIconHashes(t *testing.T) {
	url := "www.baidu.com"
	iconHash := NewIconHash()
	result := iconHash.RunFetchIconHashes(url)
	for _, r := range result {
		t.Log(r)
	}
}

func TestIconHash_RunFetchIconHashes2(t *testing.T) {
	url := "172.16.222.1:8000"
	iconHash := NewIconHash()
	result := iconHash.RunFetchIconHashes(url)
	for _, r := range result {
		t.Log(r)
	}
}

func TestIconHash_Do(t *testing.T) {
	domainConfig := domainscan.Config{Target: "800best.com"}
	subdomain := domainscan.NewSubFinder(domainConfig)
	subdomain.Do()
	t.Log(&subdomain.Result)

	ih := NewIconHash()
	ih.ResultDomainScan = &subdomain.Result
	ih.Do()
	for d, da := range ih.ResultDomainScan.DomainResult {
		t.Log(d, da)
	}
	subdomain.Result.SaveResult(subdomain.Config)
}

func TestIconHash_Do2(t *testing.T) {
	nmapConfig := portscan.Config{
		Target: "127.0.0.1",
		Port:   "80,3306,1080,8000",
		Rate:   1000,
		IsPing: false,
		Tech:   "-sS",
		CmdBin: "nmap",
	}
	nmap := portscan.NewNmap(nmapConfig)
	nmap.Do()

	httpx := NewHttpxFinger()
	httpx.ResultPortScan = &nmap.Result
	httpx.Do()

	ih := NewIconHash()
	ih.ResultPortScan = httpx.ResultPortScan
	ih.OptimizationMode = true
	ih.Do()

	for ip, r := range ih.ResultPortScan.IPResult {
		t.Log(ip, r)
		for port, p := range r.Ports {
			t.Log(port, p)
		}
	}
}
