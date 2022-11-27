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
		IsIPPortScan:       true,
		IsIPSubnetPortScan: false,
	}
	configJSON, err := json.Marshal(config)
	if err != nil {
		t.Log(err)
	}
	result, err := serverapi.NewRunTask("domainscan", string(configJSON), "", "")
	t.Log(result, err)
}

func TestDomainScanCrawler(t *testing.T) {
	config := domainscan.Config{
		Target:             "appl.800best.com",
		OrgId:              nil,
		IsSubDomainFinder:  false,
		IsSubDomainBrute:   false,
		IsCrawler:          true,
		IsHttpx:            true,
		IsIPPortScan:       true,
		IsIPSubnetPortScan: false,
	}
	configJSON, err := json.Marshal(config)
	if err != nil {
		t.Log(err)
	}
	result, err := serverapi.NewRunTask("domainscan", string(configJSON), "", "")
	t.Log(result, err)
}
func TestDomainWithPortScanScan(t *testing.T) {
	config := domainscan.Config{
		Target:             "appl.800best.com",
		OrgId:              nil,
		IsSubDomainFinder:  true,
		IsSubDomainBrute:   false,
		IsHttpx:            true,
		IsIPPortScan:       true,
		IsIPSubnetPortScan: false,
	}
	configJSON, err := json.Marshal(config)
	if err != nil {
		t.Log(err)
	}
	result, err := serverapi.NewRunTask("domainscan", string(configJSON), "", "")
	t.Log(result, err)
}
