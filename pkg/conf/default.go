package conf

// 对参数配置文件的定义和说明

var defaultYamlConfig = []byte(`
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
  host: 127.0.0.1
  port: 3306
  name: nemo
  username: nemo
  password: nemo2020
# 消息中间件，用于分布式任务
rabbitmq:
  host: localhost
  port: 5672
  username: guest
  password: guest
# 使用到的api的key，如果为空则该功能将不能正常使用
api:
  fofa:
    name: xxx
    key: xxx
  icp:
    name: chinaz
    key: xxx
# 端口扫描任务的默认参数
portscan:
  ping: false
  port: --top-ports 1000
  rate: 1000
  tech: -sS
  cmdbin: masscan
# 域名任务用到的参数
domainscan:
  # dns服务器地址文件
  resolver: resolver.txt
  # 子域名爆破用到的字典文件
  wordlist: subnames.txt
  # 子域名爆破时massdns的线程时，带宽为1M的VPS建议600左右；带宽足够最大值可到10000
  massdnsThreads: 600
# pocc验证时用到的一些参数
pocscan:
  xray:
    # xray的poc文件所在位置
    pocPath: thirdparty/xray/xray/pocs
    # 当xray可执行文件不存在时，可从github上下载，这里指定下载的版本，可保持与github版本的一致性
    latest: 1.7.1
  pocsuite:
    # pocsuite的poc文件所在位置
    pocPath: thirdparty/pocsuite/some_pocsuite
    # 调用Pocsuite验证时，每个poc同时执行的线程数
    threads: 10
`)
