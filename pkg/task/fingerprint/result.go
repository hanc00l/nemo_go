package fingerprint

import "sync"

var IgnorePort = []int{7, 9, 13, 17, 19, 21, 22, 23, 25, 26, 37, 53, 100, 106, 110, 111, 113, 119, 135, 138, 139,
	143, 144, 145, 161,
	179, 199, 389, 427, 444, 445, 514, 515, 543, 554, 631, 636, 646, 880, 902, 990, 993,
	1433, 1521, 3306, 5432, 3389, 5900, 5901, 5902, 49152, 49153, 49154, 49155, 49156, 49157,
	49158, 49159, 49160, 49161, 49163, 49165, 49167, 49175, 49176,
	13306, 11521, 15432, 11433, 13389, 15900, 15901}

const (
	fpHttpxThreadNumber   = 10
	fpWhatwebThreadNumber = 4
	fpScreenshotThreadNum = 5
	fpWappalyzerThreadNumber = 10
	fpObserverWardThreadNumber =  10
	fpIconHashThreadNumber = 10
)

type Config struct {
	Target string
	//OrgId  *int
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
