package custom

import (
	"bufio"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type BlackDomain struct {
	blackListMap map[string]struct{}
}

// NewBlackDomain 创建域名黑名单对象
func NewBlackDomain() *BlackDomain {
	b := BlackDomain{}
	b.loadBlankList()
	return &b
}

// loadBlankList 从配置文件中加载域名黑名单列表
func (b *BlackDomain) loadBlankList() {
	b.blackListMap = make(map[string]struct{})

	inputFile, err := os.Open(filepath.Join(conf.GetRootPath(), "thirdparty/custom/black_domain.txt"))
	if err != nil {
		logging.RuntimeLog.Error(err)
		logging.CLILog.Error(err)
		return
	}
	scanner := bufio.NewScanner(inputFile)
	for scanner.Scan() {
		text := strings.ToLower(scanner.Text())
		if text == "" || strings.HasPrefix(text, "#") {
			continue
		}
		if strings.HasPrefix(text, ".") == false {
			b.blackListMap["."+text] = struct{}{}
		} else {
			b.blackListMap[text] = struct{}{}
		}
	}
	inputFile.Close()
	return
}

// CheckBlack 检查一个域名是否是位于黑名单中
func (b *BlackDomain) CheckBlack(domain string) bool {
	for txt := range b.blackListMap {
		// 生成格式为.qq.com$
		regPattern := strings.ReplaceAll(txt, ".", "\\.") + "$"
		if m, _ := regexp.MatchString(regPattern, domain); m == true {
			return true
		}
	}
	return false
}

// AppendBlackDomain 增加一个黑名单
func (b *BlackDomain) AppendBlackDomain(domain string) error {
	f, err := os.OpenFile(filepath.Join(conf.GetRootPath(), "thirdparty/custom/black_domain.txt"), os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		logging.RuntimeLog.Error(err)
		logging.CLILog.Error(err)
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	w.WriteString("\n")
	w.WriteString(domain)
	w.Flush()

	return nil
}
