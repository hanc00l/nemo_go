package domainscan

import (
	"fmt"
	"github.com/hanc00l/nemo_go/v3/pkg/conf"
	"github.com/hanc00l/nemo_go/v3/pkg/utils"
	"github.com/joeguo/tldextract"
	"path"
)

type TldExtract struct {
	extract *tldextract.TLDExtract
}

// NewTldExtract 创建TldExtract对象
func NewTldExtract() TldExtract {
	tldCachePath := path.Join(conf.GetRootPath(), "thirdparty/tldextract")
	utils.MakePath(tldCachePath)

	t := TldExtract{}
	t.extract, _ = tldextract.New(path.Join(tldCachePath, "tld.cache"), false)

	return t
}

// ExtractFLD 从url或domain中提取FLD
func (t *TldExtract) ExtractFLD(url string) (fldDomain string) {
	result := t.extract.Extract(url)
	if result == nil {
		return
	}
	if result.Flag == tldextract.Domain {
		fldDomain = fmt.Sprintf("%s.%s", result.Root, result.Tld)
	}
	return
}
