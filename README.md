## Nemo

<img src="docs/image/index.png" alt="login" />

Nemo是用来进行自动化信息收集的一个简单平台，通过集成常用的信息收集工具和技术，实现对内网及互联网资产信息的自动收集，提高隐患排查和渗透测试的工作效率，用Golang完全重构了原Python版本。




## 已实现的功能

### 1、IP资产

- Masscan、Nmap端口扫描
- IP归属地（纯真离线数据）
- 自定义IP归属地、Service、蜜罐
- 导入本地的Masscan、Nmap端口扫描结果
- 导入[fscan](https://github.com/shadow1ng/fscan)扫描结果（适用于内网渗透的资产信息收集）
- CDN识别

### 2、域名资产

- [Subfinder](https://github.com/projectdiscovery/subfinder) 子域名收集
- [Massdns](https://github.com/blechschmidt/massdns) 子域名爆破
- [Crawlergo](https://github.com/Qianlitp/crawlergo) 子域名爬虫

### 3、指纹信息

- [HTTPX](https://github.com/projectdiscovery/httpx) 
- [ScreenShot](https://github.com/chromedp/chromedp) （调用chrome headless）
- [Wappalyzer](https://github.com/AliasIO/Wappalyzer) （基于[webanalyze](https://github.com/rverton/webanalyze) 代码，可[自定义指纹规则](thirdparty/wappalyzer/technologies_custom.json)）
- [ObserverWard](https://github.com/0x727/ObserverWard_0x727)  (指纹信息来源于https://github.com/0x727/FingerprintHub)
- IconHash（基于[mat/besticon](github.com/mat/besticon)和[Becivells/iconhash](github.com/Becivells/iconhash)项目）


### 4、API接口 （需提供相应的Key)

- [Fofa](https://fofa.so/) 
- [ICP备案信息](http://icp.chinaz.com/) 
- [Quake](https://quake.360.cn) [TODO]

### 5、Poc验证与目录扫描

- [Pocsuite3](https://github.com/knownsec/pocsuite3)  && [some_pocsuite](https://github.com/hanc00l/some_pocsuite) 
- [XRay](https://github.com/chaitin/xray)
- [Dirsearch](https://github.com/evilsocket/dirsearch)

### 6、扫描任务

- 分布式、异步任务执行
- 多维度任务切分
- Server与Worker通过 [RPC](https://github.com/smallnest/rpcx) 同步


### 7、团队在线协作

- [TODO]

### 8、其它

- 资产的统计、颜色标记与备忘录协作
- Docker支持

## Docker

```shell
mkdir nemo;tar xvf nemo_linux_amd64.tar -C nemo;cd nemo
docker-compose up -d
```

[更多Docker使用方式](docs/docker.md)



## Install

Tested on [ubuntu18.04/20.04 LTS](docs/install_linux.md)、[macOS](docs/install_mac.md)



## Demo

默认监听端口为5000，默认密码 **nemo** ；通过“系统设置--配置管理”更改默认密码。

<img src="docs/image/dashboard.png" alt="dashbord"  />

<img src="docs/image/iplist.png" alt="iplist"  />

<img src="docs/image/ipinfo.png" alt="ipinfo"  />

<img src="docs/image/portscan.png" alt="portscan"  />

<img src="docs/image/domainlist.png" alt="domainlist"  />

<img src="docs/image/domaininfo.png" alt="domaininfo"  />

<img src="docs/image/domainscan.png" alt="domainscan"  />

<img src="docs/image/vulnerabilitylist.png" alt="vulnerabilitylist"  />

<img src="docs/image/vulnerabilityinfo.png" alt="vulnerabilityinfo"  />

<img src="docs/image/config.png" alt="config"  />



## 版本更新

- 2.4.13：2022-1-8，增加导入fscan的扫描结果（由于只有txt方式，通过正则匹配提取IP、端口及一些重要信息，不保证百分百全部导入）；
- 2.4.12：2022-1-4，修复域名扫描同时进端口扫描时不正确创建任务导致worker退出的Bug；
- 2.4.11：2022-1-1，增加目录遍历扫描（[dirsearch](https://github.com/evilsocket/dirsearch)，建议谨慎使用）；
- 2.4.10：2021-12-31，增加子域名爬虫功能（感谢 [crawlergo](https://github.com/Qianlitp/crawlergo) 开源的代码）；
- 2.4.9：2021-12-28，增加域名和IP的CDN识别（借鉴 [github.com/timwhitez/Frog-checkCDN](https://www.github.com/timwhitez/Frog-checkCDN)）;
- 2.4.8：2021-12-13，增加根据favicon.ico获取IconHash指纹功能；
- 2.4.7：2021-12-9，增加导入本地扫描结果功能（支持nmap/masscan的XML文件），增加保存并查看任务执行结果的JSON格式的文件；优化FOFA查询后获取IP与域名的Web指纹信息；更新Xray版本为1.8.2；
- 2.4.6：2021-12-8，更新指纹识别\"侦查守卫\"的JSON结果的解析；
- 2.4.5：2021-12-7，增加调用ObserverWard获取应用系统的指纹信息，指纹信息来源于 [FingerprintHub](https://github.com/0x727/FingerprintHub)；
- 2.4.4：2021-10-18，对新建任务增加部份提示信息，便于掌握任务执行的参数；状态信息可手动刷新和查看正在执行的任务；
- 2.4.3：2021-10-13，增加IP扫描的masscan+nmap方法，masscan快速进行端口开放扫描，nmap用-sV进行详细扫描；
- 2.4.2：2021-10-9，增加IP扫描的“探测+扫描”模式任务，增加内网资产收集的便利性；去除whatweb的安装和使用（HTTPX已基本可替代其功能）；
- 2.4.1：2021-9-15，支持扫描任务按IP和端口进行多维度切分，使任务在多个worker之间均衡分布执行；
- 2.4.0：2021-9-10，使用RPC架构，优化server与worker之间的同步、server与worker的配置文件分离；增加在线的IP信息、登录验证码、按发现时间筛选资产功能。
- 2.3：2021-8-25，使用docker-compose构建Docker，修改数据库连接重试功能，完善端口扫描及任务等信息；
- 2.2：2021-8-2，增加基于Wappalyzer规则的指纹识别功能；
- 2.1：2021-7-30，用Golang完成对原python3版本的重构；



## 参考

- jeffzh3ng：https://github.com/jeffzh3ng/fuxi
- TideSec：https://github.com/TideSec/Mars