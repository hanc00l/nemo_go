package custom

import (
	"bufio"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/utils"
	"os"
	"path/filepath"
	"strings"
)

type BlackIP struct {
	blackIPMap map[string]struct{}
}

func NewBlackIP() *BlackIP {
	b := BlackIP{}
	b.loadBlackIPList()
	return &b
}

// loadBlackIPList 加载黑名单
func (b *BlackIP) loadBlackIPList() {
	b.blackIPMap = make(map[string]struct{})
	inputFile, err := os.Open(filepath.Join(conf.GetRootPath(), "thirdparty/custom/black_ip.txt"))
	if err != nil {
		return
	}
	scanner := bufio.NewScanner(inputFile)
	for scanner.Scan() {
		text := strings.ToLower(scanner.Text())
		if text == "" || strings.HasPrefix(text, "#") {
			continue
		}
		ipAndComment := strings.Split(text, " ")
		ips := utils.ParseIP(ipAndComment[0])
		for _, ip := range ips {
			if _, ok := b.blackIPMap[ip]; !ok {
				b.blackIPMap[ip] = struct{}{}
			}
		}
	}
	inputFile.Close()
	return
}

// CheckBlack 检查一个IP是否是位于黑名单中
func (b *BlackIP) CheckBlack(ip string) bool {
	_, existed := b.blackIPMap[ip]
	return existed
}
