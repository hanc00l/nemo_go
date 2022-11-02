package fingerprint

import (
	"github.com/hanc00l/nemo_go/pkg/task/portscan"
	"testing"
)

func TestScreenShot_Do(t *testing.T) {
	config := portscan.Config{
		Target: "127.0.0.1",
		Port:   "5000",
		Rate:   1000,
		Tech:   "-sS",
		CmdBin: "nmap",
	}
	nmap := portscan.NewNmap(config)
	nmap.Do()
	ss := NewScreenShot()
	ss.ResultPortScan = nmap.Result
	ss.Do()
	for k, v := range ss.ResultScreenShot.Result {
		t.Log(k)
		for _, s := range v {
			t.Log(s)
		}
	}
}
