package pocscan

import "testing"

func TestGoby_StartScan(t *testing.T) {
	g := Goby{}
	ips := []string{"127.0.0.1:8161"}
	id, api, err := g.StartScan(ips)
	t.Log(id, err)
	err = g.GetAsset(api, id)
	t.Log(err)
	err = g.GetVulnerability(api, id)
	for _, v := range g.Result {
		t.Log(v)
	}
}

func TestGoby_Do(t *testing.T) {
	g := NewGoby(Config{Target: "192.168.50.217:8161,127.0.0.1:8161"})
	g.Do()
	for _, v := range g.Result {
		t.Log(v)
	}
}
