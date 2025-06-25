package execute

import "time"

const (
	LLMAPI      = "llmapi"
	PortScan    = "portscan"
	DomainScan  = "domainscan"
	OnlineAPI   = "onlineapi"
	FingerPrint = "fingerprint"
	PocScan     = "pocscan"
	Standalone  = "standalone"
)

type MainTaskInfo struct {
	Target        string `json:"target"`
	ExcludeTarget string `json:"executeTarget,omitempty"`

	ExecutorConfig `json:"config"`
	OrgId          string `json:"org,omitempty"`
	WorkspaceId    string `json:"workspaceId,omitempty"`
	MainTaskId     string `json:"mainTaskId,omitempty"`
	IsProxy        bool   `json:"proxy,omitempty"`

	TargetSliceType int `json:"targetSliceType,omitempty"`
	TargetSliceNum  int `json:"targetSliceNum,omitempty"`

	StartTime *time.Time `json:"startTime,omitempty"`
	EndTime   *time.Time `json:"endTime,omitempty"`
}

type ExecutorTaskInfo struct {
	Executor  string `json:"executor"`
	TaskId    string `json:"taskId"`
	PreTaskId string `json:"preTaskId"`

	MainTaskInfo
}

type ExecutorConfig struct {
	PortScan    map[string]PortscanConfig       `json:"portscan,omitempty" form:"portscan"`
	DomainScan  map[string]DomainscanConfig     `json:"domainscan,omitempty" form:"domainscan"`
	OnlineAPI   map[string]OnlineAPIConfig      `json:"onlineapi,omitempty" form:"onlineapi"`
	FingerPrint map[string]FingerprintConfig    `json:"fingerprint,omitempty" form:"fingerprint"`
	PocScan     map[string]PocscanConfig        `json:"pocscan,omitempty" form:"pocscan"`
	LLMAPI      map[string]LLMAPIConfig         `json:"llmapi,omitempty" form:"llmapi"`
	Standalone  map[string]StandaloneScanConfig `json:"standalone,omitempty" form:"standalone"`
}

type PortscanConfig struct {
	Target        string `json:"target,omitempty" form:"-"`
	ExcludeTarget string `json:"executeTarget,omitempty" form:"-"`
	Port          string `json:"port,omitempty" form:"port"`
	Rate          int    `json:"rate,omitempty" form:"rate"`
	IsPing        bool   `json:"ping,omitempty" form:"ping"`
	Tech          string `json:"tech,omitempty" form:"tech"`
	// 扫描结果中单个IP最多开放的端口数量，超过该数量则跳过该IP
	MaxOpenedPortPerIp int `json:"maxOpenedPortPerIp,omitempty" form:"maxOpenedPortPerIp"`
}

type DomainscanConfig struct {
	Target               string          `json:"target,omitempty" form:"-"`
	WordlistFile         string          `json:"wordlistFile,omitempty" form:"wordlistFile"`
	IsIPPortScan         bool            `json:"portscan,omitempty" form:"portscan"`
	IsIPSubnetPortScan   bool            `json:"subnetPortscan,omitempty" form:"subnetPortscan"`
	IsIgnoreCDN          bool            `json:"ignorecdn,omitempty" form:"ignorecdn"`
	IsIgnoreChinaOther   bool            `json:"ignorechinaother,omitempty" form:"ignorechinaother"`
	IsIgnoreOutsideChina bool            `json:"ignoreoutsidechina,omitempty" form:"ignoreoutsidechina"`
	ResultPortscanBin    string          `json:"resultPortscanBin,omitempty" form:"resultPortscanBin"`
	ResultPortscanConfig *PortscanConfig `json:"resultPortscanConfig,omitempty" form:"resultPortscanConfig"`
	// 结果中同一个IP对应的域名最多数量，超过该数量则这些域名被忽略
	MaxResolvedDomainPerIP int `json:"maxResolvedDomainPerIP,omitempty" form:"maxResolvedDomainPerIP"`
}

type FingerprintConfig struct {
	Target         string `json:"target,omitempty" form:"-"`
	IsHttpx        bool   `json:"httpx,omitempty" form:"httpx"`
	IsFingerprintx bool   `json:"fingerprintx,omitempty" form:"fingerprintx"`
	IsScreenshot   bool   `json:"screenshot,omitempty" form:"screenshot"`
	IsIconHash     bool   `json:"iconhash,omitempty" form:"iconhash"`
}

type OnlineAPIConfig struct {
	Target               string `json:"target,omitempty" form:"-"`
	IsIgnoreCDN          bool   `json:"ignorecdn,omitempty" form:"ignorecdn"`
	IsIgnoreChinaOther   bool   `json:"ignorechinaother,omitempty" form:"ignorechinaother"`
	IsIgnoreOutsideChina bool   `json:"ignoreoutsidechina,omitempty" form:"ignoreoutsidechina"`
	SearchByKeyWord      bool   `json:"keywordsearch,omitempty" form:"keywordsearch"`
	SearchStartTime      string `json:"searchstarttime,omitempty" form:"searchstarttime"`
	SearchLimitCount     int    `json:"searchlimitcount,omitempty" form:"searchlimitcount"`
	SearchPageSize       int    `json:"searchpagesize,omitempty" form:"searchpagesize"`
	// 扫描结果中单个IP最多开放的端口数量，超过该数量则跳过该IP
	MaxOpenedPortPerIp int `json:"maxOpenedPortPerIp,omitempty" form:"maxOpenedPortPerIp"`
	// 结果中同一个IP对应的域名最多数量，超过该数量则这些域名被忽略
	MaxResolvedDomainPerIP int `json:"maxResolvedDomainPerIP,omitempty" form:"maxResolvedDomainPerIP"`
}

type PocscanConfig struct {
	Target              string `json:"target,omitempty" `
	PocType             string `json:"pocType" form:"pocType"`
	PocFile             string `json:"pocFile,omitempty" form:"pocFile"`
	IsBrutePassword     bool   `json:"brutePassword,omitempty" form:"brutePassword"`
	IsScanBaseWebStatus bool   `json:"baseWebStatus,omitempty" form:"baseWebStatus"`
}

type StandaloneScanConfig struct {
	Target             string `json:"target,omitempty" form:"-"`
	Port               string `json:"port,omitempty" form:"port"`
	ExcludeTarget      string `json:"executeTarget,omitempty" form:"-"`
	Rate               int    `json:"rate,omitempty" form:"rate"`
	IsPing             bool   `json:"ping,omitempty" form:"ping"`
	IsVerbose          bool   `json:"verbose,omitempty" form:"verbose"`
	IsPocScan          bool   `json:"pocscan,omitempty" form:"pocscan"`
	MaxOpenedPortPerIp int    `json:"maxOpenedPortPerIp,omitempty" form:"maxOpenedPortPerIp"`
}

type LLMAPIConfig struct {
	Target           string `json:"target,omitempty" form:"-"`
	AutoAssociateOrg bool   `json:"autoAssociateOrg,omitempty" form:"autoAssociateOrg"`
}
