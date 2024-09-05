package onlineapi

import "testing"

func TestHunter_ParseCSVContentResult(t *testing.T) {
	data := ``
	s := NewOnlineAPISearch(OnlineAPIConfig{}, "hunter")
	s.ParseContentResult([]byte(data))
	for kk, ip := range s.IpResult.IPResult {
		t.Log(kk)
		for kk, port := range ip.Ports {
			t.Log(kk, port.Status)
			t.Log(port.PortAttrs)
		}
	}
	for kk, d := range s.DomainResult.DomainResult {
		t.Log(kk)
		for kk, da := range d.DomainAttrs {
			t.Log(kk, da)
		}
	}
}
