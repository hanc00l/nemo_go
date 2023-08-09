package onlineapi

import "testing"

func TestOnlineSearch_Do(t *testing.T) {
	domain := "shansteelgroup.com"
	hunter := NewOnlineAPISearch(OnlineAPIConfig{Target: domain}, "hunter")

	hunter.Do()
	for ip, ipr := range hunter.IpResult.IPResult {
		t.Log(ip, ipr)
		for port, pat := range ipr.Ports {
			t.Log(port, pat)
		}
	}
	for d, dar := range hunter.DomainResult.DomainResult {
		t.Log(d, dar)
	}
}

func TestOnlineSearch_Do2(t *testing.T) {
	domain := "shansteelgroup.com"
	hunter := NewOnlineAPISearch(OnlineAPIConfig{Target: domain}, "quake")

	hunter.Do()
	for ip, ipr := range hunter.IpResult.IPResult {
		t.Log(ip, ipr)
		for port, pat := range ipr.Ports {
			t.Log(port, pat)
		}
	}
	for d, dar := range hunter.DomainResult.DomainResult {
		t.Log(d, dar)
	}
}

func TestOnlineSearch_Do3(t *testing.T) {
	domain := "800best.com"

	hunter := NewOnlineAPISearch(OnlineAPIConfig{Target: domain}, "fofa")

	hunter.Do()
	for ip, ipr := range hunter.IpResult.IPResult {
		t.Log(ip, ipr)
		for port, pat := range ipr.Ports {
			t.Log(port, pat)
		}
	}
	for d, dar := range hunter.DomainResult.DomainResult {
		t.Log(d, dar)
	}
}
