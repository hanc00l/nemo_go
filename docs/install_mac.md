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

### 4、conf/server.yml（根据实际情况修改）

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

### 5、conf/worker.yml

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
  pocsuite:
    pocPath: thirdparty/pocsuite/some_pocsuite
    threads: 10
  nuclei:
    pocPath: thirdparty/nuclei/nuclei-templates
    threads: 25
```

### 6、安装pocsuite3 （可选）

  ```
pip3 install pocsuite3
  ```



## 运行

 ### 1.mysql和rabbitmq

   ```
   brew services run mysql@5.7
   brew services run rabbitmq
   ```

### 2. web app

   ```
./server_darwin_amd64
   ```

### 2. worker

   ```bash
   ./daemon_worker_darwin_amd64
   ```

