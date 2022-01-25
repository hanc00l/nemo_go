package workerapi

import (
	"encoding/json"
	"github.com/hanc00l/nemo_go/pkg/task/onlineapi"
	"github.com/hanc00l/nemo_go/pkg/task/serverapi"
	"testing"
)

func TestQuake(t *testing.T) {
	config := onlineapi.QuakeConfig{}
	config.Target = "800best.com"
	config.IsIPLocation = true
	configJSON, err := json.Marshal(config)
	if err != nil {
		t.Log(err)
	}
	result, err := serverapi.NewTask("quake", string(configJSON))
	t.Log(result,err)
}