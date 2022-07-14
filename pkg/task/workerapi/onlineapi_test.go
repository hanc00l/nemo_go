package workerapi

import (
	"encoding/json"
	"github.com/hanc00l/nemo_go/pkg/task/onlineapi"
	"github.com/hanc00l/nemo_go/pkg/task/serverapi"
	"testing"
)

func TestFofa(t *testing.T) {
	orgId := 5
	config := onlineapi.OnlineAPIConfig{
		Target: "800best.com",
		OrgId:  &orgId,
		IsIPLocation: true,
	}
	configJSON, err := json.Marshal(config)
	if err != nil {
		t.Log(err)
	}
	result, err := serverapi.NewTask("fofa", string(configJSON),"")
	t.Log(result,err)
}
