package db

import (
	"testing"
	"time"
)

func TestDomain_Add(t *testing.T) {
	domain := Domain{
		DomainName:     "10086.cn",
		OrgId:          nil,
		CreateDatetime: time.Time{},
		UpdateDatetime: time.Time{},
	}
	success := domain.Add()
	t.Log(success, domain)
}

func TestDomain_GetsBySearch(t *testing.T) {
	searchMap := map[string]interface{}{}
	searchMap["domain"] = "10086"

	domain := &Domain{}
	domainLists, _ := domain.Gets(searchMap, 1, 10, false)
	for _, o := range domainLists {
		t.Log(o)
	}
}
