package pocscan

import (
	"testing"
)

func TestDirSearch(t *testing.T){

	d := NewDirsearch(Config{
		Target: "127.0.0.1:8000,172.16.80.130",
		PocFile: "php",
	})
	d.Do()
	t.Log(d.Result)
}
