package fingerprint

import (
	"github.com/hanc00l/nemo_go/pkg/task/portscan"
	"github.com/hanc00l/nemo_go/pkg/utils"
	"testing"
)

func TestDoFullScreenshot(t *testing.T) {
	src := "/Users/user/Downloads/3.png"
	dst := "/Users/user/Downloads/3_alt.png"
	DoFullScreenshot("http://192.168.3.1", "/Users/user/Downloads/3.png")
	utils.ReSizePicture(src, dst,1024,0)
}

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
