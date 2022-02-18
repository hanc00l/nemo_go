package domainscan

import "testing"

func TestCrawler_RunCrawler(t *testing.T) {
	config := Config{Target: "www.800best.com,www.10086.cn"}
	c := NewCrawler(config)
	c.Do()
	t.Log(c.Result.DomainResult)
}

func TestCrawler_CheckRequest(t *testing.T) {
	//url := "https://www.800best.com/"
	//url := "http://47.111.243.170:5000"
	//url := "http://youyierp.800best.com/"
	url := "https://www.baidu.com"
	c := NewCrawler(Config{})
	c.RunCrawler2(url)
	for _,u := range c.Result.ReqResponseList{
		t.Log(u.Method,u.URL,u.StatusCode,u.ContentLength,u.RedirectLocation)
	}
}