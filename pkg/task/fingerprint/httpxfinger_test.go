package fingerprint

import (
	"github.com/hanc00l/nemo_go/pkg/task/portscan"
	"testing"
)

func TestHttpxFinger_DoHttpxAndFingerPrint(t *testing.T) {
	v := NewHttpxFinger()
	v.ResultPortScan = portscan.Result{
		IPResult: make(map[string]*portscan.IPResult),
	}
	v.ResultPortScan.SetIP("172.16.222.1")
	v.ResultPortScan.SetPort("172.16.222.1", 8080)

	v.DoHttpxAndFingerPrint()

	//t.Log(v.ResultPortScan.IPResult)
	for ip, r := range v.ResultPortScan.IPResult {
		t.Log(ip, r)
		for port, p := range r.Ports {
			t.Log(port, p)
		}
	}
}

func TestHttpxFinger_DoHttpxAndFingerPrint2(t *testing.T) {
	v := NewHttpxFinger()
	v.ResultPortScan = portscan.Result{
		IPResult: make(map[string]*portscan.IPResult),
	}
	v.ResultPortScan.SetIP("172.16.222.1")
	v.ResultPortScan.SetPort("172.16.222.1", 8000)

	v.DoHttpxAndFingerPrint()

	//t.Log(v.ResultPortScan.IPResult)
	for ip, r := range v.ResultPortScan.IPResult {
		t.Log(ip, r)
		for port, p := range r.Ports {
			t.Log(port, p)
		}
	}
}
