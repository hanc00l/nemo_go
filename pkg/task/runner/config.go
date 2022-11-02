package runner

type PortscanRequestParam struct {
	Target             string `form:"target"`
	IsPortScan         bool   `form:"portscan"`
	IsIPLocation       bool   `form:"iplocation"`
	IsFofa             bool   `form:"fofasearch"`
	IsQuake            bool   `form:"quakesearch"`
	IsHunter           bool   `form:"huntersearch"`
	Port               string `form:"port"`
	Rate               int    `form:"rate"`
	NmapTech           string `form:"nmap_tech"`
	CmdBin             string `form:"bin"`
	OrgId              int    `form:"org_id"`
	IsHttpx            bool   `form:"httpx"`
	IsPing             bool   `form:"ping"`
	ExcludeIP          string `form:"exclude"`
	IsScreenshot       bool   `form:"screenshot"`
	IsFingerprintHub   bool   `form:"fingerprinthub"`
	IsIconHash         bool   `form:"iconhash"`
	TaskMode           int    `form:"taskmode"`
	IsTaskCron         bool   `form:"taskcron" json:"-"`
	TaskCronRule       string `form:"cronrule" json:"-"`
	TaskCronComment    string `form:"croncomment" json:"-"`
	IsLoadOpenedPort   bool   `form:"load_opened_port"`
	IsIgnoreOutofChina bool   `form:"ignoreoutofchina"`
	IsIgnoreCDN        bool   `form:"ignorecdn"`
}

type DomainscanRequestParam struct {
	Target             string `form:"target"`
	OrgId              int    `form:"org_id"`
	IsSubfinder        bool   `form:"subfinder"`
	IsSubdomainBrute   bool   `form:"subdomainbrute"`
	IsFldDomain        bool   `form:"fld_domain"`
	IsHttpx            bool   `form:"httpx"`
	IsIPPortscan       bool   `form:"portscan"`
	IsSubnetPortscan   bool   `form:"networkscan"`
	IsCrawler          bool   `form:"crawler"`
	IsFofa             bool   `form:"fofasearch"`
	IsQuake            bool   `form:"quakesearch"`
	IsHunter           bool   `form:"huntersearch"`
	IsScreenshot       bool   `form:"screenshot"`
	IsICPQuery         bool   `form:"icpquery"`
	IsWhoisQuery       bool   `form:"whoisquery"`
	IsFingerprintHub   bool   `form:"fingerprinthub"`
	IsIconHash         bool   `form:"iconhash"`
	TaskMode           int    `form:"taskmode"`
	PortTaskMode       int    `form:"porttaskmode"`
	IsTaskCron         bool   `form:"taskcron" json:"-"`
	TaskCronRule       string `form:"cronrule" json:"-"`
	TaskCronComment    string `form:"croncomment" json:"-"`
	IsIgnoreOutofChina bool   `form:"ignoreoutofchina"`
	IsIgnoreCDN        bool   `form:"ignorecdn"`
}

type PocscanRequestParam struct {
	Target           string `form:"target"`
	IsXrayVerify     bool   `form:"xrayverify"`
	XrayPocFile      string `form:"xray_poc_file"`
	IsNucleiVerify   bool   `form:"nucleiverify"`
	NucleiPocFile    string `form:"nuclei_poc_file"`
	IsDirsearch      bool   `form:"dirsearch"`
	DirsearchExtName string `form:"ext"`
	IsLoadOpenedPort bool   `form:"load_opened_port"`
	IsTaskCron       bool   `form:"taskcron" json:"-"`
	TaskCronRule     string `form:"cronrule" json:"-"`
	TaskCronComment  string `form:"croncomment" json:"-"`
}

type XScanRequestParam struct {
	XScanType string `form:"xscan_type"`
	Target    string `form:"target"`
	Port      string `form:"port"`
	OrgId     int    `form:"org_id"`
	IsCn      bool   `form:"is_CN"`

	IsXrayPocscan bool `form:"xraypoc"`
	IsFofaSearch  bool `form:"fofa"`

	IsFingerprint   bool   `form:"fingerprint"`
	IsTaskCron      bool   `form:"taskcron" json:"-"`
	TaskCronRule    string `form:"cronrule" json:"-"`
	TaskCronComment string `form:"croncomment" json:"-"`
	IsXrayPoc       bool   `form:"xraypoc"`
	PocFile         string `form:"pocfile"`

	IsOrgIP     bool
	IsOrgDomain bool
}

type taskKeySearchParam struct {
	KeyWord      string `json:"key_word"`
	SearchTime   string `json:"search_time"`
	ExcludeWords string `json:"exclude_words"`
	CheckMod     string `json:"check_mod"`
	Count        int    `json:"count"`
	IsCN         bool   `json:"is_CN"`
}

//var GlobalFilterWords = "网址大全||美食餐廳||网上开户||卡密充值||财经资讯||中工招商网||武义县公安局林警智治应用平台||哈尔滨隆腾尚云||行业信息网||公安采购商城||制造有限公司||蛋糕店配送||招标资源公共平台||网站制作||房产||广告||销售热线||博物馆||博客||销量||护肤||产品中心||产品展示||电话咨询||阻燃软包||刻章||服装厂家||品牌展示||案例欣赏||案例展示||企业文化||我们的优势||全国统一服务热线||解决方案||诚聘英才||招募热线||全国服务热线||商务合作||关注我们||关于我们||开锁服务||云领信息科技有限公司||公务员学习网||产品展示||产品与服务||专业承揽监控工程||工艺品有限公司||紧急开锁"
