## MacOS安装与配置

### **0、Unpack Release Package**

  ```
mkdir nemo;tar xvf nemo_darwin_amd64.tar -C nemo;cd nemo
  ```

### **1、rabbitmq**

  ```
brew install rabbitmq

配置rabbitmq：增加rabbitmq用户和密码
sudo rabbitmqctl add_user nemo nemo2020
sudo rabbitmqctl set_permissions -p "/" nemo ".*" ".*" ".*"
  ```

### **2、mysql**

```
brew install mysql@5.7
```


- 创建数据库

  ```
  brew services run mysql@5.7
  mysql -u root
  	>CREATE DATABASE `nemo` DEFAULT CHARACTER SET utf8mb4;
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

### 3、nmap&masscan

```
brew install nmap masscan
# 因为nmap、masscan的SYN扫描需要root权限，为避免使用sudo，设置root权限的suid
cd /usr/local/Cellar/nmap/7.80_1/bin
sudo chown root nmap
sudo chmod u+s nmap
cd /usr/local/Cellar/masscan/1.0.5/bin
sudo chown root masscan
sudo chmod u+s masscan
```

### 4、whatweb（可选）

```
git clone https://github.com/urbanadventurer/WhatWeb
cd WhatWeb
# whatwebf需要编译和安装ruby，通过make install自动安装相关的ruby依赖
make install
```

### 5、conf/server.yml（根据实际情况修改）

```yaml
web:
  host: 0.0.0.0
  port: 5000
  username: nemo
  password: 648ce596dba3b408b523d3d1189b15070123456789abcdef
  screenshotPath: /tmp/screenshot
  taskresultPath: /tmp/taskresult
rpc:
  host: 0.0.0.0
  port: 5001
  authKey: ZduibTKhcbb6Pi8W
database:
  host: 127.0.0.1
  port: 3306
  name: nemo
  username: nemo
  password: nemo2020
rabbitmq:
  host: localhost
  port: 5672
  username: guest
  password: guest
```

### 6、conf/worker.yml

```yaml
rpc:
  host: 0.0.0.0
  port: 5001
  authKey: ZduibTKhcbb6Pi8W
rabbitmq:
  host: localhost
  port: 5672
  username: guest
  password: guest
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
    latest: 1.8.2
  pocsuite:
    pocPath: thirdparty/pocsuite/some_pocsuite
    threads: 10
```

### 7、安装pocsuite3 （可选）

  ```
pip3 install pocsuite3
  ```

### 

## 运行

 ### 1.mysql和rabbitmq

   ```
   brew services run mysql@5.7
   brew services run rabbitmq
   ```

### 2. worker

   ```bash
   ./worker_darwin_amd64
   ```

### 3. web app

   ```
   ./server_darwin_amd64
   ```

