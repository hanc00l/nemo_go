package xray_v2_pocs_yml

import (
	"github.com/hanc00l/nemo_go/pkg/xraypocv2/check"
	common_structs "github.com/hanc00l/nemo_go/pkg/xraypocv2/pkg/common/structs"
	xray_requests "github.com/hanc00l/nemo_go/pkg/xraypocv2/pkg/xray/requests"
	"github.com/hanc00l/nemo_go/pkg/xraypocv2/pkg/xray/structs"
	"github.com/hanc00l/nemo_go/pkg/xraypocv2/utils"
	"sync"
	"time"
)

type xrayV2Poc struct {
	IsInit bool
	mutex  *sync.Mutex
}

var iXrayV2Poc = &xrayV2Poc{false, &sync.Mutex{}}

func InitXrayV2Poc(proxy string, ceyeapi string, ceyedomain string) *xrayV2Poc {
	iXrayV2Poc.mutex.Lock()
	defer iXrayV2Poc.mutex.Unlock()
	if !iXrayV2Poc.IsInit {
		iXrayV2Poc.initReverseAndHttpClient(proxy, ceyeapi, ceyedomain)
	}
	return iXrayV2Poc
}

func (p *xrayV2Poc) initReverseAndHttpClient(proxy string, ceyeapi string, ceyedomain string) {
	if !p.IsInit {
		common_structs.InitReversePlatform(ceyeapi, ceyedomain)
		_ = xray_requests.InitHttpClient(10, proxy, time.Duration(5)*time.Second)
	}
	p.IsInit = true
}

// 因为http302，需要传入多个Content进行匹配
func (p *xrayV2Poc) RunXrayCheckOneByQuery(target string, pocBytes []byte, contents []Content) bool {
	xrayPoc, err := utils.LoadPocByBytes(pocBytes)
	if err != nil {
		return false
	}
	if contents != nil && len(contents) > 0 {
		for _, content := range contents {
			if xrayPoc.Query == "" || xrayPoc.Query == "all" || MatchRules(*ParseRules(xrayPoc.Query), content) {
				xrayTotalReqeusts := 1
				xray_requests.InitCache(xrayTotalReqeusts)
				return check.XrayStartOne(target, xrayPoc)
			}
		}
	} else {
		xrayTotalReqeusts := 1
		xray_requests.InitCache(xrayTotalReqeusts)
		return check.XrayStartOne(target, xrayPoc)
	}
	return false
}

// 因为http302，需要传入多个Content进行匹配
func (p *xrayV2Poc) RunXrayMultiPocByQuery(target string, pocsBytes [][]byte, contents []Content) []string {
	pocs := p.LoadMultiPocs(pocsBytes)

	pocs_new := []*structs.Poc{}
	if contents != nil && len(contents) > 0 {
		for _, xrayPoc := range pocs {
			for _, content := range contents {
				if xrayPoc.Query == "" || xrayPoc.Query == "all" || MatchRules(*ParseRules(xrayPoc.Query), content) {
					pocs_new = append(pocs_new, xrayPoc)
				}
			}
		}
	} else {
		pocs_new = pocs
	}
	if len(pocs_new) > 0 {
		return p.XrayCheck(target, pocs_new)
	}
	return []string{}
}
func (p *xrayV2Poc) XrayCheckByQuery(target string, xrayPocs []*structs.Poc, contents []Content) []string {
	pocs_new := []*structs.Poc{}
	if contents != nil && len(contents) > 0 {
		for _, xrayPoc := range xrayPocs {
			for _, content := range contents {
				if xrayPoc.Query == "" || xrayPoc.Query == "all" || MatchRules(*ParseRules(xrayPoc.Query), content) {
					pocs_new = append(pocs_new, xrayPoc)
				}
			}
		}
	} else {
		pocs_new = xrayPocs
	}

	xrayTotalReqeusts := 0
	for _, poc := range pocs_new {
		ruleLens := len(poc.Rules)
		if poc.Transport == "tcp" || poc.Transport == "udp" {
			ruleLens += 1
		}
		xrayTotalReqeusts += 1 * ruleLens
	}
	if xrayTotalReqeusts == 0 {
		xrayTotalReqeusts = 1
	}
	xray_requests.InitCache(xrayTotalReqeusts)
	return check.XrayStart(target, pocs_new)
}
func (p *xrayV2Poc) RunXrayMultiPocs(target string, pocsBytes [][]byte) []string {
	pocs := p.LoadMultiPocs(pocsBytes)
	if len(pocs) > 0 {
		return p.XrayCheck(target, pocs)
	}
	return []string{}
}

func (p *xrayV2Poc) RunXrayCheckOne(target string, pocBytes []byte) bool {
	xrayPoc, err := utils.LoadPocByBytes(pocBytes)
	if err != nil {
		return false
	}
	xrayTotalReqeusts := 1
	xray_requests.InitCache(xrayTotalReqeusts)
	return check.XrayStartOne(target, xrayPoc)
}

func (p *xrayV2Poc) LoadOnePoc(pocBytes []byte) *structs.Poc {
	xrayPoc, err := utils.LoadPocByBytes(pocBytes)
	if err != nil {
		return nil
	}
	return xrayPoc
}
func (p *xrayV2Poc) LoadMultiPocs(pocsBytes [][]byte) []*structs.Poc {
	var pocs []*structs.Poc
	for _, f := range pocsBytes {
		if poc, err := utils.LoadPocByBytes(f); err == nil {
			pocs = append(pocs, poc)
		}
	}
	return pocs
}

func (p *xrayV2Poc) XrayCheck(target string, xrayPocs []*structs.Poc) []string {

	xrayTotalReqeusts := 0
	for _, poc := range xrayPocs {
		ruleLens := len(poc.Rules)
		if poc.Transport == "tcp" || poc.Transport == "udp" {
			ruleLens += 1
		}
		xrayTotalReqeusts += 1 * ruleLens
	}
	if xrayTotalReqeusts == 0 {
		xrayTotalReqeusts = 1
	}
	xray_requests.InitCache(xrayTotalReqeusts)
	return check.XrayStart(target, xrayPocs)
}
