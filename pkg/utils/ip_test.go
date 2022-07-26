package utils

import "testing"

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
}

func TestCheckIPLocationInChinaMainLand(t *testing.T) {
	data1 := "香x港 台湾 澳门"
	data2 := "德国 香港"
	data3 := "中国 香港"
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
