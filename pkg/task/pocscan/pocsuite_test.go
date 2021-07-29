package pocscan

import "testing"

func TestPocsuite_RunPocsuite(t *testing.T) {
	p := NewPocsuite(Config{Target: "172.16.80.1:7001,127.0.0.1,192.168.3.223", PocFile: "weblogic-console-2020-14882_all_rce.py"})
	p.Do()
	SaveResult(p.Result)

	t.Log(p.Result)
}
