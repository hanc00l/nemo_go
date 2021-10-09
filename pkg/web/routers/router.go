package routers

import (
	"github.com/beego/beego/v2/server/web"
	"github.com/hanc00l/nemo_go/pkg/web/controllers"
)

func init() {
	login := &controllers.LoginController{}
	web.Router("/", login, "get:IndexAction;post:LoginAction")
	web.Router("/logout", login, "get:LogoutAction")

	config := &controllers.ConfigController{}
	web.Router("/config-list", config, "get:IndexAction;post:LoadDefaultConfigAction")
	web.Router("/config-change-password", config, "post:ChangePasswordAction")
	web.Router("/custom-list", config, "get:CustomAction")
	web.Router("/custom-load", config, "post:LoadCustomConfigAction")
	web.Router("/custom-save", config, "post:SaveCustomConfigAction")
	web.Router("/config-save-taskslice", config, "post:SaveTaskSliceNumberAction")

	dashboard := &controllers.DashboardController{}
	web.Router("/dashboard", dashboard, "get:IndexAction;post:GetStatisticDataAction")
	web.Router("/dashboard-task-info", dashboard, "post:GetTaskInfoAction")
	web.Router("/worker-list", dashboard, "post:WorkerAliveListAction")
	web.Router("/onlineuser-list", dashboard, "post:OnlineUserListAction")

	ip := &controllers.IPController{}
	web.Router("/ip-list", ip, "get:IndexAction;post:ListAction")
	web.Router("/ip-info", ip, "get:InfoAction")
	web.Router("/ip-delete", ip, "post:DeleteIPAction")
	web.Router("/port-attr-delete", ip, "post:DeletePortAttrAction")
	web.Router("/ip-statistics", ip, "get:StatisticsAction")
	web.Router("/ip-memo-get", ip, "get:GetMemoAction")
	web.Router("/ip-memo-update", ip, "post:UpdateMemoAction")
	web.Router("/ip-memo-export", ip, "get:ExportMemoAction")
	web.Router("/ip-color-tag", ip, "post:MarkColorTagAction")

	domain := &controllers.DomainController{}
	web.Router("/domain-list", domain, "get:IndexAction;post:ListAction")
	web.Router("/domain-info", domain, "get:InfoAction")
	web.Router("/domain-delete", domain, "post:DeleteDomainAction")
	web.Router("/domain-attr-delete", domain, "post:DeleteDomainAttrAction")
	web.Router("/domain-fofa-attr-delete", domain, "post:DeleteDomainFofaAttrAction")
	web.Router("/domain-statistics", domain, "get:StatisticsAction")

	web.Router("/domain-memo-get", domain, "get:GetMemoAction")
	web.Router("/domain-memo-update", domain, "post:UpdateMemoAction")
	web.Router("/domain-memo-export", domain, "get:ExportMemoAction")
	web.Router("/domain-color-tag", domain, "post:MarkColorTagAction")

	vulnerability := &controllers.VulController{}
	web.Router("/vulnerability-list", vulnerability, "get:IndexAction;post:ListAction")
	web.Router("/vulnerability-info", vulnerability, "get:InfoAction")
	web.Router("/vulnerability-delete", vulnerability, "post:DeleteAction")
	web.Router("/vulnerability-load-pocsuite-pocfile", vulnerability, "post:LoadPocsuitePocFileAction")
	web.Router("/vulnerability-load-xray-pocfile", vulnerability, "post:LoadXrayPocFileAction")

	org := &controllers.OrganizationController{}
	web.Router("/org-list", org, "get:IndexAction;post:ListAction")
	web.Router("/org-get", org, "post:GetAction")
	web.Router("/org-getall", org, "post:GetAllAction")
	web.Router("/org-add", org, "get:AddIndexAction;post:AddSaveAction")
	web.Router("/org-update", org, "post:UpdateAction")
	web.Router("/org-delete", org, "post:DeleteAction")

	task := &controllers.TaskController{}
	web.Router("/task-list", task, "get:IndexAction;post:ListAction")
	web.Router("/task-info", task, "get:InfoAction")
	web.Router("/task-delete", task, "post:DeleteAction")
	web.Router("/task-stop", task, "post:StopAction")
	web.Router("/task-start-portscan", task, "post:StartPortScanTaskAction")
	web.Router("/task-start-batchscan", task, "post:StartBatchScanTaskAction")
	web.Router("/task-start-domainscan", task, "post:StartDomainScanTaskAction")
	web.Router("/task-start-vulnerability", task, "post:StartPocScanTaskAction")
}
