package fingerprint

import (
	"github.com/hanc00l/nemo_go/pkg/task/portscan"
	"testing"
)

func TestFingerprintx_RunFingerprintx(t *testing.T) {
	f := NewFingerprintx()
	result := f.RunFingerprintx("127.0.0.1:3306")
	for _, r := range result {
		t.Log(r)
	}
}

func TestFingerprintx_Do(t *testing.T) {
	v := NewHttpxAll()
	v.ResultPortScan = &portscan.Result{
		IPResult: make(map[string]*portscan.IPResult),
	}
	v.ResultPortScan.SetIP("127.0.0.1")
	v.ResultPortScan.SetPort("127.0.0.1", 8000)
	v.ResultPortScan.SetPort("127.0.0.1", 3306)
	v.ResultPortScan.SetPort("127.0.0.1", 4369)
	v.Do()

	f := NewFingerprintx()
	f.ResultPortScan = v.ResultPortScan
	f.Do()

	for ip, r := range f.ResultPortScan.IPResult {
		t.Log(ip, r)
		for port, p := range r.Ports {
			t.Log(port, p)
		}
	}
}
