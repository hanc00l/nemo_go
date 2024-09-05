package routers

import (
	beego "github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context/param"
)

func init() {

	beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:ConfigController"] = append(beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:ConfigController"],
		beego.ControllerComments{
			Method:           "ChangePassword",
			Router:           `/changepass`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:ConfigController"] = append(beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:ConfigController"],
		beego.ControllerComments{
			Method:           "LoadDefaultConfig",
			Router:           `/load-default`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:ConfigController"] = append(beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:ConfigController"],
		beego.ControllerComments{
			Method:           "SaveDefaultConfig",
			Router:           `/save-default`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:DashboardController"] = append(beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:DashboardController"],
		beego.ControllerComments{
			Method:           "GetStatisticData",
			Router:           `/statistic`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:DashboardController"] = append(beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:DashboardController"],
		beego.ControllerComments{
			Method:           "GetTaskInfo",
			Router:           `/task`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:DashboardController"] = append(beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:DashboardController"],
		beego.ControllerComments{
			Method:           "OnlineUserList",
			Router:           `/user`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:DashboardController"] = append(beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:DashboardController"],
		beego.ControllerComments{
			Method:           "ManualWorkerFileSync",
			Router:           `/worker/filesync`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:DashboardController"] = append(beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:DashboardController"],
		beego.ControllerComments{
			Method:           "WorkerAliveList",
			Router:           `/worker/list`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:DashboardController"] = append(beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:DashboardController"],
		beego.ControllerComments{
			Method:           "ManualReloadWorker",
			Router:           `/worker/reload`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:DomainController"] = append(beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:DomainController"],
		beego.ControllerComments{
			Method:           "MarkColor",
			Router:           `/color/mark`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:DomainController"] = append(beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:DomainController"],
		beego.ControllerComments{
			Method:           "DeleteDomain",
			Router:           `/delete`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:DomainController"] = append(beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:DomainController"],
		beego.ControllerComments{
			Method:           "DeleteDomainOnlineAPIAttr",
			Router:           `/domain-api-attr/delete`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:DomainController"] = append(beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:DomainController"],
		beego.ControllerComments{
			Method:           "DeleteDomainAttr",
			Router:           `/domain-attr/delete`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:DomainController"] = append(beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:DomainController"],
		beego.ControllerComments{
			Method:           "InfoHttp",
			Router:           `/http/info`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:DomainController"] = append(beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:DomainController"],
		beego.ControllerComments{
			Method:           "Info",
			Router:           `/info`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:DomainController"] = append(beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:DomainController"],
		beego.ControllerComments{
			Method:           "List",
			Router:           `/list`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:DomainController"] = append(beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:DomainController"],
		beego.ControllerComments{
			Method:           "GetMemo",
			Router:           `/memo`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:DomainController"] = append(beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:DomainController"],
		beego.ControllerComments{
			Method:           "UpdateMemo",
			Router:           `/memo/update`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:DomainController"] = append(beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:DomainController"],
		beego.ControllerComments{
			Method:           "PinTop",
			Router:           `/pintop`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:IPController"] = append(beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:IPController"],
		beego.ControllerComments{
			Method:           "MarkColor",
			Router:           `/color/mark`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:IPController"] = append(beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:IPController"],
		beego.ControllerComments{
			Method:           "DeleteIP",
			Router:           `/delete`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:IPController"] = append(beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:IPController"],
		beego.ControllerComments{
			Method:           "InfoHttp",
			Router:           `/http/info`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:IPController"] = append(beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:IPController"],
		beego.ControllerComments{
			Method:           "Info",
			Router:           `/info`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:IPController"] = append(beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:IPController"],
		beego.ControllerComments{
			Method:           "List",
			Router:           `/list`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:IPController"] = append(beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:IPController"],
		beego.ControllerComments{
			Method:           "GetMemo",
			Router:           `/memo`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:IPController"] = append(beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:IPController"],
		beego.ControllerComments{
			Method:           "UpdateMemo",
			Router:           `/memo/update`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:IPController"] = append(beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:IPController"],
		beego.ControllerComments{
			Method:           "PinTop",
			Router:           `/pintop`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:IPController"] = append(beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:IPController"],
		beego.ControllerComments{
			Method:           "DeletePortAttr",
			Router:           `/port-attr/delete`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:IPController"] = append(beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:IPController"],
		beego.ControllerComments{
			Method:           "ImportPortscanResult",
			Router:           `/result/import`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:LoginController"] = append(beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:LoginController"],
		beego.ControllerComments{
			Method:           "Capture",
			Router:           `/captcha`,
			AllowHTTPMethods: []string{"get", "post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:LoginController"] = append(beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:LoginController"],
		beego.ControllerComments{
			Method:           "Login",
			Router:           `/login`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:OrganizationController"] = append(beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:OrganizationController"],
		beego.ControllerComments{
			Method:           "DeleteOrg",
			Router:           `/delete`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:OrganizationController"] = append(beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:OrganizationController"],
		beego.ControllerComments{
			Method:           "GetAll",
			Router:           `/getall`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:OrganizationController"] = append(beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:OrganizationController"],
		beego.ControllerComments{
			Method:           "Info",
			Router:           `/info`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:OrganizationController"] = append(beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:OrganizationController"],
		beego.ControllerComments{
			Method:           "List",
			Router:           `/list`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:OrganizationController"] = append(beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:OrganizationController"],
		beego.ControllerComments{
			Method:           "SaveOrg",
			Router:           `/save`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:OrganizationController"] = append(beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:OrganizationController"],
		beego.ControllerComments{
			Method:           "UpdateOrg",
			Router:           `/update`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:TaskController"] = append(beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:TaskController"],
		beego.ControllerComments{
			Method:           "DeleteBatchTask",
			Router:           `/batch-delete`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:TaskController"] = append(beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:TaskController"],
		beego.ControllerComments{
			Method:           "DeleteCronTask",
			Router:           `/cron/delete`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:TaskController"] = append(beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:TaskController"],
		beego.ControllerComments{
			Method:           "DisableCronTask",
			Router:           `/cron/disable`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:TaskController"] = append(beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:TaskController"],
		beego.ControllerComments{
			Method:           "EnableCronTask",
			Router:           `/cron/enable`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:TaskController"] = append(beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:TaskController"],
		beego.ControllerComments{
			Method:           "InfoCronTask",
			Router:           `/cron/info`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:TaskController"] = append(beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:TaskController"],
		beego.ControllerComments{
			Method:           "ListCronTask",
			Router:           `/cron/list`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:TaskController"] = append(beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:TaskController"],
		beego.ControllerComments{
			Method:           "RunCronTask",
			Router:           `/cron/run`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:TaskController"] = append(beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:TaskController"],
		beego.ControllerComments{
			Method:           "DeleteMainTask",
			Router:           `/main/delete`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:TaskController"] = append(beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:TaskController"],
		beego.ControllerComments{
			Method:           "InfoMainTask",
			Router:           `/main/info`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:TaskController"] = append(beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:TaskController"],
		beego.ControllerComments{
			Method:           "ListMainTask",
			Router:           `/main/list`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:TaskController"] = append(beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:TaskController"],
		beego.ControllerComments{
			Method:           "DeleteRunTask",
			Router:           `/run/delete`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:TaskController"] = append(beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:TaskController"],
		beego.ControllerComments{
			Method:           "InfoRunTask",
			Router:           `/run/info`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:TaskController"] = append(beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:TaskController"],
		beego.ControllerComments{
			Method:           "StopRunTask",
			Router:           `/run/stop`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:TaskController"] = append(beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:TaskController"],
		beego.ControllerComments{
			Method:           "StartXScanTask",
			Router:           `/xscan`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:UserController"] = append(beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:UserController"],
		beego.ControllerComments{
			Method:           "DeleteUser",
			Router:           `/delete`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:UserController"] = append(beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:UserController"],
		beego.ControllerComments{
			Method:           "Info",
			Router:           `/info`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:UserController"] = append(beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:UserController"],
		beego.ControllerComments{
			Method:           "List",
			Router:           `/list`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:UserController"] = append(beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:UserController"],
		beego.ControllerComments{
			Method:           "ResetPassword",
			Router:           `/reset`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:UserController"] = append(beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:UserController"],
		beego.ControllerComments{
			Method:           "SaveUser",
			Router:           `/save`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:UserController"] = append(beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:UserController"],
		beego.ControllerComments{
			Method:           "UpdateUser",
			Router:           `/update`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:UserController"] = append(beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:UserController"],
		beego.ControllerComments{
			Method:           "ListUserWorkspace",
			Router:           `/workspace/list`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:UserController"] = append(beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:UserController"],
		beego.ControllerComments{
			Method:           "UpdateUserWorkspace",
			Router:           `/workspace/update`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:VulController"] = append(beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:VulController"],
		beego.ControllerComments{
			Method:           "DeleteVul",
			Router:           `/delete`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:VulController"] = append(beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:VulController"],
		beego.ControllerComments{
			Method:           "Info",
			Router:           `/info`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:VulController"] = append(beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:VulController"],
		beego.ControllerComments{
			Method:           "List",
			Router:           `/list`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:VulController"] = append(beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:VulController"],
		beego.ControllerComments{
			Method:           "LoadNucleiPocFile",
			Router:           `/nuclei/pocfile`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:VulController"] = append(beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:VulController"],
		beego.ControllerComments{
			Method:           "LoadXrayPocFile",
			Router:           `/xray/pocfile`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:WorkspaceController"] = append(beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:WorkspaceController"],
		beego.ControllerComments{
			Method:           "ChangeWorkspaceSelect",
			Router:           `/change`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:WorkspaceController"] = append(beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:WorkspaceController"],
		beego.ControllerComments{
			Method:           "DeleteWorkspace",
			Router:           `/delete`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:WorkspaceController"] = append(beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:WorkspaceController"],
		beego.ControllerComments{
			Method:           "Info",
			Router:           `/info`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:WorkspaceController"] = append(beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:WorkspaceController"],
		beego.ControllerComments{
			Method:           "List",
			Router:           `/list`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:WorkspaceController"] = append(beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:WorkspaceController"],
		beego.ControllerComments{
			Method:           "SaveWorkspace",
			Router:           `/save`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:WorkspaceController"] = append(beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:WorkspaceController"],
		beego.ControllerComments{
			Method:           "UpdateWorkspace",
			Router:           `/update`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:WorkspaceController"] = append(beego.GlobalControllerRouter["github.com/hanc00l/nemo_go/v2/pkg/webapi/controllers:WorkspaceController"],
		beego.ControllerComments{
			Method:           "UserWorkspace",
			Router:           `/user-ownerd`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

}
