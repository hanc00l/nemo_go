package utils

import (
	"errors"
	"fmt"
	"math"
	"net"
	"regexp"
	"strconv"
	"strings"
)

// IPToUInt32 将点分格式的IP地址转换为UINT32
func IPToUInt32(ip string) uint32 {
	bits := strings.Split(ip, ".")
	b0, _ := strconv.Atoi(bits[0])
	b1, _ := strconv.Atoi(bits[1])
	b2, _ := strconv.Atoi(bits[2])
	b3, _ := strconv.Atoi(bits[3])
	var sum uint32
	sum += uint32(b0) << 24
	sum += uint32(b1) << 16
	sum += uint32(b2) << 8
	sum += uint32(b3)

	return sum
}

func UInt32ToIP(ip uint32) string {
	return fmt.Sprintf("%d.%d.%d.%d",
		byte(ip>>24), byte(ip>>16), byte(ip>>8), byte(ip))
}

func CheckIPV4(ip string) bool {
	ipReg := `^((0|[1-9]\d?|1\d\d|2[0-4]\d|25[0-5])\.){3}(0|[1-9]\d?|1\d\d|2[0-4]\d|25[0-5])$`
	r, _ := regexp.Compile(ipReg)

	return r.MatchString(ip)

}

func CheckIPV4Subnet(ip string) bool {
	ipReg := `^((0|[1-9]\d?|1\d\d|2[0-4]\d|25[0-5])\.){3}(0|[1-9]\d?|1\d\d|2[0-4]\d|25[0-5])/\d{1,2}$`
	r, _ := regexp.Compile(ipReg)

	return r.MatchString(ip)
}

func GetOutBoundIP() (ip string, err error) {
	conn, err := net.Dial("udp", "8.8.8.8:53")
	if err != nil {
		return "", err
	}
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	ip = strings.Split(localAddr.String(), ":")[0]

	return ip, nil
}

func GetClientIp() (ip string, err error) {
	adders, err := net.InterfaceAddrs()
	if err != nil {
		return ip, err
	}
	for _, address := range adders {
		if inet, ok := address.(*net.IPNet); ok && !inet.IP.IsLoopback() {
			if inet.IP.To4() != nil {
				return inet.IP.String(), nil
			}
		}
	}

	return "", errors.New("Can not find the client ip address!")
}

func ParseIP(ip string) (ipResults []string) {
	//192.168.1.1
	if CheckIPV4(ip) {
		return []string{ip}
	}
	//192.168.1.0/24
	if CheckIPV4Subnet(ip) {
		addr, ipv4sub, err := net.ParseCIDR(ip)
		if err != nil {
			return
		}
		ones, bits := ipv4sub.Mask.Size()
		ipStart := IPToUInt32(addr.String())
		ipSize := int(math.Pow(2, float64(bits-ones)))
		for i := 0; i < ipSize; i++ {
			ipResults = append(ipResults, UInt32ToIP(uint32(i)+ipStart))
		}
		return
	}
	//192.168.1.1-192.168.1.5
	address := strings.Split(ip, "-")
	if len(address) == 2 && CheckIPV4(address[0]) && CheckIPV4(address[1]) {
		ipStart := address[0]
		ipEnd := address[1]
		for i := IPToUInt32(ipStart); i <= IPToUInt32(ipEnd); i++ {
			ipResults = append(ipResults, UInt32ToIP(i))
		}
		return
	}

	return
}

// CheckIPLocationInChinaMainLand 根据IP归属地判断是否是属于中国大陆的IP地区
func CheckIPLocationInChinaMainLand(ipLocation string) bool {
	//如果无IP定位，无法判断返回true
	if ipLocation == "" {
		return true
	}
	//检查：香港 台湾 澳门地区
	pattern2 := `香港|台湾|澳门`
	r2 := regexp.MustCompile(pattern2)
	result2 := r2.FindStringSubmatch(ipLocation)
	if len(result2) >= 1 {
		return false
	}
	//检查：X国
	pattern1 := `(.+?)国`
	r1 := regexp.MustCompile(pattern1)
	result1 := r1.FindStringSubmatch(ipLocation)
	if len(result1) >= 2 && result1[1] != "中" {
		return false
	}
	//检查：带有省、市及自治区
	pattern3 := `省|市|内蒙古|广西|西藏|宁夏|新疆`
	r3 := regexp.MustCompile(pattern3)
	result3 := r3.FindStringSubmatch(ipLocation)
	if len(result3) >= 1 {
		return true
	}
	return false
}
