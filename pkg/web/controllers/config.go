package controllers

import (
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/utils"
	"os"
	"path/filepath"
)

type ConfigController struct {
	BaseController
}

const (
	HoneyPot    string = "honeypot"
	IPLoation   string = "iplocation"
	IPLocationB string = "iplocationB"
	IPLocationC string = "iplocationC"
	Service     string = "service"
)

type DefaultConfig struct {
	CmdBin          string `json:"cmdbin"`
	Port            string `json:"port"`
	Rate            int    `json:"rate"`
	Tech            string `json:"tech"`
	IsPing          bool   `json:"ping"`
	IpSliceNumber   int    `json:"ipslicenumber"`
	PortSliceNumber int    `json:"portslicenumber"`
}

func (c *ConfigController) IndexAction() {
	c.UpdateOnlineUser()
	c.Layout = "base.html"
	c.TplName = "config.html"
}

func (c *ConfigController) CustomAction() {
	c.UpdateOnlineUser()
	c.Layout = "base.html"
	c.TplName = "custom.html"
}

// LoadDefaultConfigAction 获取默认的端口扫描配置参数
func (c *ConfigController) LoadDefaultConfigAction() {
	conf.GlobalWorkerConfig().ReloadConfig()
	portscan := conf.GlobalWorkerConfig().Portscan
	task := conf.GlobalServerConfig().Task
	data := DefaultConfig{
		CmdBin:          portscan.Cmdbin,
		Port:            portscan.Port,
		Rate:            portscan.Rate,
		Tech:            portscan.Tech,
		IsPing:          portscan.IsPing,
		IpSliceNumber:   task.IpSliceNumber,
		PortSliceNumber: task.PortSliceNumber,
	}
	c.Data["json"] = data
	c.ServeJSON()
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
	if !CheckPassword(oldPass) {
		c.FailedStatus("校验旧密码失败！")
		return
	}
	if UpdatePassword(newPass) {
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
	c.SucceededStatus("")
}

// SaveTaskSliceNumberAction 保存任务切分设置
func (c *ConfigController) SaveTaskSliceNumberAction() {
	defer c.ServeJSON()

	ipSliceNumber, err1 := c.GetInt("ipslicenumber", utils.DefaultIpSliceNumber)
	portSliceNumber, err2 := c.GetInt("portslicenumber", utils.DefaultPortSliceNumber)
	if err1 != nil || err2 != nil {
		c.FailedStatus("数量错误")
		return
	}
	conf.GlobalServerConfig().ReloadConfig()
	conf.GlobalServerConfig().Task.IpSliceNumber = ipSliceNumber
	conf.GlobalServerConfig().Task.PortSliceNumber = portSliceNumber
	err := conf.GlobalServerConfig().WriteConfig()
	if err != nil {
		c.FailedStatus(err.Error())
	}
	c.SucceededStatus("")
}

// getCustomFilename  根据类型返回自定义文件名
func getCustomFilename(customType string) (customFile string) {
	switch customType {
	case HoneyPot:
		customFile = "custom/honeypot.txt"
	case IPLoation:
		customFile = "custom/iplocation-custom.txt"
	case IPLocationB:
		customFile = "custom/iplocation-custom-B.txt"
	case IPLocationC:
		customFile = "custom/iplocation-custom-C.txt"
	case Service:
		customFile = "custom/services-custom.txt"
	}
	return
}
