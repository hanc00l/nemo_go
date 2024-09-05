package custom

import "testing"

func TestNewIpv6Location(t *testing.T) {
	datas := []string{"2409:8929:52b:36d9:8f6e:2e8b:a35:1148",
		"2405:6f00:c602::1",
		"2409:8c1e:75b0:1120::27",
		"2402:3c00:1000:4::1",
		"2408:8652:200::c101",
		"2409:8900:103f:14f:d7e:cd36:11af:be83",
		"fe80::5c12:27dc:93a4:3426",
	}
	ipv6, _ := NewIPv6Location()
	for _, d := range datas {
		r := ipv6.Find(d)
		t.Log(d, "->", r)
	}
}
