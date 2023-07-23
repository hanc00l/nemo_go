# Nemo快速上手

## v1.0

2023-7-8

## 一、Docker安装

**下载release的nemo_linux_amd64.tar**

  ```bash
  mkdir nemo;tar xvf nemo_linux_amd64.tar -C nemo;cd nemo
  docker-compose up -d
  ```
  构建完mysql、rabbitmq和nemo后，在5000端口会映射nemo的web访问接口。
  ```bash
Pulling mysql (mysql:5.7)...
Pulling rabbitmq (rabbitmq:3-management-alpine)...
Building web
[+] Building 363.7s (11/11) FINISHED
 => [internal] load build definition from Dockerfile                                                                                                                                                                                                                                                                                                                   0.0s
 => => transferring dockerfile: 37B                                                                                                                                                                                                                                                                                                                                    0.0s
 => [internal] load .dockerignore                                                                                                                                                                                                                                                                                                                                      0.0s
 => => transferring context: 2B                                                                                                                                                                                                                                                                                                                                        0.0s
 => [internal] load metadata for docker.io/library/ubuntu:22.04                                                                                                                                                                                                                                                                                                       24.5s
 => [internal] load build context                                                                                                                                                                                                                                                                                                                                      1.6s
 => => transferring context: 117.99MB                                                                                                                                                                                                                                                                                                                                  1.5s
 => [1/6] FROM docker.io/library/ubuntu:22.04@sha256:0bced47fffa3361afa981854fcabcd4577cd43cebbb808cea2b1f33a3dd7f508                                                                                                                                                                                                                                                 36.3s
 => => resolve docker.io/library/ubuntu:22.04@sha256:0bced47fffa3361afa981854fcabcd4577cd43cebbb808cea2b1f33a3dd7f508                                                                                                                                                                                                                                                  8.8s
 => => sha256:0bced47fffa3361afa981854fcabcd4577cd43cebbb808cea2b1f33a3dd7f508 1.13kB / 1.13kB                                                                                                                                                                                                                                                                         0.0s
 => => sha256:b060fffe8e1561c9c3e6dea6db487b900100fc26830b9ea2ec966c151ab4c020 424B / 424B                                                                                                                                                                                                                                                                             0.0s
 => => sha256:5a81c4b8502e4979e75bd8f91343b95b0d695ab67f241dbed0d1530a35bde1eb 2.30kB / 2.30kB                                                                                                                                                                                                                                                                         0.0s
 => => sha256:3153aa388d026c26a2235e1ed0163e350e451f41a8a313e1804d7e1afb857ab4 29.53MB / 29.53MB                                                                                                                                                                                                                                                                      35.0s
 => => extracting sha256:3153aa388d026c26a2235e1ed0163e350e451f41a8a313e1804d7e1afb857ab4                                                                                                                                                                                                                                                                              1.1s
 => [2/6] RUN set -x     && sed -i 's/archive.ubuntu.com/mirrors.ustc.edu.cn/g' /etc/apt/sources.list     && sed -i 's/security.ubuntu.com/mirrors.ustc.edu.cn/g' /etc/apt/sources.list     && apt update     && apt install -y tzdata     && ln -fs /usr/share/zoneinfo/Asia/Shanghai /etc/localtime     && echo Asia/Shanghai > /etc/timezone     && dpkg-reconfig  11.7s
 => [3/6] RUN set -x     && apt update     && apt install -y     wget curl vim net-tools  iputils-ping git unzip     nmap masscan  --fix-missing                                                                                                                                                                                                                      40.4s
 => [4/6] RUN set -x     && curl -LO https://dl.google.com/linux/direct/google-chrome-stable_current_amd64.deb     && apt install -y ./google-chrome-stable_current_amd64.deb --fix-missing     && rm google-chrome-stable_current_amd64.deb                                                                                                                         235.7s
 => [5/6] COPY . /opt/nemo                                                                                                                                                                                                                                                                                                                                             1.9s
 => [6/6] RUN set -x     && sed -i 's/host: 127.0.0.1/host: mysql/g' /opt/nemo/conf/server.yml     && sed -i 's/host: localhost/host: rabbitmq/g' /opt/nemo/conf/server.yml     && sed -i 's/host: localhost/host: rabbitmq/g' /opt/nemo/conf/worker.yml                                                                                                               0.2s
 => exporting to image                                                                                                                                                                                                                                                                                                                                                 4.2s
 => => exporting layers                                                                                                                                                                                                                                                                                                                                                4.2s
 => => writing image sha256:11413db698b301278b92b1df873fce33c67dfb2613c9cc904576747fd877053e                                                                                                                                                                                                                                                                           0.0s
 => => naming to docker.io/library/nemo:v2                                                                                                                                                                                                                                                                                                                             0.0s

Use 'docker scan' to run Snyk tests against images to find vulnerabilities and learn how to fix them
WARNING: Image for service web was built because it did not already exist. To rebuild this image you must use `docker-compose build` or `docker-compose up --build`.
Creating mysql    ... done
Creating rabbitmq ... done
Creating nemo     ... done
  ```
## 二、第三方API配置

Nemo使用的工具大部份已经集成在thirdparty目录中了。但为了更好的利用第三方提供的资源和全面地搜索资产信息，以下使用的技术手段需要配置已获得的API Token：
- FOFA
- Hunter
- Quake
- Chinaz ICP备案查询

同时为了更好地使用Nemo的任务完成的消息通知机制，建议配置以下一个或多个通知平台的Token：
- Server酱
- 钉钉群机器人
- 飞书群机器人

以上API配置可以登录Nemo后，在“Config-配置管理”中进行设置，并勾选默认要使用的API接口。

## 三、使用

+ 通过http://ip:5000，默认用户名和密码均为nemo；登录后建议第一时间在“Config”-“配置管理”更改密码密码。
+ 在导航栏点击IP或Domain，进入资源管理的列表视图；
+ 点“XScan”，在弹出的窗口中，“Targets”输入任务的IP或Domain，多个目标分多行；如果IP任务，指定要扫描的IP端口；勾选“默认指纹识别”选项；如果已配置了Fofa等API接口，可勾选“在线资产平台”；
+ 如果需要进行漏洞扫描，选择要一个或多个使用的扫描工具及使用的POC；
+ 点击“执行”后将生成待执行的主任务；如果需要周期性地执行，选中“定时任务”，并设置定时执行规则；
+ 在导航栏的“TaskRun”，可查看任务的执行状态。
