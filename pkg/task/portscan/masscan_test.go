package portscan

import (
	"github.com/hanc00l/nemo_go/pkg/logging"
	"os"
	"testing"
)

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
	t.Log(&m.Result)
	for ip, ipa := range m.Result.IPResult {
		t.Log(ip, ipa)
		for port, pa := range ipa.Ports {
			t.Log(port, pa)
		}
	}
}

func TestMasscan_ParseXMLResult(t *testing.T) {
	content, err := os.ReadFile("masscan.xml")
	if err != nil {
		logging.RuntimeLog.Error(err)
		return
	}
	i := NewImportOfflineResult("masscan")
	i.Parse(content)
	for ip, ipa := range i.IpResult.IPResult {
		t.Log(ip, ipa)
		for port, pa := range ipa.Ports {
			t.Log(port, pa)
		}
	}
}
