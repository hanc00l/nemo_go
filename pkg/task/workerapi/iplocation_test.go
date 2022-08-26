package workerapi

import (
	"encoding/json"
	"github.com/hanc00l/nemo_go/pkg/task/custom"
	"github.com/hanc00l/nemo_go/pkg/task/serverapi"
	"strings"
	"testing"
)

func TestIPLocation(t *testing.T) {
	config := custom.Config{Target: "39.98.21.49,101.37.114.100,39.96.249.185,101.37.114.107,112.13.170.75," +
		"124.90.39.17,47.96.49.97,112.13.170.59,112.13.170.43,112.13.170.24,47.92.183.90,47.98.65.38,39.101.139.234," +
		"60.205.219.245,52.139.216.101,101.201.67.227,121.196.213.232,202.107.193.29,47.110.188.170,120.78.133.119," +
		"114.55.201.35,120.77.111.48,47.97.237.236,101.37.120.146,202.107.193.74"}
	configJSON, err := json.Marshal(config)
	if err != nil {
		t.Log(err)
	}
	result, err := serverapi.NewTask("iplocation", string(configJSON), "")
	t.Log(result, err)

}

func TestIPLocation2(t *testing.T) {
	result1 := ""
	result2 := "ip:10"
	result3 := ""
	result4 := "screenshot:0"

	result := strings.Join([]string{result1, result2, result3, result4}, ",")

	t.Log(result)
	var m map[string]int
	m = nil
	t.Log(len(m))
}
