package models

import "time"

// StatusResponseData JSON的状态响应
type StatusResponseData struct {
	Status string `json:"status"`
	Msg    string `json:"msg"`
}

// IPListData IP列表显示数据
type IPListData struct {
	Id             int      `json:"id"`
	Index          int      `json:"index"`
	IP             string   `json:"ip"`
	Location       string   `json:"location"`
	Port           []string `json:"port"`
	Title          string   `json:"title"`
	Banner         string   `json:"banner"`
	ColorTag       string   `json:"color_tag"`
	MemoContent    string   `json:"memo_content"`
	Vulnerability  string   `json:"vulnerability"`
	HoneyPot       string   `json:"honeypot"`
	ScreenshotFile []string `json:"screenshot"`
	IsCDN          bool     `json:"cdn"`
	IconImage      []string `json:"iconimage"`
	WorkspaceId    int      `json:"workspace"`
	WorkspaceGUID  string   `json:"workspace_guid"`
	PinIndex       int      `json:"pinindex"`
}

// IPDataTableResponseData IP列表的返回数据
type IPDataTableResponseData struct {
	Draw            int          `json:"draw"`
	RecordsTotal    int          `json:"recordsTotal"`
	RecordsFiltered int          `json:"recordsFiltered"`
	Data            []IPListData `json:"data"`
}

// PortAttrInfo 每一个端口的详细数据
type PortAttrInfo struct {
	Id                 int
	PortId             int
	IP                 string
	Port               string
	Tag                string
	Content            string
	Source             string
	FofaLink           string
	CreateTime         string
	UpdateTime         string
	TableBackgroundSet bool
}

// ScreenshotFileInfo screenshot文件
type ScreenshotFileInfo struct {
	ScreenShotFile          string
	ScreenShotThumbnailFile string
	Tooltip                 string
}

// PortInfo 端口详细数据的集合
type PortInfo struct {
	PortNumbers      []int
	PortStatus       map[int]string
	TitleSet         map[string]struct{}
	BannerSet        map[string]struct{}
	PortAttr         []PortAttrInfo
	IconHashImageSet map[string]string
	TlsDataSet       map[string]struct{}
}

// VulnerabilityInfo 漏洞信息
type VulnerabilityInfo struct {
	Id         int
	Target     string
	Url        string
	PocFile    string
	Source     string
	Extra      string
	CreateTime string
	UpdateTime string
	Workspace  string
}

// IconHashWithFofa iconhash信息
type IconHashWithFofa struct {
	IconHash  string
	IconImage string
	FofaUrl   string
}

// IPInfo IP的详细数据的集合
type IPInfo struct {
	Id            int
	IP            string
	Organization  string
	Status        string
	Location      string
	Port          []int
	Title         []string
	Banner        []string
	PortAttr      []PortAttrInfo
	Domain        []string
	ColorTag      string
	Memo          string
	Vulnerability []VulnerabilityInfo
	CreateTime    string
	UpdateTime    string
	Screenshot    []ScreenshotFileInfo
	DisableFofa   bool
	IconHashes    []IconHashWithFofa
	TlsData       []string
	Workspace     string
	WorkspaceGUID string
	PinIndex      string
}

// TaskListData 任务的列表显示数据
type TaskListData struct {
	Id           int    `json:"id"`
	Index        string `json:"index"`
	TaskId       string `json:"task_id"`
	Worker       string `json:"worker"`
	TaskName     string `json:"task_name"`
	State        string `json:"state"`
	Result       string `json:"result"`
	KwArgs       string `json:"kwargs"`
	ReceivedTime string `json:"received"`
	StartedTime  string `json:"started"`
	CreateTime   string `json:"created"`
	UpdateTime   string `json:"updated"`
	Runtime      string `json:"runtime"`
	ResultFile   string `json:"resultfile"`
	TaskType     string `json:"tasktype"`
}

// TaskDataTableResponseData 任务的列表返回数据
type TaskDataTableResponseData struct {
	Draw            int            `json:"draw"`
	RecordsTotal    int            `json:"recordsTotal"`
	RecordsFiltered int            `json:"recordsFiltered"`
	Data            []TaskListData `json:"data"`
}

type TaskCronListData struct {
	Id          int    `json:"id"`
	Index       int    `json:"index"`
	TaskId      string `json:"task_id""`
	TaskName    string `json:"task_name"`
	Status      string `json:"status"`
	KwArgs      string `json:"kwargs"`
	CronRule    string `json:"cron_rule"`
	CreateTime  string `json:"create_time"`
	LastRunTime string `json:"lastrun_time"`
	NextRunTime string `json:"nextrun_time"`
	RunCount    int    `json:"run_count"`
	Comment     string `json:"comment"`
}

type TaskInfo struct {
	Id            int
	TaskId        string
	Worker        string
	TaskName      string
	State         string
	Result        string
	KwArgs        string
	ReceivedTime  string
	StartedTime   string
	SucceededTime string
	FailedTime    string
	RetriedTime   string
	RevokedTime   string
	Runtime       string
	CreateTime    string
	UpdateTime    string
	ResultFile    string
	RunTaskInfo   []TaskListData
	Workspace     string
}

type TaskCronInfo struct {
	Id          int
	TaskId      string
	TaskName    string
	Status      string
	KwArgs      string
	CronRule    string
	LastRunTime string
	CreateTime  string
	UpdateTime  string
	RunCount    int
	Comment     string
	Workspace   string
}

// DomainListData datable显示的每一行数据
type DomainListData struct {
	Id             int      `json:"id"`
	Index          int      `json:"index"`
	Domain         string   `json:"domain"`
	IP             []string `json:"ip"`
	Title          string   `json:"title"`
	Banner         string   `json:"banner"`
	ColorTag       string   `json:"color_tag"`
	MemoContent    string   `json:"memo_content"`
	Vulnerability  string   `json:"vulnerability"`
	HoneyPot       string   `json:"honeypot"`
	ScreenshotFile []string `json:"screenshot"`
	DomainCDN      string   `json:"domaincdn"`
	DomainCNAME    string   `json:"domaincname"`
	IsIPCDN        bool     `json:"ipcdn"`
	IconImage      []string `json:"iconimage"`
	WorkspaceId    int      `json:"workspace"`
	WorkspaceGUID  string   `json:"workspace_guid"`
	PinIndex       int      `json:"pinindex"`
}

// DomainDataTableResponseData 域名资产的列表返回数据
type DomainDataTableResponseData struct {
	Draw            int              `json:"draw"`
	RecordsTotal    int              `json:"recordsTotal"`
	RecordsFiltered int              `json:"recordsFiltered"`
	Data            []DomainListData `json:"data"`
}

// DomainInfo domain详细数据聚合
type DomainInfo struct {
	Id            int
	Domain        string
	Organization  string
	IP            []string
	Port          []int
	PortAttr      []PortAttrInfo
	Title         []string
	Banner        []string
	ColorTag      string
	Memo          string
	Vulnerability []VulnerabilityInfo
	CreateTime    string
	UpdateTime    string
	Screenshot    []ScreenshotFileInfo
	DomainAttr    []DomainAttrInfo
	DisableFofa   bool
	IconHashes    []IconHashWithFofa
	TlsData       []string
	DomainCDN     string
	DomainCNAME   string
	Workspace     string
	WorkspaceGUID string
	PinIndex      string
}

// DomainAttrInfo domain属性
type DomainAttrInfo struct {
	Id         int    `json:"id"`
	DomainId   int    `json:"domainId"`
	Port       int    `json:"port"`
	Tag        string `json:"tag"`
	Content    string `json:"content"`
	CreateTime string `json:"create_datetime"`
	UpdateTime string `json:"update_datetime"`
}

type VulnerabilityData struct {
	Id          int    `json:"id"`
	Index       int    `json:"index"`
	Target      string `json:"target"`
	Url         string `json:"url"`
	PocFile     string `json:"poc_file"`
	Source      string `json:"source"`
	CreateTime  string `json:"create_datetime"`
	UpdateTime  string `json:"update_datetime"`
	WorkspaceId int    `json:"workspace"`
}

// VulDataTableResponseData DataTable列表的返回数据
type VulDataTableResponseData struct {
	Draw            int                 `json:"draw"`
	RecordsTotal    int                 `json:"recordsTotal"`
	RecordsFiltered int                 `json:"recordsFiltered"`
	Data            []VulnerabilityData `json:"data"`
}

type PocFileList []string

type OrganizationData struct {
	Id             int    `json:"id" form:"id"`
	Index          int    `json:"index" form:"-"`
	OrgName        string `json:"org_name" form:"org_name"`
	Status         string `json:"status" form:"status"`
	SortOrder      int    `json:"sort_order" form:"sort_order"`
	CreateDatetime string `json:"create_time" form:"-"`
	UpdateDatetime string `json:"update_time" form:"-"`
	WorkspaceId    int    `json:"workspace"`
}

// OrgDataTableResponseData DataTable列表的返回数据
type OrgDataTableResponseData struct {
	Draw            int                `json:"draw"`
	RecordsTotal    int                `json:"recordsTotal"`
	RecordsFiltered int                `json:"recordsFiltered"`
	Data            []OrganizationData `json:"data"`
}

type OrganizationSelectData struct {
	Id      int    `json:"id"`
	OrgName string `json:"name"`
}

type OrganizationAllData []OrganizationSelectData

type UserData struct {
	Id              int    `json:"id" form:"id"`
	Index           int    `json:"index" form:"-"`
	UserName        string `json:"user_name" form:"user_name"`
	UserPassword    string `json:"user_password" form:"user_password"`
	UserDescription string `json:"user_description" form:"user_description"`
	UserRole        string `json:"user_role" form:"user_role"`
	State           string `json:"state" form:"state"`
	SortOrder       int    `json:"sort_order" form:"sort_order"`
	CreateDatetime  string `json:"create_time" form:"-"`
	UpdateDatetime  string `json:"update_time" form:"-"`
}

// UserDataTableResponseData DataTable列表的返回数据
type UserDataTableResponseData struct {
	Draw            int        `json:"draw"`
	RecordsTotal    int        `json:"recordsTotal"`
	RecordsFiltered int        `json:"recordsFiltered"`
	Data            []UserData `json:"data"`
}

type UserWorkspaceData []WorkspaceInfoData

type WorkspaceData struct {
	Id                   int    `json:"id" form:"id"`
	Index                int    `json:"index" form:"-"`
	WorkspaceName        string `json:"workspace_name" form:"workspace_name"`
	WorkspaceGUID        string `json:"workspace_guid" form:"workspace_guid"`
	WorkspaceDescription string `json:"workspace_description" form:"workspace_description"`
	State                string `json:"state" form:"state"`
	SortOrder            int    `json:"sort_order" form:"sort_order"`
	CreateDatetime       string `json:"create_time" form:"-"`
	UpdateDatetime       string `json:"update_time" form:"-"`
}

// WorkspaceDataTableResponseData DataTable列表的返回数据
type WorkspaceDataTableResponseData struct {
	Draw            int             `json:"draw"`
	RecordsTotal    int             `json:"recordsTotal"`
	RecordsFiltered int             `json:"recordsFiltered"`
	Data            []WorkspaceData `json:"data"`
}

type WorkspaceInfoData struct {
	WorkspaceId   string `json:"workspaceId"`
	WorkspaceName string `json:"workspaceName"`
	Enable        bool   `json:"enable"`
}

type WorkspaceInfo struct {
	WorkspaceInfoList []WorkspaceInfoData
	CurrentWorkspace  string
}

type DefaultConfigData struct {
	// portscan
	CmdBin string `json:"cmdbin"`
	Port   string `json:"port"`
	Rate   int    `json:"rate"`
	Tech   string `json:"tech"`
	IsPing bool   `json:"ping"`
	// domainscan
	Wordlist           string `json:"wordlist"`
	IsSubDomainFinder  bool   `json:"subfinder"`
	IsSubDomainBrute   bool   `json:"subdomainBrute"`
	IsSubDomainCrawler bool   `json:"subdomainCrawler"`
	IsIgnoreCDN        bool   `json:"ignoreCDN"`
	IsIgnoreOutofChina bool   `json:"ignoreOutofChina"`
	IsPortscan         bool   `json:"portscan"`
	IsWhois            bool   `json:"whois"`
	IsICP              bool   `json:"icp"`
	// fingerprint
	IsHttpx          bool `json:"httpx"`
	IsScreenshot     bool `json:"screenshot"`
	IsFingerprintHub bool `json:"fingerprinthub"`
	IsIconHash       bool `json:"iconhash"`
	IsFingerprintx   bool `json:"fingerprintx"`
	// onlineapi
	IsFofa   bool `json:"fofa"`
	IsQuake  bool `json:"quake"`
	IsHunter bool `json:"hunter"`
	// task
	IpSliceNumber   int `json:"ipslicenumber"`
	PortSliceNumber int `json:"portslicenumber"`
	// version
	Version string `json:"version"`
}

type DashboardStatisticData struct {
	IP            int `json:"ip_count"`
	Domain        int `json:"domain_count"`
	Vulnerability int `json:"vulnerability_count"`
	ActiveTask    int `json:"task_active"`
}

type WorkerStatusData struct {
	Index                    int    `json:"index"`
	WorkName                 string `json:"worker_name"`
	CreateTime               string `json:"create_time"`
	UpdateTime               string `json:"update_time"`
	TaskExecutedNumber       int    `json:"task_number"`
	EnableManualReloadFlag   bool   `json:"enable_manual_reload_flag"`
	EnableManualFileSyncFlag bool   `json:"enable_manual_file_sync_flag"`
	HeartColor               string `json:"heart_color"`
}

type TaskInfoData struct {
	TaskInfo string `json:"task_info"`
}

// OnlineUserDataTableResponseData DataTable列表的返回数据
type OnlineUserDataTableResponseData struct {
	Draw            int              `json:"draw"`
	RecordsTotal    int              `json:"recordsTotal"`
	RecordsFiltered int              `json:"recordsFiltered"`
	Data            []OnlineUserInfo `json:"data"`
}

// OnlineUserInfo 在线用户
type OnlineUserInfo struct {
	IP           string
	LoginTime    time.Time
	UpdateTime   time.Time
	UpdateNumber int64
}
