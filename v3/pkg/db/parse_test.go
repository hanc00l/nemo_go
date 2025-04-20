package db

import (
	"testing"
)

func TestParseQuery(t *testing.T) {
	//expr := `(ip=="192.168.120.1" || ip=="192.168.120.2") && status=="200"`
	//expr := `(host=="192.168.3.1" && service=="https") || (host=="192.168.3.215" && service=="https")`
	expr := `ip=="192.168.3.128/24" && new=="false"`
	query, err := ParseQuery(expr)
	if err != nil {
		t.Errorf("parse query error: %v", err)
		return
	}
	t.Log(query)
}
