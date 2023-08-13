package pocscan

import "testing"

func TestFScan_ParseContentResult(t *testing.T) {
	var fscanResult = `(ping) Target 192.168.3.183   is alive
(ping) Target 192.168.3.184   is alive
(icmp) Target 192.168.3.1     is alive
(icmp) Target 192.168.3.15    is alive
(icmp) Target 192.168.3.167   is alive
(icmp) Target 192.168.3.181   is alive
(icmp) Target 192.168.3.199   is alive
(icmp) Target 192.168.3.226   is alive
(icmp) Target 192.168.3.224   is alive
(icmp) Target 192.168.3.242   is alive
(icmp) Target 192.168.3.243   is alive
(icmp) Target 192.168.3.255   is alive
(icmp) Target 192.168.3.228   is alive
[*] Icmp alive hosts len is: 11
[*] LiveTop 192.168.3.0/24   段存活数量为: 11
192.168.3.242:80 open
192.168.3.167:80 open
192.168.3.1:53 open
192.168.3.167:23 open
192.168.3.167:443 open
192.168.3.167:631 open
[*] alive ports len is: 5
start vulscan
[*] WebTitle:http://192.168.3.242      code:200 len:86     title:NemoTest
[*] WebTitle:http://192.168.3.167:631  code:200 len:254    title:None
[*] WebTitle:https://192.168.3.167  code:200 len:254    title:HTTPS Title
[+] InfoScan:https://192.168.3.242:80     [天融信防火墙]
[+] InfoScan:http://192.168.3.167:80/login;JSESSIONID=b46eae74-5399-462c-a5c7-15ce187f4c73 [Shiro]
[+] InfoScan:http://192.168.3.242:80/login [SpringBoot]
[*]192.168.10.136
   [->]USER-20181206MW
   [->]192.168.10.136
[*] WebTitle:https://192.168.10.2      code:404 len:0      title:None
[*] WebTitle:http://192.168.10.40      code:200 len:689    title:IIS7
[+] http://192.168.10.229:9200 poc-yaml-elasticsearch-unauth
[+] http://192.168.10.232:9200 poc-yaml-elasticsearch-unauth
[+] http://192.168.10.234:9200 poc-yaml-elasticsearch-unauth
[+] http://192.168.10.233:9200 poc-yaml-elasticsearch-unauth
[+] http://192.168.10.236:8085 poc-yaml-vmware-vcenter-cve-2021-21985-rce
[+] http://192.168.10.237:8085 poc-yaml-vmware-vcenter-cve-2021-21985-rce
[+] SSH:192.168.10.236:22:root root@123
[+] mysql:192.168.10.239:3306:root 123456
[+] Redis:192.168.0.187:6379 like can write /var/spool/cron/
[+] 192.168.0.60 CVE-2020-0796 SmbGhost Vulnerable
NetInfo:
[*]192.168.0.35
   [->]WIN-A68V66RGFAF
   [->]192.168.0.35
   [->]192.168.61.1
   [->]192.168.176.1
   [->]2001:0:348b:fb58:14e0:3676:3f57:ffdc
已完成 5/5
scan end`

	i := NewImportOfflineResult("fscan", 1)
	i.Parse([]byte(fscanResult))
	for _, v := range i.VulResult {
		t.Log(v)
	}
}
