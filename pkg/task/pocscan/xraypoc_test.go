package pocscan

import (
	"testing"
)

func Test1(t *testing.T) {
	config := XrayPocConfig{
		IPPort: make(map[string][]int),
		Domain: make(map[string]struct{}),
	}
	config.IPPort["172.16.222.1"] = append(config.IPPort["172.16.222.1"], 8080)
	config.IPPort["172.16.222.1"] = append(config.IPPort["172.16.222.1"], 8000)
	config.Domain["localhost:8080"] = struct{}{}
	p := NewXrayPoc(config)
	p.Do()
	t.Log(p.VulResult)
}

func Test2(t *testing.T) {
	config := XrayPocConfig{}
	config.Domain["localhost:8080"] = struct{}{}
	p := NewXrayPoc(config)
	p.Do()
	t.Log(p.VulResult)
}
