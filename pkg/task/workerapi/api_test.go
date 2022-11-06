package workerapi

import (
	"testing"
)

func TestXScan(t *testing.T) {
	taskName := "xportscan"
	config := XScanConfig{
		OrgId:              nil,
		FofaKeyword:        "",
		FofaSearchLimit:    0,
		IPPort:             make(map[string][]int),
		IPPortString:       nil,
		Domain:             nil,
		IsSubDomainFinder:  false,
		IsSubDomainBrute:   false,
		IsIgnoreCDN:        false,
		IsIgnoreOutofChina: false,
		IsHttpx:            true,
		IsScreenshot:       true,
		IsFingerprintHub:   true,
		IsIconHash:         false,
		IsXrayPoc:          true,
	}
	config.IPPort["172.16.222.1"] = []int{8080, 8448, 8000, 3306}
	result, err := sendTask(config, taskName)
	if err != nil {
		t.Log(err)
	}
	t.Log(result)
}
