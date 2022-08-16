package onlineapi

import "testing"

func TestFofa_Run(t *testing.T) {
	config1 := OnlineAPIConfig{Target: "47.98.181.116"}
	fofa1 := NewFofa(config1)
	fofa1.Do()
	//fofa1.SaveResult()
	for ip, ipr := range fofa1.IpResult.IPResult {
		t.Log(ip, ipr)
		for port, pat := range ipr.Ports {
			t.Log(port, pat)
		}
	}
	for domain, dar := range fofa1.DomainResult.DomainResult {
		t.Log(domain, dar)
		//for _, a := range dar.DomainAttrs {
		//	t.Log(a)
		//}
	}
}
func TestFofa_Run2(t *testing.T) {
	config2 := OnlineAPIConfig{Target: "800best.com"}
	//config2 := OnlineAPIConfig{Target: "10086.cn"}
	fofa2 := NewFofa(config2)
	fofa2.Do()
	//t.Log(fofa2.SaveResult())
	t.Log(fofa2.Result)
	t.Log(fofa2.IpResult)

	for ip, ipr := range fofa2.IpResult.IPResult {
		t.Log(ip, ipr)
		for port, pat := range ipr.Ports {
			t.Log(port, pat)
		}
	}
	t.Log(fofa2.DomainResult)
	for domain, dar := range fofa2.DomainResult.DomainResult {
		t.Log(domain, dar)
		//for _, a := range dar.DomainAttrs {
		//	t.Log(a)
		//}
	}
}

func TestFofa_ParseCSVContentResult(t *testing.T) {
	data := `host,lastupdatetime,ip,port,title,domain,protocol,country,city,as_organization
yk.ellingtonstopclasscleaning.com,2022-08-02 13:00:00,154.91.183.30,80,綦江县倩痉芍电力科技有限公司,ellingtonstopclasscleaning.com,http,HK,,PEGTECHINC
banniangdun.studiomantegazza.com,2022-08-02 12:00:00,154.91.185.28,80,綦江县栏土浪电力科技有限公司,studiomantegazza.com,http,HK,,PEGTECHINC
qijiang.spbxg.com,2022-08-01 12:00:00,103.148.150.234,80,"綦江(电力,油浸式)变压器价格/厂家/6300KVA/8000KVA/10000KVA/S11/S13/SZ11/35KV  -华恒变压器有限公司",spbxg.com,http,HK,,Cloud Iv Limited
qijiang.mpppipes.com,2022-07-02 02:00:00,154.213.186.190,80,綦江MPP电力管_綦江MPP管_綦江MPP顶管_綦江MPP电缆保护管 - 綦江山东润星电力管材有限公司,mpppipes.com,http,HK,,rainbow network limited
qijiang.mpplg.com,2022-07-02 01:00:00,154.213.181.183,80,"綦江MPP电力管,MPP拖拉管,MPP拉管,MPP拖拉管厂家,綦江MPP顶管,MPP非开挖拉电力管,山东润星电力管材有限公司",mpplg.com,http,HK,,rainbow network limited
749f31.liang6131.com,2022-06-13 12:00:00,172.245.43.109,80,重庆市重庆周边綦江县测坠电力有限公司,liang6131.com,http,US,,AS-COLOCROSSING
`
	f := NewFofa(OnlineAPIConfig{})
	f.ParseCSVContentResult([]byte(data))
	for kk, ip := range f.IpResult.IPResult {
		t.Log(kk)
		for kk, port := range ip.Ports {
			t.Log(kk, port.Status)
			t.Log(port.PortAttrs)
		}
	}
	for kk, d := range f.DomainResult.DomainResult {
		t.Log(kk)
		for kk, da := range d.DomainAttrs {
			t.Log(kk, da)
		}
	}
}
