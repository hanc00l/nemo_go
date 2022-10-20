package pocscan

import (
	"testing"
)

func Test1(t *testing.T) {
	config := XrayPocConfig{
		IPPortResult: make(map[string][]int),
	}
	config.IPPortResult["172.16.222.1"] = append(config.IPPortResult["172.16.222.1"], 8080)
	config.IPPortResult["172.16.222.1"] = append(config.IPPortResult["172.16.222.1"], 8000)
	config.DomainResult = append(config.DomainResult, "localhost:8080")
	p := NewXrayPoc(config)
	p.Do()
	t.Log(p.VulResult)
}

func Test2(t *testing.T) {
	config := XrayPocConfig{}
	config.DomainResult = append(config.DomainResult, "localhost:8080")
	p := NewXrayPoc(config)
	p.Do()
	t.Log(p.VulResult)
}
