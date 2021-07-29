package pocscan

import "testing"

func TestXray_RunXray(t *testing.T) {
	config := Config{
		Target:  "172.16.80.1:7001,127.0.0.1:7001",
		PocFile: "weblogic-cve-2020-14750.yml",
	}
	xray := NewXray(config)
	xray.Do()
	t.Log(xray.Result)

}

func TestXray_CheckXrayBinFile(t *testing.T) {
	xray := NewXray(Config{})
	t.Log(xray.CheckXrayBinFile())
}