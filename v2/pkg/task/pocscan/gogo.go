package pocscan

import (
	"bytes"
	"compress/flate"
	"encoding/json"
	"fmt"
	"github.com/hanc00l/nemo_go/v2/pkg/logging"
	"io"
	"strings"
)

// Gogo 导入gogo的扫描结果
type Gogo struct {
}

// forked from https://github.com/chainreactors/gogo

type GOGOConfig struct {
	IP            string   `json:"ip"`
	IPlist        []string `json:"ips"`
	Ports         string   `json:"ports"`
	JsonFile      string   `json:"json_file"`
	ListFile      string   `json:"list_file"`
	Threads       int      `json:"threads"` // 线程数
	Mod           string   `json:"mod"`     // 扫描模式
	AliveSprayMod []string `json:"alive_spray"`
	PortSpray     bool     `json:"port_spray"`
	Exploit       string   `json:"exploit"`
	JsonType      string   `json:"json_type"`
	VersionLevel  int      `json:"version_level"`
}

type Framework struct {
	Name    string       `json:"name"`
	Version string       `json:"version,omitempty"`
	From    int          `json:"-"`
	Froms   map[int]bool `json:"froms,omitempty"`
	Tags    []string     `json:"tags,omitempty"`
	IsFocus bool         `json:"is_focus,omitempty"`
	Data    string       `json:"-"`
}

const (
	FrameFromDefault = iota
	FrameFromACTIVE
	FrameFromICO
	FrameFromNOTFOUND
	FrameFromGUESS
)

var frameFromMap = map[int]string{
	FrameFromDefault:  "finger",
	FrameFromACTIVE:   "active",
	FrameFromICO:      "ico",
	FrameFromNOTFOUND: "404",
	FrameFromGUESS:    "guess",
}

func (f *Framework) String() string {
	var s strings.Builder
	if f.IsFocus {
		s.WriteString("focus:")
	}
	s.WriteString(f.Name)

	if f.Version != "" {
		s.WriteString(":" + strings.Replace(f.Version, ":", "_", -1))
	}

	if len(f.Froms) > 1 {
		s.WriteString(":")
		for from, _ := range f.Froms {
			s.WriteString(frameFromMap[from] + " ")
		}
	} else {
		for from, _ := range f.Froms {
			if from != FrameFromDefault {
				s.WriteString(":")
				s.WriteString(frameFromMap[from])
			}
		}
	}
	return strings.TrimSpace(s.String())
}

func (fs Frameworks) String() string {
	if fs == nil {
		return ""
	}
	var frameworkStrs []string
	for _, f := range fs {
		//if NoGuess && f.IsGuess() {
		//	continue
		//}
		frameworkStrs = append(frameworkStrs, f.String())
	}
	return strings.Join(frameworkStrs, "||")
}

type Vuln struct {
	Name          string                 `json:"name"`
	Payload       map[string]interface{} `json:"payload,omitempty"`
	Detail        map[string]interface{} `json:"detail,omitempty"`
	SeverityLevel int                    `json:"severity"`
}

func (v *Vuln) String() string {
	s := v.Name
	if payload := v.GetPayload(); payload != "" {
		s += fmt.Sprintf(" payloads:%s", payload)
	}
	if detail := v.GetDetail(); detail != "" {
		s += fmt.Sprintf(" payloads:%s", detail)
	}
	return s
}
func MapToString(m map[string]interface{}) string {
	if m == nil || len(m) == 0 {
		return ""
	}
	var s string
	for k, v := range m {
		s += fmt.Sprintf(" %s:%s ", k, v.(string))
	}
	return s
}
func (v *Vuln) GetPayload() string {
	return MapToString(v.Payload)
}

func (v *Vuln) GetDetail() string {
	return MapToString(v.Detail)
}

type GOGOResults []*GOGOResult
type Vulns []*Vuln
type Frameworks map[string]*Framework

type GOGOResult struct {
	Ip         string              `json:"ip"`                   // ip
	Port       string              `json:"port"`                 // port
	Uri        string              `json:"uri,omitempty"`        // uri
	Os         string              `json:"os,omitempty"`         // os
	Host       string              `json:"host,omitempty"`       // host
	Frameworks Frameworks          `json:"frameworks,omitempty"` // framework
	Vulns      Vulns               `json:"vulns,omitempty"`
	Extracteds map[string][]string `json:"extracted,omitempty"`
	Protocol   string              `json:"protocol"` // protocol
	Status     string              `json:"status"`   // http_stat
	Language   string              `json:"language"`
	Title      string              `json:"title"`   // title
	Midware    string              `json:"midware"` // midware
}

type GOGOData struct {
	Config GOGOConfig  `json:"config"`
	IP     string      `json:"ip"`
	Data   GOGOResults `json:"data"`
}

func UnFlat(input []byte) []byte {
	rdata := bytes.NewReader(input)
	r := flate.NewReader(rdata)
	s, _ := io.ReadAll(r)
	return s
}

// loadGOGOResultData 读取并解析gogo的json类型的结果文件，支持压缩格式
func (g *Gogo) loadGOGOResultData(input []byte) (gogoData *GOGOData) {
	gogoData = &GOGOData{}
	// 先直接解析为json，如果没有报错则直接返回结果
	if err := json.Unmarshal(input, gogoData); err == nil {
		return
	}
	// 如果不是json文件，则先解压后再解析
	flatInput := UnFlat(input)
	if err := json.Unmarshal(flatInput, gogoData); err != nil {
		logging.RuntimeLog.Error(err)
		gogoData = nil
	}
	return
}

// ParseContentResult 解析gogo扫描的文本结果
func (g *Gogo) ParseContentResult(content []byte) (result []Result) {
	gogoData := g.loadGOGOResultData(content)
	if gogoData == nil {
		return
	}
	for _, r := range gogoData.Data {
		for _, v := range r.Vulns {
			result = append(result, Result{
				Target:  r.Ip,
				Url:     r.Uri,
				PocFile: v.Name,
				Source:  "gogo",
				Extra:   v.String(),
			})
		}
	}
	return
}
