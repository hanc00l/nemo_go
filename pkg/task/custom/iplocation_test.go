package custom

import (
	"testing"
)

func TestIpLocation_FindPublicIP(t *testing.T) {
	ipl := NewIPLocation()

	ip := "222.180.198.142"
	result := ipl.FindPublicIP(ip)
	t.Log(result)

	ip = "47.98.65.38"
	result = ipl.FindPublicIP(ip)
	t.Log(result)

	ip = "10.1.1.1"
	result = ipl.FindPublicIP(ip)
	t.Log(result)
}

func TestLoadCustomIP(t *testing.T) {
	//t.Log(customMap)
	//t.Log(customBMap)
	//t.Log(customCMap)
}

func TestIpLocation_FindCustomIP(t *testing.T) {
	ipl := NewIPLocation()

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
