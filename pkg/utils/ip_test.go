package utils

import (
	"fmt"
	"net/netip"
	"testing"
)

func TestCheckIPV4(t *testing.T) {
	t.Log(CheckIPV4("192.168.1.1"))
	t.Log(CheckIPV4("192.168.1.1/24"))
	t.Log(CheckIPV4("10.0.0.0/8"))
	t.Log(CheckIPV4("0.0.0.0/0"))
}

func TestCheckIPV4Subnet(t *testing.T) {
	t.Log(CheckIPV4Subnet("192.168.1.1"))
	t.Log(CheckIPV4Subnet("192.168.1.1/24"))
	t.Log(CheckIPV4Subnet("10.0.0.0/8"))
	t.Log(CheckIPV4Subnet("0.0.0.0/0"))

}

func TestGetOutBoundIP(t *testing.T) {
	ip, err := GetOutBoundIP()
	t.Log(ip)
	t.Log(err)
}

func TestGetClientIp(t *testing.T) {
	ip, err := GetClientIp()
	t.Log(ip)
	t.Log(err)
}

func TestParseIP(t *testing.T) {
	t.Log(ParseIP("192.168.1.1"))
	t.Log(ParseIP("192.168.1.1/30"))
	t.Log(ParseIP("192.168.1.3-192.168.1.10"))
	for _, v := range ParseIP("2409:8929:42d:bf31:1840:27ba:d669::/124") {
		t.Log(v)
	}
	for _, v := range ParseIP("2409:8929:42d:bf31:1840:27ba:d669:8200-2409:8929:42d:bf31:1840:27ba:d669:82ff") {
		t.Log(v)
	}

}

func TestCheckIPLocationInChinaMainLand(t *testing.T) {
	data1 := "香x港 台湾 澳门"
	data2 := "德国 香港"
	data3 := "中国"
	data4 := "江苏省"
	data5 := "上海市 浦东"
	data6 := "阿根廷"
	data7 := "内蒙古"
	t.Log(data1, CheckIPLocationInChinaMainLand(data1))
	t.Log(data2, CheckIPLocationInChinaMainLand(data2))
	t.Log(data3, CheckIPLocationInChinaMainLand(data3))
	t.Log(data4, CheckIPLocationInChinaMainLand(data4))
	t.Log(data5, CheckIPLocationInChinaMainLand(data5))
	t.Log(data6, CheckIPLocationInChinaMainLand(data6))
	t.Log(data7, CheckIPLocationInChinaMainLand(data7))

}

func TestCheckIPV6(t *testing.T) {
	datas := []string{"::ffff:127:2cde",
		"::ffff:7f01:0203",
		"0:0:0:0:0000:ffff:127.1.2.3",
		"0:0:0:0:000000:ffff:127.1.2.3",
		"192.168.3.4",
		"fe80::5c12:27dc:93a4:3426",
		"fe3f:3fa::0",
	}
	for _, d := range datas {
		fmt.Println(d, CheckIPV6(d))
	}
}

func TestCheckIPV6Subnet(t *testing.T) {
	datas := []string{"::ffff:127:2cde",
		"::ffff:7f01:0203/126",
		"0:0:0:0:0000:ffff:127.1.2.3",
		"0:0:0:0:000000:ffff:127.1.2.3",
		"192.168.3.4/24",
		"fe80::5c12:27dc:93a4:3426",
		"fe3f:3fa::0/48",
	}
	for _, d := range datas {
		fmt.Println(d, CheckIPV6Subnet(d))
	}
}

func TestIPV6Prefix64ToUInt64(t *testing.T) {
	datas := []string{
		"a3:0:0:123:0000:ffff:127:2cde",
		"fe80::5c12:27dc:93a4:3426",
		"fe3f:3fa::0",
	}
	for _, d := range datas {
		fmt.Println(d, "->", fmt.Sprintf("%x", IPV6Prefix64ToUInt64(d)))
	}
}

func TestGetIPV6FullFormat(t *testing.T) {
	datas := []string{"::ffff:127:2cde",
		"::ffff:7f01:0203",
		"fe3f:3fa::0:3456",
		"fe3f:3fa::0",
		"fe3f:3fa::",
	}
	for _, d := range datas {
		fmt.Println(d, "->", GetIPV6FullFormat(d))
	}
}

func TestGetIPV6ParsedFormat(t *testing.T) {
	datas := []string{"::ffff:127:2cde",
		"::ffff:7f01:0203",
		"fe3f:3fa::0:3456",
		"fe3f:3fa::0",
		"fe3f:3fa::",
		"0:0:0:0:0000:ffff:127:abc",
		"24ae:0:0:0:0000:ffff:127:abc",
		"192.168.3.4",
		"fe80::5c12:27dc:93a4:3426",
	}
	for _, d := range datas {
		dd := GetIPV6ParsedFormat(d)
		ddd := GetIPV6FullFormat(dd)
		fmt.Println(d, "->", dd, "->", ddd, "->", GetIPV6CIDRParsedFormat(fmt.Sprintf("%s/120", ddd)))
	}
}

func TestIPv6Subnet2(t *testing.T) {
	ipv6 := "24ae:0:0:0:0000:ffff:127:abc"
	p := netip.PrefixFrom(netip.MustParseAddr(ipv6), 120)
	t.Log(p)
}

func TestCheckIPV6Subnet2(t *testing.T) {
	data := "24ae:0:0:0:0000:ffff:127:8201"
	ipv6 := IPV6ToBigInt(data)
	t.Log(ipv6)
	t.Log(BigIntToIPV6(ipv6))
}

func TestFindIPV4(t *testing.T) {
	data := `部门：输入部门名称
        汇报人：输入“@+人名”提及相关人
        时间：11月1日-11月7日
        本周工作总结
        
        1.2.3.4 172.16.222.123
a3.4.5.6
		1a.35.test.10086.cn
		this is a test.
        做的好的
        `

	t.Log(FindIPV4(data))
}
