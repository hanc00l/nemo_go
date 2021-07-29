package domainscan

import "testing"

func TestMassdns_Do(t *testing.T) {
	config := Config{Target: "800best.com"}
	m := NewMassdns(config)
	m.Do()
	resolve := NewResolve(Config{})
	resolve.Result.DomainResult = m.Result.DomainResult
	resolve.Do()
	for k,v :=range m.Result.DomainResult{
		t.Log(k)
		t.Log(v)
	}
	t.Log(len(m.Result.DomainResult))
}
