# Nemo安装手册

## v1.3

2024-1-3


Nemo分为**Server**端和**Worker**端两部份。Server提供Http访问、API接口、RPC接口、消息中间件服务以及文件同步接口。Worker是通过消息中间件从Worker接收任务并执行，通过RPC接口上传任务的结果，并通过文件同步接口接收Server的文件。

**Server需要安装的组件：**
- MySQL 
- Rabbitmq

**Worker需要安装的组件：**
- Nmap
- Masscan
- Chrome

**Worker其它使用工具已集成到thirdparty目录中：**
- httpx
- subfinder
- massdns
- nuclei
- xray
- goby
- fingerprintx
- gogo

Nemo目前可运行在**x86-AMD**平台的MacOS、Linux及Windows平台，其它平台目前未做测试。本文档均以Ubuntu Linux版本进行介绍安装步骤，在Mac、Windows平台及其它请参考相应的安装步骤。

**Server运行后，默认会开启以下端口：**
- 5000：Http，web界面
- 5001：RPC接口，用于worker将任务结果保存到server
- 5002：文件同步，在server与worker之间同步文件
- 5010：本地socks5转发，用于chrome-headless支持带验证的socks5代理；如果该端口被占用，会自动递增1个可使用的端口
- 5672：rabbitmq消息中间件监听端口，用于Nemo的异步和分布式任务执行

如果启用API接口（需单独运行），会开启以下端口：
- 5003

Worker不会开启任务监听端口（启用goby服务端模式除外），worker会根据conf/worker.yml配置的参数与server进行连接。

## Docker安装

按使用需求不同，Docker有两种使用方式：

- 单Docker（Docker里同时安装Server与Worker，适用临时使用）
- 分别构建Server与Worker的Docker（适用动态调整Worker数量及分布式部署Worker）


### 一、单Docker使用 

- **下载release的nemo_linux_amd64.tar**

  ```bash
  mkdir nemo;tar xvf nemo_linux_amd64.tar -C nemo;cd nemo
  docker-compose up -d
  ```

### 二、分别构建Server与Worker的Docker

#### 1 、下载release的nemo_linux_amd64.tar

  ```bash
  mkdir nemo;tar xvf nemo_linux_amd64.tar -C nemo;cd nemo
  ```

#### 2、Server

- **修改docker-compose.server.yml中默认rabbitmq用户和密码**：

  ```dockerfile
  rabbitmq:
      image: rabbitmq:3-management-alpine
      container_name: rabbitmq
      environment:
          # Docker构建server时，需要对外公开rabbitmq供worker连接，建议更改此默认密码
          # 同时修改conf/server.yml和worker.yml
          RABBITMQ_DEFAULT_USER: nemo
          RABBITMQ_DEFAULT_PASS: nemo2020
      ports:
          - 5672:5672
  ```

- **修改conf/server.yml配置文件中，RPC与fileSync的authkey（由worker认证）、rabbitmq的IP、用户和密码**

  ```yaml
  # rpc配置
  rpc:
    authKey: ZduibTKhcbb6Pi8W
  # 消息中间件配置，与docker-compose.server.yml一致
  rabbitmq: 
    username: nemo
    password: nemo2020
  # 文件同步验证
  fileSync:
    authKey: ZduibTKhcbb6Pi8W
  ```
  

- **构建Docker并启动**

  ```bash
  docker-compose -f docker-compose.server.yml up -d
  ```

#### 3、Worker

- **修改conf/worker.yml配置文件中，RPC与fileSync的IP、authkey与rabbitmq的IP、用户和密码**

  ```yaml
  # rpc配置
  rpc:
    host: x.x.x.x(server所在的vps地址）
    authKey: ZduibTKhcbb6Pi8W
  # 消息中间件配置
  rabbitmq:
    host: x.x.x.x(server所在的vps地址）
    username: nemo
    password: nemo2020
  # 文件同步验证
  fileSync:
    host: x.x.x.x(server所在的vps地址）
    authKey: ZduibTKhcbb6Pi8W
  ```
  
- **构建Docker并启动**

  ```bash
  docker-compose -f docker-compose.worker.yml up -d # 默认启动1个worker
  docker-compose -f docker-compose.worker.yml up -d --scale worker=2  #启动指定个worker
  ```
#### 4、关于文件同步

由于server与worker的文件自动同步机制，worker与server的conf/worker.yml配置应先确保一致后，再分别构建Server与Worker的docker镜像，否则可能会导致worker的worker.yml被不正确同步。


## Linux安装

推荐使用Ubuntu22.04 LTS版本安装（以下的安装方法均是在Ubuntu上进行）。根据大家的测试和反馈，也支持其它的Linux发行版本的安装（比如centos），只是各组件的步骤和细节稍有不同，可以参照进行相应的调整。

### 一、Server 

- **创建安装目录并解压tar包**

  ```bash
  mkdir nemo;tar xvf nemo_linux_amd64.tar -C nemo;cd nemo
  ```

- **安装环境和依赖**

  ```bash
  sudo apt-get update \
    && sudo apt-get install vim \
    mysql-server rabbitmq-server --fix-missing -y
  ```

- **导入并配置mysql数据库**：

  由于mysql5.7版本后默认安装后不能使用root和空密码在本机登录，系统自动生成的用户名密码位于/etc/mysql/debian.cnf，请替换并使用配置文件中的user和password：

  ```
  user@localhost:/etc/mysql$ sudo cat debian.cnf
  # Automatically generated for Debian scripts. DO NOT TOUCH!
  [client]
  host     = localhost
  user     = debian-sys-maint
  password = BtRH1gaoI5lPqZpk
  socket   = /var/run/mysqld/mysqld.sock
  ```

  或者直接使用sudo的方式免密码登录。
  导入nemo.sql：

  ```bash
  sudo mysql -e 'CREATE DATABASE `nemo` DEFAULT CHARACTER SET utf8mb4;' \
    && sudo mysql -e 'CREATE USER "nemo"@"%" IDENTIFIED BY "nemo2020";GRANT ALL PRIVILEGES ON nemo.* TO "nemo"@"%";FLUSH PRIVILEGES;' \
    && sudo mysql nemo < docker/mysql/initdb.d/nemo.sql 
  ```

- **配置rabbitmq**：增加rabbitmq用户和密码（建议更换默认密码，并相应在配置文件里同步更改）

  ```bash
  sudo rabbitmqctl add_user nemo nemo2020
  sudo rabbitmqctl set_permissions -p "/" nemo ".*" ".*" ".*"
  ```

- 配置文件

  **conf/server.yml**

  ```yaml
  web:
    # web server 监听IP和地址
    host: 0.0.0.0
    port: 5000
    # v2.9：支持多用户和角色，用户管理在超级管理员登录后在System-User中进行管理
    # webfiles 在用于保存屏幕截图、Icon、任务执行结果等本地保存位置，需与app.conf中与staticdir映射地址保持一致
    webfiles: /tmp/webfiles
  # rpc监听地址和端口、auth
  rpc: 
    host: 0.0.0.0
    port: 5001
    authKey: ZduibTKhcbb6Pi8W
  # 文件同步
  fileSync:
    host: 0.0.0.0
    port: 5002
    authKey: ZduibTKhcbb6Pi8W
  # API接口：
  api:
    host: 0.0.0.0
    port: 5003
  # 数据库配置，server端可默认使用127.0.0.1或localhost
  database:
    host: 127.0.0.1
    port: 3306
    name: nemo
    username: nemo
    password: nemo2020
  # 消息中间件配置，server端可默认使用localhost和guest帐号
  rabbitmq: 
    host: localhost
    port: 5672
    username: guest
    password: guest
  ```

  
    **重要：修改默认的RPC authKey、Rabbitmq消息中间件、数据库及文件同步的密码。**
  
    **conf/app.conf：**
  
    ``` config
    # web映射的目录，static请勿修改；webfiles需和server.yml保持一致
    staticdir = static:web/static webfiles:/tmp/webfiles
    ```


### 二、Worker

- **创建安装目录并解压tar包**

  ```bash
  mkdir nemo;tar xvf worker_linux_amd64.tar -C nemo;cd nemo
  ```

- **安装环境和依赖**

  ```bash
  sudo apt-get update \
    && sudo apt-get install vim nmap masscan --fix-missing -y
  #docker ubuntu
  curl -LO https://dl.google.com/linux/direct/google-chrome-stable_current_amd64.deb \
    && sudo apt install -y ./google-chrome-stable_current_amd64.deb \
    && rm google-chrome-stable_current_amd64.deb
  ```


- **nmap&masscan：** 因为nmap、masscan的SYN扫描需要root权限，为避免使用sudo，设置root权限的suid（如果默认是root则可跳过）

  ```bash
  cd /usr/bin \
    && sudo chown root nmap masscan && sudo chmod u+s nmap masscan 
  ```

- **配置文件**

  **conf/worker.yml** ：（RPC、Rabbitmq用户名和密码应与服务端保持一致）

  ```yaml
   # RPC 调用的server监听地址和端口、auth，host地址和port必须能访问，authKey与server端配置一致
  rpc:
    host: x.x.x.x
    port: 5001
    authKey: ZduibTKhcbb6Pi8W
  # 文件同步，host地址和port必须能访问，authKey与server端配置一致
  fileSync:
    host: x.x.x.x
    port: 5002
    authKey: ZduibTKhcbb6Pi8W
  # 消息中间件，host地址和port必须能访问，用户名与密码与server端配置一致
  rabbitmq: 
    host: x.x.x.x
    port: 5672
    username: nemo
    password: nemo2020
  ```
- **注意：由于chrome-headless因安全的原因，不推荐在root权限下运行，建议worker使用普通用户运行。**

## 运行

 ### 一. Server

```bash
./server_linux_amd64
```
可选参数：
```bash
  -tls
    	use TLS for web、RPC and filesync
  -cert string
    	TLS cert file (default "server.crt")
  -key string
    	TLS private key file (default "server.key")
  -nf
    	disable file sync
  -nr
    	disable rpc
```

**启用TLS**

Server的Web（5000）、RPC（5001）及文件同步（5002）默认不使用TLS；为提高安全性，可配置好SSL证书和私钥文件后，通过命令行-tls启用HTTPS和TLS加密。 如果没有配置默认的server.crt和server.key，将生成并使用自签名证书。

**特别提醒：**

如果Server启用了-tls参数，Worker的daemon也必须启用-tls参数。

### 二. Worker

```bash
./daemon_worker_linux_amd64
```
#### 1、可选参数
```bash
  -c int
    	concurrent number of tasks (default 3)
  -m string
    	worker run task mode; 0: all, 1:active, 2:finger, 3:passive, 4:pocscan, 5:custom; run multiple mode separated by "," (default "0")
  -ma string
    	manual file sync auth key
  -mh string
    	manual file sync host address
  -mp string
    	manual file sync port,default is 5002
  -nf
    	disable file sync (default true)
  -p int
    	worker performance,default is autodetect (0:autodetect, 1:high, 2:normal)
  -tls
    	use TLS for RPC and filesync
  -w string
    	workspace guid for custom task; multiple workspace separated by ","
  -f string
    	worker default config file (default "conf/worker.yml")  
  -np
    	disable proxy configuration,include socks5 proxy and socks5forward
```
- -c worker并发的任务数量，默认为3。
- -mh、mp及ma Server文件同步的host、port及authKey，如果同时指定这三个参数，将执行文件同步功能，从server同步文件到worker。
- -nf 禁用文件同步功能。
- -p worker的的性能模式，默认为0；根据worker的性能模式（1：高性能，2：普通）不同，在任务的并发线程数会有所区别；参数为0则自动判断，判断规则为CPU>=4核、内存>=4G为高性能模式。
- -m worker执行的任务类型
- -w worker执行自定义任务（-m 5）时，自定义任务所在的工作空间GUID
- -tls 启用TLS加密（server也必须使用-tls）
- -f worker配置文件，默认为conf/worker.yml
- -np 禁用socks5代理

#### 2、Goby的服务端部署模式
需在thirdparty/goby目录下运行：（Docker已自动运行）
```bash
 ./goby-cmd-linux -mode api -bind 127.0.0.1:8361 -apiauth goby:goby
```
- 默认只在127.0.0.1上进行监听
- 默认的api验证用户名和密码为goby

同时需要conf/worker.yml配置文件中，指定goby监听的参数（默认为127.0.0.1:8362及goby/goby）：
```yaml
pocscan:
  goby:
    authUser: goby
    authPass: goby
    api:
      - http://127.0.0.1:8361
```
- 可以列表的方式，指定多个不同的goby服务端。
- 除了本地部署外，可将goby远程部署。
- goby同时只能并发执行一个扫描任务，因此多个任务下worker会查找配置文件中空闲可用的goby，并通过sleep的方式每隔一定间隔测试是否可用。
- 如果goby不可用，goby任务将一直显示执行中，不会被自动结束。

#### 3、worker任务模式

Nemo将任务分为5种类型，worker启动时通过参数-m指定worker执行的任务类型，可以指定一种或多种任务类型；参数分别用1-5，如果为0则表示可执行类型为1-4的任务。
对自定义的任务：-m 5，需要用-w参数指定任务关联的工作空间GUID。

任务类型及数值：
- 1：Active，主动扫描类的任务
- 2：Finger，获取指纹类的任务
- 3：Passive，被动收集信息的任务
- 4：Pocscan，漏洞验证类的任务
- 5：Custom，自定义任务
- 0：同时包含1-4种类型的任务

|  |1:Active|2:Finger|3:Passive|4:Pocscan|  
| --- | --- | --- | --- | --- |
|portscan| √ |  |  |  |          
|batchscan|  √|  |  |  |         
|domainscan| √ |  |  |  |        
|subfinder|  |  |  √|  |         
|subdomainbrute|  |  | √ |  |    
|subdomaincrawler| √ |  |  |  |  
|iplocation|  |  | √ |  |        
|fofa|  |  | √ |  |              
|quake|  |  | √ |  |             
|hunter|  |  | √ |  |            
|xray|  |  |  |  √|              
|nuclei|  |  |  | √ |            
|goby|  |  |  | √ |              
|icpquery|  |  | √ |  |          
|whoisquery|  |  |  √|  |        
|fingerprint|  |  |  |  |       
|xportscan| √ |  |  |  |         
|xonlineapi|  |  | √ |  |        
|xfofa|  |  |√  |  |             
|xquake|  |  | √ |  |            
|xhunter|  |  | √ |  |           
|xdomainscan|  | |  √ |  |       
|xsubfinder|  |  | √  |  |        
|xsubdomainbrute|  |  | √  |  |   
|xsubdomaincralwer|  √ |  |  |  | 
|xfingerprint|  |  √ |  |  |      
|xxray|  |  |  | √  |             
|xnuclei|  |  |  | √  |           
|xgoby|  |  |  | √  |             
|xorgscan|  |  |  | √  |          

Worker默认启动时参数为-m 0，将会执行所有类型（除custom）的任务；分布式部署的vps可以合理分配资源，如专用于扫描类vps：-m 1，被动信息搜索指纹可以同时执行任务：-m 2,3。

如果需要使用自定义任务，需进行以下配置：
- 在Config-自定义任务工作空间配置中，指定工作空间的GUID并保存；格式为：GUID 备注，如：1a0ca919-7960-4067-9981-9abcb4eaa735 172网段；
- 在命令行启动worker时，指定任务模式为-m 5，同时-w参数指定在上一步中配置的工作空间的GUID；
- 在Nemo的IP或Domain列表视图中，切换到第一步配置的工作空间，在新建任务或XScan任务后，只有启动命令为：-m 5 -w 1a0ca919-7960-4067-9981-9abcb4eaa735的worker才会收到任务并执行。


## 分布式部署的典型架构

![nemo_vps](./image/nemo_vps.png)


## Worker使用自定义的socks5代理

为提高Worker部署和扫描的灵活性，在v2.11版本后，Worker的部份任务功能支持sock5代理，包括：
- gogo的端口扫描
- 指纹获取
- 在线API接口
- 子域名任务被动收集（subfinder）和爬虫
- Nuclei漏洞验证

代理设置：Config-配置管理，支持同时配置多个socks5代理地址（地址格式为socks5://user:pass@host:port），多个地址将由worker每次任务时随机选择。

由于获取网站截图时调用的chrome-headless不支持验证功能的socks5代理，因此worker在启动时默认在127.0.0.1:5010地址进行代理转发到设置的socks5地址。

worker可通过命令行参数-np关闭代理功能。如果前端任务指定了代理扫描选项，但未配置socks5地址或worker关闭了代理功能，任务将会由不使用代理直接执行。


## Worker使用全局Socks5代理

在实战的高强度对抗中，使用代理可以降低IP被安设备BAN导致的信息收集不全。由于Nemo的特点是大量使用第三方组件，每个组件对代理的支持的方式和能力都不尽相同，因此建议通过第三方工具和技术，通过全局代理的方式来达到代理Worker的网络流量的目的。

目前从实战中总结来说，经过比较和测试，比较推荐的方式为：
- Windows/MacOS：proxifier+socks5代理
- Linux：iptables+redsocks+socks5代理

proxifier在实战对抗中通过反向代理进行内网渗透的利器，具有GUI界面的易用性、规则配置的灵活性，具体使用方法可参照网上的教程。

**iptables+redsocks的使用方法**

- 安装：

```bash
sudo apt-get install redsocks
```

- 配置：

```bash
配置文件：/etc/redsocks.conf

redsocks下的local_ip, local_port，这两个是本地映射出的tcp端口，默认127.0.0.1:12345，可不用修改。
修改redsocks下的ip, port, type，这三个是远程代理服务器的配置，type默认为socks5，需要修改为实际使用的楼socks代理服务器的地址和端口；如果有认证，需修改login和password。
```
- 重启服务：
```bash
sudo systemctl restart redsocks
```

- 配置iptables规则文件（redsocks.rules）：
```bash
# Transparent SOCKS proxy
# See: http://darkk.net.ru/redsocks/

*nat
:PREROUTING ACCEPT [0:0]
:INPUT ACCEPT [0:0]
:OUTPUT ACCEPT [0:0]
:POSTROUTING ACCEPT [0:0]
:REDSOCKS - [0:0]

# Redirect all output through redsocks
-A OUTPUT -p tcp -j REDSOCKS

# Whitelist LANs and some other reserved addresses.
# https://en.wikipedia.org/wiki/Reserved_IP_addresses#Reserved_IPv4_addresses
-A REDSOCKS -d 0.0.0.0/8 -j RETURN
-A REDSOCKS -d 10.0.0.0/8 -j RETURN
-A REDSOCKS -d 127.0.0.0/8 -j RETURN
-A REDSOCKS -d 169.254.0.0/16 -j RETURN
-A REDSOCKS -d 172.16.0.0/12 -j RETURN
-A REDSOCKS -d 192.168.0.0/16 -j RETURN
-A REDSOCKS -d 224.0.0.0/4 -j RETURN
-A REDSOCKS -d 240.0.0.0/4 -j RETURN

# 这里改为socks5代理服务器的地址
# -A REDSOCKS -d socks5_proxy_ip -j RETURN

# 将Nemo的Server地址，也要加入到白名单中
# -A REDSOCKS -d nemo_server_ip -j RETURN

# Redirect everything else to redsocks port
-A REDSOCKS -p tcp -j REDIRECT --to-ports 12345

COMMIT

```
- 应用规则：
```bash
sudo iptables-restore ./redsocks.rules
```

- 查看规则：
```bash
sudo iptables -t nat -L

Chain REDSOCKS (1 references)
target     prot opt source               destination
RETURN     all  --  anywhere             0.0.0.0/8
RETURN     all  --  anywhere             10.0.0.0/8
RETURN     all  --  anywhere             localhost/8
RETURN     all  --  anywhere             169.254.0.0/16
RETURN     all  --  anywhere             172.16.0.0/12
RETURN     all  --  anywhere             192.168.0.0/16
RETURN     all  --  anywhere             base-address.mcast.net/4
RETURN     all  --  anywhere             240.0.0.0/4
RETURN     all  --  anywhere             x.x.x.x
REDIRECT   tcp  --  anywhere             anywhere             redir ports 12345
```

- 清空规则：
```bash
sudo iptables -t nat -F
```

> 参考文档：https://www.coder4.com/archives/7290

关于使用代理的注意事项 ：
- socks5代理为网络层代理，主要用于TCP协议代理，如果需代理udp，请修改redsocks.rules配置以及iptables规则文件，具体请自行参考文档；
- nmap与masscan的SYN扫描不支持在socks5代理下使用，所以使用代理的worker尽量不要分配active类型的任务（-m 参数为1或0）；如果必须要使用，只能使用nmap的-sT的扫描类型；
- 对于在线API接口的任务（比如FOFA、Hunter、ICP查询等），不建议使用代理功能，防止因源IP导致访问被限；
- 使用代理功能时，建议合理分配worker的任务类型；
- 推荐使用gost实现自已搭建的多个代理的负载均衡；如果要更多的代理池功能，建议购买第三方的代理服务（比如快代理的隧道代理）；
- 以上测试只在UbuntuLTS 22.04中测试稳定运行，在其它linux版本及docker中请自行参考网上文档。