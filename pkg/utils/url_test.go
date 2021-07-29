package utils

import "testing"

func TestHostStrip(t *testing.T) {
	urls := []string{"http://www.sina.com.cn/","china.gov.cn","https://api.baidu.com:8080/user","114.114.114.114"}
	for _,u := range urls{
		t.Log(HostStrip(u))
	}
}

func TestCheckDomain(t *testing.T) {
	urls := []string{"http://www.sina.com.cn/","china.gov.cn","https://api.baidu.com:8080/user","114.114.114.114","x.y.info-1.hello.art","../../etc/passwd"}
	for _,u := range urls{
		t.Log(u,CheckDomain(u))
	}
}
