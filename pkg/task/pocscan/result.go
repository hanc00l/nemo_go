package pocscan

import (
	"fmt"
	"github.com/hanc00l/nemo_go/pkg/db"
	"github.com/hanc00l/nemo_go/pkg/utils"
)

type Config struct {
	Target  string `json:"target"`
	PocFile string `json:"poc_file"`
	CmdBin  string `json:"cmdbin"`
}

type Result struct {
	Target  string `json:"target"`
	Url     string `json:"url"`
	PocFile string `json:"poc_file""`
	Source  string `json:"source"`
	Extra   string `json:"extra"`
}

type xrayJSONResult struct {
	Target struct {
		Url string `json:"url"`
	} `json:"target"`
	Plugin string `json:"plugin"`
	Detail struct {
		Addr     string     `json:"addr"`
		Payload  string     `json:"payload"`
		Snapshot [][]string `json:"snapshot"`
	} `json:"detail"`
}

// SaveResult 保存结果
func SaveResult(result []Result) string {
	var resultCount int
	for _, r := range result {
		target := utils.HostStrip(r.Target)
		extra := r.Extra
		if len(r.Extra) > 2000 {
			extra = r.Extra[:2000] + "..."
		}
		vul := db.Vulnerability{
			Target:  target,
			Url:     r.Url,
			PocFile: r.PocFile,
			Source:  r.Source,
			Extra:   extra,
		}
		if vul.SaveOrUpdate() {
			resultCount++
		}
	}
	return fmt.Sprintf("vulnerability:%d", resultCount)
}
