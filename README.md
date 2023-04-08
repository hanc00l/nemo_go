## Nemo

<img src="docs/image/index.png" alt="login" />

Nemo是用来进行自动化信息收集的一个简单平台，通过集成常用的信息收集工具和技术，实现对内网及互联网资产信息的自动收集，提高隐患排查和渗透测试的工作效率，用Golang完全重构了原Python版本。




## 功能

### 1、IP资产

- Masscan、Nmap端口扫描
- IP归属地（纯真离线数据）
- 自定义IP归属地、Service、蜜罐
- 导入本地的Masscan、Nmap端口扫描结果
- 导入[fscan](https://github.com/shadow1ng/fscan)、[gogo](https://github.com/chainreactors/gogo)、[Naabu]( https://github.com/projectdiscovery/naabu)、[Httpx]( https://github.com/projectdiscovery/httpx)及TXPortmap的扫描结果（适用于内网渗透的资产信息收集）
- 导入FOFA、Hunter及0Zone的查询结果导出的资产文件
- CDN识别

### 2、域名资产

- [Subfinder](https://github.com/projectdiscovery/subfinder) 子域名收集
- [Massdns](https://github.com/blechschmidt/massdns) 子域名爆破
- [Crawlergo](https://github.com/Qianlitp/crawlergo) 子域名爬虫
- [Whois](https://github.com/likexian/whois)

### 3、指纹信息

- [HTTPX](https://github.com/projectdiscovery/httpx) 
- [ScreenShot](https://github.com/chromedp/chromedp) （调用chrome headless）
- [ObserverWard](https://github.com/0x727/ObserverWard_0x727)  (指纹信息来源于https://github.com/0x727/FingerprintHub)
- IconHash（基于[mat/besticon](github.com/mat/besticon)和[Becivells/iconhash](github.com/Becivells/iconhash)项目）


### 4、API接口 （需提供相应的Key)

- [Fofa](https://fofa.info/) 
- [ICP备案信息](http://icp.chinaz.com/) 
- [Quake](https://quake.360.cn)
- [Hunter](https://hunter.qianxin.com/)

### 5、Poc验证与目录扫描

- [XRay](https://github.com/chaitin/xray)
- [Nuclei](https://github.com/projectdiscovery/nuclei) && [Nuclei-Templates](https://github.com/projectdiscovery/nuclei-templates)
- [Dirsearch](https://github.com/evilsocket/dirsearch)

### 6、分布式任务

- 分布式、异步任务执行
- 多维度任务切分
- 定时任务执行
- Server与Worker通过 [RPC](https://github.com/smallnest/rpcx) 同步
- Server与Worker文件自动同步
- 任务执行完成消息通知（钉钉、飞书群机器人及Server酱）


### 7、团队在线协作

- 多用户/角色、多工作空间（项目）支持
- 资产颜色标记、置顶、备忘录协作

### 8、其它

- Docker支持
- 资产流程化扫描（XSCAN）

## Docker

```shell
mkdir nemo;tar xvf nemo_linux_amd64.tar -C nemo;cd nemo
docker-compose up -d
```

[更多Docker使用方式](docs/docker.md)



## Install

Tested on [ubuntu18.04/20.04/22.04 LTS](docs/install_linux.md)、[macOS](docs/install_mac.md)、Windows（Test In Win10）



## Demo

- 默认监听端口为5000，默认用户**nemo（超级管理员）**、密码 **nemo** ；通过“Config--配置管理”更改默认密码，通过“System--User“创建和管理用户权限。
- 用户角色分为superadmin、admin和guest三种； **superadmin（超级管理员）** 可管理用户和工作空间，**admin（管理员）** 可管理资源、任务和参数配置，**guest（普通用户）** 只有资源和任务的查看权限。
- 工作空间对资源（IP、Domain、任务、组织及漏洞）进行隔离，每个工作空间可分配给不同的用户访问权限。

<img src="docs/image/dashboard.png" alt="dashbord"  />

<img src="docs/image/iplist.png" alt="iplist"  />

<img src="docs/image/ipinfo.png" alt="ipinfo"  />

<img src="docs/image/portscan.png" alt="portscan"  />

<img src="docs/image/domainlist.png" alt="domainlist"  />

<img src="docs/image/domaininfo.png" alt="domaininfo"  />

<img src="docs/image/domainscan.png" alt="domainscan"  />

<img src="docs/image/vulnerabilitylist.png" alt="vulnerabilitylist"  />



<img src="docs/image/config.png" alt="config"  />

![xscan2](docs/image/xscan1.png)



![xscan2](docs/image/xscan2.png)



## [CHANGELOG](CHANGELOG.md)


## 参考

- [jeffzh3ng](https://github.com/jeffzh3ng/fuxi)
- [TideSec](https://github.com/TideSec/Mars)