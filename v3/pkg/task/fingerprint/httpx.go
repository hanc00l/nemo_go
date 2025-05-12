package fingerprint

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/chainreactors/fingers"
	"github.com/hanc00l/nemo_go/v3/pkg/conf"
	"github.com/hanc00l/nemo_go/v3/pkg/core"
	"github.com/hanc00l/nemo_go/v3/pkg/logging"
	"github.com/hanc00l/nemo_go/v3/pkg/resource"
	"github.com/hanc00l/nemo_go/v3/pkg/task/execute"
	"github.com/hanc00l/nemo_go/v3/pkg/utils"
	"github.com/spaolacci/murmur3"
	"io"
	"net/http"
	"path/filepath"
)

type Httpx struct {
	Config  execute.FingerprintConfig
	IsProxy bool
	// 被动指纹引擎
	chainreactorsEngine  *fingers.Engine
	fingerprintHubEngine *FingerprintHubEngine
}

func (h *Httpx) GetRequiredResources() (re []core.RequiredResource) {
	re = append(re, core.RequiredResource{
		Category: resource.HttpxCategory,
		Name:     utils.GetThirdpartyBinNameByPlatform(utils.Httpx),
	})
	re = append(re, core.RequiredResource{
		Category: resource.DictCategory,
		Name:     "web_fingerprint_v4.json",
	})
	re = append(re, core.RequiredResource{
		Category: resource.DictCategory,
		Name:     "web_poc_map_v2.json",
	})
	return
}

func (h *Httpx) IsExecuteFromCmd() bool {
	return true
}

func (h *Httpx) GetExecuteCmd() string {
	return filepath.Join(conf.GetRootPath(), "thirdparty/httpx", utils.GetThirdpartyBinNameByPlatform(utils.Httpx))
}

func (h *Httpx) GetExecuteArgs(inputTempFile, outputTempFile string) (cmdArgs []string) {
	cmdArgs = append(cmdArgs,
		"-random-agent", "-l", inputTempFile, "-o", outputTempFile,
		"-retries", "3", "-threads", "50", "-timeout", "5", "-disable-update-check",
		"-title", "-server", "-status-code", "-content-type", "-follow-redirects", "-json", "-silent", "-no-color", "-tls-grab", "-jarm",
		"-ehb", "-irrb",
		// -esb, -exclude-screenshot-bytes  enable excluding screenshot bytes from json output
		// -ehb, -exclude-headless-body     enable excluding headless header from json output
		// -irrb, -include-response-base64     include base64 encoded http request/response in JSON output (-json only)
	)
	if h.Config.IsScreenshot {
		cmdArgs = append(cmdArgs, "-screenshot", "--system-chrome")
	}
	if h.Config.IsIconHash {
		cmdArgs = append(cmdArgs, "-favicon")
	}
	// 由于chrome不支持带认证的socks5代理，因此httpx及chrome使用本地的socks5转发
	if h.IsProxy {
		if proxy := conf.GetProxyConfig(); proxy != "" {
			if conf.Socks5ForwardAddr != "" {
				cmdArgs = append(cmdArgs, "-http-proxy", fmt.Sprintf("socks5://%s", conf.Socks5ForwardAddr))
			}
		} else {
			logging.RuntimeLog.Warning("获取代理配置失败或禁用了代理功能，代理被跳过")
		}
	}
	return
}

func (h *Httpx) Run(target []string) (result Result) {
	//TODO implement me
	panic("implement me")
}

func (h *Httpx) ParseContentResult(content []byte) (result Result) {
	result.FingerResults = make(map[string]interface{})
	h.loadChainReactorsFingers()
	h.loadFingerprintHubEngine()

	lines := bytes.Split(content, []byte{'\n'})
	for _, line := range lines {
		if len(line) > 0 {
			r := h.ParseHttpxJson(line)
			if len(r.Input) > 0 {
				result.FingerResults[r.Input] = r
			}
		}
	}

	return
}

// ParseHttpxJson 解析一条httpx的JSON记录
func (h *Httpx) ParseHttpxJson(content []byte) (resultJSON HttpxResult) {
	err := json.Unmarshal(content, &resultJSON)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return
	}
	if resultJSON.FaviconURL != "" && resultJSON.IconHash != "" {
		resultJSON.FaviconContent, _ = h.getFaviconImage(resultJSON.IconHash, resultJSON.FaviconURL)
	}
	// 被动指纹匹配
	if len(resultJSON.RawHeader) > 0 {
		var header, body []byte
		header, _ = base64.StdEncoding.DecodeString(resultJSON.RawHeader)
		body, _ = base64.StdEncoding.DecodeString(resultJSON.Body)
		resultJSON.Fingers = append(resultJSON.Fingers, h.MatchChainReactorsFingers(header, body)...)
		resultJSON.Fingers = append(resultJSON.Fingers, h.fingerprintHubEngine.Match(resultJSON.IconHash, string(header), string(body))...)
		resultJSON.Fingers = utils.RemoveDuplicatesAndSort(resultJSON.Fingers)
	}

	return
}

// getFaviconImage 获取favicon图片
func (h *Httpx) getFaviconImage(iconHash string, iconUrlPathFile string) ([]byte, error) {
	//获取icon
	request, err := http.NewRequest(http.MethodGet, iconUrlPathFile, nil)
	if err != nil {
		//fmt.Println(err)
		return nil, err
	}
	resp, err := utils.GetProxyHttpClient(h.IsProxy).Do(request)
	if err != nil {
		//fmt.Println(err)
		return nil, err
	}
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		//fmt.Println(err)
		return nil, err
	}
	err = resp.Body.Close()
	if err != nil {
		return nil, err
	}
	if len(content) == 0 {
		return nil, fmt.Errorf("获取icon:%s失败", iconUrlPathFile)
	}
	// 使用Murmur3算法计算hash值
	hash := murmurhash(content)
	if fmt.Sprintf("%d", hash) != iconHash {
		return nil, fmt.Errorf("计算icon:%s的hash值不匹配", iconUrlPathFile)
	}

	return content, nil
}

// MatchChainReactorsFingers ChainReactorsFingers指纹识别
func (h *Httpx) MatchChainReactorsFingers(header, body []byte) (fingers []string) {
	var buffer bytes.Buffer
	buffer.Write(header)
	buffer.Write([]byte("\r\n\r\n"))
	buffer.Write(body)
	frames, err := h.chainreactorsEngine.DetectContent(buffer.Bytes())
	if err != nil {
		logging.RuntimeLog.Errorf("调用chainreactors指纹检测失败，err:%v", err)
		return
	}
	for _, f := range frames.List() {
		fingers = append(fingers, f.String())
	}

	return
}

// loadChainReactorsFingers 加载ChainReactorsFingers指纹引擎
func (h *Httpx) loadChainReactorsFingers() {
	if h.chainreactorsEngine != nil {
		return
	}
	var err error
	h.chainreactorsEngine, err = fingers.NewEngine(fingers.FingersEngine)

	if err != nil {
		logging.RuntimeLog.Errorf("调用chainreactors指纹检测失败，err:%v", err)
		return
	}
	logging.CLILog.Info("已加载chainreactors指纹引擎")
}

// loadFingerprintHubEngine 加载FingerprintHub指纹引擎
func (h *Httpx) loadFingerprintHubEngine() {
	if h.fingerprintHubEngine != nil {
		return
	}
	h.fingerprintHubEngine = NewFingerprintHubEngine()
	err := h.fingerprintHubEngine.LoadFromFile(filepath.Join(conf.GetAbsRootPath(), "thirdparty/dict/web_fingerprint_v4.json"))
	if err != nil {
		logging.RuntimeLog.Error(err)
		return
	}
	logging.CLILog.Info("已加载FingerprintHub指纹引擎")
}

func InsertInto(s string, interval int, sep rune) string {
	var buffer bytes.Buffer
	before := interval - 1
	last := len(s) - 1
	for i, char := range s {
		buffer.WriteRune(char)
		if i%interval == before && i != last {
			buffer.WriteRune(sep)
		}
	}
	buffer.WriteRune(sep)
	return buffer.String()
}

func murmurhash(data []byte) int32 {
	stdBase64 := base64.StdEncoding.EncodeToString(data)
	stdBase64 = InsertInto(stdBase64, 76, '\n')
	hasher := murmur3.New32WithSeed(0)
	hasher.Write([]byte(stdBase64))
	return int32(hasher.Sum32())
}
