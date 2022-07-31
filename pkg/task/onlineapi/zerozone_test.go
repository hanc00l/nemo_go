package onlineapi

import (
	"testing"
)

func TestZeroZone_ParseCSVContentResult(t *testing.T) {
	data := `序号,url,端口,域名,IP,Title,状态码,端口组件(web容器),CMS,运营商,所属地区,标签,所属公司
1,https://vpn.sz.csg.cn:8443,8443,,58.249.28.33,SSL Error pages,200,unknown,,联通,中国广东广州,[],中国南方电网有限责任公司
2,http://cgysd.dlzb.com,80,cgysd.dlzb.com,118.212.233.107,中国南方电网有限责任公司超高压输电公司-首页,200,,,联通,中国江西南昌,[],中电中招（北京）招标有限公司
3,https://mail.pgc.csg.cn,443,mail.pgc.csg.cn,218.19.148.195,CSG企业邮件,200,nginx,,电信,中国广东广州,['登录页'],中国南方电网有限责任公司
4,http://ps.csg.cn:8000,8000,,112.94.221.2,首页,200,nginx,,联通,中国广东广州,[],中国南方电网有限责任公司
5,http://xxqa.gmp.csg.cn:8000,8000,,112.94.221.2,首页,200,nginx,,联通,中国广东广州,[],中国南方电网有限责任公司
6,https://ailabel.csg.cn,443,ailabel.csg.cn,183.63.250.104,抱歉，出错了 - 标注工具客户端 - 人工智能平台 - 中国南方电网有限责任公司,200,nginx,,电信,中国广东广州,[],中国南方电网有限责任公司
7,http://vpn.yn.csg.cn:443,443,,58.249.28.225,200 OK,200,unknown,,联通,中国广东广州,[],中国南方电网有限责任公司
8,https://app.csg.cn:8801,8801,,121.32.27.111,eLink,200,nginx,,电信,中国广东广州,[],中国南方电网有限责任公司
9,http://oss.rhib.csg.cn,80,oss.rhib.csg.cn,121.32.27.121,Document,200,unknown,,电信,中国广东广州,[],中国南方电网有限责任公司
10,https://help.sz.csg.cn,443,help.sz.csg.cn,58.251.73.11,下载页面-&#21319;&#32423;&#26381;&#21153;,200,,,联通,中国广东深圳,[],中国南方电网有限责任公司
11,http://app.guangzhou.csg.cn:443,443,,112.94.223.228,200 OK,200,unknown,,联通,中国广东广州,[],中国南方电网有限责任公司
12,http://vs.csg.cn,80,vs.csg.cn,121.32.27.123,二维码寻车,200,nginx,,电信,中国广东广州,[],中国南方电网有限责任公司
13,https://woa.hn.csg.cn,443,woa.hn.csg.cn,120.234.181.163,海南电网有限责任公司,200,Apache httpd,,移动,中国广东汕头,[],中国南方电网有限责任公司
14,http://vpn8.yn.csg.cn:443,443,,58.249.28.226,200 OK,200,unknown,,联通,中国广东广州,[],中国南方电网有限责任公司
15,http://vpn7.yn.csg.cn:443,443,,120.234.181.226,200 OK,200,unknown,,移动,中国广东汕头,[],中国南方电网有限责任公司
16,http://spa.rhib.csg.cn,80,spa.rhib.csg.cn,121.32.27.121,瑞恒微信应用服务,200,unknown,,电信,中国广东广州,[],中国南方电网有限责任公司
17,http://opr.rhib.csg.cn,80,opr.rhib.csg.cn,121.32.27.121,瑞恒保险经纪公司互联网销售平台业务管理系统,200,unknown,,电信,中国广东广州,[],中国南方电网有限责任公司
18,https://vpn.yn.csg.cn,443,vpn.yn.csg.cn,120.234.181.225,EasyConnect,200,unknown,,移动,中国广东汕头,[],中国南方电网有限责任公司
19,http://mail.guangzhou.csg.cn,80,mail.guangzhou.csg.cn,218.19.148.195,CSG企业邮件,200,nginx,,电信,中国广东广州,[],中国南方电网有限责任公司
20,https://vpn.csg.cn,443,vpn.csg.cn,120.236.186.171,登陆,200,unknown,,移动,中国广东广州,[],中国南方电网有限责任公司
21,https://vpn.pgc.csg.cn:8888,8888,,58.249.28.65,Download EasyConnect,200,unknown,,联通,中国广东广州,[],中国南方电网有限责任公司
22,http://ancuser.mos.csg.cn,80,ancuser.mos.csg.cn,121.32.27.102,南方区域电力现货技术支持系统,200,unknown,,电信,中国广东广州,[],中国南方电网有限责任公司
23,https://ancuser.mos.csg.cn,443,ancuser.mos.csg.cn,121.32.27.102,南方区域电力现货技术支持系统,200,nginx,,电信,中国广东广州,[],中国南方电网有限责任公司
24,http://mail.pgc.csg.cn,80,mail.pgc.csg.cn,218.19.148.195,CSG企业邮件,200,nginx,,电信,中国广东广州,['登录页'],中国南方电网有限责任公司
25,https://mail.guangzhou.csg.cn,443,mail.guangzhou.csg.cn,218.19.148.195,CSG企业邮件,200,nginx,,电信,中国广东广州,['登录页'],中国南方电网有限责任公司
26,https://www.bidding.csg.cn:9090,9090,,120.236.186.149,首页,200,nginx,,移动,中国广东广州,[],中国南方电网有限责任公司
27,http://zhaopin.csg.cn:8080,8080,,120.236.186.144,中国南方电网员工招聘系统,200,unknown,,移动,中国广东广州,[],中国南方电网有限责任公司
28,http://app.csg.cn:8087,8087,,112.94.221.4,安运在线,200,unknown,,联通,中国广东广州,[],中国南方电网有限责任公司
29,https://ywjk.csg.cn:20443,20443,,218.19.148.202,Welcome to tengine!,200,nginx,,电信,中国广东广州,[],中国南方电网有限责任公司
30,http://leadership.csg.cn:443,443,,120.236.186.155,Welcome to nginx!,200,unknown,,移动,中国广东广州,[],中国南方电网有限责任公司
31,https://app.ehv.csg.cn,443,app.ehv.csg.cn,120.234.181.10,Download EasyConnect,200,unknown,,移动,中国广东汕头,[],中国南方电网有限责任公司
32,http://iat.gd.csg.cn:8989,8989,,120.234.183.50,,200,nginx,,移动,中国广东汕头,[],中国南方电网有限责任公司
33,https://cgysd.dlzb.com:444,444,,117.23.61.254,中国南方电网有限责任公司超高压输电公司-首页,200,unknown,,电信,中国陕西西安,[],中电中招（北京）招标有限公司
34,https://cgysd.dlzb.com:7080,7080,,117.23.61.254,中国南方电网有限责任公司超高压输电公司-首页,200,unknown,,电信,中国陕西西安,[],中电中招（北京）招标有限公司
35,http://cgysd.dlzb.com:7000,7000,,117.23.61.254,中国南方电网有限责任公司超高压输电公司-首页,200,unknown,,电信,中国陕西西安,[],中电中招（北京）招标有限公司
36,https://cgysd.dlzb.com:8443,8443,,117.23.61.254,中国南方电网有限责任公司超高压输电公司-首页,200,unknown,,电信,中国陕西西安,[],中电中招（北京）招标有限公司
37,https://cgysd.dlzb.com:8010,8010,,117.23.61.254,中国南方电网有限责任公司超高压输电公司-首页,200,unknown,,电信,中国陕西西安,[],中电中招（北京）招标有限公司
38,http://cgysd.dlzb.com:8003,8003,,117.23.61.254,中国南方电网有限责任公司超高压输电公司-首页,200,,,电信,中国陕西西安,[],中电中招（北京）招标有限公司
39,https://cgysd.dlzb.com:8081,8081,,117.23.61.254,中国南方电网有限责任公司超高压输电公司-首页,200,unknown,,电信,中国陕西西安,[],中电中招（北京）招标有限公司
40,https://cgysd.dlzb.com:7443,7443,,117.23.61.254,中国南方电网有限责任公司超高压输电公司-首页,200,,,电信,中国陕西西安,[],中电中招（北京）招标有限公司
41,https://cgysd.dlzb.com:7001,7001,,117.23.61.254,中国南方电网有限责任公司超高压输电公司-首页,200,,,电信,中国陕西西安,[],中电中招（北京）招标有限公司
42,http://cgysd.dlzb.com:82,82,,117.23.61.254,中国南方电网有限责任公司超高压输电公司-首页,200,,,电信,中国陕西西安,[],中电中招（北京）招标有限公司
43,https://dljy.gd.csg.cn,443,dljy.gd.csg.cn,120.234.183.4,广东电力交易中心-首页,200,,,移动,中国广东汕头,[],中国南方电网有限责任公司
44,http://mx.mail.csg.cn,80,mx.mail.csg.cn,120.236.148.57,CSG企业邮件,200,nginx,,移动,中国广东广州,['登录页'],中国南方电网有限责任公司
45,http://mail.gz.csg.cn,80,mail.gz.csg.cn,120.236.148.57,CSG企业邮件,200,nginx,,移动,中国广东广州,['登录页'],中国南方电网有限责任公司
46,http://m.sz.csg.cn,80,m.sz.csg.cn,120.236.148.57,CSG企业邮件,200,nginx,,移动,中国广东广州,['登录页'],中国南方电网有限责任公司
47,http://mail.ehv.csg.cn,80,mail.ehv.csg.cn,120.236.148.57,CSG企业邮件,200,nginx,,移动,中国广东广州,['登录页'],中国南方电网有限责任公司
48,https://mail.gx.csg.cn,443,mail.gx.csg.cn,120.236.148.57,CSG企业邮件,200,nginx,,移动,中国广东广州,['登录页'],中国南方电网有限责任公司
49,https://oss.rhib.csg.cn,443,oss.rhib.csg.cn,120.236.186.158,Document,200,nginx,,移动,中国广东广州,[],中国南方电网有限责任公司
50,http://pay.csg.cn,80,pay.csg.cn,120.236.148.84,,200,,,移动,中国广东广州,[],中国南方电网有限责任公司
817,https://120.236.186.173,443,120.236.186.173,120.236.186.173,南方电网商旅通,200,nginx,unknown,移动,中国广东广州,['中国南方电网有限责任公司'],中国南方电网有限责任公司
818,https://120.236.186.158,443,120.236.186.158,120.236.186.158,瑞恒保险经纪有限责任公司,200,nginx,unknown,移动,中国广东广州,[],中国南方电网有限责任公司
819,https://120.236.186.139,443,120.236.186.139,120.236.186.139,,200,nginx,unknown,移动,中国广东广州,[],中国南方电网有限责任公司
820,https://120.236.186.143,443,120.236.186.143,120.236.186.143,,200,nginx,unknown,移动,中国广东广州,[],中国南方电网有限责任公司
821,https://120.236.186.162,443,120.236.186.162,120.236.186.162,中国南方电网,200,nginx,unknown,移动,中国广东广州,[],中国南方电网有限责任公司`
	/*
	   1 [1 https://vpn.sz.csg.cn:8443 8443  58.249.28.33 SSL Error pages 200 unknown  联通 中国广东广州 [] 中国南方电网有限责任公司]
	   2 [2 http://cgysd.dlzb.com 80 cgysd.dlzb.com 118.212.233.107 中国南方电网有限责任公司超高压输电公司-首页 200   联通 中国江西南昌 [] 中电中招（北京）招标有限公司]
	   3 [3 https://mail.pgc.csg.cn 443 mail.pgc.csg.cn 218.19.148.195 CSG企业邮件 200 nginx  电信 中国广东广州 ['登录页'] 中国南方电网有限责任公司]
	   4 [4 http://ps.csg.cn:8000 8000  112.94.221.2 首页 200 nginx  联通 中国广东广州 [] 中国南方电网有限责任公司]
	   5 [5 http://xxqa.gmp.csg.cn:8000 8000  112.94.221.2 首页 200 nginx  联通 中国广东广州 [] 中国南方电网有限责任公司]
	*/
	z := NewZeroZone(OnlineAPIConfig{})
	z.ParseCSVContentResult([]byte(data))
	for kk, ip := range z.IpResult.IPResult {
		t.Log(kk)
		for kk, port := range ip.Ports {
			t.Log(kk, port.Status)
			t.Log(port.PortAttrs)
		}
	}
	for kk, d := range z.DomainResult.DomainResult {
		t.Log(kk)
		for kk, da := range d.DomainAttrs {
			t.Log(kk, da)
		}
	}
}
