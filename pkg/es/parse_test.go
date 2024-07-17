package es

import (
	"encoding/json"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/search"
	"testing"
)

func TestParseQuery(t *testing.T) {
	expr := `(ip=="192.168.120.1" || ip=="192.168.120.2") && status>200`
	query, err := ParseQuery(expr)
	if err != nil {
		t.Errorf("parse query error: %v", err)
		return
	}
	a := NewAssets("test1")
	res, err := a.Search(query, 1, 10)
	if err != nil {
		t.Errorf("search error: %v", err)
		return
	}
	outputDoc(res, t)
}

func outputSimple(res *search.Response, t *testing.T) {
	t.Logf("total:%d", res.Hits.Total.Value)
	for _, hit := range res.Hits.Hits {
		var doc Document
		json.Unmarshal(hit.Source_, &doc)
		t.Logf("Host:%s, Domain:%s, IP:%v", doc.Host, doc.Domain, doc.Ip)
	}
}

func TestParseQuery1(t *testing.T) {
	ws := map[int]string{
		1: "b0c79065-7ff7-32ae-cc18-864ccd8f7717",
	}
	expr := `domain=="fofa.info"`
	query, err := ParseQuery(expr)
	if err != nil {
		t.Errorf("parse query error: %v", err)
		return
	}
	a := NewAssets(ws[1])
	res, err := a.Search(query, 1, 10)
	if err != nil {
		t.Errorf("search error: %v", err)
		return
	}
	outputSimple(res, t)
}
