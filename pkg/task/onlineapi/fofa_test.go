package onlineapi

import "testing"

func TestFofa_Run(t *testing.T) {
	config1 := FofaConfig{Target: "47.98.181.116"}
	fofa1 := NewFofa(config1)
	fofa1.Do()
	fofa1.SaveResult()
}
func TestFofa_Run2(t *testing.T) {
	config2 := FofaConfig{Target: "800best.com"}
	//config2 := FofaConfig{Target: "10086.cn"}
	fofa2 := NewFofa(config2)
	fofa2.Do()
	//t.Log(fofa2.SaveResult())
	t.Log(fofa2.Result)
	t.Log(fofa2.IpResult)

	for ip, ipr := range fofa2.IpResult.IPResult {
		t.Log(ip, ipr)
		for port, pat := range ipr.Ports {
			t.Log(port, pat)
		}
	}
	t.Log(fofa2.DomainResult)
	for domain, dar := range fofa2.DomainResult.DomainResult {
		t.Log(domain, dar)
		//for _, a := range dar.DomainAttrs {
		//	t.Log(a)
		//}
	}
}
