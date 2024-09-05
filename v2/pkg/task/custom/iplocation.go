package custom

import (
	"fmt"
	"github.com/hanc00l/nemo_go/v2/pkg/conf"
	"github.com/hanc00l/nemo_go/v2/pkg/logging"
	"github.com/hanc00l/nemo_go/v2/pkg/utils"
	"os"
	"path/filepath"
	"strings"
)

var cloudNameList = []string{
	"阿里云", "华为云", "腾讯云",
	"天翼云", "金山云", "UCloud", "青云", "QingCloud", "百度云", "盛大云", "世纪互联蓝云",
	"Azure", "Amazon", "Microsoft", "Google", "vultr", "CloudFlare", "Choopa",
}

type Config struct {
	Target string `json:"target"`
	OrgId  *int   `json:"orgId"`
}

type Ipv4Location struct {
	customMap  map[string]string
	customBMap map[string]string
	customCMap map[string]string
}

// NewIPv4Location 创建iplocation对象
func NewIPv4Location() *Ipv4Location {
	ipl := &Ipv4Location{
		customMap:  make(map[string]string),
		customBMap: make(map[string]string),
		customCMap: make(map[string]string),
	}
	ipl.loadCustomIP()
	ipl.loadQQwry()
	return ipl
}

// FindPublicIP 查询纯真数据库获取公网IP归属地
func (ipl *Ipv4Location) FindPublicIP(ip string) string {
	qqWry := NewQQwry()
	result := qqWry.Find(ip)

	if len(result.Area) > 0 {
		for _, v := range cloudNameList {
			if strings.Index(result.Area, v) >= 0 {
				return fmt.Sprintf("%s [%s]", result.Country, result.Area)
			}
		}
	}
	return result.Country
}

// FindCustomIP 查询自定义IP归属地
func (ipl *Ipv4Location) FindCustomIP(ip string) string {
	result, ok := ipl.customMap[ip]
	if ok {
		return result
	}
	ipBytes := strings.Split(ip, ".")
	if len(ipBytes) != 4 {
		return ""
	}
	result, ok = ipl.customCMap[strings.Join([]string{ipBytes[0], ipBytes[1], ipBytes[2], "0"}, ".")]
	if ok {
		return result
	}
	result, ok = ipl.customBMap[strings.Join([]string{ipBytes[0], ipBytes[1], "0", "0"}, ".")]
	if ok {
		return result
	}

	return ""
}

// loadQQwry 加载纯真IP数据库
func (ipl *Ipv4Location) loadQQwry() {
	IPData.FilePath = filepath.Join(conf.GetRootPath(), "thirdparty/qqwry/qqwry.dat")
	res := IPData.InitIPData()

	if v, ok := res.(error); ok {
		logging.RuntimeLog.Error(v)
		logging.CLILog.Error(v)
	} else {
		//logging.RuntimeLog.Infof("纯真IP库加载完成,共加载:%d 条 Domain 记录", IPData.IPNum)
	}
}

// loadCustomIP 加载自定义IP归属地库
func (ipl *Ipv4Location) loadCustomIP() {
	content, err := os.ReadFile(filepath.Join(conf.GetRootPath(), "thirdparty/custom/iplocation-custom-B.txt"))
	if err != nil {
		logging.RuntimeLog.Error(err)
		logging.CLILog.Error(err)
	} else {
		for _, line := range strings.Split(string(content), "\n") {
			txt := strings.TrimSpace(line)
			if txt == "" || strings.HasPrefix(txt, "#") {
				continue
			}
			ipLocationArrays := strings.Split(txt, " ")
			if len(ipLocationArrays) < 2 {
				continue
			}
			ips := strings.Split(ipLocationArrays[0], ".")
			if len(ips) != 4 {
				continue
			}
			ipl.customBMap[strings.Join([]string{ips[0], ips[1], "0", "0"}, ".")] = ipLocationArrays[1]
		}
	}

	content, err = os.ReadFile(filepath.Join(conf.GetRootPath(), "thirdparty/custom/iplocation-custom-C.txt"))
	if err != nil {
		logging.RuntimeLog.Error(err)
		logging.CLILog.Error(err)
	} else {
		for _, line := range strings.Split(string(content), "\n") {
			txt := strings.TrimSpace(line)
			if txt == "" || strings.HasPrefix(txt, "#") {
				continue
			}
			ipLocationArrays := strings.Split(txt, " ")
			if len(ipLocationArrays) < 2 {
				continue
			}
			ips := strings.Split(ipLocationArrays[0], ".")
			if len(ips) != 4 {
				continue
			}
			ipl.customCMap[strings.Join([]string{ips[0], ips[1], ips[2], "0"}, ".")] = ipLocationArrays[1]
		}
	}

	content, err = os.ReadFile(filepath.Join(conf.GetRootPath(), "thirdparty/custom/iplocation-custom.txt"))
	if err != nil {
		logging.RuntimeLog.Error(err)
		logging.CLILog.Error(err)
	} else {
		for _, line := range strings.Split(string(content), "\n") {
			txt := strings.TrimSpace(line)
			if txt == "" || strings.HasPrefix(txt, "#") {
				continue
			}
			ipLocationArrays := strings.Split(txt, " ")
			if len(ipLocationArrays) < 2 {
				continue
			}
			ips := utils.ParseIP(ipLocationArrays[0])
			for _, ip := range ips {
				ipl.customMap[ip] = ipLocationArrays[1]
			}
		}
	}
}
