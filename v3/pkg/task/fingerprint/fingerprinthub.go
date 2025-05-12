package fingerprint

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

// FingerprintHubInfo 指纹基本信息
type FingerprintHubInfo struct {
	Name     string `json:"name"`
	Author   string `json:"author"`
	Tags     string `json:"tags"`
	Severity string `json:"severity"`
}

// Matcher 匹配器定义
type Matcher struct {
	Type            string           `json:"type"`
	Words           []string         `json:"words,omitempty"`
	Hash            []string         `json:"hash,omitempty"`
	Part            string           `json:"part,omitempty"`
	CaseInsensitive bool             `json:"case-insensitive,omitempty"`
	Negative        bool             `json:"negative,omitempty"`
	Condition       string           `json:"condition,omitempty"`
	Regex           []string         `json:"regex,omitempty"`
	compiledRegexps []*regexp.Regexp // 新增字段：预编译的正则表达式
}

// HTTPDefinition HTTP定义
type HTTPDefinition struct {
	Method   string    `json:"method"`
	Path     []string  `json:"path"`
	Matchers []Matcher `json:"matchers"`
}

// FingerprintHub 指纹定义
type FingerprintHub struct {
	ID   string             `json:"id"`
	Info FingerprintHubInfo `json:"info"`
	HTTP []HTTPDefinition   `json:"http"`
}

// MatchResult 匹配结果
type MatchResult struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// FingerprintHubEngine 指纹匹配引擎
type FingerprintHubEngine struct {
	fingerprints []FingerprintHub
}

// NewFingerprintHubEngine 创建新引擎实例
func NewFingerprintHubEngine() *FingerprintHubEngine {
	return &FingerprintHubEngine{}
}

// LoadFromFile 从文件加载指纹数据
func (e *FingerprintHubEngine) LoadFromFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open fingerprint file: %v", err)
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read fingerprint file: %v", err)
	}

	return e.LoadFromBytes(bytes)
}

// LoadFromBytes 从字节数据加载指纹
func (e *FingerprintHubEngine) LoadFromBytes(data []byte) error {
	var fingerprints []FingerprintHub
	if err := json.Unmarshal(data, &fingerprints); err != nil {
		return fmt.Errorf("failed to parse fingerprint JSON: %v", err)
	}

	e.fingerprints = preprocessFingerprints(fingerprints)
	return nil
}

// 预处理指纹数据并预编译正则表达式
func preprocessFingerprints(fps []FingerprintHub) []FingerprintHub {
	for i := range fps {
		for j := range fps[i].HTTP {
			for k := range fps[i].HTTP[j].Matchers {
				matcher := &fps[i].HTTP[j].Matchers[k]
				if matcher.Type == "regex" {
					// 转换正则表达式标志
					for l := range matcher.Regex {
						matcher.Regex[l] = convertRegexFlags(matcher.Regex[l])
					}
					// 预编译所有正则表达式
					matcher.compiledRegexps = make([]*regexp.Regexp, 0, len(matcher.Regex))
					for _, pattern := range matcher.Regex {
						if re, err := regexp.Compile(pattern); err == nil {
							matcher.compiledRegexps = append(matcher.compiledRegexps, re)
						}
					}
				}
			}
		}
	}
	return fps
}

// 转换正则表达式标志
func convertRegexFlags(pattern string) string {
	replacements := map[string]string{
		"(?mi)": "(?im)",
		"(?i)":  "(?i)",
		"(?m)":  "(?m)",
	}

	for rustFlag, goFlag := range replacements {
		if strings.HasPrefix(pattern, rustFlag) {
			return goFlag + strings.TrimPrefix(pattern, rustFlag)
		}
	}
	return pattern
}

// Match 执行指纹匹配
func (e *FingerprintHubEngine) Match(faviconHash, header, body string) []string {
	if len(e.fingerprints) == 0 {
		return nil //, fmt.Errorf("no fingerprints loaded")
	}

	var results []string

	for _, fp := range e.fingerprints {
		if e.matchFingerprint(fp, faviconHash, header, body) {
			results = append(results, fp.Info.Name)
		}
	}

	return results //, nil
}

// 匹配单个指纹
func (e *FingerprintHubEngine) matchFingerprint(fp FingerprintHub, faviconHash, header, body string) bool {
	for _, httpDef := range fp.HTTP {
		if e.matchHTTPDefinition(httpDef, faviconHash, header, body) {
			return true
		}
	}
	return false
}

// 匹配HTTP定义
func (e *FingerprintHubEngine) matchHTTPDefinition(httpDef HTTPDefinition, faviconHash, header, body string) bool {
	for _, matcher := range httpDef.Matchers {
		if e.matchSingleMatcher(matcher, faviconHash, header, body) {
			return true
		}
	}
	return false
}

// 匹配单个匹配器
func (e *FingerprintHubEngine) matchSingleMatcher(matcher Matcher, faviconHash, header, body string) bool {
	matched := false

	switch matcher.Type {
	case "favicon":
		matched = e.matchFavicon(matcher, faviconHash)
	case "word":
		matched = e.matchWord(matcher, header, body)
	case "regex":
		matched = e.matchRegex(matcher, header, body)
	}

	if matcher.Negative {
		return !matched
	}
	return matched
}

// 匹配favicon
func (e *FingerprintHubEngine) matchFavicon(matcher Matcher, faviconHash string) bool {
	if faviconHash == "" {
		return false
	}

	for _, h := range matcher.Hash {
		if h == faviconHash {
			return true
		}
	}
	return false
}

// 匹配关键词（内部处理condition）
func (e *FingerprintHubEngine) matchWord(matcher Matcher, header, body string) bool {
	searchText := body
	if matcher.Part == "header" {
		searchText = header
	}

	// 处理AND条件（所有words都必须匹配）
	if matcher.Condition == "and" {
		for _, word := range matcher.Words {
			found := false
			if matcher.CaseInsensitive {
				found = strings.Contains(strings.ToLower(searchText), strings.ToLower(word))
			} else {
				found = strings.Contains(searchText, word)
			}
			if !found {
				return false
			}
		}
		return len(matcher.Words) > 0 // 确保至少有一个word
	}

	// 默认OR条件（任意word匹配即可）
	for _, word := range matcher.Words {
		if matcher.CaseInsensitive {
			if strings.Contains(strings.ToLower(searchText), strings.ToLower(word)) {
				return true
			}
		} else if strings.Contains(searchText, word) {
			return true
		}
	}
	return false
}

// 匹配正则表达式（内部处理condition）
func (e *FingerprintHubEngine) matchRegex(matcher Matcher, header, body string) bool {
	searchText := body
	if matcher.Part == "header" {
		searchText = header
	}

	// 处理AND条件（所有正则都必须匹配）
	if matcher.Condition == "and" {
		for _, re := range matcher.compiledRegexps {
			if !re.MatchString(searchText) {
				return false
			}
		}
		return len(matcher.compiledRegexps) > 0 // 确保至少有一个正则
	}

	// 默认OR条件（任意正则匹配即可）
	for _, re := range matcher.compiledRegexps {
		if re.MatchString(searchText) {
			return true
		}
	}
	return false
}
