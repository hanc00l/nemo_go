package fingerprint

import (
	"github.com/hanc00l/nemo_go/pkg/task/portscan"
	"testing"
)

func TestFingerprintHub_RunObserverWard(t *testing.T) {
	f := NewFingerprintHub(Config{Target: "183.141.20.136:8082"})
	rs := f.RunObserverWard("183.141.20.136:8082")
	for _,fp := range rs{
		t.Log(fp)
		for _,n := range fp.WhatWebName {
			t.Log(n)
		}
	}
}

func  TestFingerprintHub_Do(t *testing.T) {
	nmapConfig := portscan.Config{
		Target:       "183.60.156.84",
		Port:         "8088",
		Rate:         1000,
		IsPing:       false,
		Tech:         "-sS",
		CmdBin:       "nmap",
	}
	nmap := portscan.NewNmap(nmapConfig)
	nmap.Do()
	t.Log(nmap.Result)

	fp :=  NewFingerprintHub(Config{})
	fp.ResultPortScan = nmap.Result
	fp.Do()
	for _,r := range fp.ResultPortScan.IPResult{
		for port,p := range r.Ports{
			t.Log(port,p)
		}
	}
}
