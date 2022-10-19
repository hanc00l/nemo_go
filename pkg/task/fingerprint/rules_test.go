package fingerprint

import "testing"

func TestRule(t *testing.T) {
	result := ParseRules("(app=\"JIRA\"||body=\"System Dashboard\")")
	println(MatchRules(*result, Content{Body: "asdasda正在加System DashboardUFIDA载 Jeecbody=sys/common/pdfg-Boot 快速开发sys/comlogo/images/mon/static平台,请耐心等待ssadasdasdsad"}))
	//content := Content{Header: "Server: SimpleHTTP/0.6 Python/3.9.14", Body: "<html>\n    <header>\n        <title>for test o:)</title>\n    </header>\n    <body>\n        <p><h1>\n            正在加System DashboardUFIDA载 Jeecbody=sys/common/pdfg-Boot 快速开发sys/comlogo/images/mon/static平台,请耐心等待\n        </h1></p>\n    </body>\n</html>"}
	//result2 := ParseRules("header=\"eHTTP\"||title=\"ProCurve Switch\"||(banner=\"HP \"&&banner=\"ProCurve Switch\")||banner=\"HP ProCurve 1810G\"||(banner=\"ProCurve\"&&banner=\"switch\"")
	//t.Log(MatchRules(*result2, content))
	//result3 := ParseRules("title=\"科来网络回溯\"||((body=\"科来软件 版权所有\"||body=\"i18ninit.min.js\")&&body!=\"nfr=\"true\"\"")
	//t.Log(MatchRules(*result3, content))
	//result4 := ParseRules("(body=\"UFIDA\"&&body=\"logo/images/\")||(body=\"logo/images/ufida_nc.png\")||title=\"Yonyou NC\"||body=\"<div id=\"nc_text\">\"||body=\"<div id=\"nc_img\" onmouseover=\"overImage\"||(title==\"产品登录界面\"&&body=\"UFIDA NC\")")
	//t.Log(MatchRules(*result4, content))
	//	result5 := ParseRules("header=\"python\"||header=\"Django\"")
	//	t.Log(MatchRules(*result5, content))
}
