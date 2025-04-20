package custom

import (
	"fmt"
	"github.com/hanc00l/nemo_go/v3/pkg/conf"
	"github.com/hanc00l/nemo_go/v3/pkg/db"
	"github.com/hanc00l/nemo_go/v3/pkg/logging"
	"github.com/hanc00l/nemo_go/v3/pkg/utils"
	"path/filepath"
	"strconv"
	"strings"
)

var cloudNameList = []string{
	"阿里云", "华为云", "腾讯云",
	"天翼云", "金山云", "UCloud", "青云", "QingCloud", "百度云", "盛大云", "世纪互联蓝云",
	"Azure", "Amazon", "Microsoft", "Google", "vultr", "CloudFlare", "Choopa",
}

type Ipv4Location struct {
	customMap  map[string]string
	customBMap map[string]string
	customCMap map[string]string
	// 用于自定义IP归属地库
	WorkspaceId string
}

// NewIPv4Location 创建iplocation对象
func NewIPv4Location(workspaceId string) *Ipv4Location {
	ipl := &Ipv4Location{
		customMap:   make(map[string]string),
		customBMap:  make(map[string]string),
		customCMap:  make(map[string]string),
		WorkspaceId: workspaceId,
	}
	ipl.loadCustomIP()
	ipl.loadQQwry()
	return ipl
}

// Find 查询IP归属地，先查询自定义IP归属地库，再查询公网的纯真IP数据库
func (ipl *Ipv4Location) Find(ip string) string {
	result := ipl.FindCustomIP(ip)
	if result != "" {
		return result
	}
	return ipl.FindPublicIP(ip)
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

func (ipl *Ipv4Location) loadCustomIP() {
	// 如果指定了workspaceId, 则从数据库加载自定义IP归属地库
	if ipl.WorkspaceId == "" {
		return
	}
	mongoClient, err := db.GetClient()
	if err != nil {
		logging.RuntimeLog.Error(err)
		return
	}
	defer db.CloseClient(mongoClient)

	cd := db.NewCustomData(ipl.WorkspaceId, mongoClient)
	docs, _ := cd.Find(db.CategoryCustomIPLocation)
	for _, doc := range docs {
		lines := strings.Split(doc.Data, "\n")
		for _, line := range lines {
			ips := strings.TrimSpace(line)
			if strings.HasPrefix(ips, "#") {
				continue
			}
			//格式为：IP 归属地
			ipLocationArrays := strings.Split(ips, " ")
			if len(ipLocationArrays) < 2 {
				continue
			}
			ip := strings.Split(ipLocationArrays[0], ".")
			if len(ip) != 4 {
				continue
			}
			if ip[3] == "/8" {
				// A段
				logging.RuntimeLog.Warningf("not support custom ip location for A segment")
				continue
			}
			if ip[3] == "/16" {
				// B段
				ipl.customBMap[strings.Join([]string{ip[0], ip[1], "0", "0"}, ".")] = ipLocationArrays[1]
			} else if ip[3] == "/24" {
				// C段
				ipl.customCMap[strings.Join([]string{ip[0], ip[1], ip[2], "0"}, ".")] = ipLocationArrays[1]
			} else {
				// 单IP或其它CIDR
				otherIP := ipLocationArrays[0]
				if strings.HasPrefix(ipLocationArrays[1], "/") {
					otherIP = strings.ReplaceAll(otherIP, "/", "")
					n, err := strconv.Atoi(otherIP)
					if err != nil {
						continue
					}
					// CIDR掩码位数不在24-32之间
					if n < 24 || n > 32 {
						continue
					}
				}
				ips2 := utils.ParseIP(otherIP)
				for _, ip2 := range ips2 {
					ipl.customMap[ip2] = ipLocationArrays[1]
				}
			}
		}
	}
	return
}
