package controllers

import (
	"fmt"
	"github.com/beego/beego/v2/client/cache"
	"github.com/beego/beego/v2/server/web/captcha"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/db"
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
	cpt.ChallengeNums = 6
	cpt.StdWidth = 200
	cpt.StdHeight = 40
}

// IndexAction 登录首页
func (c *LoginController) IndexAction() {
	c.TplName = "login.html"
}

// LoginAction 登录验证
func (c *LoginController) LoginAction() {
	if conf.RunMode == conf.Release && !cpt.VerifyReq(c.Ctx.Request) {
		c.Redirect("/", http.StatusFound)
		return
	}
	userName := c.GetString("username")
	password := c.GetString("password")
	if userName != "" && password != "" {
		// 校验用户名、密码
		status, userData := ValidLoginUser(userName, password)
		if status {
			// 获取用户的可用工作空间
			userWorkspace := db.UserWorkspace{}
			userWorkspaceData := userWorkspace.GetsByUserId(userData.Id)
			var enabledUserWorkspaceData []db.Workspace
			for n := range userWorkspaceData {
				w := db.Workspace{Id: userWorkspaceData[n].WorkspaceId}
				if w.Get() && w.State == "enable" {
					enabledUserWorkspaceData = append(enabledUserWorkspaceData, w)
				}
			}
			// superadmin：允许同时管理多个workspace资源；普通管理员及guest，必须设置一个默认的workspace
			if len(enabledUserWorkspaceData) <= 0 {
				if userData.UserRole == SuperAdmin {
					c.SetSession("Workspace", 0)
				} else {
					logging.RuntimeLog.Infof("%s login from ip:%s,no available workspace set!", userData.UserName, c.Ctx.Input.IP())
					logging.CLILog.Infof("%s login from ip:%s,no available workspace set!", userData.UserName, c.Ctx.Input.IP())
					c.Redirect("/", http.StatusFound)
				}
			} else {
				// 默认关联第一个工作空间
				c.SetSession("Workspace", enabledUserWorkspaceData[0].Id)
			}
			c.SetSession("User", userData.UserName)
			c.SetSession("UserRole", userData.UserRole)
			logging.RuntimeLog.Infof("%s login from ip:%s", userData.UserName, c.Ctx.Input.IP())
			logging.CLILog.Infof("%s login from ip:%s", userData.UserName, c.Ctx.Input.IP())
			c.UpdateOnlineUser()
			c.Redirect("/dashboard", http.StatusFound)
		}
	}
	c.Redirect("/", http.StatusFound)
}

// LogoutAction 退出登录
func (c *LoginController) LogoutAction() {
	logging.RuntimeLog.Infof("logout from ip:%s", c.Ctx.Input.IP())
	logging.CLILog.Infof("logout from ip:%s", c.Ctx.Input.IP())
	c.DeleteOnlineUser()
	c.DelSession("User")
	c.DelSession("UserRole")
	c.DelSession("Workspace")
	c.Redirect("/", http.StatusFound)
}

// ValidLoginUser 校验用户名和密码
func ValidLoginUser(username, password string) (bool, db.User) {
	user := db.User{UserName: username}
	if user.GetByUsername() == false {
		return false, user
	}
	if user.State != "enable" {
		return false, user
	}

	configuredPassword := user.UserPassword
	hash := configuredPassword[:32]
	salt := configuredPassword[32:]
	checkedPass := utils.MD5V3(fmt.Sprintf("%s%s", password, salt))
	if checkedPass == hash {
		return true, user
	}

	return false, user
}

// UpdatePassword 更新密码
func UpdatePassword(userName, oldPassword, newPassword string) bool {
	validOldPassword, user := ValidLoginUser(userName, oldPassword)
	if validOldPassword {
		updateMap := make(map[string]interface{})
		updateMap["user_password"] = ProcessPasswordHash(newPassword)
		return user.Update(updateMap)
	}
	return false
}

// ProcessPasswordHash 根据明文密码生成带salt的hash密码
func ProcessPasswordHash(password string) string {
	salt := utils.GetRandomString2(16)
	hash := utils.MD5V3(fmt.Sprintf("%s%s", password, salt))
	return fmt.Sprintf("%s%s", hash, salt)
}

//password:648ce596dba3b408b523d3d1189b15070123456789abcdef -> nemo
