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
	//config2 := OnlineAPIConfig{Target: "800best.com"}
	config2 := OnlineAPIConfig{Target: "shansteelgroup.com"}
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

func TestFofa_Run3(t *testing.T) {
	config3 := OnlineAPIConfig{}
	config3.Target = `title="国网" && body!="帝国|中国|外国|三国网页|全国网络|国网站|网约车|全国网安|战国网络|国网络安全|情感故事|全国网上|全国网红|苏州国网电子|二手车|全国网约房|伊顿EIP|全国网益通系统|一个充满思想的平台"&& country="CN" && region!="HK" && region!="TW"  && region!="MO"&& body !="网址大全"&& body !="美食餐廳"&& body !="网上开户"&& body !="卡密充值"&& body !="财经资讯"&& body !="中工招商网"&& body !="武义县公安局林警智治应用平台"&& body !="哈尔滨隆腾尚云"&& body !="行业信息网"&& body !="公安采购商城"&& body !="制造有限公司"&& body !="蛋糕店配送"&& body !="招标资源公共平台"&& body !="网站制作"&& body !="房产"&& body !="广告"&& body !="销售热线"&& body !="博物馆"&& body !="博客"&& body !="销量"&& body !="护肤"&& body !="产品中心"&& body !="产品展示"&& body !="电话咨询"&& body !="阻燃软包"&& body !="刻章"&& body !="服装厂家"&& body !="品牌展示"&& body !="案例欣赏"&& body !="案例展示"&& body !="企业文化"&& body !="我们的优势"&& body !="全国统一服务热线"&& body !="解决方案"&& body !="诚聘英才"&& body !="招募热线"&& body !="全国服务热线"&& body !="商务合作"&& body !="关注我们"&& body !="关于我们"&& body !="开锁服务"&& body !="云领信息科技有限公司"&& body !="公务员学习网"&& body !="产品展示"&& body !="产品与服务"&& body !="专业承揽监控工程"&& body !="工艺品有限公司"&& body !="紧急开锁" && after="2022-10-10"`
	config3.SearchLimitCount = 100
	config3.SearchByKeyWord = true
	fofa3 := NewFofa(config3)
	fofa3.Do()
	//t.Log(fofa2.SaveResult())
	t.Log(len(fofa3.Result))
	t.Log(fofa3.Result)
	t.Log(fofa3.IpResult)

	for ip, ipr := range fofa3.IpResult.IPResult {
		t.Log(ip, ipr)
		for port, pat := range ipr.Ports {
			t.Log(port, pat)
		}
	}
	t.Log(fofa3.DomainResult)
	for domain, dar := range fofa3.DomainResult.DomainResult {
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
