package workerapi

import (
	"encoding/json"
	"github.com/hanc00l/nemo_go/pkg/task/portscan"
	"github.com/hanc00l/nemo_go/pkg/task/serverapi"
	"testing"
)

func TestPortScan(t *testing.T) {
	config := portscan.Config{
		Target:        "192.168.3.0/24",
		ExcludeTarget: "",
		Port:          "--top-ports 100",
		OrgId:         nil,
		Rate:          1000,
		IsPing:        true,
		Tech:          "-sS",
		IsIpLocation:  true,
		IsHttpx:       true,
		IsWhatWeb:     false,
		CmdBin:        "masnmap",
	}

	configJSON, err := json.Marshal(config)
	if err != nil {
		t.Log(err)
	}

	result, err := serverapi.NewTask("portscan", string(configJSON))
	if err != nil {
		t.Log(err)
	}
	t.Log(result)
}
