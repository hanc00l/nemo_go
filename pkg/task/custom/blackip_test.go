package custom

import "testing"

func TestBlackIP_CheckBlack(t *testing.T) {
	ips := []string{"127.0.0.1", "172.16.3.1", "114.114.114.115", "192.168.1.65"}
	b := NewBlackIP()
	for _, d := range ips {
		t.Log(d, b.CheckBlack(d))
	}
}
