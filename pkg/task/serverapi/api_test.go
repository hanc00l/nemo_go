package serverapi

import (
	"encoding/json"
	"github.com/hanc00l/nemo_go/pkg/task/pocscan"
	"testing"
)

func TestNewTask(t *testing.T) {
	taskName := "xraypoc"
	config := pocscan.XrayPocConfig{
		IPPortResult: make(map[string][]int),
	}
	config.IPPortResult["172.16.222.1"] = append(config.IPPortResult["172.16.222.1"], 8080)
	config.IPPortResult["172.16.222.1"] = append(config.IPPortResult["172.16.222.1"], 8000)
	config.DomainResult = append(config.DomainResult, "localhost:8080")

	configJSON, _ := json.Marshal(config)
	taskId, err := NewTask(taskName, string(configJSON), "")
	if err != nil {
		t.Log(err)
	}
	t.Log(taskId)
}
