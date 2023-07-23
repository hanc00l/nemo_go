# Nemo常见问题

##v1.0

2023-7-8

### 1、Docker安装后无法使用

请从github的Release下载最新的nemo_linux_amd64.tar；clone或下载的源码，不包括编译后的可执行二进制文件及thirdparty中的第三方二进制程序。

### 2、Docker安装后看不到worker

由于Server与worker的文件同步机制，worker在每次启动时会自动与Server进行文件同步，server同时会监控Server目录下的文件更改并自动发起同步。worker的配置文件conf/worker.yml如果与server目录下的不一致从而导致worker的配置文件被同步覆盖，从而导致worker不能正常与server进行通信。
因此在构建docker前（特别是worker是分布式部署情况下），请确保Server与worker的conf/yml配置文件已正确配置后再进行构建。

### 3、Subfinder增加对第三方API的Key的支持

Nemo集成的subfinder只支持不需要第三方Key的平台的API接口。
如果需要更全面地收集子域名信息，建议申请和配置第三方的API接口的Key，配置文件位于thirdpary/dict/provider-config.yml，具体配置方法请参考[subfinder](https://github.com/projectdiscovery/subfinder)。

### 4、能否支持资产的导出

目前Nemo提供了简单的统计功能（用于将收集的信息用于其它工具），还没有直接将资产导出的功能。

### 5、能否支持Arm平台

目前Nemo支持和测试的平台包括x86-amd64平台的Linux、Macos和Windows。理论上golang支持向更多平台的移植，但依赖的其它工具也需要同步集成到thirdparty目录下，Nemo在调用时会按照运行的平台自动选择匹配的文件，文件的命名格式可参考pkg/utils/fileutil.go：
```golang
// GetThirdpartyBinNameByPlatform 根据当前运行平台及架构，生成指定的文件名称
func GetThirdpartyBinNameByPlatform(binShortName BinShortName) (binPlatformName string) {
	binPlatformName = fmt.Sprintf("%s_%s_%s", binShortName, runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		binPlatformName += ".exe"
	}
	/*
		https://go.dev/doc/install/source#environment
			$GOOS	$GOARCH
			android   arm
			darwin    386
			darwin    amd64
			darwin    arm
			darwin    arm64
			dragonfly amd64
			freebsd   386
			freebsd   amd64
			freebsd   arm
			linux     386
			linux     amd64
			linux     arm
			linux     arm64
			linux     ppc64
			linux     ppc64le
			linux     mips
			linux     mipsle
			linux     mips64
			linux     mips64le
			netbsd    386
			netbsd    amd64
			netbsd    arm
			openbsd   386
			openbsd   amd64
			openbsd   arm
			plan9     386
			plan9     amd64
			solaris   amd64
			windows   386
			windows   amd64
	*/
	return
}
```

如果有其它平台的需求，可以参考进行移植。

### 6、能支持其它漏洞扫描工具吗

Nemo更多的聚焦在资产信息的收集工具和手段，以及为下一步渗透测试、实战攻防作好参考。漏洞的扫描和利用，只是集成专业了相关工具，并且也只是作为资产的一个辅助。目前的代码主要还是满足个人及同事在实际工作中最主要的需求来写的，因能力、精力和时间有限，我们也知道可能无法满足所有人的功能需求。

从实战的角度来讲，漏洞类的扫描（比如dirsearch、Nuclei的POC等），因为具有比较显示的攻击流量特征，在有流量及安全设备的情况下，极容易触发封IP条件，从而导致worker无法进一步的对目标进行资产收集。在具有授权白名单或内网中，使用漏洞扫描工具是具有可行性的。因此基于实战考虑，漏洞扫描请大家在使用时根据情况慎重使用。

### 7、重启server后，用户名密码正确有时无法正确登录

为了提高安全性，Nemo在登录时使用了RSA加密用户名和密码，并且在每次server启动时会重新生成公钥和私钥对，公钥写入login.js文件，私钥保存在内存中。

在确保用户名、密码及验证正确的情况下，如果无法登录到Nemo，原因是由于浏览器缓存的存在导致js文件中的私钥没有更新。可手工清除浏览器的缓存文件即可。

### 8、在VPS上后台运行的server与worker的推荐方式

```bash
#在命令行下运行：
screen ./server_linux_amd64
#或者
screen ./daemon_worker_linux_amd64

#server或worker的进程不会因为关掉终端而被kill，通过执行命令可以恢复到命令行
screen -r

#如果没有screen命令，可以通过apt install screen安装；详细使用方式请网上搜索参考文档。
```