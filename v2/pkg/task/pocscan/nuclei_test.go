package pocscan

import (
	"testing"
)

func TestNuclei_Do(t *testing.T) {
	config := Config{
		// v3起不需要指定scheme了；使用http://xxxx会导致checkAndFormatUrl不正确解析
		Target: "http://127.0.0.1:7001,127.0.0.1:8000",
		//PocFile: "http/cves/2020/CVE-2020-2551.yaml",
		PocFile: "http/technologies/springboot-actuator.yaml",
	}
	n := NewNuclei(config)
	n.Do()
	for _, r := range n.Result {
		t.Log(r)
	}
}

func TestNuclei_LoadPocFile(t *testing.T) {
	n := NewNuclei(Config{})
	pocs := n.LoadPocFile()
	t.Log(pocs)
}
