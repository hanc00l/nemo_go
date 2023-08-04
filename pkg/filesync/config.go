package filesync

import (
	"strings"
)

// syncFileList 需要同步的文件白名单
var syncFileList = []string{"worker_linux_amd64", "version.txt", "conf", "thirdparty"}

// syncFileBlackList 不需要、禁止同步的文件黑名单
var syncFileBlackList = []string{"thirdparty/massdns/temp", "conf/server.yml", "conf/app.conf"}

var (
	// TLSEnabled 是否启用TLS加密
	TLSEnabled  bool
	TLSCertFile string
	TLSKeyFile  string
)

// checkFileIsSyncWhileList 同步文件的白名单校验
func checkFileIsSyncWhileList(filePathName string) bool {
	if strings.Contains(filePathName, "..") {
		return false
	}
	for _, f := range syncFileBlackList {
		if strings.HasPrefix(filePathName, f) {
			return false
		}
	}
	for _, f := range syncFileList {
		if strings.HasPrefix(filePathName, f) {
			return true
		}
	}

	return false
}

// checkFileIsMonitorWhileList 自动监测文件的白名单校验
func checkFileIsMonitorWhileList(filePathName string) bool {
	// 将配置文件所在目录加入到自动检测中
	if filePathName == "conf" {
		return true
	}
	return checkFileIsSyncWhileList(filePathName)
}
