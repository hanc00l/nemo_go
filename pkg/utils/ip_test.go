package utils

import "testing"

func TestCheckIPV4(t *testing.T) {
	t.Log(CheckIPV4("192.168.1.1"))
	t.Log(CheckIPV4("192.168.1.1/24"))
	t.Log(CheckIPV4("10.0.0.0/8"))
	t.Log(CheckIPV4("0.0.0.0/0"))
}

func TestCheckIPV4Subnet(t *testing.T) {
	t.Log(CheckIPV4Subnet("192.168.1.1"))
	t.Log(CheckIPV4Subnet("192.168.1.1/24"))
	t.Log(CheckIPV4Subnet("10.0.0.0/8"))
	t.Log(CheckIPV4Subnet("0.0.0.0/0"))

}

func TestGetOutBoundIP(t *testing.T) {
	ip,err := GetOutBoundIP()
	t.Log(ip)
	t.Log(err)
}

func TestGetClientIp(t *testing.T) {
	ip,err := GetClientIp()
	t.Log(ip)
	t.Log(err)
}

func TestParseIP(t *testing.T) {
	t.Log(ParseIP("192.168.1.1"))
	t.Log(ParseIP("192.168.1.1/30"))
	t.Log(ParseIP("192.168.1.3-192.168.1.10"))
}