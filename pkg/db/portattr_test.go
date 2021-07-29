package db

import (
	"testing"
)

func TestPortAttr_Add(t *testing.T) {
	obj := PortAttr{
		RelatedId: 5,
		Source:    "nmap",
		Tag:       "service",
		Content:   "sqlserver",
	}
	success := obj.Add()
	t.Log(success, obj)
}

func TestPortAttr_GetsByRelatedId(t *testing.T) {
	obj := PortAttr{RelatedId: 6}
	objLists := obj.GetsByRelatedId()
	for _, o := range objLists {
		t.Log(o)
	}
}
