package custom

import (
	"bufio"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type BlackDomain struct {
	blackList []string
}

// NewBlackDomain 创建域名黑名单对象
func NewBlackDomain() *BlackDomain {
	b := BlackDomain{}
	b.loadBlankList()
	return &b
}

// loadBlankList 从配置文件中加载域名黑名单列表
func (b *BlackDomain) loadBlankList() {
	inputFile, err := os.Open(filepath.Join(conf.GetRootPath(), "thirdparty/custom/black_domain.txt"))
	if err != nil {
		return
	}
	scanner := bufio.NewScanner(inputFile)
	for scanner.Scan() {
		text := strings.ToLower(scanner.Text())
		if text == "" || strings.HasPrefix(text, "#") {
			continue
		}
		if strings.HasPrefix(text, ".") == false {
			b.blackList = append(b.blackList, "."+text)
		} else {
			b.blackList = append(b.blackList, text)
		}
	}
	inputFile.Close()
	return
}

// CheckBlack 检查一个域名是否是位于黑名单中
func (b *BlackDomain) CheckBlack(domain string) bool {
	for _, txt := range b.blackList {
		// 生成格式为.qq.com$
		regPattern := strings.ReplaceAll(txt, ".", "\\.") + "$"
		if m, _ := regexp.MatchString(regPattern, domain); m == true {
			return true
		}
	}
	return false
}
