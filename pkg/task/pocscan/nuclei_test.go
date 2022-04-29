package pocscan

import (
	"testing"
)

func TestNuclei_Do(t *testing.T) {
	config := Config{
		Target:  "http://127.0.0.1:7001,localhost:7001",
		PocFile: "cves/2020/CVE-2020-2551.yaml",
		//PocFile: "cves/2020",
	}
	n := NewNuclei(config)
	n.Do()
	for _,r := range n.Result{
		t.Log(r)
	}
}

func TestNuclei_LoadPocFile(t *testing.T) {
	n := NewNuclei(Config{})
	pocs := n.LoadPocFile()
	t.Log(pocs)
}