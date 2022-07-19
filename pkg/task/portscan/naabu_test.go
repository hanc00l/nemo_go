package portscan

import "testing"

func TestNaabu_ParseTxtContentResult(t *testing.T) {
	data := `172.16.222.1:1080
172.16.222.1:8000`
	naabu := NewNaabu(Config{})
	naabu.ParseTxtContentResult([]byte(data))
	for k, ip := range naabu.Result.IPResult {
		t.Log(k)
		for kk, port := range ip.Ports {
			t.Log(kk, port.Status)
			t.Log(port.PortAttrs)
		}
	}
}
