## Nemo安装

**Tested on ubuntu18.04 LTS**

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

  mysql默认只监听127.0.0.1，需修改/etc/mysql/mysql.conf.d的bind-address，供worker远程连接（如果不需要分布式则不需要该配置）

  ```ini
  # Default Homebrew MySQL server config
  [mysqld]
  # Only allow connections from localhost
  # bind-address = 127.0.0.1
  bind-address = 0.0.0.0
  ```

  由于mysql5.7新版本默认安装后不能使用root和空密码在本机登录，系统自动生成的用户名密码位于/etc/mysql/debian.cnf，请替换并使用配置文件中的user和password：

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
      && mysql  -u debian-sys-maint -p nemo < docs/nemo.sql 
  ```

- **配置rabbitmq**：增加rabbitmq用户和密码

  ```bash
  sudo rabbitmqctl add_user nemo nemo2020
  sudo rabbitmqctl set_permissions -p "/" nemo ".*" ".*" ".*"
  ```

- 配置文件

  **conf/config.yml**

  ```yaml
  web:
    # host,port: server监听地址；worker用于keepalive和upload的地址
    host: 0.0.0.0
    port: 5000
    # 登录用户名和密码
    username: nemo
    password: 648ce596dba3b408b523d3d1189b15070123456789abcdef
    # keepalive和upload的地址使用的密钥
    encryptKey: ZduibTKhcbb6Pi8W
    # Server保存screenshot的位置，同时需要修改conf/app.conf里对应表staticdir映射
    screenshotPath: /tmp/screenshot
  # server和worker访问的数据库
  database:
    host: localhost
    port: 3306
    name: nemo
    username: nemo
    password: nemo2020
  # 消息中间件，用于分布式任务
  rabbitmq:
    host: localhost
    port: 5672
    username: nemo
    password: nemo2020
  ```

  **conf/app.conf**

  ```yaml
  #screenshot默认保存位置，与config.yml保持一致
  staticdir = static:web/static screenshot:/tmp/screenshot
  ```



### 2、Worker

- **创建安装目录并解压tar包**

  ```bash
  mkdir nemo;tar xvf nemo_linux_amd64.tar -C nemo;cd nemo
  ```

- **安装环境和依赖**

  ```bash
  sudo apt-get update \
      && sudo apt-get install vim git python3-pip python3-setuptools \
      nmap whatweb masscan chromium-browser --fix-missing
  ```


- **nmap&masscan：**因为nmap、masscan的SYN扫描需要root权限，为避免使用sudo，设置root权限的suid

  ```bash
  cd /usr/bin \
      && sudo chown root nmap masscan && sudo chmod u+s nmap masscan 
  ```

- **python3与pocsuite**

  ```
  python3 -m pip install -U pip -i https://mirrors.aliyun.com/pypi/simple/ --user \
      && python3 -m pip install -U requests pocsuite3 -i https://mirrors.aliyun.com/pypi/simple/
  ```

- **配置文件**

  **conf/config.yml** ：

  ```yaml
  web:
    # host,port: server监听地址；worker用于keepalive和upload的地址
    host: 172.16.80.1
    port: 5000
    # keepalive和upload的地址使用的密钥
    encryptKey: ZduibTKhcbb6Pi8W
  # server和worker访问的数据库
  database:
    host: 172.16.80.1
    port: 3306
    name: nemo
    username: nemo
    password: nemo2020
  # 消息中间件，用于分布式任务
  rabbitmq:
    host: 172.16.80.1
    port: 5672
    username: nemo
    password: nemo2020
  # 使用到的api的key，如果为空则该功能将不能正常使用
  api:
    fofa:
      name: xxx
      key: xxx
    icp:
      name: chinaz
      key: xxx
  ```

  