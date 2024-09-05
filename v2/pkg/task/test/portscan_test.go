package test

import (
	"github.com/hanc00l/nemo_go/v2/pkg/task/portscan"
	"testing"
)

func init() {
	/*
		 nemo_test用于测试的server端口
		 go test -v -count=1 .
			21 ftp
			22 ssh
			23 telnet
			80 http
			443 https
			5900 vnc
			9200 elasticsearch
	*/
}

func TestNmap_test(t *testing.T) {
	doPortscan("nmap", t)
}

//func TestMasscan_test(t *testing.T) {
//	doPortscan("masscan", t)
//}

func TestGogo_test(t *testing.T) {
	doPortscan("gogo", t)
}

func doPortscan(cmdBin string, t *testing.T) {
	var rPortscan = map[string]map[int]bool{
		"127.0.0.1": {
			21:   false,
			22:   false,
			23:   false,
			80:   false,
			443:  false,
			5900: false,
			9200: false,
		},
	}
	config := portscan.Config{
		Target:        "127.0.0.1,172.16.222.1",
		ExcludeTarget: "",
		Port:          "--top-ports 1000",
		Rate:          1000,
		IsPing:        true,
		Tech:          "-sS",
		CmdBin:        cmdBin,
	}
	var result map[string]*portscan.IPResult
	if cmdBin == "nmap" {
		nmap := portscan.NewNmap(config)
		nmap.Do()
		result = nmap.Result.IPResult
	} else if cmdBin == "masscan" {
		masscan := portscan.NewMasscan(config)
		masscan.Do()
		result = masscan.Result.IPResult
	} else if cmdBin == "gogo" {
		gogo := portscan.NewGogo(config)
		gogo.Do()
		result = gogo.Result.IPResult
	} else {
		t.Errorf("invalid cmdbin:%s", cmdBin)
		return
	}

	for ip, ipa := range result {
		//t.Log(ip, ipa)
		for port, _ := range ipa.Ports {
			//t.Log(port, result[ip].Ports[port])
			if ip == "127.0.0.1" {
				if _, exist := rPortscan[ip][port]; exist {
					rPortscan[ip][port] = true
				}
			}
		}
	}
	//check result
	for port, status := range rPortscan["127.0.0.1"] {
		if !status {
			t.Errorf("port %d not checked...", port)
		}
	}
}
