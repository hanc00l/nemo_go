## Nemo

<img src="docs/image/index.png" alt="login" />

Nemo是用来进行自动化信息收集的一个简单平台，通过集成常用的信息收集工具和技术，实现对内网及互联网资产信息的自动收集，提高隐患排查和渗透测试的工作效率，用Golang完全重构了原Python版本。




## 已实现的功能

### 1、IP资产

- Masscan、Nmap端口扫描
- IP归属地（纯真离线数据）
- 自定义IP归属地、Service、蜜罐

### 2、域名资产

- [Subfinder](https://github.com/projectdiscovery/subfinder) 子域名收集
- [Massdns](https://github.com/blechschmidt/massdns) 子域名爆破
- JSFinder [TODO]

### 3、指纹信息

- [HTTPX](https://github.com/projectdiscovery/httpx) 
- [ScreenShot](https://github.com/chromedp/chromedp) （调用chrome headless）
- [Wappalyzer](https://github.com/AliasIO/Wappalyzer) （基于[webanalyze](https://github.com/rverton/webanalyze) 代码，可[自定义指纹规则](thirdparty/wappalyzer/technologies_custom.json)）

### 4、API接口 （需提供相应的Key)

- [Fofa](https://fofa.so/) 
- [ICP备案信息](http://icp.chinaz.com/) 
- [Quake](https://quake.360.cn) [TODO]

### 5、Poc验证

- [Pocsuite3](https://github.com/knownsec/pocsuite3)  && [some_pocsuite](https://github.com/hanc00l/some_pocsuite) 
- [XRay](https://github.com/chaitin/xray)

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

正式使用建议独立VPS分布式部署。



## Install

Tested on [ubuntu18.04 LTS](docs/install_linux.md)、[macOS](docs/install_mac.md)



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