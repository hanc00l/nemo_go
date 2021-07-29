package onlineapi

import (
	"testing"
)

func TestICPQuery_Do(t *testing.T) {
	config := ICPQueryConfig{Target: "10086.cn"}
	icp := NewICPQuery(config)
	icp.Do()
	t.Log(icp.UploadICPInfo())
	icpInfo := icp.LookupICP("800best.com")
	t.Log(*icpInfo)
}
