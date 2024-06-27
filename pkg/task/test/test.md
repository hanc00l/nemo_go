# Nemo 自动化测试

## 1. 模拟环境：nemo_test

## 2. 运行方式

```bash
$ cd nemo/pkg/task/test
$ go test -v -count=1  .
```

## 3. 自动化测试项目：

- [x] 3.1. 端口扫描：nmap、masscan
- [x] 3.2. 域名任务：subfinder、massdns
- [x] 3.3. 指纹识别：httpx（网站标题、状态码、指纹识别）、iconhash（网站图标指纹）、screenshot、fingerprintx、fingerprinthub被动识别
- [x] 3.4. 漏洞扫描：nuclei、xray

## 测试结果样例：
```text
=== RUN   TestMassdns_test
cp.v.news.cn
player.v.news.cn
see.v.news.cn
monitor1.v.news.cn
yunbo.v.news.cn
show.v.news.cn
interact.v.news.cn
live.v.news.cn
stat.v.news.cn
review.v.news.cn
cloud.v.news.cn
[INF] Finished resolving. Hack the Planet!
    domainscan_test.go:54: massdns find subdomain total:11
--- PASS: TestMassdns_test (165.71s)
=== RUN   TestSubfinder
vodpub1.v.news.cn
oldvod2.v.news.cn
vodpub2.v.news.cn
review.v.news.cn
live.v.news.cn
source10hls.v.news.cn
show.v.news.cn
player.v.news.cn
source08.v.news.cn
source07.v.news.cn
--- PASS: TestSubfinder (38.21s)
=== RUN   TestHttp_test
time="2024-06-27 20:27:15" level=info msg="Load fingerprinthub total:3487"
time="2024-06-27 20:27:15" level=info msg="httpx output tempdir is /var/folders/0p/pd6qx8z12bbgt_pkl_y03d6w0000gn/T/b1907833ac47aed5.dir"
{"timestamp":"2024-06-27T20:27:31.354294+08:00","hash":{"body_md5":"be81a470fbf6469c719bbbd20e6a9192","body_mmh3":"1540059936","body_sha256":"36f34b6897c72b9ae6a8d9b44c27ffd26361267ec0072fd49721f879c445f197","body_simhash":"9825648629383921534","header_md5":"a8e28bd723b2f6a59daaab7fbb234ba4","header_mmh3":"-82235458","header_sha256":"58f2451857e2a330f4434b94b4ab0a9ada5d061abb67be694c6245b4bafca1ee","header_simhash":"15614743101849712621"},"port":"9200","url":"http://127.0.0.1:9200","input":"127.0.0.1:9200","scheme":"http","content_type":"application/json","method":"GET","host":"127.0.0.1","path":"/","favicon_path":"/favicon.ico","time":"4.31252ms","a":["127.0.0.1"],"words":122,"lines":12,"status_code":200,"content_length":345,"failed":false,"stored_response_path":"/var/folders/0p/pd6qx8z12bbgt_pkl_y03d6w0000gn/T/b1907833ac47aed5.dir/response/127.0.0.1_9200/6e08d1bf80e75641f576a59d57d45d64684410f3.txt","screenshot_path":"/var/folders/0p/pd6qx8z12bbgt_pkl_y03d6w0000gn/T/b1907833ac47aed5.dir/screenshot/127.0.0.1_9200/6e08d1bf80e75641f576a59d57d45d64684410f3.png","screenshot_path_rel":"127.0.0.1_9200/6e08d1bf80e75641f576a59d57d45d64684410f3.png","knowledgebase":{"PageType":"nonerror","pHash":10312996234928906464}}
{"timestamp":"2024-06-27T20:27:32.343332+08:00","hash":{"body_md5":"bda2113ce2adcdaa40dba090d50c4608","body_mmh3":"-1922502038","body_sha256":"d8e1e2b10843c149d6e8b1bd0cb949c70206941eeab052256eaea2d09333a1b7","body_simhash":"16227071425473497477","header_md5":"9e4fe6a4d61351c9823d696318dadb64","header_mmh3":"370843095","header_sha256":"d3cb1dda4c8b30ab3287a8b4909f80569f6f8cdafed40243d210b64cc6806909","header_simhash":"10998551284930033645"},"port":"80","url":"http://127.0.0.1:80","input":"127.0.0.1:80","title":"在线资产管理平台","scheme":"http","webserver":"PHP/8.3","content_type":"text/html","method":"GET","host":"127.0.0.1","path":"/","favicon":"673152537","favicon_path":"/favicon.ico","time":"4.283618ms","a":["127.0.0.1"],"words":5,"lines":1,"status_code":200,"content_length":145,"failed":false,"stored_response_path":"/var/folders/0p/pd6qx8z12bbgt_pkl_y03d6w0000gn/T/b1907833ac47aed5.dir/response/127.0.0.1_80/57b32cdd2eb04520e58112870b9e32a4cfb6114c.txt","screenshot_path":"/var/folders/0p/pd6qx8z12bbgt_pkl_y03d6w0000gn/T/b1907833ac47aed5.dir/screenshot/127.0.0.1_80/57b32cdd2eb04520e58112870b9e32a4cfb6114c.png","screenshot_path_rel":"127.0.0.1_80/57b32cdd2eb04520e58112870b9e32a4cfb6114c.png","knowledgebase":{"PageType":"error","pHash":9732013877122502415}}
time="2024-06-27 20:27:32" level=info msg="http://127.0.0.1:80/favicon.ico"
{"timestamp":"2024-06-27T20:27:32.391073+08:00","tls":{"host":"127.0.0.1","port":"443","probe_status":true,"tls_version":"tls13","cipher":"TLS_AES_128_GCM_SHA256","self_signed":true,"not_before":"2024-05-29T02:32:47Z","not_after":"2034-05-27T02:32:47Z","subject_dn":"CN=127.0.0.1","subject_cn":"127.0.0.1","subject_an":["localhost"],"issuer_dn":"CN=127.0.0.1","issuer_cn":"127.0.0.1","fingerprint_hash":{"md5":"717a453537c6e7aaf73f05363e3134cd","sha1":"a12eb630593c9999dad0117d8f3eab692442738a","sha256":"963f7bc86b3b7a87c7fd26f018cbdc9471a56992dee76f259097a2547e581c2d"},"tls_connection":"ctls"},"hash":{"body_md5":"4a1e4ef9be70031719578293adb36b51","body_mmh3":"-485498080","body_sha256":"bd4499b44a1fe493a0b0e78cc38dae98755c71bd328a4ffd6f4913ac435fcc4e","body_simhash":"17668238709599592854","header_md5":"0786813e6cb14e08500120661cd09517","header_mmh3":"-987518008","header_sha256":"1c7cf93496803119219a3eb54fd8008e297c89872b2d27188463730eddceba3b","header_simhash":"10967026087412610540"},"port":"443","url":"https://127.0.0.1:443","input":"127.0.0.1:443","scheme":"https","webserver":"ASP.Net/3.5","content_type":"text/html","method":"GET","host":"127.0.0.1","path":"/","favicon":"673152537","favicon_path":"/favicon.ico","time":"1.412761315s","jarm":"3fd3fd20d00000000043d3fd3fd43d684d61a135bd962c8dd9c541ddbaefa8","a":["127.0.0.1"],"words":2,"lines":1,"status_code":200,"content_length":63,"failed":false,"stored_response_path":"/var/folders/0p/pd6qx8z12bbgt_pkl_y03d6w0000gn/T/b1907833ac47aed5.dir/response/127.0.0.1_443/24d4b6b9d4be1be238fb42c67f4a068fba01b7eb.txt","screenshot_path":"/var/folders/0p/pd6qx8z12bbgt_pkl_y03d6w0000gn/T/b1907833ac47aed5.dir/screenshot/127.0.0.1_443/24d4b6b9d4be1be238fb42c67f4a068fba01b7eb.png","screenshot_path_rel":"127.0.0.1_443/24d4b6b9d4be1be238fb42c67f4a068fba01b7eb.png","knowledgebase":{"PageType":"nonerror","pHash":9295714417410834431}}
time="2024-06-27 20:27:32" level=info msg="https://127.0.0.1:443/favicon.ico"
--- PASS: TestHttp_test (17.42s)
=== RUN   TestNuclei_test
--- PASS: TestNuclei_test (4.04s)
=== RUN   TestXray_test
[Vuln: phantasm]
Target           "http://127.0.0.1:9200"
VulnType         "poc-yaml-elasticsearch-unauth/default"
Author           "p0wd3r"
Links            ["https://yq.aliyun.com/articles/616757"]
level            "high"

--- PASS: TestXray_test (1.53s)
=== RUN   TestNmap_test
Starting Nmap 7.95 ( https://nmap.org ) at 2024-06-27 20:27 CST
Nmap scan report for 127.0.0.1
Host is up (0.0000070s latency).
Not shown: 990 closed tcp ports (reset)
PORT     STATE SERVICE
21/tcp   open  ftp
22/tcp   open  ssh
23/tcp   open  telnet
80/tcp   open  http
443/tcp  open  https
5900/tcp open  vnc
9200/tcp open  wap-wsp

Nmap scan report for 127.0.0.1
Host is up (0.000011s latency).
Not shown: 992 closed tcp ports (reset)
PORT     STATE SERVICE
21/tcp   open  ftp
22/tcp   open  ssh
23/tcp   open  telnet
80/tcp   open  http
443/tcp  open  https
5900/tcp open  vnc
9200/tcp open  wap-wsp

Nmap done: 2 IP addresses (2 hosts up) scanned in 0.08 seconds
--- PASS: TestNmap_test (0.14s)
=== RUN   TestMasscan_test
--- PASS: TestMasscan_test (13.49s)
PASS
ok  	github.com/hanc00l/nemo_go/pkg/task/test	241.929s
```