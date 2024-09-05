package test

import (
	"github.com/hanc00l/nemo_go/v2/pkg/task/pocscan"
	"testing"
)

var rNucleiPocResult = map[string]map[string]bool{
	"127.0.0.1:18080": {
		"springboot-actuator": false,
	},
	"127.0.0.1:9200": {
		"elasticsearch": false,
	},
}

var rXrayPocResult = map[string]map[string]bool{
	"http://127.0.0.1:9200": {
		"poc-yaml-elasticsearch-unauth": false,
	},
}

func TestNuclei_test(t *testing.T) {
	config := pocscan.Config{
		// v3起不需要指定scheme了；使用http://xxxx会导致checkAndFormatUrl不正确解析
		Target:  "127.0.0.1:9200,127.0.0.1:18080",
		PocFile: "http/technologies/springboot-actuator.yaml",
	}
	n := pocscan.NewNuclei(config)
	n.Do()
	for _, r := range n.Result {
		if _, ok := rNucleiPocResult[r.Url]; ok {
			if _, ok2 := rNucleiPocResult[r.Url][r.PocFile]; ok2 {
				rNucleiPocResult[r.Url][r.PocFile] = true
			}
		}
	}
	config2 := pocscan.Config{
		// v3起不需要指定scheme了；使用http://xxxx会导致checkAndFormatUrl不正确解析
		Target:  "127.0.0.1:9200,127.0.0.1:18080",
		PocFile: "http/misconfiguration/elasticsearch.yaml",
	}
	n = pocscan.NewNuclei(config2)
	n.Do()
	for _, r := range n.Result {
		if _, ok := rNucleiPocResult[r.Url]; ok {
			if _, ok2 := rNucleiPocResult[r.Url][r.PocFile]; ok2 {
				rNucleiPocResult[r.Url][r.PocFile] = true
			}
		}
	}
	for url, r := range rNucleiPocResult {
		for poc, v := range r {
			if !v {
				t.Errorf("Nuclei poc test failed:%s %s ", url, poc)
			}
		}
	}
}

func TestXray_test(t *testing.T) {
	config := pocscan.Config{
		Target:  "127.0.0.1:9200",
		PocFile: "poc-yaml-elasticsearch-unauth",
	}
	xray := pocscan.NewXray(config)
	xray.Do()
	//t.Log(xray.Result)

	for _, r := range xray.Result {
		if _, ok := rXrayPocResult[r.Url]; ok {
			if _, ok2 := rXrayPocResult[r.Url][r.PocFile]; ok2 {
				rXrayPocResult[r.Url][r.PocFile] = true
			}
		}
	}
	for url, r := range rXrayPocResult {
		for poc, v := range r {
			if !v {
				t.Errorf("Xray poc test failed:%s %s ", url, poc)
			}
		}
	}
}
