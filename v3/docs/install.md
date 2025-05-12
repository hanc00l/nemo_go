# Nemo安装手册

## v1.0

2025-4-15

Nemo分为**Server**端和**Worker**
端两部份。Server提供Http访问、RPC接口、消息中间件服务。Worker是通过消息中间件从Worker接收任务并执行，通过RPC接口上传任务的结果，并通过文件同步接口接收Server的文件。

**Server需要安装的组件：**

- Mongodb
- Redis

**Worker需要安装的组件：**

- Nmap
- Masscan
- Chrome

**Worker其它使用工具已集成到thirdparty目录中：**

- httpx
- subfinder
- massdns
- nuclei
- fingerprintx
- gogo

Nemo目前可运行在**x86-AMD**平台的MacOS、Linux及Windows平台，其它平台目前未做测试。本文档均以Ubuntu
Linux版本进行介绍安装步骤，在Mac、Windows平台及其它请参考相应的安装步骤。

**Server运行后，默认会开启以下端口：**

- 5000：Http，web界面
- 5001：RPC接口，用于worker将任务结果保存到server
- 5002：Redis安全代理隧道，用于Worker与Server通信
- 27017：mongodb数据库，用于数据存储，默认在127.0.0.1上开放；
- 6379：服务器Redis，用于存储任务队列，默认在127.0.0.1上开放；使用Redis安全代理隧道后，该端口不需要暴露到公网

Worker不会开启对外的任务监听端口，但默认开启本地socks5转发和本地Redis转发，监听地址将为本地127.0.0.1。

- 5010：本地socks5转发，用于chrome-headless支持带验证的socks5代理；如果该端口被占用，会自动递增1个可使用的端口
- 16379：本地Redis转发，worker通过与Server建立Redis安全隧道后，会在该端口进行redis代理

## 在VPS安装

(在ubuntu 22.04LTS上测试，其它linux版本请自行测试)

**一、server端**

- 1、将nemo_linux_amd64.tar解压到nemo目录
- 2、运行./server_install.sh
- 3、修改server.yml中的mongodb、redis的地址和端口，与实际情况保持一致
- 4、server参数：
```bash
 Usage:
  main [OPTIONS]

Application Options:
      --web            web service
      --cron           cron service
      --service        rpc service
      --redis-tunnel   redis tunnel service

web-option:
      --tls_cert_file= TLS certificate file (default: server.crt)
      --tls_key_file=  TLS key file (default: server.key)

Help Options:
  -h, --help           Show this help message

```

**默认安装下，mongodb和redis只会在本地127.0.0.1上运行，如无必要请不要开启外网访问。**

**二、worker端安装**

- 1、创建nemo目录，将daemon_worker_linux_amd64和worker_install.sh放入nemo目录
- 2、运行./worker_install.sh
- 3、v3版本的worker启动参数传递有三种方式：通过命令行指定、通过配置文件指定和通过环境变量指定。通过命令行指定server的service地址、端口、authKey，其余参数与service连接后由运行中获取；通过配置文件的方式所有参数都在conf/worker.ymal中提前配置后，通过-f指定文件。
- 4、如果worker支持ipv6，请在启动参数中添加--ipv6参数。
- 5、worker的参数
```bash
Usage:
  main [OPTIONS]

Application Options:
  -f, --config-file=          config file
      --no-proxy              disable proxy configuration,include socks5 proxy and socks5forward
      --no-redis-proxy        disable redis proxy configuration
      --ipv6                  support ipv6 portscan

services:
      --service=              Service host
      --port=                 Service port (default: 5001)
      --auth=                 Service auth

worker-tasks:
  -c, --concurrency=          Number of concurrent workers (default: 2)
  -p, --worker-performance=   worker performance,default is autodetect (0:autodetect, 1:high, 2:normal) (default: 0)
  -m, --worker-run-task-mode= worker run task mode; 0: all, 1:active, 2:finger, 3:passive, 4:pocscan, 5:custom; run multiple mode separated by "," (default: 0)

Help Options:
  -h, --help                  Show this help message 
```
**三、修改默认配置密码**

- 1、进入nemo/conf目录，分别修改server.yml和worker.yml中Service、redisTunnel的authKey，建议16位以上随机字符串。
- 2、建议redis增加密码验证（可选，具体请参考相应文档）。


**四、启动Server**

- 1、启动Server：./nemo_server
```bash
  ./nemo_linux_amd64  --web --service --cron --redis-tunnel 
````
**五、启动Worker**
- 启动Worker，请将Server的 service地址、端口、authKey替换为实际情况。
```bash
  ./worker_linux_amd64 --service x.x.x.x --port 5001 --auth da8Ae0e295eba72a7rgb1d34b13d86b
```

## Docker安装

- **下载release的nemo_linux_amd64.tar**

  ```bash
  mkdir nemo;tar xvf nemo_linux_amd64.tar -C nemo;cd nemo
  docker-compose up -d
  ```

#### worker任务模式

Nemo将任务分为4种类型，worker启动时通过参数-m指定worker执行的任务类型，可以指定一种或多种任务类型；参数分别用1-4，如果为0则表示可执行类型为1-4的任务。

任务类型及数值：

- 1：Active，主动扫描类的任务
- 2：Finger，获取指纹类的任务
- 3：Passive，被动收集信息的任务
- 4：Pocscan，漏洞验证类的任务
- 0：同时包含1-4种类型的任务

|              | 1:Active | 2:Finger | 3:Passive | 4:Pocscan |  
|--------------|----------|----------|-----------|-----------|
| nmap         | √        |          |           |           |          
| masscan      | √        |          |           |           |          
| gogo         | √        |          |           |           |          
| subfinder    |          |          | √         |           |         
| massdns      |          |          | √         |           |    
| iplocation   |          |          | √         |           |        
| fofa         |          |          | √         |           |              
| quake        |          |          | √         |           |             
| hunter       |          |          | √         |           |            
| llmapi       |          |          | √         |           |            
| icp          |          |          | √         |           |          
| whois        |          |          | √         |           |        
| httpx        |          | √        |           |           |       
| fingerprintx |          | √        |           |           |       
| nuclei       |          |          |           | √         |           

Worker默认启动时参数为-m 0，将会执行所有类型（除custom）的任务；分布式部署的vps可以合理分配资源，如专用于扫描类vps：-m
1，被动信息搜索指纹可以同时执行任务：-m 2,3。


## Worker使用自定义的socks5代理

为提高Worker部署和扫描的灵活性，在v2.11版本后，Worker的部份任务功能支持sock5代理，包括：

- gogo的端口扫描
- 指纹获取
- 在线API接口
- 子域名任务被动收集（subfinder）
- Nuclei漏洞验证

代理设置：Config-配置管理，支持同时配置多个socks5代理地址（地址格式为socks5://user:pass@host:port），多个地址将由worker每次任务时随机选择。

由于获取网站截图时调用的chrome-headless不支持验证功能的socks5代理，因此worker在启动时默认在127.0.0.1:
5010地址进行代理转发到设置的socks5地址。

worker可通过命令行参数-np关闭代理功能。如果前端任务指定了代理扫描选项，但未配置socks5地址或worker关闭了代理功能，任务将会由不使用代理直接执行。



关于使用代理的注意事项 ：

- socks5代理为网络层代理，主要用于TCP协议代理，如果需代理udp，请修改redsocks.rules配置以及iptables规则文件，具体请自行参考文档；
- nmap与masscan的SYN扫描不支持在socks5代理下使用，所以使用代理的worker尽量不要分配active类型的任务（-m
  参数为1或0）；如果必须要使用，只能使用nmap的-sT的扫描类型；
- 对于在线API接口的任务（比如FOFA、Hunter、ICP查询等），不建议使用代理功能，防止因源IP导致访问被限；
- 使用代理功能时，建议合理分配worker的任务类型；
- 推荐使用gost实现自已搭建的多个代理的负载均衡；如果要更多的代理池功能，建议购买第三方的代理服务（比如快代理的隧道代理）；
- 以上测试只在UbuntuLTS 22.04中测试稳定运行，在其它linux版本及docker中请自行参考网上文档。