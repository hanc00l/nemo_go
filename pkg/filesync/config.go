package filesync

import "strings"

var syncFileList = []string{"worker_darwin_amd64", "version.txt", "conf", "thirdparty"}

// checkFileIsSyncWhileList 同步文件的白名单校验
func checkFileIsSyncWhileList(filePathName string) bool {
	if strings.Contains(filePathName, "..") {
		return false
	}
	for _, f := range syncFileList {
		if strings.HasPrefix(filePathName, f) {
			return true
		}
	}

	return false
}
