package db

import (
	"testing"
)

func TestIp_Add(t *testing.T) {
	ipObj := Ip{
		IpName:   "114.114.114.114",
		OrgId:    nil,
		Location: "江苏省南京",
		Status:   "open",
	}
	success := ipObj.Add()
	t.Log(success, ipObj)
}

func TestIp_GetByIp(t *testing.T) {
	ipObj := &Ip{IpName: "192.168.1.5"}
	if ipObj.GetByIp() {
		t.Log(ipObj.IpName, ipObj.Location, ipObj.OrgId)
		t.Log(ipObj.CreateDatetime.Format("2006-01-02 15:04"))
	}
	t.Log(ipObj)
}

func TestIp_Gets(t *testing.T) {
	ipObj := Ip{}
	searchMap := map[string]interface{}{}
	searchMap["domain"] = "10086"
	//searchMap["ip"] = "0.0.0.0/0"
	//searchMap["port"] = "80,443,445,3306,1433,33306"
	//searchMap["port_status"] = "200"
	//searchMap["date_delta"] = 2

	ipLists, count := ipObj.Gets(searchMap, 1, 10)
	t.Log(count)
	for _, ip := range ipLists {
		t.Log(ip)
	}
}
