package domainscan

import (
	"testing"
)

func TestRun(t *testing.T) {
	config := Config{
		Target: "appl.800best.com",
	}
	subdomain := NewSubFinder(config)
	subdomain.Do()
	resolve := NewResolve(config)
	resolve.Result = subdomain.Result
	resolve.Do()

	for domain, da := range resolve.Result.DomainResult {
		t.Log(domain, da)
	}
	//subdomain.Result.SaveResult(config)
}
