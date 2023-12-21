package controllers

import (
	"github.com/hanc00l/nemo_go/pkg/conf"
	ctrl "github.com/hanc00l/nemo_go/pkg/web/controllers"
	"github.com/hanc00l/nemo_go/pkg/webapi/models"
	"os"
	"path/filepath"
	"strings"
)

type ConfigController struct {
	ctrl.ConfigController
}

// @Title LoadDefaultConfig
// @Description 获取默认的配置
// @Param authorization		header string true "token"
// @Success 200 {object} models.DefaultConfigData
// @router /load-default [post]
func (c *ConfigController) LoadDefaultConfig() {
	c.IsServerAPI = true
	defer c.ServeJSON()

	// worker config
	err := conf.GlobalWorkerConfig().ReloadConfig()
	if err != nil {
		c.Data["json"] = models.DefaultConfigData{}
		return
	}
	portscan := conf.GlobalWorkerConfig().Portscan
	fingerprint := conf.GlobalWorkerConfig().Fingerprint
	domainscan := conf.GlobalWorkerConfig().Domainscan
	onlineapi := conf.GlobalWorkerConfig().OnlineAPI
	// server config
	err = conf.GlobalServerConfig().ReloadConfig()
	if err != nil {
		c.Data["json"] = models.DefaultConfigData{}
		return
	}
	task := conf.GlobalServerConfig().Task

	data := models.DefaultConfigData{
		// portscan
		CmdBin: portscan.Cmdbin,
		Port:   portscan.Port,
		Rate:   portscan.Rate,
		Tech:   portscan.Tech,
		IsPing: portscan.IsPing,
		// fingerprint
		IsHttpx:          fingerprint.IsHttpx,
		IsScreenshot:     fingerprint.IsScreenshot,
		IsFingerprintHub: fingerprint.IsFingerprintHub,
		IsIconHash:       fingerprint.IsIconHash,
		IsFingerprintx:   fingerprint.IsFingerprintx,
		// domainscan
		Wordlist:           domainscan.Wordlist,
		IsSubDomainFinder:  domainscan.IsSubDomainFinder,
		IsSubDomainBrute:   domainscan.IsSubDomainBrute,
		IsSubDomainCrawler: domainscan.IsSubdomainCrawler,
		IsIgnoreCDN:        domainscan.IsIgnoreCDN,
		IsIgnoreOutofChina: domainscan.IsIgnoreOutofChina,
		IsPortscan:         domainscan.IsPortScan,
		IsWhois:            domainscan.IsWhois,
		IsICP:              domainscan.IsICP,
		//onlineAPI:
		IsFofa:   onlineapi.IsFofa,
		IsHunter: onlineapi.IsHunter,
		IsQuake:  onlineapi.IsQuake,
		// task
		IpSliceNumber:   task.IpSliceNumber,
		PortSliceNumber: task.PortSliceNumber,
	}
	if fileContent, err1 := os.ReadFile(filepath.Join(conf.GetRootPath(), "version.txt")); err1 == nil {
		data.Version = strings.TrimSpace(string(fileContent))
	}

	c.Data["json"] = data
}

// @Title SaveDefaultConfig
// @Description 保存默认的配置
// @Param authorization		header string true "token"
// @Param cmdbin			formData string true "端口扫描使用的程序"
// @Param port				formData string true "默认端口（nmap支持和格式，如--top-ports 1000）"
// @Param rate				formData int true "速率（默认1000）"
// @Param tech				formData string true "扫描技术（nmap支持的格式，如-sS，-sT，-sV），masscan只支持-sS"
// @Param ping				formData bool true "是否Ping（只支持nmap）"
// @Param wordlist			formData string true "Brute使用的子域名字典文件（默认：subnames.txt，9万条记录；较大的字典：subnames_medium.txt，88万条记录）"
// @Param subfinder			formData bool true "是否进行子域名枚举"
// @Param subdomainBrute	formData bool true "是否进行子域名Brute"
// @Param subdomainCrawler	formData bool true "是否进行子域名爬虫"
// @Param ignoreCDN			formData bool true "是否忽略CDN"
// @Param ignoreOutofChina	formData bool true "是否忽略非中国大陆IP"
// @Param portscan			formData bool true "是否对域名收集结果的IP进行端口扫描"
// @Param whois				formData bool true "是否执行whois"
// @Param icp				formData bool true "是否执行icp备案查询"
// @Param httpx				formData bool true "是否使用httpx获取指纹"
// @Param fingerprinthub	formData bool true "是否使用fingerprinthub获取指纹"
// @Param screenshot		formData bool true "是否进行屏幕截图"
// @Param iconhash			formData bool true "是否获取icon的哈希值"
// @Param ipslicenumber		formData int true "ip拆分的数量"
// @Param portslicenumber	formData int true "端口拆分的数量"
// @Param fofa				formData bool true "是否执行fofa"
// @Param hunter			formData bool true "是否执行hunter"
// @Param quake				formData bool true "是否执行quake"
// @Success 200 {object} models.WorkspaceDataTableResponseData
// @router /save-default [post]
func (c *ConfigController) SaveDefaultConfig() {
	c.IsServerAPI = true
	defer c.ServeJSON()
	if c.CheckMultiAccessRequest([]ctrl.RequestRole{ctrl.SuperAdmin, ctrl.Admin}, false) == false {
		c.FailedStatus("当前用户权限不允许！")
		return
	}

	data := models.DefaultConfigData{}
	err := c.ParseForm(&data)
	if err != nil {
		c.FailedStatus(err.Error())
		return
	}
	// server config
	err = conf.GlobalServerConfig().ReloadConfig()
	if err != nil {
		c.FailedStatus(err.Error())
		return
	}
	conf.GlobalServerConfig().Task.IpSliceNumber = data.IpSliceNumber
	conf.GlobalServerConfig().Task.PortSliceNumber = data.PortSliceNumber
	err = conf.GlobalServerConfig().WriteConfig()
	if err != nil {
		c.FailedStatus(err.Error())
		return
	}
	// worker config
	err = conf.GlobalWorkerConfig().ReloadConfig()
	if err != nil {
		c.FailedStatus(err.Error())
		return
	}
	conf.GlobalWorkerConfig().Portscan.Cmdbin = "masscan"
	if data.CmdBin == "nmap" {
		conf.GlobalWorkerConfig().Portscan.Cmdbin = "nmap"
	}
	//portscan
	conf.GlobalWorkerConfig().Portscan.Port = data.Port
	conf.GlobalWorkerConfig().Portscan.Rate = data.Rate
	conf.GlobalWorkerConfig().Portscan.Tech = data.Tech
	conf.GlobalWorkerConfig().Portscan.IsPing = data.IsPing
	//domainscan
	conf.GlobalWorkerConfig().Domainscan.Wordlist = data.Wordlist
	conf.GlobalWorkerConfig().Domainscan.IsSubDomainFinder = data.IsSubDomainFinder
	conf.GlobalWorkerConfig().Domainscan.IsSubDomainBrute = data.IsSubDomainBrute
	conf.GlobalWorkerConfig().Domainscan.IsSubdomainCrawler = data.IsSubDomainCrawler
	conf.GlobalWorkerConfig().Domainscan.IsIgnoreCDN = data.IsIgnoreCDN
	conf.GlobalWorkerConfig().Domainscan.IsIgnoreOutofChina = data.IsIgnoreOutofChina
	conf.GlobalWorkerConfig().Domainscan.IsPortScan = data.IsPortscan
	conf.GlobalWorkerConfig().Domainscan.IsICP = data.IsICP
	conf.GlobalWorkerConfig().Domainscan.IsWhois = data.IsWhois
	//fingerprint
	conf.GlobalWorkerConfig().Fingerprint.IsHttpx = data.IsHttpx
	conf.GlobalWorkerConfig().Fingerprint.IsFingerprintHub = data.IsFingerprintHub
	conf.GlobalWorkerConfig().Fingerprint.IsScreenshot = data.IsScreenshot
	conf.GlobalWorkerConfig().Fingerprint.IsIconHash = data.IsIconHash
	conf.GlobalWorkerConfig().Fingerprint.IsFingerprintx = data.IsFingerprintx
	//onlineapi
	conf.GlobalWorkerConfig().OnlineAPI.IsFofa = data.IsFofa
	conf.GlobalWorkerConfig().OnlineAPI.IsQuake = data.IsQuake
	conf.GlobalWorkerConfig().OnlineAPI.IsHunter = data.IsHunter
	err = conf.GlobalWorkerConfig().WriteConfig()
	if err != nil {
		c.FailedStatus(err.Error())
		return
	}
	c.SucceededStatus("save ok")
}

// @Title ChangePassword
// @Description 修改密码
// @Param authorization		header string true "token"
// @Param oldpass			formData string true "旧的密码"
// @Param newpass			formData string true "新的密码"
// @Success 200 {object} models.WorkspaceDataTableResponseData
// @router /changepass [post]
func (c *ConfigController) ChangePassword() {
	c.IsServerAPI = true
	c.ChangePasswordAction()
}
