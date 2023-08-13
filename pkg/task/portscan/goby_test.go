package portscan

import "testing"

func TestGoby_ParseContentResult(t *testing.T) {
	gobyAssertResult := `{"statusCode":200,"messages":"","data":{"query":"taskId=20230813112031","options":{"TaskID":"20230813112031","Page":1,"size":100000,"OrderField":"","OrderASC":""},"taskId":"20230813112031","query_total":{"ips":1,"ports":1,"protocols":2,"assets":3,"vulnerabilities":2,"dist_ports":1,"dist_protocols":2,"dist_assets":3,"dist_vulnerabilities":2},"total":{"assets":3,"ips":1,"ports":1,"vulnerabilities":2,"allassets":0,"allips":0,"allports":0,"allvulnerabilities":0,"scan_ips":0,"scan_ports":0},"ips":[{"ip":"127.0.0.1","mac":"","os":"LinuxKit","hostname":"","honeypot":"0","ipTag":"","ports":[{"port":"8161","baseprotocol":"tcp"}],"protocols":{"127.0.0.1:8161":{"port":"8161","hostinfo":"127.0.0.1:8161","url":"","product":"Log4j2|Jetty","protocol":"http","json":"","fid":["BeXtuZI3tG69/8LJRGs3YlV/qTd9FmG3"],"products":["Log4j2","Jetty","APACHE-ActiveMQ"],"protocols":["http","web"]}},"tags":[{"rule_id":"223","product":"Jetty","company":"Eclipse Foundation, Inc.","level":"3","category":"Service","parent_category":"Support System","softhard":"2","version":"8.1.16.v20140903"},{"rule_id":"855434","product":"Log4j2","company":"其他","level":"4","category":"Component","parent_category":"Support System","softhard":"2","version":""},{"rule_id":"2471","product":"APACHE-ActiveMQ","company":"Apache Software Foundation.","level":"4","category":"Other Enterprise Application","parent_category":"Enterprise Application","softhard":"2","version":""}],"vulnerabilities":[{"hostinfo":"127.0.0.1:8161","name":"ActiveMQ Arbitrary File Write Vulnerability (CVE-2016-3088)","filename":"ActiveMQ_RCE_CVE_2016_3088.json","level":"3","vulurl":"","keymemo":"","hasexp":true},{"hostinfo":"127.0.0.1:8161","name":"ActiveMQ default admin account","filename":"ActiveMQ_default_account.json","level":"2","vulurl":"http://admin:admin@127.0.0.1:8161/admin/","keymemo":"","hasexp":false}],"screenshots":null,"favicons":[{"hostinfo":"127.0.0.1:8161","imgpath":"/screenshots/20230813112031/127.0.0.1-8161-f.ico","imgsize":"3638","phash":"1766699363"}],"hostnames":[""]}],"products":{"software":{"total_assets":1,"risk_assets":0,"lists":[{"name":"Jetty","company":"Eclipse Foundation, Inc.","total_assets":1,"risk_assets":0}]},"hardware":{"total_assets":0,"risk_assets":0,"lists":null}},"companies":{"software":{"total_assets":1,"risk_assets":0,"lists":[{"name":"Eclipse Foundation, Inc.","total_assets":1,"risk_assets":0}]},"hardware":{"total_assets":0,"risk_assets":0,"lists":null}}}}`
	i := NewImportOfflineResult("goby")
	i.Parse([]byte(gobyAssertResult))
	for ip, ipa := range i.IpResult.IPResult {
		t.Log(ip, ipa)
		for port, pa := range ipa.Ports {
			t.Log(port, pa)
		}
	}
}
