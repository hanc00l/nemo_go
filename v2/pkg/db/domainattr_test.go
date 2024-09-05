package db

import (
	"testing"
)


func TestDomainAttr_GetsByRelatedId(t *testing.T) {
	domainattr := DomainAttr{RelatedId: 4175}

	dLists := domainattr.GetsByRelatedId()
	for _,o := range dLists{
		t.Log(o)
	}
}