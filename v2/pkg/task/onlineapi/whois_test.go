package onlineapi

import (
	"encoding/json"
	"testing"
)

func TestWhois_Do(t *testing.T) {
	w := NewWhois(WhoisQueryConfig{Target: "10086.cn,mirages.tech"})
	w.Do()
	r1 :=w.LookupWhois("10086.cn")
	r1Txt,_:=json.Marshal(r1)
	t.Log(string(r1Txt))
}

