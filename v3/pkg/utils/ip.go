package utils

import (
	"errors"
	"fmt"
	"github.com/hanc00l/nemo_go/v3/pkg/logging"
	"math"
	"math/big"
	"net"
	"net/netip"
	"regexp"
	"strconv"
	"strings"
)

// IPV4ToUInt32 将点分格式的IP地址转换为UINT32
func IPV4ToUInt32(ip string) uint32 {
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

// UInt32ToIPV4 将UINT32格式的IP地址转换为点分格式
func UInt32ToIPV4(ip uint32) string {
	return fmt.Sprintf("%d.%d.%d.%d",
		byte(ip>>24), byte(ip>>16), byte(ip>>8), byte(ip))
}

// IPV6ToBigInt 将点分格式的IP地址转换为大整数
func IPV6ToBigInt(ipString string) *big.Int {
	ip := net.ParseIP(ipString)
	if ip == nil {
		return nil
	}
	i := big.NewInt(0)
	i.SetBytes(ip)
	return i
}

// IPV6ToDoubleInt64 ipv6转换为两个int64
func IPV6ToDoubleInt64(ipString string) (uint64, uint64) {
	ip := net.ParseIP(ipString)
	if ip == nil {
		return 0, 0
	}
	var highSum, lowSum uint64
	for i := 0; i < 8; i++ {
		highSum += uint64(ip[i]) << (8 * (8 - i - 1))
	}
	for i := 8; i < 16; i++ {
		lowSum += uint64(ip[i]) << (8 * (16 - i - 1))
	}
	return highSum, lowSum
}

func IPV6DoubleInt64ToString(highSum, lowSum uint64) string {
	var ip net.IP
	buf := make([]byte, 16)
	for i := 0; i < 8; i++ {
		buf[i] = byte(highSum >> (8 * (8 - i - 1)))
	}
	for i := 8; i < 16; i++ {
		buf[i] = byte(lowSum >> (8 * (16 - i - 1)))
	}
	ip = buf
	return ip.String()
}

// BigIntToIPV4 将大整数格式的IP地址转换为点分格式
func BigIntToIPV6(i *big.Int) string {
	var ip net.IP
	buf := make([]byte, 16)
	ip = i.FillBytes(buf)
	return ip.String()
}

func CheckIPV4(ip string) bool {
	ipReg := `^((0|[1-9]\d?|1\d\d|2[0-4]\d|25[0-5])\.){3}(0|[1-9]\d?|1\d\d|2[0-4]\d|25[0-5])$`
	r, _ := regexp.Compile(ipReg)

	return r.MatchString(ip)

}

// CheckIPV4Subnet 检查是否是ipv4地址段
func CheckIPV4Subnet(ip string) bool {
	ipReg := `^((0|[1-9]\d?|1\d\d|2[0-4]\d|25[0-5])\.){3}(0|[1-9]\d?|1\d\d|2[0-4]\d|25[0-5])/\d{1,2}$`
	r, _ := regexp.Compile(ipReg)

	return r.MatchString(ip)
}

// GetOutBoundIP 获取本机出口IP
func GetOutBoundIP() (ip string, err error) {
	conn, err := net.Dial("udp", "8.8.8.8:53")
	if err != nil {
		return "", err
	}
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	ip = strings.Split(localAddr.String(), ":")[0]

	return ip, nil
}

// GetClientIp 获取客户端IP
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

	return "", errors.New("can not find the client ip address")
}

// ParseIP 将IP地址、IP地址段、IP地址范围解析为IP地址列表，支持ipv4/ipv6
func ParseIP(ip string) (ipResults []string) {
	//192.168.1.1
	//2409:8929:42d:bf31:1840:27ba:d669:823f
	if CheckIPV4(ip) || CheckIPV6(ip) {
		return []string{ip}
	}
	//192.168.1.0/24
	if CheckIPV4Subnet(ip) {
		addr, ipv4sub, err := net.ParseCIDR(ip)
		if err != nil {
			return
		}
		ones, bits := ipv4sub.Mask.Size()
		ipStart := IPV4ToUInt32(addr.String())
		ipSize := int(math.Pow(2, float64(bits-ones)))
		for i := 0; i < ipSize; i++ {
			ipResults = append(ipResults, UInt32ToIPV4(uint32(i)+ipStart))
		}
		return
	}
	//2409:8929:42d:bf31:1840:27ba:d669:8200/120
	if CheckIPV6Subnet(ip) {
		ipv6Prefix, err := netip.ParsePrefix(ip)
		if err != nil {
			return
		}
		ipStart := ipv6Prefix.Addr()
		bits := ipv6Prefix.Bits()
		// 为防止IP爆炸，最多支持解析一个A段的地址
		if bits < 104 {
			logging.RuntimeLog.Warningf("ipv6 subnet too large to discard parse:%s", ip)
			//ipResults = append(ipResults, ip)
			return
		}
		for {
			ipResults = append(ipResults, ipStart.String())
			ipStart = ipStart.Next()
			if !ipv6Prefix.Contains(ipStart) {
				break
			}
		}
		return
	}
	//192.168.1.1-192.168.1.5
	//2409:8929:42d:bf31:1840:27ba:d669:8200-2409:8929:42d:bf31:1840:27ba:d669:82ff
	address := strings.Split(ip, "-")
	if len(address) == 2 {
		if CheckIPV4(address[0]) && CheckIPV4(address[1]) {
			ipStart := address[0]
			ipEnd := address[1]
			for i := IPV4ToUInt32(ipStart); i <= IPV4ToUInt32(ipEnd); i++ {
				ipResults = append(ipResults, UInt32ToIPV4(i))
			}
			return
		}
		if CheckIPV6(address[0]) && CheckIPV6(address[1]) {
			ipStart, err1 := netip.ParseAddr(address[0])
			ipEnd, err2 := netip.ParseAddr(address[1])
			if err1 != nil || err2 != nil {
				return
			}
			for {
				ipResults = append(ipResults, ipStart.String())
				if ipStart.Compare(ipEnd) >= 0 {
					break
				}
				ipStart = ipStart.Next()
			}
			return
		}
	}
	return
}

// CheckIPLocationInChinaMainLand 根据IP归属地判断是否是属于中国大陆的IP地区
func CheckIPLocationInChinaMainLand(ipLocation string) bool {
	// 省、市、自治区
	regions := []string{
		"河北", "山西", "辽宁", "吉林", "黑龙江",
		"江苏", "浙江", "安徽", "福建", "江西",
		"山东", "河南", "湖北", "湖南", "广东",
		"海南", "四川", "贵州", "云南", "陕西",
		"甘肃", "青海",
		"内蒙古", "广西", "西藏", "宁夏", "新疆",
		"北京", "天津", "上海", "重庆",
	}
	for _, region := range regions {
		if strings.Contains(ipLocation, region) {
			return true
		}
	}

	return false
}

// CheckIPLocationInChinaOther 根据IP归属地判断是否是属于中国非大陆地区
func CheckIPLocationInChinaOther(ipLocation string) bool {
	if len(ipLocation) == 0 {
		return false
	}
	regions := []string{
		"香港", "台湾", "澳门"}
	for _, region := range regions {
		if strings.Contains(ipLocation, region) {
			return true
		}
	}
	return false
}

// CheckIPLocationOutsideChina 根据IP归属地判断是否是属于非中国区域
func CheckIPLocationOutsideChina(ipLocation string) bool {
	if len(ipLocation) == 0 {
		return false
	}
	if !CheckIPLocationInChinaMainLand(ipLocation) && !CheckIPLocationInChinaOther(ipLocation) {
		return true
	}

	return false
}

// CheckIPV6 检查是否是ipv6地址
func CheckIPV6(ip string) bool {
	ipv6Regex := `^(([0-9a-fA-F]{1,4}:){7,7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(ffff(:0{1,4}){0,1}:){0,1}((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9]))$`
	match, _ := regexp.MatchString(ipv6Regex, ip)

	return match
}

func CheckIPV6Subnet(ip string) bool {
	ipv6Regex := `^(([0-9a-fA-F]{1,4}:){7,7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(ffff(:0{1,4}){0,1}:){0,1}((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9]))/\d{1,3}$`
	match, _ := regexp.MatchString(ipv6Regex, ip)

	return match
}

// GetIPV4ParsedFormat 将IPv6地址进行格式化（缩写）
func GetIPV6ParsedFormat(ip string) string {
	ipv6Addr, err := netip.ParseAddr(ip)
	if err != nil {
		return ""
	} else {
		return ipv6Addr.String()
	}
}

// GetIPV6CIDRParsedFormat 将IPv6地址段进行格式化（缩写）
func GetIPV6CIDRParsedFormat(ip string) string {
	ipv6Prefix, err := netip.ParsePrefix(ip)
	if err != nil {
		return ""
	} else {
		return ipv6Prefix.String()
	}
}

// GetIPV6FullFormat 将IPv6地址进行格式化（完全展开格式）
func GetIPV6FullFormat(ip string) string {
	ipv6Addr, err := netip.ParseAddr(ip)
	if err != nil {
		return ""
	}
	return ipv6Addr.StringExpanded()
}

// IPV6Prefix64ToUInt64 将IPv6地址段的Prefix段（前64位）转换为uint64
func IPV6Prefix64ToUInt64(ip string) uint64 {
	var sum uint64
	dataArray := strings.Split(GetIPV6FullFormat(ip), ":")
	if len(dataArray) == 8 {
		b0, _ := strconv.ParseUint(dataArray[0], 16, 16)
		b1, _ := strconv.ParseUint(dataArray[1], 16, 16)
		b2, _ := strconv.ParseUint(dataArray[2], 16, 16)
		b3, _ := strconv.ParseUint(dataArray[3], 16, 16)
		sum += b0 << 48
		sum += b1 << 32
		sum += b2 << 16
		sum += b3
	}
	return sum
}

// CheckIP 通过正则检查是否是ipv4或ipv6地址
func CheckIP(ip string) bool {
	if CheckIPV4(ip) || CheckIPV6(ip) {
		return true
	}
	return false
}

// CheckIPOrSubnet 通过正则检查是否是ipv4、ipv6地址或CIDR
func CheckIPOrSubnet(ip string) bool {
	if CheckIPV4(ip) || CheckIPV6(ip) || CheckIPV4Subnet(ip) || CheckIPV6Subnet(ip) {
		return true
	}
	return false
}

// GetIPV6SubnetC 生成IPv6的C段掩码地址
func GetIPV6SubnetC(ip string) string {
	ipArray := strings.Split(GetIPV6FullFormat(ip), ":")
	return GetIPV6CIDRParsedFormat(fmt.Sprintf("%s:%s:%s:%s:%s:%s:%s:%s00/120",
		ipArray[0], ipArray[1], ipArray[2], ipArray[3], ipArray[4], ipArray[5], ipArray[6], ipArray[7][0:2]))
}

func GetDomainName(domain string) string {
	if strings.Contains(domain, ":") {
		return domain[:strings.Index(domain, ":")]
	}
	return domain
}
func GetDomainPort(domain string) int {
	if strings.Contains(domain, ":") {
		if port, err := strconv.Atoi(domain[strings.Index(domain, ":")+1:]); err == nil {
			return port
		}
	}
	return 0
}

// FindIPV4 从字符串中提取出所有的ipv4地址
//func FindIPV4(content string) []string {
//	// forked from https://github.com/mingrammer/commonregex
//
//	var IPv4Pattern = `(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)`
//	//var IPv6Pattern = `(?:(?:(?:[0-9A-Fa-f]{1,4}:){7}(?:[0-9A-Fa-f]{1,4}|:))|(?:(?:[0-9A-Fa-f]{1,4}:){6}(?::[0-9A-Fa-f]{1,4}|(?:(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(?:\.(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3})|:))|(?:(?:[0-9A-Fa-f]{1,4}:){5}(?:(?:(?::[0-9A-Fa-f]{1,4}){1,2})|:(?:(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(?:\.(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3})|:))|(?:(?:[0-9A-Fa-f]{1,4}:){4}(?:(?:(?::[0-9A-Fa-f]{1,4}){1,3})|(?:(?::[0-9A-Fa-f]{1,4})?:(?:(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(?:\.(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(?:(?:[0-9A-Fa-f]{1,4}:){3}(?:(?:(?::[0-9A-Fa-f]{1,4}){1,4})|(?:(?::[0-9A-Fa-f]{1,4}){0,2}:(?:(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(?:\.(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(?:(?:[0-9A-Fa-f]{1,4}:){2}(?:(?:(?::[0-9A-Fa-f]{1,4}){1,5})|(?:(?::[0-9A-Fa-f]{1,4}){0,3}:(?:(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(?:\.(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(?:(?:[0-9A-Fa-f]{1,4}:){1}(?:(?:(?::[0-9A-Fa-f]{1,4}){1,6})|(?:(?::[0-9A-Fa-f]{1,4}){0,4}:(?:(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(?:\.(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(?::(?:(?:(?::[0-9A-Fa-f]{1,4}){1,7})|(?:(?::[0-9A-Fa-f]{1,4}){0,5}:(?:(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(?:\.(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:)))(?:%.+)?\s*`
//	//var IPPattern = IPv4Pattern + `|` + IPv6Pattern
//	var IPv4Regex = regexp.MustCompile(IPv4Pattern)
//	//var IPv6Regex = regexp.MustCompile(IPv6Pattern)
//	//var IPRegex = regexp.MustCompile(IPPattern)
//
//	allResult := IPv4Regex.FindAllString(content, -1)
//	//去重：
//	ips := make(map[string]struct{})
//	for _, ip := range allResult {
//		if _, existed := ips[ip]; !existed {
//			ips[ip] = struct{}{}
//		}
//	}
//
//	return SetToSlice(ips)
//}
