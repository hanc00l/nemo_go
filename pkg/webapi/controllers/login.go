package controllers

import (
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/db"
	"github.com/hanc00l/nemo_go/pkg/logging"
	ctrl "github.com/hanc00l/nemo_go/pkg/web/controllers"
)

type LoginController struct {
	ctrl.LoginController
}

// @Title Capture
// @Description 创建一个验证码id和对应的验证码（返回验证码id，通过请求/captcha/<id>.png来显示图形化的验证码）
// @Success 200 {object} models.StatusResponseData
// @router /captcha [get,post]
func (c *LoginController) Capture() {
	defer c.ServeJSON()

	captureId, err := ctrl.Cpt.CreateCaptcha()
	if err != nil || len(captureId) == 0 {
		c.FailedStatus("create capture fail")
		return
	}
	c.SucceededStatus(captureId)
}

// @Title Login
// @Description 用户登录
// @Param captcha_id 	formData string true "capture_id for capture verify"
// @Param captcha 		formData string true "capture code for capture verify"
// @Param username 		formData string true "the user name for login"
// @Param password 		formData string true "the password for login"
// @Success 200 {object} models.StatusResponseData
// @router /login [post]
func (c *LoginController) Login() {
	defer c.ServeJSON()

	if conf.RunMode == conf.Release && !ctrl.Cpt.VerifyReq(c.Ctx.Request) {
		c.FailedStatus("login fail")
		return
	}
	userName := c.GetString("username")
	password := c.GetString("password")
	if userName != "" && password != "" {
		// 校验用户名、密码
		status, userData := ctrl.ValidLoginUser(userName, password)
		if status {
			jwtData := ctrl.TokenData{User: userData.UserName, UserRole: userData.UserRole}
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
				if userData.UserRole == ctrl.SuperAdmin {
					jwtData.Workspace = 0
				} else {
					logging.RuntimeLog.Infof("%s login from ip:%s,no available workspace set!", userData.UserName, c.Ctx.Input.IP())
					logging.CLILog.Infof("%s login from ip:%s,no available workspace set!", userData.UserName, c.Ctx.Input.IP())
					c.FailedStatus("login fail,no available workspace set!")
					return
				}
			} else {
				// 默认关联第一个工作空间
				jwtData.Workspace = enabledUserWorkspaceData[0].Id
			}
			// 生成token并通过header返回
			token, _ := ctrl.GenerateToken(jwtData)
			c.Ctx.Output.Header("Authorization", ctrl.SetTokenValueToHeader(token))

			logging.RuntimeLog.Infof("%s login by web api from ip:%s", userData.UserName, c.Ctx.Input.IP())
			logging.CLILog.Infof("%s login by web api from ip:%s", userData.UserName, c.Ctx.Input.IP())
			c.UpdateOnlineUser()
			c.SucceededStatus("login success")
			return
		}
	}
	c.FailedStatus("login fail")
}
