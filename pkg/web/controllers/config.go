package controllers

import (
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/notify"
	"github.com/hanc00l/nemo_go/pkg/task/onlineapi"
	"github.com/hanc00l/nemo_go/pkg/utils"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type ConfigController struct {
	BaseController
}

const (
	HoneyPot    string = "honeypot"
	IPLocation  string = "iplocation"
	IPLocationB string = "iplocationB"
	IPLocationC string = "iplocationC"
	Service     string = "service"
	Xray        string = "xray.yaml"
	XrayConfig  string = "config.yaml"
	XrayPlugin  string = "plugin.xray.yaml"
	XrayModule  string = "module.xray.yaml"
)

type DefaultConfig struct {
	CmdBin           string `json:"cmdbin"`
	Port             string `json:"port"`
	Rate             int    `json:"rate"`
	Tech             string `json:"tech"`
	IsPing           bool   `json:"ping"`
	IpSliceNumber    int    `json:"ipslicenumber"`
	PortSliceNumber  int    `json:"portslicenumber"`
	Version          string `json:"version"`
	IsHttpx          bool   `json:"httpx"`
	IsScreenshot     bool   `json:"screenshot"`
	IsFingerprintHub bool   `json:"fingerprinthub"`
	IsIconHash       bool   `json:"iconhash"`
	ServerChanToken  string `json:"serverchan"`
	DingTalkToken    string `json:"dingtalk"`
	FeishuToken      string `json:"feishu"`
	FofaUser         string `json:"fofauser"`
	FofaToken        string `json:"fofatoken"`
	HunterToken      string `json:"huntertoken"`
	QuakeToken       string `json:"quaketoken"`
	ChinazToken      string `json:"chinaztoken"`
}

func (c *ConfigController) IndexAction() {
	c.UpdateOnlineUser()
	c.Layout = "base.html"
	c.TplName = "config.html"
}

func (c *ConfigController) CustomAction() {
	c.CheckMultiAccessRequest([]RequestRole{SuperAdmin, Admin}, true)

	c.UpdateOnlineUser()
	c.Layout = "base.html"
	c.TplName = "custom.html"
}

// LoadDefaultConfigAction 获取默认的端口扫描配置参数
func (c *ConfigController) LoadDefaultConfigAction() {
	defer c.ServeJSON()

	err := conf.GlobalWorkerConfig().ReloadConfig()
	if err != nil {
		c.FailedStatus(err.Error())
		return
	}
	portscan := conf.GlobalWorkerConfig().Portscan
	task := conf.GlobalServerConfig().Task
	fingerprint := conf.GlobalWorkerConfig().Fingerprint
	notify := conf.GlobalServerConfig().Notify
	apiConfig := conf.GlobalWorkerConfig().API
	data := DefaultConfig{
		CmdBin:           portscan.Cmdbin,
		Port:             portscan.Port,
		Rate:             portscan.Rate,
		Tech:             portscan.Tech,
		IsPing:           portscan.IsPing,
		IpSliceNumber:    task.IpSliceNumber,
		PortSliceNumber:  task.PortSliceNumber,
		IsHttpx:          fingerprint.IsHttpx,
		IsScreenshot:     fingerprint.IsScreenshot,
		IsFingerprintHub: fingerprint.IsFingerprintHub,
		IsIconHash:       fingerprint.IsIconHash,
		ServerChanToken:  notify["serverchan"].Token,
		DingTalkToken:    notify["dingtalk"].Token,
		FeishuToken:      notify["feishu"].Token,
		FofaUser:         apiConfig.Fofa.Name,
		FofaToken:        apiConfig.Fofa.Key,
		HunterToken:      apiConfig.Hunter.Key,
		QuakeToken:       apiConfig.Quake.Key,
		ChinazToken:      apiConfig.ICP.Key,
	}
	if fileContent, err1 := os.ReadFile(filepath.Join(conf.GetRootPath(), "version.txt")); err1 == nil {
		data.Version = string(fileContent)
	}
	c.Data["json"] = data
}

// ChangePasswordAction 修改密码
func (c *ConfigController) ChangePasswordAction() {
	defer c.ServeJSON()

	oldPass := c.GetString("oldpass", "")
	newPass := c.GetString("newpass", "")
	if oldPass == "" || newPass == "" {
		c.FailedStatus("密码为空！")
		return
	}
	userName := c.GetSession("User").(string)
	if len(userName) == 0 {
		c.FailedStatus("修改密码失败！")
		return
	}
	if UpdatePassword(userName, oldPass, newPass) {
		c.SucceededStatus("OK！")
	} else {
		c.FailedStatus("修改密码失败！")
	}
}

// LoadCustomConfigAction 加载一个自定义文件
func (c *ConfigController) LoadCustomConfigAction() {
	defer c.ServeJSON()

	customType := c.GetString("type", "")
	if customType == "" {
		c.FailedStatus("未指定类型")
		return
	}
	customFile := getCustomFilename(customType)
	if customFile == "" {
		c.FailedStatus("错误的类型")
		return
	}
	content, err := os.ReadFile(filepath.Join(conf.GetRootPath(), "thirdparty", customFile))
	if err != nil {
		c.FailedStatus(err.Error())
		return
	}
	c.SucceededStatus(string(content))
}

// SaveCustomConfigAction 保存一个自定义文件
func (c *ConfigController) SaveCustomConfigAction() {
	defer c.ServeJSON()
	if c.CheckMultiAccessRequest([]RequestRole{SuperAdmin, Admin}, false) == false {
		c.FailedStatus("当前用户权限不允许！")
		return
	}

	customType := c.GetString("type", "")
	customContent := c.GetString("content", "")
	if customType == "" || customContent == "" {
		c.FailedStatus("未指定类型或内容")
		return
	}
	customFile := getCustomFilename(customType)
	if customFile == "" {
		c.FailedStatus("错误的类型")
		return
	}
	err := os.WriteFile(filepath.Join(conf.GetRootPath(), "thirdparty", customFile), []byte(customContent), 0666)
	if err != nil {
		c.FailedStatus(err.Error())
		return
	}
	c.SucceededStatus("保存配置成功")
}

// SaveTaskSliceNumberAction 保存任务切分设置
func (c *ConfigController) SaveTaskSliceNumberAction() {
	defer c.ServeJSON()
	if c.CheckMultiAccessRequest([]RequestRole{SuperAdmin, Admin}, false) == false {
		c.FailedStatus("当前用户权限不允许！")
		return
	}

	ipSliceNumber, err1 := c.GetInt("ipslicenumber", utils.DefaultIpSliceNumber)
	portSliceNumber, err2 := c.GetInt("portslicenumber", utils.DefaultPortSliceNumber)
	if err1 != nil || err2 != nil {
		c.FailedStatus("数量错误")
		return
	}
	err := conf.GlobalServerConfig().ReloadConfig()
	if err != nil {
		c.FailedStatus(err.Error())
		return
	}
	conf.GlobalServerConfig().Task.IpSliceNumber = ipSliceNumber
	conf.GlobalServerConfig().Task.PortSliceNumber = portSliceNumber
	err = conf.GlobalServerConfig().WriteConfig()
	if err != nil {
		c.FailedStatus(err.Error())
	}
	c.SucceededStatus("保存配置成功")
}

// SaveTaskNotifyAction 保存任务通知的Token设置
func (c *ConfigController) SaveTaskNotifyAction() {
	defer c.ServeJSON()
	if c.CheckMultiAccessRequest([]RequestRole{SuperAdmin, Admin}, false) == false {
		c.FailedStatus("当前用户权限不允许！")
		return
	}

	serverChanToken := c.GetString("token_serverchan", "")
	dingtalkToken := c.GetString("token_dingtalk", "")
	feishuToken := c.GetString("token_feishu", "")

	err := conf.GlobalServerConfig().ReloadConfig()
	if err != nil {
		c.FailedStatus(err.Error())
		return
	}
	if conf.GlobalServerConfig().Notify == nil {
		conf.GlobalServerConfig().Notify = make(map[string]conf.Notify)
	}
	conf.GlobalServerConfig().Notify["serverchan"] = conf.Notify{Token: serverChanToken}
	conf.GlobalServerConfig().Notify["dingtalk"] = conf.Notify{Token: dingtalkToken}
	conf.GlobalServerConfig().Notify["feishu"] = conf.Notify{Token: feishuToken}

	err = conf.GlobalServerConfig().WriteConfig()
	if err != nil {
		c.FailedStatus(err.Error())
	}
	c.SucceededStatus("保存配置成功")
}

// TestTaskNotifyAction 测试任务通知
func (c *ConfigController) TestTaskNotifyAction() {
	defer c.ServeJSON()
	if c.CheckMultiAccessRequest([]RequestRole{SuperAdmin, Admin}, false) == false {
		c.FailedStatus("当前用户权限不允许！")
		return
	}

	message := "这是一个测试消息，来自Nemo的配置管理！"
	notify.Send(message)

	c.SucceededStatus("已发送测试通知，请确认消息是否正确！")
}

// SaveAPITokenAction 保存API的Token
func (c *ConfigController) SaveAPITokenAction() {
	defer c.ServeJSON()
	if c.CheckMultiAccessRequest([]RequestRole{SuperAdmin, Admin}, false) == false {
		c.FailedStatus("当前用户权限不允许！")
		return
	}

	fofaUser := c.GetString("fofa_user", "")
	fofaToken := c.GetString("fofa_token", "")
	hunterToken := c.GetString("hunter_token", "")
	quakeToken := c.GetString("quake_token", "")
	chinazToken := c.GetString("chinaz_token", "")

	err := conf.GlobalWorkerConfig().ReloadConfig()
	if err != nil {
		c.FailedStatus(err.Error())
		return
	}
	conf.GlobalWorkerConfig().API.Fofa.Name = fofaUser
	conf.GlobalWorkerConfig().API.Fofa.Key = fofaToken
	conf.GlobalWorkerConfig().API.Hunter.Key = hunterToken
	conf.GlobalWorkerConfig().API.Quake.Key = quakeToken
	conf.GlobalWorkerConfig().API.ICP.Key = chinazToken

	err = conf.GlobalWorkerConfig().WriteConfig()
	if err != nil {
		c.FailedStatus(err.Error())
	}
	c.SucceededStatus("保存配置成功")
}

// OnlineTestAPITokenAction 在线测试API的Token是否可用
func (c *ConfigController) OnlineTestAPITokenAction() {
	defer c.ServeJSON()
	if c.CheckMultiAccessRequest([]RequestRole{SuperAdmin, Admin}, false) == false {
		c.FailedStatus("当前用户权限不允许！")
		return
	}
	sb := strings.Builder{}
	//FOFA
	config := onlineapi.OnlineAPIConfig{Target: "fofa.info"}
	fofa := onlineapi.NewFofa(config)
	fofa.Do()
	if len(fofa.DomainResult.DomainResult) > 0 || len(fofa.IpResult.IPResult) > 0 {
		sb.WriteString("Fofa: OK!\n")
	} else {
		sb.WriteString("Fofa: Fail!\n")
	}
	//HUNTER
	hunter := onlineapi.NewHunter(config)
	hunter.Do()
	if len(hunter.DomainResult.DomainResult) > 0 || len(hunter.IpResult.IPResult) > 0 {
		sb.WriteString("Hunter: OK!\n")
	} else {
		sb.WriteString("Hunter: Fail!\n")
	}
	//Quake
	quake := onlineapi.NewQuake(config)
	quake.Do()
	if len(quake.DomainResult.DomainResult) > 0 || len(quake.IpResult.IPResult) > 0 {
		sb.WriteString("Quake: OK!\n")
	} else {
		sb.WriteString("Quake: Fail!\n")
	}
	//ICP
	icp := onlineapi.NewICPQuery(onlineapi.ICPQueryConfig{})
	if icp.RunICPQuery("10086.cn") != nil {
		sb.WriteString("ICP: OK!\n")
	} else {
		sb.WriteString("ICP: Fail!\n")
	}

	c.SucceededStatus(sb.String())
}

// SaveFingerprintAction 保存默认指纹设置
func (c *ConfigController) SaveFingerprintAction() {
	defer c.ServeJSON()
	if c.CheckMultiAccessRequest([]RequestRole{SuperAdmin, Admin}, false) == false {
		c.FailedStatus("当前用户权限不允许！")
		return
	}

	httpx, err1 := c.GetBool("httpx", true)
	fingerprinthub, err2 := c.GetBool("fingerprinthub", true)
	screenshot, err3 := c.GetBool("screenshot", true)
	iconhash, err4 := c.GetBool("iconhash", true)

	if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
		c.FailedStatus("配置参数错误！")
		return
	}
	err := conf.GlobalWorkerConfig().ReloadConfig()
	if err != nil {
		c.FailedStatus(err.Error())
		return
	}

	conf.GlobalWorkerConfig().Fingerprint.IsHttpx = httpx
	conf.GlobalWorkerConfig().Fingerprint.IsScreenshot = screenshot
	conf.GlobalWorkerConfig().Fingerprint.IsFingerprintHub = fingerprinthub
	conf.GlobalWorkerConfig().Fingerprint.IsIconHash = iconhash
	err = conf.GlobalWorkerConfig().WriteConfig()
	if err != nil {
		c.FailedStatus(err.Error())
	}
	c.SucceededStatus("保存配置成功")
}

// SavePortscanAction 保存默认端口扫描设置
func (c *ConfigController) SavePortscanAction() {
	defer c.ServeJSON()
	if c.CheckMultiAccessRequest([]RequestRole{SuperAdmin, Admin}, false) == false {
		c.FailedStatus("当前用户权限不允许！")
		return
	}

	cmdbin := c.GetString("cmdbin", "masscan")
	port := c.GetString("port", "--top-ports 1000")
	rate, err1 := c.GetInt("rate", 1000)
	tech := c.GetString("tech", "-sS")
	ping, err2 := c.GetBool("ping", false)
	if err1 != nil || err2 != nil {
		c.FailedStatus("配置参数错误！")
		return
	}
	err := conf.GlobalWorkerConfig().ReloadConfig()
	if err != nil {
		c.FailedStatus(err.Error())
		return
	}

	conf.GlobalWorkerConfig().Portscan.Cmdbin = "masscan"
	if cmdbin == "nmap" {
		conf.GlobalWorkerConfig().Portscan.Cmdbin = "nmap"
	}
	conf.GlobalWorkerConfig().Portscan.Port = port
	conf.GlobalWorkerConfig().Portscan.Rate = rate
	conf.GlobalWorkerConfig().Portscan.Tech = tech
	conf.GlobalWorkerConfig().Portscan.IsPing = ping
	err = conf.GlobalWorkerConfig().WriteConfig()
	if err != nil {
		c.FailedStatus(err.Error())
	}
	c.SucceededStatus("保存配置成功")
}

// UploadPocAction xraypoc的上传
func (c *ConfigController) UploadPocAction() {
	defer c.ServeJSON()
	if c.CheckMultiAccessRequest([]RequestRole{SuperAdmin, Admin}, false) == false {
		c.FailedStatus("当前用户权限不允许！")
		return
	}

	pocType := c.GetString("type", "")
	if len(pocType) == 0 {
		logging.RuntimeLog.Error("get poc file error ")
		return
	}
	// 获取上传信息
	f, h, err := c.GetFile("file")
	if err != nil {
		logging.RuntimeLog.Error("get file err ", err)
		c.FailedStatus(err.Error())
		return
	}
	defer f.Close()
	// 检查文件后缀
	fileExt := path.Ext(h.Filename)
	if fileExt != ".yml" && fileExt != ".yaml" {
		logging.RuntimeLog.Error("invalid file type!")
		c.FailedStatus("invalid file type!")
		return
	}
	// 保存到poc目录下
	var pocSavedPathName string
	if pocType == "xray" {
		pocSavedPathName = filepath.Join(conf.GetRootPath(), conf.GlobalWorkerConfig().Pocscan.Xray.PocPath, h.Filename)
	} else if pocType == "nuclei" {
		pocSavedPathName = filepath.Join(conf.GetRootPath(), conf.GlobalWorkerConfig().Pocscan.Nuclei.PocPath, h.Filename)
	}
	if pocSavedPathName == "" {
		logging.RuntimeLog.Error("invalid poc type!")
		c.FailedStatus("invalid poc type!")
		return
	}
	err = c.SaveToFile("file", pocSavedPathName)
	if err != nil {
		logging.RuntimeLog.Error("save file err ", err)
		c.FailedStatus(err.Error())
		return
	}
	c.SucceededStatus("上传成功")
}

// getCustomFilename  根据类型返回自定义文件名
func getCustomFilename(customType string) (customFile string) {
	switch customType {
	case HoneyPot:
		customFile = "custom/honeypot.txt"
	case IPLocation:
		customFile = "custom/iplocation-custom.txt"
	case IPLocationB:
		customFile = "custom/iplocation-custom-B.txt"
	case IPLocationC:
		customFile = "custom/iplocation-custom-C.txt"
	case Service:
		customFile = "custom/services-custom.txt"
	case Xray:
		customFile = "xray/xray.yaml"
	case XrayConfig:
		customFile = "xray/config.yaml"
	case XrayModule:
		customFile = "xray/module.xray.yaml"
	case XrayPlugin:
		customFile = "xray/plugin.xray.yaml"

	}
	return
}
