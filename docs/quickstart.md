# Nemo快速上手

## v1.0

2023-7-8

## 一、Docker安装Nemo

**下载release的nemo_linux_amd64.tar**

  ```bash
  mkdir nemo;tar xvf nemo_linux_amd64.tar -C nemo;cd nemo
  docker-compose up -d
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

## 三、使用Nemo

+ 通过http://ip:5000，默认用户名和密码均为nemo；登录后建议第一时间在“Config”-“配置管理”更改密码密码。
+ 在导航栏点击IP或Domain，进入资源管理的列表视图；
+ 点“XScan”，在弹出的窗口中，“Targets”输入任务的IP或Domain，多个目标分多行；如果IP任务，指定要扫描的IP端口；勾选“默认指纹识别”选项；如果已配置了Fofa等API接口，可勾选“在线资产平台”；
+ 如果需要进行漏洞扫描，选择要一个或多个使用的扫描工具及使用的POC；
+ 点击“执行”后将生成待执行的主任务；如果需要周期性地执行，选中“定时任务”，并设置定时执行规则；
+ 在导航栏的“TaskRun”，可查看任务的执行状态。
