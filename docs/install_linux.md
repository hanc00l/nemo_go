## Nemo安装

**Tested on ubuntu18.04/20.04 LTS**

### 1、Server 

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

  导入nemo.sql：

  ```bash
  mysql -u debian-sys-maint -p -e 'CREATE DATABASE `nemo` DEFAULT CHARACTER SET utf8mb4;' \
      && mysql -u debian-sys-maint -p -e 'CREATE USER "nemo"@"%" IDENTIFIED BY "nemo2020";GRANT ALL PRIVILEGES ON nemo.* TO "nemo"@"%";FLUSH PRIVILEGES;' \
      && mysql  -u debian-sys-maint -p nemo < docker/mysql/initdb.d/nemo.sql 
  ```

- **配置rabbitmq**：增加rabbitmq用户和密码

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
    # 登录用户名密码
    username: nemo
    password: 648ce596dba3b408b523d3d1189b15070123456789abcdef
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
  # 数据库配置
  database:
    host: 127.0.0.1
    port: 3306
    name: nemo
    username: nemo
    password: nemo2020
  # 消息中间件配置
  rabbitmq: 
    host: localhost
    port: 5672
    username: guest
    password: guest
  ```

  
    **重要：记得要修改默认的RPC authKey和Rabbitmq消息中间件密码。**
  
    **conf/app.conf：**
  
    ```yaml
    # beego配置文件不区别大小写
    appname = nemo
    runmode = prod
    # web映射的目录，static请勿修改；webfiles需和server.yml保持一致
    staticdir = static:web/static webfiles:/tmp/webfiles
    viewspath = web/views
    accesslogs = true
    filelinenum = false
    CopyRequestBody = true
    # beego v2.0.3后在配置文件里启用session，在代码里死活不行，暂不明确原因
    SessionOn = true
    SessionName = "sessionID"
    ```



### 2、Worker

- **创建安装目录并解压tar包**

  ```bash
  mkdir nemo;tar xvf worker_linux_amd64.tar -C nemo;cd nemo
  ```

- **安装环境和依赖**

  ```bash
  sudo apt-get update \
      && sudo apt-get install vim git python3-pip python3-setuptools \
      nmap masscan --fix-missing
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
   # RPC 调用的server监听地址和端口、auth
  rpc:
    host: 127.0.0.1
    port: 5001
    authKey: ZduibTKhcbb6Pi8W
  # 文件同步
  fileSync:
    host: 127.0.0.1
    port: 5002
    authKey: ZduibTKhcbb6Pi8W
  # 消息中间件
  rabbitmq: 
    host: localhost
    port: 5672
    username: guest
    password: guest
  # 使用的API接口用户、密码，如果为空则该api不无使用
  api:
    fofa:
      name:
      key:
    icp:
      name: chinaz
      key:
    quake:
      key:
    hunter:
      key:
  # 任务使用的参数
  portscan:
    ping: false
    port: --top-ports 1000
    rate: 1000
    tech: -sS
    cmdbin: masscan
  domainscan:
    resolver: resolver.txt
    wordlist: subnames.txt
    massdnsThreads: 600
    providerConfig: provider-config.yml
  pocscan:
    xray:
      pocPath: thirdparty/xray/xray/pocs
      latest: 1.8.2
    nuclei:
      pocPath: thirdparty/nuclei/nuclei-templates
      threads: 25
  ```

## 运行

 ### 1. web app

   ```
./server_linux_amd64
   ```

### 2. worker

   ```bash
./daemon_worker_linux_amd64
   ```

