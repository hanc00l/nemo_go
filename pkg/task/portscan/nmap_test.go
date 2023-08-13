package portscan

import (
	"github.com/hanc00l/nemo_go/pkg/logging"
	"os"
	"testing"
)

func TestNmap_Run(t *testing.T) {
	config := Config{
		Target:        "127.0.0.1,172.16.222.1",
		ExcludeTarget: "",
		Port:          "--top-ports 100",
		Rate:          1000,
		IsPing:        true,
		Tech:          "-sS",
		CmdBin:        "nmap",
	}
	nmap := NewNmap(config)
	nmap.Do()
	//nmap.Result.SaveResult(nmap.Config)

	t.Log(nmap.Result)
	for ip, ipa := range nmap.Result.IPResult {
		t.Log(ip, ipa)
		for port, pa := range ipa.Ports {
			t.Log(port, pa)
		}
	}
}

func TestNmap_ParseXMLResult(t *testing.T) {
	i := NewImportOfflineResult("nmap")
	content, err := os.ReadFile("/Users/user/Downloads/nmap2.xml")
	if err != nil {
		logging.RuntimeLog.Error(err)
		return
	}
	i.Parse(content)
	for ip, ipa := range i.IpResult.IPResult {
		t.Log(ip, ipa)
		for port, pa := range ipa.Ports {
			t.Log(port, pa)
		}
	}
}
