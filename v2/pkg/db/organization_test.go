package db

import (
	"testing"
)

func TestOrganization_Get(t *testing.T) {
	obj := Organization{Id: 5}

	if result := obj.Get(); result {
		t.Log(obj)
		t.Log(obj.OrgName)
		t.Log(obj.CreateDatetime.Format("2006-01-02 15:04"))
	}
}

func TestOrganization_Gets(t *testing.T) {
	obj := &Organization{}

	searchMap := map[string]interface{}{}
	searchMap["org_name like ?"] = "%电%"
	searchMap["sort_order"] = 200
	orgs := obj.Gets(searchMap, 2, 2)

	for _, org := range orgs {
		t.Log(org.OrgName)
		t.Log(org.CreateDatetime.Format("2006-01-02 15:04"))
	}
}

func TestOrganization_Add(t *testing.T) {
	obj := &Organization{
		OrgName:   "test",
		Status:    "disable",
		SortOrder: 200,
	}
	success := obj.Add()
	t.Log(success, obj)
}

func TestOrganization_Update(t *testing.T) {
	obj := &Organization{Id: 8}
	updatedMap := map[string]interface{}{}
	updatedMap["org_name"] = "test2.0"
	updatedMap["status"] = "enable"

	n := obj.Update(updatedMap)
	t.Log(n)
}

func TestOrganization_Count(t *testing.T) {
	obj := &Organization{}
	searchMap := map[string]interface{}{"org_name like ?": "%电%", "sort_order": 200}
	count := obj.Count(searchMap)
	t.Log(count)
}
