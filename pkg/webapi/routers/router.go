// @APIVersion 1.0.0
// @Title beego Test API
// @Description beego has a very cool tools to autogenerate documents for your API
// @Contact astaxie@gmail.com
// @TermsOfServiceUrl http://beego.me/
// @License Apache 2.0
// @LicenseUrl http://www.apache.org/licenses/LICENSE-2.0.html
package routers

import (
	beego "github.com/beego/beego/v2/server/web"
	"github.com/hanc00l/nemo_go/pkg/webapi/controllers"
)

func init() {
	ns := beego.NewNamespace("/v1",
		beego.NSNamespace("/login",
			beego.NSInclude(
				&controllers.LoginController{},
			),
		),
		beego.NSNamespace("/ip",
			beego.NSInclude(
				&controllers.IPController{},
			),
		),
		beego.NSNamespace("/domain",
			beego.NSInclude(
				&controllers.DomainController{},
			),
		),
		beego.NSNamespace("/vul",
			beego.NSInclude(
				&controllers.VulController{},
			),
		),
		beego.NSNamespace("/org",
			beego.NSInclude(
				&controllers.OrganizationController{},
			),
		),
		beego.NSNamespace("/task",
			beego.NSInclude(
				&controllers.TaskController{},
			),
		),
		beego.NSNamespace("/user",
			beego.NSInclude(
				&controllers.UserController{},
			),
		),
		beego.NSNamespace("/workspace",
			beego.NSInclude(
				&controllers.WorkspaceController{},
			),
		),
		beego.NSNamespace("/config",
			beego.NSInclude(
				&controllers.ConfigController{},
			),
		),
		beego.NSNamespace("/dashboard",
			beego.NSInclude(
				&controllers.DashboardController{},
			),
		),
	)
	beego.AddNamespace(ns)
}
