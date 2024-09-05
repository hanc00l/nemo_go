package test

import (
	"github.com/hanc00l/nemo_go/v2/pkg/task/fingerprint"
	"github.com/hanc00l/nemo_go/v2/pkg/task/portscan"
	"testing"
)

var rFinger = map[string]map[int]map[string]map[string]bool{
	"127.0.0.1": {
		80: {
			"server": {
				"PHP/8.3": false,
			},
			"service": {
				"http": false,
			},
			"favicon": {
				"673152537 | http://127.0.0.1:80/favicon.ico": false,
			},
			"title": {
				"在线资产管理平台": false,
			},
			// 测试fingerprinthub的被动解析
			"fingerprint": {
				"奇安信-天擎": false,
			},
			"status": {
				"200": false,
			},
		},
		443: {
			"tlsdata": {
				`{"subject_an":["localhost"],"subject_cn":"127.0.0.1","subject_dn":"CN=127.0.0.1","issuer_dn":"CN=127.0.0.1"}`: false,
			},
		},
	},
}

var rFingeprintx = map[string]map[int]map[string]map[string]bool{
	"127.0.0.1": {
		22: {
			"service": {
				"ssh": false,
			},
		},
		3306: {
			"service": {
				"mysql": false,
			},
			"version": {
				"5.7.44": false,
			},
		},
	},
}

func TestHttp_test(t *testing.T) {
	v := fingerprint.NewHttpxAll()
	v.ResultPortScan = &portscan.Result{
		IPResult: make(map[string]*portscan.IPResult),
	}
	v.ResultPortScan.SetIP("127.0.0.1")
	v.ResultPortScan.SetPort("127.0.0.1", 80)
	v.ResultPortScan.SetPort("127.0.0.1", 443)
	v.ResultPortScan.SetPort("127.0.0.1", 9200)
	v.Do()

	for _, r := range v.ResultPortScan.IPResult {
		for port, p := range r.Ports {
			//t.Log(port, p)
			if _, ok := rFinger["127.0.0.1"][port]; ok {
				for _, pa := range p.PortAttrs {
					if _, ok3 := rFinger["127.0.0.1"][port][pa.Tag]; ok3 {
						if _, ok4 := rFinger["127.0.0.1"][port][pa.Tag][pa.Content]; ok4 {
							rFinger["127.0.0.1"][port][pa.Tag][pa.Content] = true
						}
					}
				}
			}
		}
	}
	//t.Log(rFinger)
	for ip, r := range rFinger {
		for port, p := range r {
			for tag, pa := range p {
				for content, ok := range pa {
					if !ok {
						t.Errorf("finger not found %s:%d %s:%s", ip, port, tag, content)
					}
				}
			}
		}
	}

	if len(v.ResultScreenShot.Result["127.0.0.1"]) < 3 {
		t.Errorf("%s screenshot failed", "127.0.0.1")
	}
}

func TestFingeprintx_test(t *testing.T) {

	ipResult := &portscan.Result{
		IPResult: make(map[string]*portscan.IPResult),
	}
	ipResult.SetIP("127.0.0.1")
	ipResult.SetPort("127.0.0.1", 22)
	ipResult.SetPort("127.0.0.1", 3306)

	f := fingerprint.NewFingerprintx()
	f.ResultPortScan = ipResult
	f.Do()

	for _, r := range f.ResultPortScan.IPResult {
		for port, p := range r.Ports {
			//t.Log(port, p)
			if _, ok := rFingeprintx["127.0.0.1"][port]; ok {
				for _, pa := range p.PortAttrs {
					if _, ok3 := rFingeprintx["127.0.0.1"][port][pa.Tag]; ok3 {
						if _, ok4 := rFingeprintx["127.0.0.1"][port][pa.Tag][pa.Content]; ok4 {
							rFingeprintx["127.0.0.1"][port][pa.Tag][pa.Content] = true
						}
					}
				}
			}
		}
	}
	//t.Log(rFingeprintx)
	for ip, r := range rFingeprintx {
		for port, p := range r {
			for tag, pa := range p {
				for content, ok := range pa {
					if !ok {
						t.Errorf("fingeprintx not found %s:%d %s:%s", ip, port, tag, content)
					}
				}
			}
		}
	}
}
