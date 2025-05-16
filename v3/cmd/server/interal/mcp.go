package interal

import (
	"fmt"
	"github.com/hanc00l/nemo_go/v3/pkg/conf"
	"github.com/hanc00l/nemo_go/v3/pkg/core"
	"github.com/hanc00l/nemo_go/v3/pkg/logging"
	"github.com/mark3labs/mcp-go/mcp"
)

func StartMCPServer() {
	s := core.NewMCPServer()
	// Add helloTool
	helloTool := mcp.NewTool("hello_nemo",
		mcp.WithDescription("NemoV3的MCP Server，返回Nemo的欢迎信息，并输出版本信息！"),
	)
	s.AddTool(helloTool, HelloHandler)
	//
	s.AddTool(mcp.NewTool("portscan",
		mcp.WithDescription("使用nemo创建一个端口扫描任务，扫描指定IP地址的存活端口，并进行指纹识别；创建任务后，可通过SSE实时获取任务状态和结果"),
		mcp.WithString("ip", mcp.Required(), mcp.Description("要扫描的IP地址，也可以是IP段，如192.168.0.0/24，一次只能指定一个IP地址或IP段")),
		mcp.WithString("port", mcp.DefaultString("--top-ports 1000"), mcp.Description("要扫描的端口，如80,443,10000-10010，或者--top-ports 1000（100或者10）；默认为--top-ports 1000")),
		mcp.WithString("rate", mcp.DefaultString("1000"), mcp.Description("扫描速率，每秒发送多少个包，默认1000")),
		mcp.WithString("bin", mcp.DefaultString("nmap"), mcp.Description("扫描工具，可选nmap、masscan、gogo")),
		mcp.WithBoolean("pocscan", mcp.DefaultBool(false), mcp.Description("是否开启POC和漏洞扫描，开启后会根据扫描的指纹自动匹配poc，扫描常见漏洞的POC，默认关闭")),
	), PortscanHandler)

	s.AddTool(mcp.NewTool("domainscan",
		mcp.WithDescription("使用nemo创建一个域名扫描任务，对域名执行子域名爆破和枚举操作，并进行指纹识别；创建任务后，可通过SSE实时获取任务状态和结果"),
		mcp.WithString("domain", mcp.Required(), mcp.Description("要扫描的域名，如example.com，一次只能指定一个域名")),
		mcp.WithBoolean("pocscan", mcp.DefaultBool(false), mcp.Description("是否开启POC和漏洞扫描，开启后会根据扫描的指纹自动匹配poc，扫描常见漏洞的POC，默认关闭")),
	), DomainscanHandler)

	s.AddTool(mcp.NewTool("onlineapi_scan",
		mcp.WithDescription("使用nemo创建一个调用在线API查询资产信息的任务，并进行指纹识别；创建任务后，可通过SSE实时获取任务状态和结果"),
		mcp.WithString("target", mcp.Required(), mcp.Description("要查询的资产信息，可以是IP地址或者域名，一次只能指定一个资产信息")),
		mcp.WithString("api", mcp.DefaultString("fofa"), mcp.Description("要调用的在线API，支持fofa、hunter、quake，请选择其中一个；一次只能指定一个API，默认为fofa")),
		mcp.WithBoolean("pocscan", mcp.DefaultBool(false), mcp.Description("是否开启POC和漏洞扫描，开启后会根据扫描的指纹自动匹配poc，扫描常见漏洞的POC，默认关闭")),
	), OnlineAPIHandler)

	s.AddTool(mcp.NewTool("query_task",
		mcp.WithDescription("查询nemo任务状态，需指定创建任务时返回的任务ID"),
		mcp.WithString("task_id", mcp.Required(), mcp.Description("nemo任务ID")),
	), QueryTaskHandler)

	s.AddTool(mcp.NewTool("query_result",
		mcp.WithDescription("查询nemo执行任务结果，需指定创建任务时返回的任务ID"),
		mcp.WithString("task_id", mcp.Required(), mcp.Description("nemo任务ID")),
		mcp.WithString("page", mcp.DefaultString("1"), mcp.Description("查询结果的页数，从1开始")),
		mcp.WithString("page_size", mcp.DefaultString("20"), mcp.Description("每次查询任务结果的最大条数，默认20")),
	), QueryResultHandler)

	s.AddTool(mcp.NewTool("query_asset",
		mcp.WithDescription("查询nemo数据库中存储的已收集到的资产信息，可根据主机名、IP、端口、标题、指纹等进行查询，可以指定多个查询条件，多个条件之间是AND关系"),
		mcp.WithString("host", mcp.DefaultString(""), mcp.Description("主机名，支持模糊查询")),
		mcp.WithString("ip", mcp.DefaultString(""), mcp.Description("IP地址，单个IP或带掩码的IP段")),
		mcp.WithString("port", mcp.DefaultString(""), mcp.Description("端口，支持单个端口或端口范围，如80,443,10000-10010")),
		mcp.WithString("title", mcp.DefaultString(""), mcp.Description("标题，支持模糊查询")),
		mcp.WithString("fingerprint", mcp.DefaultString(""), mcp.Description("指纹，支持模糊查询")),
		mcp.WithString("page", mcp.DefaultString("1"), mcp.Description("查询结果的页数，从1开始")),
		mcp.WithString("page_size", mcp.DefaultString("20"), mcp.Description("每次查询任务结果的最大条数，默认20")),
	), QueryAssetHandler)

	// Start MCP server
	serverAddr := fmt.Sprintf("%s:%d", conf.GlobalServerConfig().MCPServer.Host, conf.GlobalServerConfig().MCPServer.Port)
	logging.CLILog.Infof("start MCP Server：%s", serverAddr)
	logging.RuntimeLog.Infof("start MCP Server：%s", serverAddr)
	if conf.GlobalServerConfig().MCPServer.TLSEnable {
		if err := core.StartMCPServerWithTLS(s, serverAddr, conf.GlobalServerConfig().MCPServer.TLSCert, conf.GlobalServerConfig().MCPServer.TLSKey); err != nil {
			logging.CLILog.Error(err)
			return
		}
	} else {
		if err := core.StartMCPServer(s, serverAddr); err != nil {
			logging.CLILog.Error(err)
			return
		}
	}
}
