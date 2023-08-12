package portscan

import "testing"

func TestMasscan_Run(t *testing.T) {
	config := Config{
		Target:        "192.168.120.1",
		ExcludeTarget: "",
		Port:          "--top-ports 1000",
		Rate:          1000,
		IsPing:        true,
		Tech:          "-sS",
		CmdBin:        "masscan",
	}
	m := NewMasscan(config)
	m.Do()
	t.Log(m.Result)
	for ip, ipa := range m.Result.IPResult {
		t.Log(ip, ipa)
		for port, pa := range ipa.Ports {
			t.Log(port, pa)
		}
	}
}

func TestMasscan_ParseXMLResult(t *testing.T) {
	m := NewMasscan(Config{})
	m.ParseXMLResult("/Users/user/Downloads/masscan.xml")
	for ip, ipa := range m.Result.IPResult {
		t.Log(ip, ipa)
		for port, pa := range ipa.Ports {
			t.Log(port, pa)
		}
	}
}
