package controllers

import (
	"fmt"
	"github.com/beego/beego/v2/client/cache"
	"github.com/beego/beego/v2/server/web/captcha"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/utils"
	"net/http"
)

type LoginController struct {
	BaseController
}

var cpt *captcha.Captcha

func init() {
	// use beego cache system store the captcha data
	store := cache.NewMemoryCache()
	cpt = captcha.NewWithFilter("/captcha/", store)
}

// IndexAction 登录首页
func (c *LoginController) IndexAction() {
	c.TplName = "login.html"
}

// LoginAction 登录验证
func (c *LoginController) LoginAction() {
	if conf.RunMode == conf.Release && !cpt.VerifyReq(c.Ctx.Request) {
		c.Redirect("/", http.StatusFound)
	}
	password := c.GetString("password")
	if password != "" && CheckPassword(password) {
		logging.RuntimeLog.Infof("login from ip:%s", c.Ctx.Input.IP())
		logging.CLILog.Infof("login from ip:%s", c.Ctx.Input.IP())
		c.UpdateOnlineUser()
		c.SetSession("IsLogin", true)
		c.Redirect("/dashboard", http.StatusFound)
	}
	c.Redirect("/", http.StatusFound)
}

// LogoutAction 退出登录
func (c *LoginController) LogoutAction() {
	logging.RuntimeLog.Infof("logout from ip:%s", c.Ctx.Input.IP())
	logging.CLILog.Infof("logout from ip:%s", c.Ctx.Input.IP())
	c.DeleteOnlineUser()
	c.DelSession("IsLogin")
	c.Redirect("/", http.StatusFound)
}

// CheckPassword 校验密码
func CheckPassword(password string) bool {
	conf.Nemo.ReloadConfig()

	configuredPassword := conf.Nemo.Web.Password
	hash := configuredPassword[:32]
	salt := configuredPassword[32:]
	checkedPass := utils.MD5V3(fmt.Sprintf("%s%s", password, salt))
	if checkedPass == hash {
		return true
	}
	return false
}

// UpdatePassword 更新密码
func UpdatePassword(newPassword string) bool {
	salt := utils.GetRandomString2(16)
	hash := utils.MD5V3(fmt.Sprintf("%s%s", newPassword, salt))

	conf.Nemo.ReloadConfig()
	conf.Nemo.Web.Password = fmt.Sprintf("%s%s", hash, salt)
	if err := conf.Nemo.WriteConfig(); err != nil {
		return false
	}
	return true
}

//password:648ce596dba3b408b523d3d1189b15070123456789abcdef -> nemo
