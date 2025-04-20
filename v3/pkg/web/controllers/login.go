package controllers

import (
	"fmt"
	"github.com/beego/beego/v2/client/cache"
	"github.com/beego/beego/v2/server/web/captcha"
	"github.com/hanc00l/nemo_go/v3/pkg/conf"
	"github.com/hanc00l/nemo_go/v3/pkg/db"
	"github.com/hanc00l/nemo_go/v3/pkg/logging"
	"github.com/hanc00l/nemo_go/v3/pkg/utils"
	"go.mongodb.org/mongo-driver/v2/bson"
	"net/http"
)

type LoginController struct {
	BaseController
}

var Cpt *captcha.Captcha

func init() {
	// use beego cache system store the captcha data
	store := cache.NewMemoryCache()
	Cpt = captcha.NewWithFilter("/captcha/", store)
	Cpt.ChallengeNums = 6
	Cpt.StdWidth = 200
	Cpt.StdHeight = 40
}

// IndexAction 登录首页
func (c *LoginController) IndexAction() {
	c.TplName = "login.html"
}

// LoginAction 登录验证
func (c *LoginController) LoginAction() {
	if conf.RunMode == conf.Release && !Cpt.VerifyReq(c.Ctx.Request) {
		c.Redirect("/", http.StatusFound)
		return
	}
	userName := c.GetString("username", "")
	password := c.GetString("password", "")
	if len(userName) == 0 || len(password) == 0 {
		c.Redirect("/", http.StatusFound)
		return
	}
	// 校验用户名、密码
	success, userDoc := ValidLoginUser(userName, password)
	if success {
		workspaceList := GetUserAvailableWorkspaceList(userDoc.Username)
		if len(workspaceList) == 0 && userDoc.Role == SuperAdmin {
			workspaceList = GetSuperAdminDefaultWorkspaceList()
		}
		if len(workspaceList) == 0 {
			logging.RuntimeLog.Infof("%s login from ip:%s,no available workspace set!", userDoc.Username, c.Ctx.Input.IP())
			c.Redirect("/", http.StatusFound)
			return
		}
		_ = c.SetSession(Workspace, workspaceList[0].WorkspaceId)
		_ = c.SetSession(User, userDoc.Username)
		_ = c.SetSession(UserRole, userDoc.Role)
		logging.RuntimeLog.Infof("%s login from ip:%s", userDoc.Username, c.Ctx.Input.IP())
		//c.UpdateOnlineUser()
		c.Redirect("/dashboard", http.StatusFound)
	}

	c.Redirect("/", http.StatusFound)
}

// LogoutAction 退出登录
func (c *LoginController) LogoutAction() {
	userName := c.GetCurrentUser()
	if len(userName) == 0 {
		c.Redirect("/", http.StatusFound)
		return
	}
	logging.RuntimeLog.Infof("%s logout from ip:%s", userName, c.Ctx.Input.IP())
	//c.DeleteOnlineUser()
	_ = c.DelSession(User)
	_ = c.DelSession(UserRole)
	_ = c.DelSession(Workspace)
	c.Redirect("/", http.StatusFound)
}

// ValidLoginUser 校验用户名和密码
func ValidLoginUser(username, password string) (bool, *db.UserDocument) {
	mongoClient, err := db.GetClient()
	if err != nil {
		logging.RuntimeLog.Error(err)
		return false, nil
	}
	defer db.CloseClient(mongoClient)

	doc, err := db.NewUser(mongoClient).GetByName(username)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return false, nil
	}
	if doc.Username != username {
		return false, nil
	}
	if doc.Status != "enable" {
		return false, nil
	}
	if PasswordEncrypt(password) != doc.Password {
		return false, nil
	}

	return true, &doc
}

func GetUserAvailableWorkspaceList(userName string) (workspaceList []WorkspaceList) {
	mongoClient, err := db.GetClient()
	if err != nil {
		logging.RuntimeLog.Error(err)
		return
	}
	defer db.CloseClient(mongoClient)
	userDoc, err := db.NewUser(mongoClient).GetByName(userName)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return
	}
	workspace := db.NewWorkspace(mongoClient)
	for _, id := range userDoc.WorkspaceId {
		workspaceDocument, err := workspace.GetByWorkspaceId(id)
		if err != nil {
			continue
		}
		if workspaceDocument.Status == "enable" {
			workspaceList = append(workspaceList, WorkspaceList{Name: workspaceDocument.Name, WorkspaceId: workspaceDocument.WorkspaceId})
		}
	}
	return
}

func GetSuperAdminDefaultWorkspaceList() (workspaceList []WorkspaceList) {
	mongoClient, err := db.GetClient()
	if err != nil {
		logging.RuntimeLog.Error(err)
		return
	}
	defer db.CloseClient(mongoClient)

	workspace := db.NewWorkspace(mongoClient)
	workspaceData, err := workspace.Find(bson.M{"status": "enable"}, 0, 0)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return
	}
	for _, doc := range workspaceData {
		workspaceList = append(workspaceList, WorkspaceList{Name: doc.Name, WorkspaceId: doc.WorkspaceId})
	}
	return
}

func PasswordEncrypt(password string) string {
	return utils.SHA512(fmt.Sprintf("%s%s", SecuritySalt, password))
}
