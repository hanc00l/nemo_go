package db

import (
	"testing"
)

func TestIpAttr_Add(t *testing.T) {
	obj := IpAttr{
		RelatedId: 10,
		Source:    "httpx",
		Tag:       "httpx",
		Content:   "cms test",
	}
	success := obj.Add()
	t.Log(success, obj)
}

func TestIpAttr_Update(t *testing.T) {
	obj := IpAttr{Id: 1}
	updatedMap := map[string]interface{}{"source": "masscan"}

	obj.Update(updatedMap)
}

func TestIpAttr_GetsByRelatedId(t *testing.T) {
	obj := IpAttr{RelatedId: 10}
	ipattrLists := obj.GetsByRelatedId()
	for _, ipa := range ipattrLists {
		t.Log(ipa.Id, ipa.RelatedId, ipa.Tag, ipa.Source, ipa.Content, ipa.Hash, ipa.CreateDatetime)
	}
}
