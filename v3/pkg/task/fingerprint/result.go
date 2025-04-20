package fingerprint

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/hanc00l/nemo_go/v3/pkg/core"
	"github.com/hanc00l/nemo_go/v3/pkg/db"
	"github.com/hanc00l/nemo_go/v3/pkg/logging"
	"github.com/hanc00l/nemo_go/v3/pkg/task/domainscan"
	"github.com/hanc00l/nemo_go/v3/pkg/task/execute"
	"github.com/hanc00l/nemo_go/v3/pkg/utils"
	"strconv"
	"strings"
	"sync"
)

type Result struct {
	sync.RWMutex
	FingerResults map[string]interface{}
}
type HttpxResult struct {
	Input       string   `json:"input"`
	Host        string   `json:"host"`
	Port        string   `json:"port"`
	Url         string   `json:"url"`
	Scheme      string   `json:"scheme"`
	A           []string `json:"a,omitempty"`
	CName       []string `json:"cname,omitempty"`
	Title       string   `json:"title,omitempty"`
	WebServer   string   `json:"webserver,omitempty"`
	ContentType string   `json:"content_type,omitempty"`
	StatusCode  int      `json:"status_code,omitempty"`
	TLSData     *TLS     `json:"tls,omitempty"`
	Jarm        string   `json:"jarm,omitempty"`
	IconHash    string   `json:"favicon,omitempty"`
	FaviconURL  string   `json:"favicon_url,omitempty"`
	Tech        []string `json:"tech,omitempty"`
	Body        string   `json:"body,omitempty"`
	RawHeader   string   `json:"raw_header,omitempty"`
	ScreenBytes string   `json:"screenshot_bytes,omitempty"`

	// icon图标和fingers指纹是单独实现（不由httpx获取）
	FaviconContent []byte   `json:"favicon_content,omitempty"`
	Fingers        []string `json:"fingers,omitempty"`
}

type TLS struct {
	Host                     string            `json:"host,omitempty"`
	Port                     string            `json:"port,omitempty"`
	TLSVersion               string            `json:"tls_version,omitempty"`
	SubjectDNSName           []string          `json:"subject_an,omitempty"`
	SubjectCommonName        string            `json:"subject_cn,omitempty"`
	SubjectDistinguishedName string            `json:"subject_dn,omitempty"`
	SubjectOrganization      []string          `json:"subject_org,omitempty"`
	IssuerDistinguishedName  string            `json:"issuer_dn,omitempty"`
	IssuerCommonName         string            `json:"issuer_cn,omitempty"`
	IssuerOrganization       []string          `json:"issuer_org,omitempty"`
	FingerprintHash          map[string]string `json:"fingerprint_hash,omitempty"`
}

type FingerprintxResult struct {
	Host      string          `json:"host,omitempty"`
	IP        string          `json:"ip"`
	Port      int             `json:"port"`
	Protocol  string          `json:"protocol"`
	TLS       bool            `json:"tls"`
	Transport string          `json:"transport"`
	Version   string          `json:"version,omitempty"`
	Raw       json.RawMessage `json:"metadata"`
}

type IconHashResult struct {
	Authority      string
	IconHash       string
	IconFileSuffix string
	IconBytes      []byte
}

const (
	MaxWidth       = 1280 //1980
	MinHeight      = 800  //1080
	SavedWidth     = 1280 //1024
	SavedHeight    = 800  //768
	thumbnailWidth = 120
)

func ParseResult(config execute.ExecutorTaskInfo, result *Result) (docs []db.AssetDocument, screenshot []core.ScreenShotResultArgs) {
	tldExacter := domainscan.NewTldExtract()

	for input, result := range result.FingerResults {
		if fr, ok := result.(FingerprintxResult); ok {
			doc := parseFingerprintxResult(config, input, fr, &tldExacter)
			docs = append(docs, doc)
		} else if hr, ok := result.(HttpxResult); ok {
			doc, ss := parseHttpxResult(config, input, hr, &tldExacter)
			docs = append(docs, doc)
			if ss != nil {
				screenshot = append(screenshot, *ss)
			}
		}
	}
	return
}

func parseFingerprintxResult(config execute.ExecutorTaskInfo, authority string, result FingerprintxResult, tldExacter *domainscan.TldExtract) (doc db.AssetDocument) {
	doc.Authority = authority
	doc.Host = result.Host
	doc.Port = result.Port
	doc.OrgId = config.OrgId
	doc.TaskId = config.MainTaskId
	if len(result.IP) > 0 {
		if utils.CheckIPV4(result.IP) {
			doc.Ip.IpV4 = append(doc.Ip.IpV4, db.IPV4{
				IPName: result.IP,
			})
			doc.Category = db.CategoryIPv4
		} else if utils.CheckIPV6(result.IP) {
			doc.Ip.IpV6 = append(doc.Ip.IpV6, db.IPV6{
				IPName: result.IP,
			})
			doc.Category = db.CategoryIPv6
		} else {
			doc.Domain = tldExacter.ExtractFLD(result.Host)
			if len(doc.Domain) > 0 {
				doc.Category = db.CategoryDomain
			}
		}
	}
	getFingerprintxFingers(&result, &doc)

	return
}

func parseHttpxResult(config execute.ExecutorTaskInfo, authority string, result HttpxResult, tldExacter *domainscan.TldExtract) (doc db.AssetDocument, screenshot *core.ScreenShotResultArgs) {
	doc.Authority = authority
	doc.OrgId = config.OrgId
	doc.TaskId = config.MainTaskId
	doc.Port, _ = strconv.Atoi(result.Port)
	// 处理子域名任务的authority的特殊情况：默认massdns或subfinder的输出authority不带端口，在fingerprint时需要补上，从而单独作为一条记录
	hostAndPort := strings.Split(doc.Authority, ":")
	if len(hostAndPort) == 1 && doc.Port > 0 {
		doc.Authority = hostAndPort[0] + ":" + strconv.Itoa(doc.Port)
	}
	// 注意:httpx会将域名解析为ip作为host，所以这里要单独处理
	doc.Host = hostAndPort[0]

	if utils.CheckIPV4(doc.Host) {
		doc.Category = db.CategoryIPv4
	} else if utils.CheckIPV6(doc.Host) {
		doc.Category = db.CategoryIPv6
	} else {
		doc.Domain = tldExacter.ExtractFLD(doc.Host)
		if len(doc.Domain) > 0 {
			doc.Category = db.CategoryDomain
		}
	}
	if len(result.A) > 0 {
		for _, a := range result.A {
			if utils.CheckIPV4(a) {
				doc.Ip.IpV4 = append(doc.Ip.IpV4, db.IPV4{
					IPName: a,
				})
			} else if utils.CheckIPV6(a) {
				doc.Ip.IpV6 = append(doc.Ip.IpV6, db.IPV6{
					IPName: a,
				})
			}
		}
	}
	if result.StatusCode > 0 {
		doc.HttpStatus = fmt.Sprintf("%d", result.StatusCode)
	}
	if len(result.RawHeader) > 0 {
		header, err := base64.StdEncoding.DecodeString(result.RawHeader)
		if err == nil {
			doc.HttpHeader = string(header)
		}
		body, err := base64.StdEncoding.DecodeString(result.Body)
		if err == nil {
			doc.HttpBody = string(body)
		}
	}
	if result.TLSData != nil {
		tlsData, err := json.Marshal(result.TLSData)
		if err == nil {
			doc.Cert = string(tlsData)
		}
	}
	getHttpxFingers(&result, &doc)
	doc.IconHash = result.IconHash
	if len(result.FaviconContent) > 0 {
		iconFile := utils.GetFaviconSuffixUrlFile(result.FaviconURL)
		if len(iconFile) > 0 {
			doc.IconHashFile = iconFile
			doc.IconHashBytes = result.FaviconContent
		}
	}
	if len(result.ScreenBytes) > 0 {
		content, err := base64.StdEncoding.DecodeString(result.ScreenBytes)
		if err != nil {
			logging.RuntimeLog.Errorf("decode screenshot base64 fail:%v", err)

		}
		screenshot = &core.ScreenShotResultArgs{
			WorkspaceId:    config.WorkspaceId,
			Host:           doc.Host, //注意：这里的host是指httpx的输入，而不是指httpx的host
			Port:           result.Port,
			Scheme:         result.Scheme,
			ScreenshotByte: content,
		}
	}

	return
}

func getHttpxFingers(httpxResult *HttpxResult, assetDocument *db.AssetDocument) {
	assetDocument.Title = httpxResult.Title
	assetDocument.Server = httpxResult.WebServer
	assetDocument.Service = httpxResult.Scheme
	if len(httpxResult.Fingers) > 0 {
		for _, f := range httpxResult.Fingers {
			assetDocument.App = append(assetDocument.App, f)
		}
	}
	return
}

func getFingerprintxFingers(service *FingerprintxResult, assetDocument *db.AssetDocument) {
	banner, err := json.Marshal(service.Raw)
	if err == nil {
		assetDocument.Banner = string(banner)
	}
	if service.Protocol != "" {
		assetDocument.Service = service.Protocol
	}
	return
}
