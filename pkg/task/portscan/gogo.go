package portscan

import (
	"bytes"
	"compress/flate"
	"encoding/json"
	"fmt"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/custom"
	"github.com/hanc00l/nemo_go/pkg/task/pocscan"
	"github.com/hanc00l/nemo_go/pkg/utils"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// Gogo 导入gogo的扫描结果
type Gogo struct {
	Config    Config
	Result    Result
	VulResult []pocscan.Result
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

// NewGogo 创建gogo对象
func NewGogo(config Config) *Gogo {
	g := &Gogo{Config: config}
	g.Config.CmdBin = filepath.Join(conf.GetRootPath(), "thirdparty/gogo", utils.GetThirdpartyBinNameByPlatform(utils.Gogo))

	return g
}

func (g *Gogo) Do() {
	g.Result.IPResult = make(map[string]*IPResult)

	inputTargetFile := utils.GetTempPathFileName()
	resultTempFile := utils.GetTempPathFileName()
	defer os.Remove(inputTargetFile)
	defer os.Remove(resultTempFile)

	btc := custom.NewBlackTargetCheck(custom.CheckIP)
	var targets []string
	for _, target := range strings.Split(g.Config.Target, ",") {
		t := strings.TrimSpace(target)
		if btc.CheckBlack(t) {
			logging.RuntimeLog.Warningf("%s is in blacklist,skip...", t)
			continue
		}
		targets = append(targets, t)
	}
	err := os.WriteFile(inputTargetFile, []byte(strings.Join(targets, "\n")), 0666)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		return
	}
	var cmdArgs []string
	cmdArgs = append(
		cmdArgs,
		"-q", "-l", inputTargetFile, "-f", resultTempFile,
		// gogo的速度，用线程来代替（与nmap和masscan有所区别）
		"-t", strconv.Itoa(g.Config.Rate),
	)
	if strings.HasPrefix(g.Config.Port, "--top-ports") {
		cmdArgs = append(cmdArgs, "-p")
		switch strings.Split(g.Config.Port, " ")[1] {
		case "1000":
			cmdArgs = append(cmdArgs, utils.TopPorts1000)
		case "100":
			cmdArgs = append(cmdArgs, utils.TopPorts100)
		case "10":
			cmdArgs = append(cmdArgs, utils.TopPorts10)
		default:
			cmdArgs = append(cmdArgs, utils.TopPorts100)
		}
	} else {
		cmdArgs = append(cmdArgs, "-p", g.Config.Port)
	}
	if g.Config.ExcludeTarget != "" {
		cmdArgs = append(cmdArgs, "--exclude", g.Config.ExcludeTarget)
	}
	if g.Config.IsPing {
		cmdArgs = append(cmdArgs, "--ping")
	}
	if g.Config.Tech == "--sV" {
		cmdArgs = append(cmdArgs, "-v")
	}
	if g.Config.IsProxy {
		if proxy := conf.GetProxyConfig(); proxy != "" {
			cmdArgs = append(cmdArgs, "--proxy", proxy)
		} else {
			logging.RuntimeLog.Warning("get proxy config fail or disabled by worker,skip proxy!")
			logging.CLILog.Warning("get proxy config fail or disabled by worker,skip proxy!")
		}
	}
	cmd := exec.Command(g.Config.CmdBin, cmdArgs...)
	var stderr bytes.Buffer
	cmd.Stdout = os.Stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		logging.RuntimeLog.Error(err, stderr)
		logging.CLILog.Error(err, stderr)
		return
	}
	g.parsResult(resultTempFile)
	FilterIPHasTooMuchPort(&g.Result, false)
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

// parsResult 解析结果
func (g *Gogo) parsResult(outputTempFile string) {
	//gogo结果比较特殊，需要-F进行转换
	resultTempFile := utils.GetTempPathFileName()
	defer os.Remove(resultTempFile)
	var cmdArgs []string
	cmdArgs = append(
		cmdArgs,
		"-F", outputTempFile, "-o", "json", "-f", resultTempFile,
	)
	cmd := exec.Command(g.Config.CmdBin, cmdArgs...)
	var stderr bytes.Buffer
	cmd.Stdout = os.Stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		logging.RuntimeLog.Error(err, stderr)
		logging.CLILog.Error(err, stderr)
		return
	}
	content, err := os.ReadFile(resultTempFile)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return
	}
	result := g.ParseContentResult(content)
	for ip, ipr := range result.IPResult {
		g.Result.IPResult[ip] = ipr
	}
}

// ParseContentResult 解析gogo扫描的文本结果
func (g *Gogo) ParseContentResult(content []byte) (result Result) {
	result.IPResult = make(map[string]*IPResult)

	gogoData := g.loadGOGOResultData(content)
	if gogoData == nil {
		return
	}
	for _, r := range gogoData.Data {
		if !result.HasIP(r.Ip) {
			result.SetIP(r.Ip)
		}
		port, err := strconv.Atoi(r.Port)
		if err != nil {
			continue
		}
		if !result.HasPort(r.Ip, port) {
			result.SetPort(r.Ip, port)
		}
		if len(r.Title) > 0 {
			result.SetPortAttr(r.Ip, port, PortAttrResult{
				Source:  "portscan",
				Tag:     "title",
				Content: r.Title,
			})
		}
		if len(r.Protocol) > 0 {
			result.SetPortAttr(r.Ip, port, PortAttrResult{
				Source:  "portscan",
				Tag:     "service",
				Content: r.Protocol,
			})
		}
		if (r.Protocol == "http" || r.Protocol == "https") && len(r.Status) > 0 {
			result.IPResult[r.Ip].Ports[port].Status = r.Status
		}
		if len(r.Midware) > 0 {
			result.SetPortAttr(r.Ip, port, PortAttrResult{
				Source:  "portscan",
				Tag:     "midware",
				Content: r.Midware,
			})
		}
		for _, f := range r.Frameworks {
			result.SetPortAttr(r.Ip, port, PortAttrResult{
				Source:  "portscan",
				Tag:     "banner",
				Content: f.String(),
			})
		}
		for _, v := range r.Vulns {
			g.VulResult = append(g.VulResult, pocscan.Result{
				Target:      r.Ip,
				Url:         r.Uri,
				PocFile:     v.Name,
				Source:      "gogo",
				Extra:       v.String(),
				WorkspaceId: g.Config.WorkspaceId,
			})
		}
	}
	return
}
