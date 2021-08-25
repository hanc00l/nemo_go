## 分布式运行方式配置

### 0、分布式配置

1台主服务器+N台任务服务器

- **主服务器：（以IP为172.16.80.1为例）**

  组件：mysql + rabbitmq

- **任务服务器（能够访问主服务器的mysql和rabbitmq）**

### **1、rabbitmq**

- 设置监听的IP地址（供worker远程访问），修改/usr/local/Cellar/rabbitmq/{VERSION}/sbin/rabbitmq-env（ubuntu18：/usr/lib/rabbitmq/bin/rabbitmq-env），增加

  ```
  NODE_IP_ADDRESS=172.16.80.1 (rabbitmq所在服务器地址，供worker远程连接，如果不需要分布式则不需要该配置)
  ```

- 增加rabbitmq用户和密码

  ```
  rabbitmqctl add_user nemo nemo2020
  rabbitmqctl set_permissions -p "/" nemo ".*" ".*" ".*"
  ```

### **2、mysql**

- mysql的bind-address（brew 安装默认是127.0.0.1，创建~/.my.cnf文件并设置bind-address；如果不需要分布式则不需要该配置）

  ```
  # Default Homebrew MySQL server config
  [mysqld]
  # Only allow connections from localhost
  # bind-address = 127.0.0.1
  bind-address = 172.16.80.1
  ```


- 创建数据库

  ```
  CREATE DATABASE `nemo` DEFAULT CHARACTER SET utf8mb4;
  ```

- 导入nemo.sql

  ```
  mysql -u root nemo < docker/mysql/initdb.d/nemo.sql
  ```

- 创建用户并授权

  ```
  CREATE USER 'nemo'@'%' IDENTIFIED BY 'nemo2020';
  GRANT ALL PRIVILEGES ON nemo.* TO 'nemo'@'%';
  FLUSH PRIVILEGES;
  ```

### 3、conf/config.yaml

  ```
  web:
  host: 0.0.0.0
  port: 5000
  username: nemo
  password: 648ce596dba3b408b523d3d1189b15070123456789abcdef
  encryptKey: ZduibTKhcbb6Pi8W
  screenshotPath: /tmp/screenshot
database:
  host: 172.16.80.1
  port: 3306
  name: nemo
  username: nemo
  password: nemo2020
rabbitmq:
  host: 172.16.80.1
  port: 5672
  username: nemo
  password: nemo2020
api:
  fofa:
    name:
    key:
  icp:
    name: chinaz
    key: 
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
pocscan:
  xray:
    pocPath: thirdparty/xray/xray/pocs
    latest: 1.7.1
  pocsuite:
    pocPath: thirdparty/pocsuite/some_pocsuite
    threads: 10
  ```



## 运行

**主服务器**

1. 启动mysql和rabbitmq

2. 启动server

   ```
   ./server_darwin_amd64
   ```
   

**分布式任务**

1. 启动worker

   ```bash
   ./worker_linux_amd64
   ```

