# Nemo安装手册

## v1.0

2023-7-8


Nemo分为**Server**端和**Worker**端两部份。Server提供Http访问、API接口、RPC接口、消息中间件服务以及文件同步接口。Worker是通过消息中间件从Worker接收任务并执行，通过RPC接口上传任务的结果，并通过文件同步接口接收Server的文件。

**Server需要安装的组件：**
- MySQL 
- Rabbitmq

**Worker需要安装的组件：**
- Nmap
- Masscan
- Chrome

**Worker其它使用工具已集成到thirdparty目录中：**
- Observer_Ward
- httpx
- subfinder
- massdns
- nuclei
- xray
- goby

Nemo目前可运行在**x86-AMD**平台的MacOS、Linux及Windows平台，其它平台目前未做测试。本文档均以Ubuntu Linux版本进行介绍安装步骤，在Mac、Windows平台及其它请参考相应的安装步骤。

**Server运行后，默认会开启以下端口：**
- 5000：Http，web界面
- 5001：RPC接口，用于worker将任务结果保存到server
- 5002：文件同步，在server与worker之间同步文件
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

推荐使用Ubuntu18.04/20.04 LTS版本安装（以下的安装方法均是在Ubuntu上进行）。根据大家的测试和反馈，也支持其它的Linux发行版本的安装（比如centos），只是各组件的步骤和细节稍有不同，可以参照进行相应的调整。

### 一、Server 

- **创建安装目录并解压tar包**

  ```bash
  mkdir nemo;tar xvf nemo_linux_amd64.tar -C nemo;cd nemo
  ```

- **安装环境和依赖**

  ```bash
  sudo apt-get update \
    && sudo apt-get install vim \
    mysql-server rabbitmq-server --fix-missing
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
    && sudo apt-get install vim nmap masscan --fix-missing
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

## 运行

 ### 一. Server

```bash
./server_linux_amd64
```
可选参数：
```bash
  -nf
    	disable file sync
  -nr
    	disable rpc
```


### 二. Worker

```bash
./daemon_worker_linux_amd64
```
1、可选参数：
```bash
  -c int
    	concurrent number of tasks (default 3)
  -ma string
    	manual file sync auth key
  -mh string
    	manual file sync host address
  -mp string
    	manual file sync port,default is 5002
  -nf
    	disable file sync
  -p int
    	worker performance,default is autodetect (0: autodetect,1:high,2:normal)
```
- -c参数：worker并发的任务数量，默认为3。
- -mh、mp及ma参数：Server文件同步的host、port及authKey，如果同时指定这三个参数，将执行文件同步功能，从server同步文件到worker。
- nf参数：指定参数则禁用文件同步功能。
- -p参数：worker的的性能模式，默认为0；根据worker的性能模式（1：高性能，2：普通）不同，在任务的并发线程数会有所区别；参数为0则自动判断，判断规则为CPU>=4核、内存>=4G为高性能模式。

2、Goby的服务端部署模式，需在thirdparty/goby目录下运行：（Docker已自动运行）
```bash
 ./goby-cmd-linux -mode api -bind 127.0.0.1:8362 -apiauth goby:goby
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
