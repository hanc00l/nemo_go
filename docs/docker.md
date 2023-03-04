## Docker使用Nemo



按使用需求不同，Docker有两种使用方式：

- 单Docker（适用临时使用）
- 分别构建Server与Worker的Docker（适用于长期使用、动态调整Worker数量及分布式部署Worker）



### 一、单Docker使用 

- **下载release的nemo_linux_amd64.tar**

  ```bash
  mkdir nemo;tar xvf nemo_linux_amd64.tar -C nemo;cd nemo
  docker-compose up -d
  ```



### 二、分别构建Server与Worker的Docker

#### 0 、下载release的nemo_linux_amd64.tar

  ```bash
  mkdir nemo;tar xvf nemo_linux_amd64.tar -C nemo;cd nemo
  ```

#### 1、Server

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

#### 2、Worker

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
#### 3、关于文件同步

#### 由于server与worker的文件自动同步机制，worker与server的conf/worker.yml配置应先确保一致后，再构建docker镜像，否则可能会导致worker的worker.yml被不正确同步。
