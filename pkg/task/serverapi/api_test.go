package serverapi

import (
	"encoding/json"
	"github.com/hanc00l/nemo_go/pkg/task/pocscan"
	"testing"
)

func TestNewTask(t *testing.T) {
	taskName := "xraypoc"
	config := pocscan.XrayPocConfig{
		IPPort: make(map[string][]int),
		Domain: make(map[string]struct{}),
	}
	config.IPPort["172.16.222.1"] = append(config.IPPort["172.16.222.1"], 8080)
	config.IPPort["172.16.222.1"] = append(config.IPPort["172.16.222.1"], 8000)
	config.Domain["localhost:8080"] = struct{}{}

	configJSON, _ := json.Marshal(config)
	taskId, err := NewRunTask(taskName, string(configJSON), "", "")
	if err != nil {
		t.Log(err)
	}
	t.Log(taskId)
}
