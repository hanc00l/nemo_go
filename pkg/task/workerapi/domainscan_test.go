package workerapi

import (
	"encoding/json"
	"github.com/hanc00l/nemo_go/pkg/task/domainscan"
	"github.com/hanc00l/nemo_go/pkg/task/serverapi"
	"testing"
)

func TestDomainScan(t *testing.T) {
	config := domainscan.Config{
		Target:             "appl.800best.com",
		OrgId:              nil,
		IsSubDomainFinder:  true,
		IsSubDomainBrute:   false,
		IsHttpx:            true,
		IsWhatWeb:          false,
		IsIPPortScan:       true,
		IsIPSubnetPortScan: false,
	}
	configJSON, err := json.Marshal(config)
	if err != nil {
		t.Log(err)
	}
	result, err := serverapi.NewTask("domainscan", string(configJSON))
	t.Log(result,err)
}

func TestDomainWithPortScanScan(t *testing.T) {
	config := domainscan.Config{
		Target:             "appl.800best.com",
		OrgId:              nil,
		IsSubDomainFinder:  true,
		IsSubDomainBrute:   false,
		IsHttpx:            true,
		IsWhatWeb:          false,
		IsIPPortScan:       true,
		IsIPSubnetPortScan: false,
	}
	configJSON, err := json.Marshal(config)
	if err != nil {
		t.Log(err)
	}
	result, err := serverapi.NewTask("domainscan", string(configJSON))
	t.Log(result,err)
}