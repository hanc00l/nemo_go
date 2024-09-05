package pocscan

import (
	"testing"
)

func TestXray_RunXray(t *testing.T) {
	config := Config{
		Target:  "172.16.80.1:7001,127.0.0.1:7001",
		PocFile: "weblogic-cve-2020-14750.yml",
	}
	xray := NewXray(config)
	xray.Do()
	t.Log(xray.Result)

}

func TestXray_RunXray2(t *testing.T) {
	config := Config{
		Target:  "172.16.222.1:8848",
		PocFile: "*",
	}
	xray := NewXray(config)
	xray.Do()
	t.Log(xray.Result)

}

func TestXray_RunXray3(t *testing.T) {
	config := Config{
		Target:  "172.16.222.1:8161",
		PocFile: "default|",
	}
	xray := NewXray(config)
	xray.Do()
	t.Log(xray.Result)
}
