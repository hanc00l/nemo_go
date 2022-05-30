package runner

type PortscanRequestParam struct {
	Target           string `form:"target"`
	IsPortScan       bool   `form:"portscan"`
	IsIPLocation     bool   `form:"iplocation"`
	IsFofa           bool   `form:"fofasearch"`
	IsQuake          bool   `form:"quakesearch"`
	IsHunter         bool   `form:"huntersearch"`
	Port             string `form:"port"`
	Rate             int    `form:"rate"`
	NmapTech         string `form:"nmap_tech"`
	CmdBin           string `form:"bin"`
	OrgId            int    `form:"org_id"`
	IsWhatweb        bool   `form:"whatweb"`
	IsHttpx          bool   `form:"httpx"`
	IsPing           bool   `form:"ping"`
	ExcludeIP        string `form:"exclude"`
	IsScreenshot     bool   `form:"screenshot"`
	IsWappalyzer     bool   `form:"wappalyzer"`
	IsFingerprintHub bool   `form:"fingerprinthub"`
	IsIconHash       bool   `form:"iconhash"`
	TaskMode         int    `form:"taskmode"`
	IsTaskCron       bool   `form:"taskcron" json:"-"`
	TaskCronRule     string `form:"cronrule" json:"-"`
	TaskCronComment  string `form:"croncomment" json:"-"`
}

type DomainscanRequestParam struct {
	Target           string `form:"target"`
	OrgId            int    `form:"org_id"`
	IsSubfinder      bool   `form:"subfinder"`
	IsSubdomainBrute bool   `form:"subdomainbrute"`
	IsFldDomain      bool   `form:"fld_domain"`
	IsWhatweb        bool   `form:"whatweb"`
	IsHttpx          bool   `form:"httpx"`
	IsIPPortscan     bool   `form:"portscan"`
	IsSubnetPortscan bool   `form:"networkscan"`
	IsCrawler        bool   `form:"crawler"`
	IsFofa           bool   `form:"fofasearch"`
	IsQuake          bool   `form:"quakesearch"`
	IsHunter         bool   `form:"huntersearch"`
	IsScreenshot     bool   `form:"screenshot"`
	IsICPQuery       bool   `form:"icpquery"`
	IsWappalyzer     bool   `form:"wappalyzer"`
	IsFingerprintHub bool   `form:"fingerprinthub"`
	IsIconHash       bool   `form:"iconhash"`
	TaskMode         int    `form:"taskmode"`
	PortTaskMode     int    `form:"porttaskmode"`
	IsTaskCron       bool   `form:"taskcron" json:"-"`
	TaskCronRule     string `form:"cronrule" json:"-"`
	TaskCronComment  string `form:"croncomment" json:"-"`
}

type PocscanRequestParam struct {
	Target           string `form:"target"`
	IsPocsuiteVerify bool   `form:"pocsuite3verify"`
	PocsuitePocFile  string `form:"pocsuite3_poc_file"`
	IsXrayVerify     bool   `form:"xrayverify"`
	XrayPocFile      string `form:"xray_poc_file"`
	IsNucleiVerify   bool   `form:"nucleiverify"`
	NucleiPocFile    string `form:"nuclei_poc_file"`
	IsDirsearch      bool   `form:"dirsearch"`
	DirsearchExtName string `form:"ext"`
	LoadOpenedPort   bool   `form:"load_opened_port"`
	IsTaskCron       bool   `form:"taskcron" json:"-"`
	TaskCronRule     string `form:"cronrule" json:"-"`
	TaskCronComment  string `form:"croncomment" json:"-"`
}
