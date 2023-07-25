package onlineapi

import (
	"os"
	"testing"
)

func TestQuake_ParseQuakeSearchResult(t *testing.T) {
	results, _ := os.ReadFile("/Users/user/Downloads/5.json")
	q := Quake{}
	q.Result, _, _ = q.parseQuakeSearchResult(results)
	q.parseResult()
	for k, ip := range q.IpResult.IPResult {
		t.Log(k)
		for p, pa := range ip.Ports {
			t.Log(p, pa)
		}
	}
}

func TestQuake_ParseQuakeSearchResult2(t *testing.T) {
	results, _ := os.ReadFile("/Users/user/Downloads/4.json")
	q := Quake{}
	q.Result, _, _ = q.parseQuakeSearchResult(results)
	q.parseResult()
	for k, domain := range q.DomainResult.DomainResult {
		t.Log(k)
		for p, pa := range domain.DomainAttrs {
			t.Log(p, pa)
		}
	}
}

func TestQuake_RunQuake(t *testing.T) {
	q := Quake{}
	q.RunQuake("47.98.181.116")
	for _, r := range q.Result {
		t.Log(r)
	}
}

func TestQuake_RunQuake2(t *testing.T) {
	q := Quake{}
	q.RunQuake("800best.com")
	for _, r := range q.Result {
		t.Log(r)
	}
}

func TestQuake_Do(t *testing.T) {
	config := OnlineAPIConfig{}
	config.Target = "47.98.181.116"
	q := NewQuake(config)
	q.Do()
	q.SaveResult()
}

func TestQuake_Do2(t *testing.T) {
	domain := "shansteelgroup.com"
	quake := NewQuake(OnlineAPIConfig{})

	quake.RunQuake(domain)
	for ip, ipr := range quake.IpResult.IPResult {
		t.Log(ip, ipr)
		for port, pat := range ipr.Ports {
			t.Log(port, pat)
		}
	}
	for d, dar := range quake.DomainResult.DomainResult {
		t.Log(d, dar)
	}
}
