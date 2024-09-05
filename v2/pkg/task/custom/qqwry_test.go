package custom

import (
	"github.com/hanc00l/nemo_go/v2/pkg/conf"
	"path/filepath"
	"testing"
)

func TestDownloadQqwry(t *testing.T) {
	IPData.FilePath = filepath.Join(conf.GetRootPath(), "thirdparty/qqwry/qqwry.dat")
	res := IPData.InitIPData()

	if v, ok := res.(error); ok {
		t.Log(v.Error())
	} else {
		t.Logf("纯真IP库加载完成,共加载:%d 条 Domain 记录", IPData.IPNum)
	}
}
