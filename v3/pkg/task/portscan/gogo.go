package portscan

import (
	"bytes"
	"compress/flate"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/chainreactors/utils/encode"
	"github.com/chainreactors/utils/iutils"
	"github.com/hanc00l/nemo_go/v3/pkg/conf"
	"github.com/hanc00l/nemo_go/v3/pkg/core"
	"github.com/hanc00l/nemo_go/v3/pkg/db"
	"github.com/hanc00l/nemo_go/v3/pkg/logging"
	"github.com/hanc00l/nemo_go/v3/pkg/resource"
	"github.com/hanc00l/nemo_go/v3/pkg/task/execute"
	"github.com/hanc00l/nemo_go/v3/pkg/utils"
	"io"
	"math"
	"path/filepath"
	"strconv"
	"strings"
)

type Gogo struct {
	Config  execute.PortscanConfig
	Result  Result
	IsProxy bool
}

func (g *Gogo) IsExecuteFromCmd() bool {
	return true
}

func (g *Gogo) GetExecuteCmd() string {
	return filepath.Join(conf.GetRootPath(), "thirdparty/gogo", utils.GetThirdpartyBinNameByPlatform(utils.Gogo))
}

func (g *Gogo) GetRequiredResources() (re []core.RequiredResource) {
	re = append(re, core.RequiredResource{
		Category: resource.GogoCategory,
		Name:     utils.GetThirdpartyBinNameByPlatform(utils.Gogo),
	})
	return
}

func (g *Gogo) GetExecuteArgs(inputTempFile, outputTempFile string, ipv6 bool) (cmdArgs []string) {
	cmdArgs = append(
		cmdArgs,
		"-q", "-l", inputTempFile, "-f", outputTempFile,
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
	if g.IsProxy {
		if proxy := conf.GetProxyConfig(); proxy != "" {
			cmdArgs = append(cmdArgs, "--proxy", proxy)
		} else {
			logging.RuntimeLog.Warning("获取代理配置失败或禁用了代理功能，代理被跳过")
		}
	}
	return
}

func (g *Gogo) Run(target []string, ipv6 bool) {
	//TODO implement me
	panic("implement me")
}

func (g *Gogo) ParseContentResult(content []byte) (result Result) {
	result.IPResult = make(map[string]*IPResult)

	rd := g.loadGOGOResultData(content)
	g.ParseGogoResultData(&result, rd)

	return
}

func (g *Gogo) ParseGogoResultData(result *Result, rd *GOGOData) {
	for _, r := range rd.Data {
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
				Tag:     db.FingerTitle,
				Content: r.Title,
			})
		}
		if len(r.Protocol) > 0 {
			result.SetPortAttr(r.Ip, port, PortAttrResult{
				Source:  "portscan",
				Tag:     db.FingerService,
				Content: r.Protocol,
			})
		}
		if (r.Protocol == "http" || r.Protocol == "https") && len(r.Status) > 0 {
			result.IPResult[r.Ip].Ports[port].HttpStatus = r.Status
		}
		if len(r.Midware) > 0 {
			result.SetPortAttr(r.Ip, port, PortAttrResult{
				Source:  "portscan",
				Tag:     db.FingerServer,
				Content: r.Midware,
			})
		}
		for _, f := range r.Frameworks {
			result.SetPortAttr(r.Ip, port, PortAttrResult{
				Source:  "portscan",
				Tag:     db.FingerApp,
				Content: f.String(),
			})
		}
	}
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
	Tags          []string               `json:"tags,omitempty"`
	Payload       map[string]interface{} `json:"payload,omitempty"`
	Detail        map[string][]string    `json:"detail,omitempty"`
	SeverityLevel int                    `json:"severity"`
	Framework     *Framework             `json:"-"`
}

func (v *Vuln) HasTag(tag string) bool {
	for _, t := range v.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

func (v *Vuln) GetPayload() string {
	return iutils.MapToString(v.Payload)
}

func (v *Vuln) GetDetail() string {
	var s strings.Builder
	for k, v := range v.Detail {
		s.WriteString(fmt.Sprintf(" %s:%s ", k, strings.Join(v, ",")))
	}
	return s.String()
}

func (v *Vuln) String() string {
	s := v.Name
	if payload := v.GetPayload(); payload != "" {
		s += fmt.Sprintf(" payloads:%s", iutils.AsciiEncode(payload))
	}
	if detail := v.GetDetail(); detail != "" {
		s += fmt.Sprintf(" payloads:%s", iutils.AsciiEncode(detail))
	}
	return s
}

type Vulns map[string]*Vuln
type GOGOResults []*GOGOResult
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

func shannonEntropy(data []byte) float64 {
	freq := make(map[byte]int)
	for _, b := range data {
		freq[b]++
	}

	entropy := 0.0
	dataLength := len(data)
	for _, count := range freq {
		freqRatio := float64(count) / float64(dataLength)
		entropy -= freqRatio * math.Log2(freqRatio)
	}

	return entropy
}

func DecryptFile(content []byte, keys []byte) []byte {
	decoded, err := base64.StdEncoding.DecodeString(string(content))
	if err == nil {
		// try to base64 decode, if decode successfully, return data
		return bytes.TrimSpace(decoded)
	}
	// else try to unflate
	decrypted := encode.XorEncode(content, keys, 0)
	if shannonEntropy(content) < 5.5 {
		// 使用香农熵判断是否是deflate后的文件, 测试了几个数据集, 数据量较大的时候接近4, 数据量较小时接近5. deflate后的文件大多都在6以上
		return content
	} else {
		return bytes.TrimSpace(encode.MustDeflateDeCompress(decrypted))
	}
}

// loadGOGOResultData 读取并解析gogo的json类型的结果文件，使用gogo代码的解密方式，支持压缩格式
func (g *Gogo) loadGOGOResultData(input []byte) (gogoData *GOGOData) {
	content := DecryptFile(input, []byte(""))
	lines := bytes.Split(content, []byte{'\n'})
	if len(lines) <= 2 {
		return
	}
	gogoData = &GOGOData{}
	_ = json.Unmarshal(lines[0], &gogoData.Config)
	// 第二行开始（去除最后一行）:
	for i := 1; i < len(lines)-1; i++ {
		var resultLine GOGOResult
		err := json.Unmarshal(lines[i], &resultLine)
		if err != nil {
			logging.RuntimeLog.Error(err)
			continue
		}
		gogoData.Data = append(gogoData.Data, &resultLine)
	}
	return
}
