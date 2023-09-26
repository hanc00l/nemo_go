package utils

import "testing"

func TestHostStrip(t *testing.T) {
	urls := []string{"http://www.sina.com.cn/", "china.gov.cn", "https://api.baidu.com:8080/user", "114.114.114.114", "3.4.5.6:8080", "[2400:dd01:103a:4041::101]:80"}
	for _, u := range urls {
		t.Log(u, "->", ParseHost(u))
	}
}

func TestCheckDomain(t *testing.T) {
	urls := []string{"testdomain", "http://www.sina.com.cn/", "china.gov.cn", "https://api.baidu.com:8080/user", "114.114.114.114", "x.y.info-1.hello.art", "../../etc/passwd"}
	for _, u := range urls {
		t.Log(u, CheckDomain(u))
	}
}

func TestHostStripV6(t *testing.T) {
	urls := []string{"[2400:dd01:103a:4041::101]"}
	for _, u := range urls {
		t.Log(ParseHost(u))
	}
}

func TestParseIP2(t *testing.T) {
	urls := []string{"[2400:dd01:103a:4041::101]", "[2400:dd01:103a:4041::101]:8080", "2400:dd01:103a:4041::101", "192.168.1.1", "192.168.2.3:3306", "[192.168.3.1]:8080"}
	for _, u := range urls {
		isIpv6, ip, port := ParseHostUrl(u)
		t.Log(u, "->", isIpv6, ip, port)
	}
}
