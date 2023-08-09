package onlineapi

import "testing"

func TestHunter_ParseCSVContentResult(t *testing.T) {
	data := `url,资产标签,IP,IP标签,端口,网站标题,域名,高危协议,协议,通讯协议,网站状态码,应用/组件,操作系统,备案单位,备案号,国家,省份,市区,探查时间,Web资产,运营商,注册机构
http://video.qlid.cn:9900,"远程会议系统,登录页面",58.17.157.11,,9900,网动统一通信平台(Active UC),video.qlid.cn,,http,tcp,200,"jQuery,网动统一通信平台",,庆铃汽车股份有限公司,渝ICP备10202574号-8,中国,重庆市,重庆市,2022-08-14,是,中国联通,中国联通
http://video.qlid.cn:11011,,58.17.157.11,,11011,,video.qlid.cn,,http,tcp,200,,,庆铃汽车股份有限公司,渝ICP备10202574号-8,中国,重庆市,重庆市,2022-08-14,是,中国联通,中国联通
http://www.qlid.net:7547,,58.17.157.10,,7547,,www.qlid.net,,http,tcp,200,,,庆铃汽车股份有限公司,渝ICP备10202574号-7,中国,重庆市,重庆市,2022-08-12,是,中国联通,中国联通
http://qlid.net:7547,,58.17.157.10,,7547,,qlid.net,,http,tcp,200,,,庆铃汽车股份有限公司,渝ICP备10202574号-7,中国,重庆市,重庆市,2022-08-12,是,中国联通,中国联通
http://rg09.qlid.cn,,183.67.39.60,,80,重庆精耕带路MES系统,rg09.qlid.cn,,http,tcp,200,"Bootstrap,Blazor,Microsoft ASP.NET,Nginx/1.17.5",,庆铃汽车股份有限公司,渝ICP备10202574号-8,中国,重庆市,重庆市,2022-08-12,是,中国电信,中国电信
https://www.qlid.net:6666,"VPN,下载页面",58.17.157.10,,6666,欢迎访问SSLVPN,www.qlid.net,,https,tcp,200,SANGFOR 深信服 SSL VPN,,庆铃汽车股份有限公司,渝ICP备10202574号-7,中国,重庆市,重庆市,2022-08-12,是,中国联通,中国联通
http://fr.qlid.cn,"信息页面,Default Page",183.67.39.60,,80,Apache Tomcat/8.5.57,fr.qlid.cn,,http,tcp,200,"Nginx/1.17.5,Tomcat Default Page",,庆铃汽车股份有限公司,渝ICP备10202574号-8,中国,重庆市,重庆市,2022-08-12,是,中国电信,中国电信
https://qlid.net:6666,"VPN,下载页面",58.17.157.10,,6666,欢迎访问SSLVPN,qlid.net,,https,tcp,200,SANGFOR 深信服 SSL VPN,,庆铃汽车股份有限公司,渝ICP备10202574号-7,中国,重庆市,重庆市,2022-08-12,是,中国联通,中国联通
http://hgz.ivi.cn,,183.67.39.60,,80,庆铃车辆管理系统,hgz.ivi.cn,,http,tcp,200,"jQuery/1.11.0,Microsoft ASP.NET,Nginx/1.17.5",,庆铃汽车股份有限公司,渝ICP备10202574号-9,中国,重庆市,重庆市,2022-08-12,是,中国电信,中国电信
http://office.qlid.cn,,183.67.39.60,,80,,office.qlid.cn,,http,tcp,200,"Microsoft ASP.NET,Nginx/1.17.5",,庆铃汽车股份有限公司,渝ICP备10202574号-8,中国,重庆市,重庆市,2022-08-12,是,中国电信,中国电信
http://srm.qlid.cn,,183.67.39.60,,80,云采,srm.qlid.cn,,http,tcp,200,"Font Awesome,Nginx/1.17.5,Microsoft ASP.NET,SweetAlert,Bootstrap,jQuery Migrate,jQuery/1.11.3",,庆铃汽车股份有限公司,渝ICP备10202574号-8,中国,重庆市,重庆市,2022-08-12,是,中国电信,中国电信
http://fpapi.qlid.cn,,183.67.39.60,,80,,fpapi.qlid.cn,,http,tcp,200,Nginx/1.17.5,,庆铃汽车股份有限公司,渝ICP备10202574号-8,中国,重庆市,重庆市,2022-08-12,是,中国电信,中国电信
http://qltmqg.qlcv.cn,Default Page,183.67.39.60,,80,Welcome to nginx!,qltmqg.qlcv.cn,,http,tcp,200,"Nginx/1.17.5,Nginx Default Page",,庆铃汽车股份有限公司,渝ICP备10202574号-12,中国,重庆市,重庆市,2022-08-12,是,中国电信,中国电信
http://qlid.cn,,183.67.39.60,,80,庆铃车辆管理系统,qlid.cn,,http,tcp,200,"Microsoft ASP.NET,jQuery/1.11.0,Nginx/1.17.5",,庆铃汽车股份有限公司,渝ICP备10202574号-8,中国,重庆市,重庆市,2022-08-12,是,中国电信,中国电信
http://soft.qlid.cn,"网络附加存储设备,登录页面",183.67.39.60,,80,QGB-NAS - Synology NAS,soft.qlid.cn,,http,tcp,200,"Prototype/1.7.2,ExtJS,Nginx/1.17.5,Synology 群晖 DiskStation Manager",,庆铃汽车股份有限公司,渝ICP备10202574号-8,中国,重庆市,重庆市,2022-08-12,是,中国电信,中国电信
http://dmsqg.qlcv.cn,,183.67.39.60,,80,庆铃汽车股份有限公司DMS,dmsqg.qlcv.cn,,http,tcp,200,"Java Servlet/3.0,jQuery/1.7.2,Nginx/1.17.5,Java",,庆铃汽车股份有限公司,渝ICP备10202574号-12,中国,重庆市,重庆市,2022-08-12,是,中国电信,中国电信
http://yun.qlid.cn,,183.67.39.60,,80,AnyShare Enterprise,yun.qlid.cn,,http,tcp,200,Nginx/1.17.5,,庆铃汽车股份有限公司,渝ICP备10202574号-8,中国,重庆市,重庆市,2022-08-12,是,中国电信,中国电信
https://office.qlid.cn,,183.67.39.60,,443,,office.qlid.cn,,https,tcp,200,"Nginx/1.17.5,Microsoft ASP.NET",,庆铃汽车股份有限公司,渝ICP备10202574号-8,中国,重庆市,重庆市,2022-08-12,是,中国电信,中国电信
https://www.qlid.cn,,183.67.39.60,,443,,www.qlid.cn,,https,tcp,200,Nginx/1.17.5,,庆铃汽车股份有限公司,渝ICP备10202574号-8,中国,重庆市,重庆市,2022-08-12,是,中国电信,中国电信
https://dmsqg.qlcv.cn,,183.67.39.60,,443,,dmsqg.qlcv.cn,,https,tcp,200,Nginx/1.17.5,,庆铃汽车股份有限公司,渝ICP备10202574号-12,中国,重庆市,重庆市,2022-08-12,是,中国电信,中国电信
`
	s := NewOnlineAPISearch(OnlineAPIConfig{}, "hunter")
	s.ParseContentResult([]byte(data))
	for kk, ip := range s.IpResult.IPResult {
		t.Log(kk)
		for kk, port := range ip.Ports {
			t.Log(kk, port.Status)
			t.Log(port.PortAttrs)
		}
	}
	for kk, d := range s.DomainResult.DomainResult {
		t.Log(kk)
		for kk, da := range d.DomainAttrs {
			t.Log(kk, da)
		}
	}
}
