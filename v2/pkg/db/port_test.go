package db

import (
	"testing"
)

func TestPort_Add(t *testing.T) {
	obj := Port{
		IpId:           19,
		PortNum:        443,
		Status:         "200",
	}
	success := obj.Add()
	t.Log(success,obj)
}


func TestPort_Update(t *testing.T) {
	obj := Port{Id:6}
	//obj.Get()
	//t.Log(obj)

	updatedMap := map[string]interface{}{}
	updatedMap["status"] = "200"
	obj.Update(updatedMap)

	//obj.Get()
	t.Log(obj)
}

