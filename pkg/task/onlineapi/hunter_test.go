package onlineapi

import "testing"

func TestHunter_RunHunter(t *testing.T) {
	domain := "47.98.181.116"
	hunter := NewHunter(OnlineAPIConfig{})

	hunter.RunHunter(domain)
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


func TestHunter_RunHunter2(t *testing.T) {
	domain := "shansteelgroup.com"
	hunter := NewHunter(OnlineAPIConfig{})

	hunter.RunHunter(domain)
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
