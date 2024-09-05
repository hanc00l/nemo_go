package test

import (
	"github.com/hanc00l/nemo_go/v2/pkg/task/domainscan"
	"testing"
)

var rSubfinder = map[string]map[string]bool{
	"v.news.cn": {
		"player.v.news.cn": false,
		"review.v.news.cn": false,
		"live.v.news.cn":   false,
		"show.v.news.cn":   false,
	},
}

func init() {
	/*
		subfinder:
		v.news.cn
		player.v.news.cn
		review.v.news.cn
		live.v.news.cn
		source10hls.v.news.cn
		show.v.news.cn
		source07.v.news.cn
		vodpub1.v.news.cn
		source08.v.news.cn
		oldvod2.v.news.cn
		vodpub2.v.news.cn

		massdns：
		live.v.news.cn
		cp.v.news.cn
		player.v.news.cn
		stat.v.news.cn
		show.v.news.cn
		see.v.news.cn
		review.v.news.cn
		monitor1.v.news.cn
		yunbo.v.news.cn
		interact.v.news.cn
		cloud.v.news.cn
	*/
}
func TestMassdns_test(t *testing.T) {
	//t.Log("skip test massdns...")
	//return

	config := domainscan.Config{Target: "v.news.cn"}
	m := domainscan.NewMassdns(config)
	m.Do()
	resultCount := len(m.Result.DomainResult)
	t.Logf("massdns find subdomain total:%d", resultCount)

	if resultCount < 5 {
		t.Error("find subdomain too title...")
	}
}

func TestSubfinder(t *testing.T) {
	config := domainscan.Config{
		Target: "v.news.cn",
	}
	subdomain := domainscan.NewSubFinder(config)
	subdomain.Do()

	for domain := range subdomain.Result.DomainResult {
		if _, ok := rSubfinder["v.news.cn"][domain]; ok {
			rSubfinder["v.news.cn"][domain] = true
		}
	}

	// check result
	if len(subdomain.Result.DomainResult) < 5 {
		t.Error("subfinder find subdomain too title...")
	}
	for domain, status := range rSubfinder["v.news.cn"] {
		if !status {
			t.Errorf("subfinder not find:%s", domain)
		}
	}
}
