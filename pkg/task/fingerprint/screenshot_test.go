package fingerprint

import (
	"github.com/hanc00l/nemo_go/pkg/task/portscan"
	"testing"
)

func TestScreenShot_Do(t *testing.T) {
	config := portscan.Config{
		Target: "127.0.0.1",
		Port:   "8000,3306,1080,5000",
		Rate:   1000,
		Tech:   "-sS",
		CmdBin: "nmap",
	}
	nmap := portscan.NewNmap(config)
	nmap.Do()

	httpx := NewHttpxFinger()
	httpx.ResultPortScan = &nmap.Result
	httpx.Do()

	ss := NewScreenShot()
	ss.ResultPortScan = httpx.ResultPortScan
	ss.OptimizationMode = true
	ss.Do()

	for k, v := range ss.ResultScreenShot.Result {
		t.Log(k)
		for _, s := range v {
			t.Log(s)
		}
	}
}

func TestScreenShot_IPV6(t *testing.T) {
	config := portscan.Config{
		Target: "2400:dd01:103a:4041::101",
		Port:   "80",
		Rate:   1000,
		Tech:   "-sS",
		CmdBin: "masscan",
	}
	n := portscan.NewMasscan(config)
	n.Do()
	ss := NewScreenShot()
	ss.ResultPortScan = &n.Result
	ss.Do()
	for k, v := range ss.ResultScreenShot.Result {
		t.Log(k)
		for _, s := range v {
			t.Log(s)
		}
	}
}
