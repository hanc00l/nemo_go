package fingerprint

import (
	"github.com/hanc00l/nemo_go/pkg/task/domainscan"
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
