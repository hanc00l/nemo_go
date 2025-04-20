package custom

import (
	"testing"
)

func TestIpLocation_FindPublicIP(t *testing.T) {
	ipl := NewIPv4Location("")
	datas := []string{"222.180.198.142", "47.98.65.138", "10.1.1.1", "175.178.37.3", "124.71.62.214", "106.75.17.76",
		"20.189.67.202", "52.205.21.252", "13.64.112.131", "35.229.255.130", "172.66.41.25", "104.238.132.205", "124.95.191.7",
	}

	for _, d := range datas {
		t.Log(d, "->", ipl.FindPublicIP(d))
	}
}

func TestIpLocation_FindCustomIP(t *testing.T) {
	ipl := NewIPv4Location("test")

	ip := "172.16.8.13"
	result := ipl.FindCustomIP(ip)
	t.Log(result)

	ip = "10.1.0.3"
	result = ipl.FindCustomIP(ip)
	t.Log(result)

	ip = "192.168.120.220"
	result = ipl.FindCustomIP(ip)
	t.Log(result)
}
