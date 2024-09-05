package portscan

import (
	"encoding/json"
	"github.com/hanc00l/nemo_go/v2/pkg/logging"
	"strconv"
)

type Goby struct {
}

// GobyAssertPortInfo 端口信息
type GobyAssertPortInfo struct {
	Port      string   `json:"port"`
	Hostinfo  string   `json:"hostinfo"`
	Url       string   `json:"url"`
	Product   string   `json:"product"`
	Protocol  string   `json:"protocol"`
	Json      string   `json:"json"`
	Fid       []string `json:"fid"`
	Products  []string `json:"products"`
	Protocols []string `json:"protocols"`
}

// GobyAssetSearchResponse 扫描端口信息返回
type GobyAssetSearchResponse struct {
	StatusCode int    `json:"statusCode"`
	Messages   string `json:"messages"`
	Data       struct {
		Ips []struct {
			Ip       string `json:"ip"`
			Mac      string `json:"mac"`
			Os       string `json:"os"`
			Hostname string `json:"hostname"`
			Honeypot string `json:"honeypot,omitempty"`
			Ports    []struct {
				Port         string `json:"port"`
				Baseprotocol string `json:"baseprotocol"`
			} `json:"ports"`
			Protocols       map[string]GobyAssertPortInfo `json:"protocols"`
			Vulnerabilities []struct {
				Hostinfo string `json:"hostinfo"`
				Name     string `json:"name"`
				Filename string `json:"filename"`
				Level    string `json:"level"`
				Vulurl   string `json:"vulurl"`
				Keymemo  string `json:"keymemo"`
				Hasexp   bool   `json:"hasexp "`
			} `json:"vulnerabilities"`
			Screenshots interface{} `json:"screenshots"`
			Favicons    interface{} `json:"favicons"`
			Hostnames   []string    `json:"hostnames"`
		} `json:"ips"`
	} `json:"data"`
}

// ParseContentResult 获取资产扫描结果
func (g *Goby) ParseContentResult(content []byte) (result Result) {
	result.IPResult = make(map[string]*IPResult)

	var resp GobyAssetSearchResponse
	err := json.Unmarshal(content, &resp)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return
	}
	for _, ipAsset := range resp.Data.Ips {
		if !result.HasIP(ipAsset.Ip) {
			result.SetIP(ipAsset.Ip)
		}
		for _, v := range ipAsset.Protocols {
			port, err := strconv.Atoi(v.Port)
			if err != nil {
				continue
			}
			if !result.HasPort(ipAsset.Ip, port) {
				result.SetPort(ipAsset.Ip, port)
			}
			if len(v.Protocol) > 0 {
				result.SetPortAttr(ipAsset.Ip, port, PortAttrResult{
					Source:  "goby",
					Tag:     "service",
					Content: v.Protocol,
				})
			}
			if len(v.Product) > 0 {
				result.SetPortAttr(ipAsset.Ip, port, PortAttrResult{
					Source:  "goby",
					Tag:     "banner",
					Content: v.Product,
				})
			}
		}
	}
	return
}
