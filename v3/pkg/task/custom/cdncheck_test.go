package custom

import "testing"

func TestCDNCheck_CheckIP(t *testing.T) {
	ips :=[]string{"114.114.114.114","125.39.46.3","14.215.177.39","173.245.48.12","119.84.174.52","116.55.250.136"}
	c := NewCDNCheck()
	for _,ip := range ips{
		t.Log(ip,c.CheckIP(ip))
	}
}

func TestCDNCheck_CheckASN(t *testing.T) {
	ips :=[]string{"114.114.114.114","125.39.46.3","14.215.177.39","173.245.48.12","119.84.174.52"}
	c := NewCDNCheck()
	for _,ip := range ips{
		t.Log(ip,c.CheckASN(ip))
	}
}

func TestCDNCheck_CheckCName(t *testing.T) {
	domains :=[]string{"www.mafengwo.cn","www.cdn.net","www.amazonaws.com","liveplay.mafengwo.cn"}
	c := NewCDNCheck()
	for _,domain := range domains{
		isCDN,CDNName,CName := c.CheckCName(domain)
		t.Log(isCDN,CDNName,CName)
	}
}

func TestCDNCheck_CheckCName2(t *testing.T) {
	domains :=[]string{"www.baidu.com"}
	c := NewCDNCheck()
	for _,domain := range domains{
		isCDN,CDNName,CName := c.CheckCName(domain)
		t.Log(isCDN,CDNName,CName)
	}
}
