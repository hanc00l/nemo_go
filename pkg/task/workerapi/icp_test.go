package workerapi

import (
	"encoding/json"
	"github.com/hanc00l/nemo_go/pkg/task/onlineapi"
	"github.com/hanc00l/nemo_go/pkg/task/serverapi"
	"testing"
)

func TestICPQuery(t *testing.T) {
	config := onlineapi.ICPQueryConfig{
		Target: "163.com",
	}
	configJSON, err := json.Marshal(config)
	if err != nil {
		t.Log(err)
	}
	result, err := serverapi.NewTask("icpquery", string(configJSON))
	t.Log(result,err)
}
