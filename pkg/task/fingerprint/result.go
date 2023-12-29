package fingerprint

import (
	"fmt"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/task/domainscan"
	"github.com/hanc00l/nemo_go/pkg/task/portscan"
	"strings"
	"sync"
)

var IgnorePort = []int{7, 9, 13, 17, 19, 21, 22, 23, 25, 26, 37, 53, 100, 106, 110, 111, 113, 119, 135, 138, 139,
	143, 144, 145, 161,
	179, 199, 389, 427, 444, 445, 514, 515, 543, 554, 631, 636, 646, 880, 902, 990, 993,
	1433, 1521, 3306, 5432, 3389, 5900, 5901, 5902, 49152, 49153, 49154, 49155, 49156, 49157,
	49158, 49159, 49160, 49161, 49163, 49165, 49167, 49175, 49176,
	13306, 11521, 15432, 11433, 13389, 15900, 15901}

var blankPort = map[int]struct{}{}

var (
	fpHttpxThreadNumber        = make(map[string]int)
	fpScreenshotThreadNum      = make(map[string]int)
	fpObserverWardThreadNumber = make(map[string]int)
	fpIconHashThreadNumber     = make(map[string]int)
	fpFingerprintxThreadNumber = make(map[string]int)
)

func init() {
	fpHttpxThreadNumber[conf.HighPerformance] = 8
	fpHttpxThreadNumber[conf.NormalPerformance] = 4
	//
	fpScreenshotThreadNum[conf.HighPerformance] = 6
	fpScreenshotThreadNum[conf.NormalPerformance] = 3
	//
	fpObserverWardThreadNumber[conf.HighPerformance] = 8
	fpObserverWardThreadNumber[conf.NormalPerformance] = 4
	//
	fpIconHashThreadNumber[conf.HighPerformance] = 8
	fpIconHashThreadNumber[conf.NormalPerformance] = 4
	//
	fpFingerprintxThreadNumber[conf.HighPerformance] = 6
	fpFingerprintxThreadNumber[conf.NormalPerformance] = 3
}

type FingerAttrResult struct {
	Tag     string
	Content string
}

type ScreenshotInfo struct {
	Port         int
	Protocol     string
	FilePathName string
}

type ScreenshotResult struct {
	sync.RWMutex
	Result map[string][]ScreenshotInfo
}

func (r *ScreenshotResult) HasDomain(domain string) bool {
	r.RLock()
	defer r.RUnlock()

	_, ok := r.Result[domain]
	return ok
}

func (r *ScreenshotResult) SetDomain(domain string) {
	r.Lock()
	defer r.Unlock()

	r.Result[domain] = []ScreenshotInfo{}
}

func (r *ScreenshotResult) SetScreenshotInfo(domain string, si ScreenshotInfo) {
	r.Lock()
	defer r.Unlock()

	if _, ok := r.Result[domain]; !ok {
		r.Result[domain] = []ScreenshotInfo{}
	}
	r.Result[domain] = append(r.Result[domain], si)
}

func init() {
	blankPort = make(map[int]struct{})
	for _, p := range IgnorePort {
		blankPort[p] = struct{}{}
	}
}

func CheckForHttpxFingerResult(ip string, domain string, port int, resultPortscan *portscan.Result, resultDomainScan *domainscan.Result) bool {
	if ip != "" && resultPortscan != nil && resultPortscan.IPResult != nil {
		for _, par := range resultPortscan.IPResult[ip].Ports[port].PortAttrs {
			if par.Source == "httpx" {
				return true
			}
		}
	}
	if domain != "" && resultDomainScan != nil && resultDomainScan.DomainResult != nil {
		for _, dar := range resultDomainScan.DomainResult[domain].DomainAttrs {
			if dar.Source == "httpx" && dar.Tag == "httpx" {
				url := fmt.Sprintf("%s:%d", domain, port)
				if strings.Contains(dar.Content, url) {
					return true
				}
			}
		}
	}
	return false
}
