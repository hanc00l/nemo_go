package workerapi

import (
	"encoding/json"
	"github.com/hanc00l/nemo_go/v2/pkg/task/pocscan"
	"github.com/hanc00l/nemo_go/v2/pkg/task/serverapi"
	"testing"
)

func TestXray(t *testing.T) {
	config := pocscan.Config{
		Target:  "172.16.80.1:7001,127.0.0.1:7001,192.168.120.160:7001",
		PocFile: "weblogic-cve-2020-14750.yml",
		CmdBin:  "xray",
	}
	configJSON, err := json.Marshal(config)
	if err != nil {
		t.Log(err)
	}
	result, err := serverapi.NewRunTask("pocscan", string(configJSON), "", "")
	t.Log(result, err)
}
