package custom

import (
	"bufio"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/utils"
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
	BlackMapList map[string]struct{}
}

type DomainTarget struct {
	BlackMapList map[string]struct{}
}

type BlackTargetCheck struct {
	checkType       BlackCheckType
	blackTargetList map[BlackCheckType]BlackTarget
}

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

func (c *BlackTargetCheck) GetBlackFileName() string {
	if c.checkType == CheckIP || c.checkType == CheckDomain {
		return c.blackTargetList[c.checkType].GetBlackFileName()
	}
	return "black_error.txt"
}

func (t *IPTarget) GetBlackFileName() string {
	return "black_ip.txt"
}

func (t *IPTarget) LoadBlackTargetList() error {
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
		ipAndComment := strings.Split(text, " ")
		ips := utils.ParseIP(ipAndComment[0])
		for _, ip := range ips {
			if _, ok := t.BlackMapList[ip]; !ok {
				t.BlackMapList[ip] = struct{}{}
			}
		}
	}
	inputFile.Close()

	return nil
}

func (t *IPTarget) CheckBlackTarget(target string) bool {
	if utils.CheckIPV4(target) {
		_, existed := t.BlackMapList[target]
		return existed
	}
	return false
}

func (t *DomainTarget) GetBlackFileName() string {
	return "black_domain.txt"

}

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
