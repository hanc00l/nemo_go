package domainscan

import "testing"

func TestCrawler_RunCrawler(t *testing.T) {
	config := Config{Target: "www.800best.com,www.10086.cn"}
	c := NewCrawler(config)
	c.Do()
	t.Log(c.Result.DomainResult)
}
