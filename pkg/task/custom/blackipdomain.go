package custom

import (
	"bufio"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/utils"
	"net/netip"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type BlackTarget interface {
	GetBlackFileName() string
	LoadBlackTargetList() error
	CheckBlackTarget(target string) bool
}

type BlackCheckType int

const (
	CheckIP BlackCheckType = iota
	CheckDomain
	CheckAll
)

type IPTarget struct {
	Ipv4BlackMapList map[string]struct{}
	Ipv6BlackMapList []netip.Prefix
}

type DomainTarget struct {
	BlackMapList map[string]struct{}
}

type BlackTargetCheck struct {
	checkType       BlackCheckType
	blackTargetList map[BlackCheckType]BlackTarget
}

// NewBlackTargetCheck 创建黑名单检查
func NewBlackTargetCheck(checkType BlackCheckType) *BlackTargetCheck {
	c := &BlackTargetCheck{checkType: checkType, blackTargetList: make(map[BlackCheckType]BlackTarget)}

	if checkType == CheckAll || checkType == CheckIP {
		ipt := new(IPTarget)
		ipt.LoadBlackTargetList()
		c.blackTargetList[CheckIP] = ipt
	}
	if checkType == CheckAll || checkType == CheckDomain {
		dt := new(DomainTarget)
		dt.LoadBlackTargetList()
		c.blackTargetList[CheckDomain] = dt
	}

	return c
}

// CheckBlack 检查黑名单
func (c *BlackTargetCheck) CheckBlack(target string) bool {
	domain := strings.TrimSpace(target)
	if domain == "" {
		return false
	}
	for _, t := range c.blackTargetList {
		if t.CheckBlackTarget(domain) {
			return true
		}
	}

	return false
}

// AppendBlackTarget 追加黑名单
func (c *BlackTargetCheck) AppendBlackTarget(target string) error {
	f, err := os.OpenFile(filepath.Join(conf.GetRootPath(), "thirdparty/custom", c.GetBlackFileName()), os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		logging.RuntimeLog.Error(err)
		logging.CLILog.Error(err)
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	w.WriteString("\n")
	w.WriteString(target)
	w.Flush()

	return nil
}

// GetBlackFileName 获取黑名单文件名
func (c *BlackTargetCheck) GetBlackFileName() string {
	if c.checkType == CheckIP || c.checkType == CheckDomain {
		return c.blackTargetList[c.checkType].GetBlackFileName()
	}
	return "black_error.txt"
}

// GetBlackFileName 获取黑名单文件名
func (t *IPTarget) GetBlackFileName() string {
	return "black_ip.txt"
}

// LoadBlackTargetList 加载黑名单列表
func (t *IPTarget) LoadBlackTargetList() error {
	t.Ipv4BlackMapList = make(map[string]struct{})

	inputFile, err := os.Open(filepath.Join(conf.GetRootPath(), "thirdparty/custom/", t.GetBlackFileName()))
	if err != nil {
		logging.RuntimeLog.Error(err)
		logging.CLILog.Error(err)
		return err
	}
	scanner := bufio.NewScanner(inputFile)
	for scanner.Scan() {
		text := strings.ToLower(scanner.Text())
		if text == "" || strings.HasPrefix(text, "#") {
			continue
		}
		ipAndComment := strings.Split(text, " ")
		if utils.CheckIPV4(ipAndComment[0]) || utils.CheckIPV4Subnet(ipAndComment[0]) {
			ips := utils.ParseIP(ipAndComment[0])
			for _, ip := range ips {
				if _, ok := t.Ipv4BlackMapList[ip]; !ok {
					t.Ipv4BlackMapList[ip] = struct{}{}
				}
			}
		} else if utils.CheckIPV6(ipAndComment[0]) || utils.CheckIPV6Subnet(ipAndComment[0]) {
			ipv6Prefix, err := netip.ParsePrefix(ipAndComment[0])
			if err != nil {
				continue
			}
			t.Ipv6BlackMapList = append(t.Ipv6BlackMapList, ipv6Prefix)
		}
	}
	inputFile.Close()

	return nil
}

// CheckBlackTarget 检查黑名单
func (t *IPTarget) CheckBlackTarget(target string) bool {
	if utils.CheckIPV4(target) {
		_, existed := t.Ipv4BlackMapList[target]
		return existed
	}
	if utils.CheckIPV6(target) {
		for _, ipv6Prefix := range t.Ipv6BlackMapList {
			ipv6, err := netip.ParseAddr(target)
			if err == nil && ipv6Prefix.Contains(ipv6) {
				return true
			}
		}
	}
	return false
}

// GetBlackFileName 获取黑名单文件名
func (t *DomainTarget) GetBlackFileName() string {
	return "black_domain.txt"
}

// LoadBlackTargetList 加载黑名单列表
func (t *DomainTarget) LoadBlackTargetList() error {
	t.BlackMapList = make(map[string]struct{})

	inputFile, err := os.Open(filepath.Join(conf.GetRootPath(), "thirdparty/custom/", t.GetBlackFileName()))
	if err != nil {
		logging.RuntimeLog.Error(err)
		logging.CLILog.Error(err)
		return err
	}
	scanner := bufio.NewScanner(inputFile)
	for scanner.Scan() {
		text := strings.ToLower(scanner.Text())
		if text == "" || strings.HasPrefix(text, "#") {
			continue
		}
		if strings.HasPrefix(text, ".") == false {
			t.BlackMapList["."+text] = struct{}{}
		} else {
			t.BlackMapList[text] = struct{}{}
		}
	}
	inputFile.Close()
	return nil
}

// CheckBlackTarget 检查黑名单
func (t *DomainTarget) CheckBlackTarget(target string) bool {
	if !utils.CheckIPV4(target) && utils.CheckDomain(target) {
		for txt := range t.BlackMapList {
			// 生成格式为.qq.com$
			regPattern := strings.ReplaceAll(txt, ".", "\\.") + "$"
			if m, _ := regexp.MatchString(regPattern, target); m == true {
				return true
			}
		}
	}
	return false
}
