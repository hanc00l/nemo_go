package utils

import (
	"errors"
	"fmt"
	"github.com/hanc00l/nemo_go/v3/pkg/logging"
	"math/big"
	"net"
	"net/netip"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var (
	TopPorts1000 = "1,3-4,6-7,9,13,17,19-26,30,32-33,37,42-43,49,53,70,79-85,88-90,99-100,106,109-111,113,119,125,135,139,143-144,146,161,163,179,199,211-212,222,254-256,259,264,280,301,306,311,340,366,389,406-407,416-417,425,427,443-445,458,464-465,481,497,500,512-515,524,541,543-545,548,554-555,563,587,593,616-617,625,631,636,646,648,666-668,683,687,691,700,705,711,714,720,722,726,749,765,777,783,787,800-801,808,843,873,880,888,898,900-903,911-912,981,987,990,992-993,995,999-1002,1007,1009-1011,1021-1100,1102,1104-1108,1110-1114,1117,1119,1121-1124,1126,1130-1132,1137-1138,1141,1145,1147-1149,1151-1152,1154,1163-1166,1169,1174-1175,1183,1185-1187,1192,1198-1199,1201,1213,1216-1218,1233-1234,1236,1244,1247-1248,1259,1271-1272,1277,1287,1296,1300-1301,1309-1311,1322,1328,1334,1352,1417,1433-1434,1443,1455,1461,1494,1500-1501,1503,1521,1524,1533,1556,1580,1583,1594,1600,1641,1658,1666,1687-1688,1700,1717-1721,1723,1755,1761,1782-1783,1801,1805,1812,1839-1840,1862-1864,1875,1900,1914,1935,1947,1971-1972,1974,1984,1998-2010,2013,2020-2022,2030,2033-2035,2038,2040-2043,2045-2049,2065,2068,2099-2100,2103,2105-2107,2111,2119,2121,2126,2135,2144,2160-2161,2170,2179,2190-2191,2196,2200,2222,2251,2260,2288,2301,2323,2366,2381-2383,2393-2394,2399,2401,2492,2500,2522,2525,2557,2601-2602,2604-2605,2607-2608,2638,2701-2702,2710,2717-2718,2725,2800,2809,2811,2869,2875,2909-2910,2920,2967-2968,2998,3000-3001,3003,3005-3007,3011,3013,3017,3030-3031,3052,3071,3077,3128,3168,3211,3221,3260-3261,3268-3269,3283,3300-3301,3306,3322-3325,3333,3351,3367,3369-3372,3389-3390,3404,3476,3493,3517,3527,3546,3551,3580,3659,3689-3690,3703,3737,3766,3784,3800-3801,3809,3814,3826-3828,3851,3869,3871,3878,3880,3889,3905,3914,3918,3920,3945,3971,3986,3995,3998,4000-4006,4045,4111,4125-4126,4129,4224,4242,4279,4321,4343,4443-4446,4449,4550,4567,4662,4848,4899-4900,4998,5000-5004,5009,5030,5033,5050-5051,5054,5060-5061,5080,5087,5100-5102,5120,5190,5200,5214,5221-5222,5225-5226,5269,5280,5298,5357,5405,5414,5431-5432,5440,5500,5510,5544,5550,5555,5560,5566,5631,5633,5666,5678-5679,5718,5730,5800-5802,5810-5811,5815,5822,5825,5850,5859,5862,5877,5900-5904,5906-5907,5910-5911,5915,5922,5925,5950,5952,5959-5963,5987-5989,5998-6007,6009,6025,6059,6100-6101,6106,6112,6123,6129,6156,6346,6389,6502,6510,6543,6547,6565-6567,6580,6646,6666-6669,6689,6692,6699,6779,6788-6789,6792,6839,6881,6901,6969,7000-7002,7004,7007,7019,7025,7070,7100,7103,7106,7200-7201,7402,7435,7443,7496,7512,7625,7627,7676,7741,7777-7778,7800,7911,7920-7921,7937-7938,7999-8002,8007-8011,8021-8022,8031,8042,8045,8080-8090,8093,8099-8100,8180-8181,8192-8194,8200,8222,8254,8290-8292,8300,8333,8383,8400,8402,8443,8500,8600,8649,8651-8652,8654,8701,8800,8873,8888,8899,8994,9000-9003,9009-9011,9040,9050,9071,9080-9081,9090-9091,9099-9103,9110-9111,9200,9207,9220,9290,9415,9418,9485,9500,9502-9503,9535,9575,9593-9595,9618,9666,9876-9878,9898,9900,9917,9929,9943-9944,9968,9998-10004,10009-10010,10012,10024-10025,10082,10180,10215,10243,10566,10616-10617,10621,10626,10628-10629,10778,11110-11111,11967,12000,12174,12265,12345,13456,13722,13782-13783,14000,14238,14441-14442,15000,15002-15004,15660,15742,16000-16001,16012,16016,16018,16080,16113,16992-16993,17877,17988,18040,18101,18988,19101,19283,19315,19350,19780,19801,19842,20000,20005,20031,20221-20222,20828,21571,22939,23502,24444,24800,25734-25735,26214,27000,27352-27353,27355-27356,27715,28201,30000,30718,30951,31038,31337,32768-32785,33354,33899,34571-34573,35500,38292,40193,40911,41511,42510,44176,44442-44443,44501,45100,48080,49152-49161,49163,49165,49167,49175-49176,49400,49999-50003,50006,50300,50389,50500,50636,50800,51103,51493,52673,52822,52848,52869,54045,54328,55055-55056,55555,55600,56737-56738,57294,57797,58080,60020,60443,61532,61900,62078,63331,64623,64680,65000,65129,65389"
	TopPorts100  = "7,9,13,21-23,25-26,37,53,79-81,88,106,110-111,113,119,135,139,143-144,179,199,389,427,443-445,465,513-515,543-544,548,554,587,631,646,873,990,993,995,1025-1029,1110,1433,1720,1723,1755,1900,2000-2001,2049,2121,2717,3000,3128,3306,3389,3986,4899,5000,5009,5051,5060,5101,5190,5357,5432,5631,5666,5800,5900,6000-6001,6646,7070,8000,8008-8009,8080-8081,8443,8888,9100,9999-10000,32768,49152-49157"
	TopPorts10   = "21-23,80,139,443,445,3306,3389,8080"
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

// ipToUint32 converts an IP address to uint32
func ipToUint32(ip net.IP) uint32 {
	ip = ip.To4()
	return uint32(ip[0])<<24 | uint32(ip[1])<<16 | uint32(ip[2])<<8 | uint32(ip[3])
}

// uint32ToIP converts a uint32 back to an IP address
func uint32ToIP(i uint32) net.IP {
	return net.IP{
		byte(i >> 24),
		byte(i >> 16),
		byte(i >> 8),
		byte(i),
	}
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
		// Parse the CIDR notation
		_, ipNet, err := net.ParseCIDR(ip)
		if err != nil {
			return nil
		}
		// Get the IP and mask
		ipx := ipNet.IP
		mask := ipNet.Mask
		// Calculate the number of IPs in the subnet
		ones, bits := mask.Size()
		numIPs := 1 << uint(bits-ones) // 2^(32-ones) for IPv4
		// Convert IP to uint32 for arithmetic
		ipInt := ipToUint32(ipx)
		ips := make([]string, 0, numIPs)
		// Generate all IPs in the subnet
		for i := uint32(0); i < uint32(numIPs); i++ {
			newIP := uint32ToIP(ipInt + i)
			ips = append(ips, newIP.String())
		}
		return ips
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

func SplitPort(portList string, sliceNumber int) []string {
	// 处理top-ports情况
	if strings.HasPrefix(portList, "--top-ports") {
		parts := strings.Fields(portList)
		if len(parts) < 2 {
			portList = TopPorts10
		} else {
			switch parts[1] {
			case "1000":
				portList = TopPorts1000
			case "100":
				portList = TopPorts100
			case "10":
				portList = TopPorts10
			default:
				portList = TopPorts10
			}
		}
	}

	// 解析端口列表为单个端口号
	ports := ParsePortList(portList)
	if len(ports) == 0 {
		return []string{}
	}

	// 排序端口号
	sort.Ints(ports)

	// 分片处理
	var result []string
	for i := 0; i < len(ports); i += sliceNumber {
		end := i + sliceNumber
		if end > len(ports) {
			end = len(ports)
		}

		slicePorts := ports[i:end]
		result = append(result, FormatPorts(slicePorts))
	}

	return result
}

// ParsePortList 解析端口列表字符串为单个端口号
func ParsePortList(portList string) []int {
	var ports []int
	parts := strings.Split(portList, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.Contains(part, "-") {
			// 处理端口范围
			rangeParts := strings.Split(part, "-")
			if len(rangeParts) != 2 {
				continue
			}
			start, err1 := strconv.Atoi(rangeParts[0])
			end, err2 := strconv.Atoi(rangeParts[1])
			if err1 != nil || err2 != nil || start > end {
				continue
			}
			for i := start; i <= end; i++ {
				ports = append(ports, i)
			}
		} else {
			// 处理单个端口
			port, err := strconv.Atoi(part)
			if err == nil {
				ports = append(ports, port)
			}
		}
	}

	return ports
}

// FormatPorts 将端口号切片格式化为端口列表字符串
func FormatPorts(ports []int) string {
	if len(ports) == 0 {
		return ""
	}

	var result []string
	start := ports[0]
	prev := ports[0]

	for i := 1; i < len(ports); i++ {
		if ports[i] == prev+1 {
			prev = ports[i]
		} else {
			if start == prev {
				result = append(result, strconv.Itoa(start))
			} else {
				result = append(result, fmt.Sprintf("%d-%d", start, prev))
			}
			start = ports[i]
			prev = ports[i]
		}
	}

	// 处理最后一个范围
	if start == prev {
		result = append(result, strconv.Itoa(start))
	} else {
		result = append(result, fmt.Sprintf("%d-%d", start, prev))
	}

	return strings.Join(result, ",")
}

func FormatTargetByPortList(target string, portList []int) string {
	var results []string

	// 分割原始target
	targets := strings.Split(target, ",")

	for _, t := range targets {
		// 分割host和可能的端口
		hostPort := strings.Split(t, ":")
		host := hostPort[0] // 总是有至少一个元素

		// 如果原始target已经有端口，忽略它，只用host部分
		for _, port := range portList {
			results = append(results, fmt.Sprintf("%s:%d", host, port))
		}
	}

	return strings.Join(results, ",")
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
