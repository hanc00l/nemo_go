package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/db"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/ampq"
	"github.com/hanc00l/nemo_go/pkg/task/custom"
	"github.com/hanc00l/nemo_go/pkg/task/domainscan"
	"github.com/hanc00l/nemo_go/pkg/task/onlineapi"
	"github.com/hanc00l/nemo_go/pkg/task/pocscan"
	"github.com/hanc00l/nemo_go/pkg/task/portscan"
	"github.com/hanc00l/nemo_go/pkg/task/serverapi"
	"github.com/hanc00l/nemo_go/pkg/utils"
	"regexp"
	"strings"
	"time"
)

type TaskController struct {
	BaseController
}

type taskRequestParam struct {
	DatableRequestParam
	State  string `form:"task_state"`
	Name   string `form:"task_name"`
	KwArgs string `form:"task_args"`
	Worker string `form:"task_worker"`
}

type TaskListData struct {
	Id           int    `json:"id"`
	Index        int    `json:"index"`
	TaskId       string `json:"task_id""`
	Worker       string `json:"worker"`
	TaskName     string `json:"task_name"`
	State        string `json:"state"`
	Result       string `json:"result"`
	KwArgs       string `json:"kwargs"`
	ReceivedTime string `json:"received"`
	StartedTime  string `json:"started"`
	Runtime      string `json:"runtime"`
}

type TaskInfo struct {
	Id            int
	TaskId        string
	Worker        string
	TaskName      string
	State         string
	Result        string
	KwArgs        string
	ReceivedTime  string
	StartedTime   string
	SucceededTime string
	FailedTime    string
	RetriedTime   string
	RevokedTime   string
	Runtime       string
	CreateTime    string
	UpdateTime    string
}

type portscanRequestParam struct {
	Target       string `form:"target"`
	IsPortScan   bool   `form:"portscan"`
	IsIPLocation bool   `form:"iplocation"`
	IsFofa       bool   `form:"fofasearch"`
	Port         string `form:"port"`
	Rate         int    `form:"rate"`
	NmapTech     string `form:"nmap_tech"`
	CmdBin       string `form:"bin"`
	OrgId        int    `form:"org_id"`
	IsWhatweb    bool   `form:"whatweb"`
	IsHttpx      bool   `form:"httpx"`
	IsPing       bool   `form:"ping"`
	ExcludeIP    string `form:"exclude"`
	IsScreenshot bool   `form:"screenshot"`
	IsWappalyzer bool   `form:"wappalyzer"`
	TaskMode     int    `form:"taskmode"`
}

type domainscanRequestParam struct {
	Target           string `form:"target"`
	OrgId            int    `form:"org_id"`
	IsSubfinder      bool   `form:"subfinder"`
	IsSubdomainBrute bool   `form:"subdomainbrute"`
	IsFldDomain      bool   `form:"fld_domain"`
	IsWhatweb        bool   `form:"whatweb"`
	IsHttpx          bool   `form:"httpx"`
	IsIPPortscan     bool   `form:"portscan"`
	IsSubnetPortscan bool   `form:"networkscan"`
	IsJSFinder       bool   `form:"jsfinder"`
	IsFofa           bool   `form:"fofasearch"`
	IsScreenshot     bool   `form:"screenshot"`
	IsICPQuery       bool   `form:"icpquery"`
	IsWappalyzer     bool   `form:"wappalyzer"`
	TaskMode         int    `form:"taskmode"`
	PortTaskMode     int    `form:"porttaskmode"`
}

type pocscanRequestParam struct {
	Target           string `form:"target"`
	IsPocsuiteVerify bool   `form:"pocsuite3verify"`
	PocsuitePocFile  string `form:"pocsuite3_poc_file"`
	IsXrayVerify     bool   `form:"xrayverify"`
	XrayPocFile      string `form:"xray_poc_file"`
}

func (c *TaskController) IndexAction() {
	c.UpdateOnlineUser()
	c.Layout = "base.html"
	c.TplName = "task-list.html"
}

// ListAction 漏洞列表的数据
func (c *TaskController) ListAction() {
	c.UpdateOnlineUser()
	defer c.ServeJSON()

	req := taskRequestParam{}
	err := c.ParseForm(&req)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
	}
	c.validateRequestParam(&req)
	resp := c.getTaskListData(req)
	c.Data["json"] = resp
}

// InfoAction 显示一个漏洞的详情
func (c *TaskController) InfoAction() {
	var taskInfo TaskInfo

	taskId := c.GetString("task_id")
	if taskId != "" {
		taskInfo = getTaskInfo(taskId)
	}
	c.Data["task_info"] = taskInfo
	c.Layout = "base.html"
	c.TplName = "task-info.html"
}

// DeleteAction 删除一个记录
func (c *TaskController) DeleteAction() {
	defer c.ServeJSON()

	id, err := c.GetInt("id")
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		c.FailedStatus(err.Error())
	} else {
		task := db.Task{Id: id}
		c.MakeStatusResponse(task.Delete())
	}
}

// StopAction 取消一个未开始执行的任务
func (c *TaskController) StopAction() {
	defer c.ServeJSON()

	taskId := c.GetString("task_id")
	if taskId != "" {
		isRevoked, _ := serverapi.RevokeUnexcusedTask(taskId)
		c.MakeStatusResponse(isRevoked)
		return
	}
	c.MakeStatusResponse(false)
}

// StartPortScanTaskAction 端口扫描任务
func (c *TaskController) StartPortScanTaskAction() {
	defer c.ServeJSON()
	// 解析参数
	var req portscanRequestParam
	err := c.ParseForm(&req)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		c.FailedStatus(err.Error())
		return
	}
	if req.Target == "" {
		c.FailedStatus("no target")
		return
	}
	if req.Port == "" {
		req.Port = conf.GlobalWorkerConfig().Portscan.Port
	}
	ts := utils.NewTaskSlice()
	ts.TaskMode = req.TaskMode
	ts.IpTarget = req.Target
	ts.Port = req.Port
	tc := conf.GlobalServerConfig().Task
	ts.IpSliceNumber = tc.IpSliceNumber
	ts.PortSliceNumber = tc.PortSliceNumber
	targets, ports := ts.DoIpSlice()
	var taskId string
	for _, t := range targets {
		for _, p := range ports {
			// 端口扫描
			if req.IsPortScan {
				if taskId, err = c.doPortscan(t, p, req); err != nil {
					c.FailedStatus(err.Error())
					return
				}
			}
			// IP归属地：如果有端口执行任务，则IP归属地任务在端口扫描中执行，否则单独执行
			if !req.IsPortScan && req.IsIPLocation {
				if taskId, err = c.doIPLocation(t, &req.OrgId); err != nil {
					c.FailedStatus(err.Error())
					return
				}
			}
			// FOFA
			if req.IsFofa {
				if taskId, err = c.doFofa(t, &req.OrgId, req.IsIPLocation); err != nil {
					c.FailedStatus(err.Error())
					return
				}
			}
		}
	}
	c.SucceededStatus(taskId)
}

// StartBatchScanTaskAction 探测+扫描任务
func (c *TaskController) StartBatchScanTaskAction() {
	defer c.ServeJSON()
	// 解析参数
	var req portscanRequestParam
	err := c.ParseForm(&req)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		c.FailedStatus(err.Error())
		return
	}
	if req.Target == "" {
		c.FailedStatus("no target")
		return
	}
	ts := utils.NewTaskSlice()
	ts.TaskMode = req.TaskMode
	ts.IpTarget = req.Target
	ts.Port = req.Port
	tc := conf.GlobalServerConfig().Task
	ts.IpSliceNumber = tc.IpSliceNumber
	ts.PortSliceNumber = tc.PortSliceNumber
	targets, ports := ts.DoIpSlice()
	var taskId string
	for _, t := range targets {
		for _, p := range ports {
			// 端口扫描
			if taskId, err = c.doBatchScan(t, p, req); err != nil {
				c.FailedStatus(err.Error())
				return
			}
		}
	}
	c.SucceededStatus(taskId)
}

// StartDomainScanTaskAction 域名任务
func (c *TaskController) StartDomainScanTaskAction() {
	defer c.ServeJSON()

	// 解析参数
	var req domainscanRequestParam
	err := c.ParseForm(&req)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		c.FailedStatus(err.Error())
		return
	}
	if req.Target == "" {
		c.FailedStatus("no target")
		return
	}
	allTarget := req.Target
	// 域名的FLD
	if req.IsFldDomain {
		fldList := getDomainFLD(req.Target)
		if len(fldList) > 0 {
			allTarget = req.Target + "," + strings.Join(fldList, ",")
		}
	}
	ts := utils.NewTaskSlice()
	ts.TaskMode = req.TaskMode
	ts.DomainTarget = allTarget
	targets := ts.DoDomainSlice()
	var taskId string
	for _, t := range targets {
		// 每个获取子域名的方式采用独立任务，以提高速度
		var taskStarted bool
		if req.IsSubfinder {
			subConfig := req
			subConfig.IsSubdomainBrute = false
			subConfig.IsJSFinder = false
			if taskId, err = c.doDomainscan(t, subConfig); err != nil {
				c.FailedStatus(err.Error())
				return
			}
			taskStarted = true
		}
		if req.IsSubdomainBrute {
			subConfig := req
			subConfig.IsSubfinder = false
			subConfig.IsJSFinder = false
			if taskId, err = c.doDomainscan(t, subConfig); err != nil {
				c.FailedStatus(err.Error())
				return
			}
			taskStarted = true
		}
		if req.IsJSFinder {
			//TODO
		}
		// 如果没有子域名任务，则至少启动一个域名解析任务
		if !taskStarted {
			if taskId, err = c.doDomainscan(t, req); err != nil {
				c.FailedStatus(err.Error())
				return
			}
		}
		if req.IsFofa {
			if taskId, err = c.doFofa(t, &req.OrgId, true); err != nil {
				c.FailedStatus(err.Error())
				return
			}
		}
		if req.IsICPQuery {
			if taskId, err = c.doICPQuery(t); err != nil {
				c.FailedStatus(err.Error())
				return
			}
		}
	}
	c.SucceededStatus(taskId)
}

// StartPocScanTaskAction pocscan任务
func (c *TaskController) StartPocScanTaskAction() {
	defer c.ServeJSON()

	// 解析参数
	var req pocscanRequestParam
	err := c.ParseForm(&req)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		c.FailedStatus(err.Error())
		return
	}
	// 格式化Target
	if req.Target == "" {
		c.FailedStatus("no target")
		return
	}
	var targetList []string
	for _, t := range strings.Split(req.Target, "\n") {
		if tt := strings.TrimSpace(t); tt != "" {
			targetList = append(targetList, tt)
		}
	}
	var taskId string
	if req.IsPocsuiteVerify && req.PocsuitePocFile != "" {
		config := pocscan.Config{Target: strings.Join(targetList, ","), PocFile: req.PocsuitePocFile, CmdBin: "pocsuite"}
		configJSON, _ := json.Marshal(config)
		taskId, err = serverapi.NewTask("pocsuite", string(configJSON))
		if err != nil {
			c.FailedStatus(err.Error())
			return
		}
	}
	if req.IsXrayVerify && req.XrayPocFile != "" {
		config := pocscan.Config{Target: strings.Join(targetList, ","), PocFile: req.XrayPocFile, CmdBin: "xray"}
		configJSON, _ := json.Marshal(config)
		taskId, err = serverapi.NewTask("xray", string(configJSON))
		if err != nil {
			c.FailedStatus(err.Error())
			return
		}
	}
	c.SucceededStatus(taskId)
}

// doDomainscan 域名任务
func (c *TaskController) doDomainscan(target string, req domainscanRequestParam) (taskId string, err error) {
	config := domainscan.Config{
		Target:             target,
		OrgId:              &req.OrgId,
		IsSubDomainFinder:  req.IsSubfinder,
		IsSubDomainBrute:   req.IsSubdomainBrute,
		IsJSFinder:         req.IsJSFinder,
		IsHttpx:            req.IsHttpx,
		IsWhatWeb:          req.IsWhatweb,
		IsIPPortScan:       req.IsIPPortscan,
		IsIPSubnetPortScan: req.IsSubnetPortscan,
		IsScreenshot:       req.IsScreenshot,
		IsWappalyzer:       req.IsWappalyzer,
		PortTaskMode:       req.PortTaskMode,
	}
	// config.OrgId 为int，默认为0
	// db.Organization.OrgId为指针，默认nil
	if *config.OrgId == 0 {
		config.OrgId = nil
	}
	configJSON, err := json.Marshal(config)
	if err != nil {
		logging.RuntimeLog.Errorf("start domainscan fail:%s", err.Error())
		return "", err
	}
	taskId, err = serverapi.NewTask("domainscan", string(configJSON))
	if err != nil {
		logging.RuntimeLog.Errorf("start domainscan fail:%s", err.Error())
		return "", err
	}
	return taskId, nil
}

// doPortscan 端口扫描
func (c *TaskController) doPortscan(target string, port string, req portscanRequestParam) (taskId string, err error) {
	config := portscan.Config{
		Target:        target,
		ExcludeTarget: req.ExcludeIP,
		Port:          port,
		OrgId:         &req.OrgId,
		Rate:          req.Rate,
		IsPing:        req.IsPing,
		Tech:          req.NmapTech,
		IsIpLocation:  req.IsIPLocation,
		IsHttpx:       req.IsHttpx,
		IsWhatWeb:     req.IsWhatweb,
		IsScreenshot:  req.IsScreenshot,
		IsWappalyzer:  req.IsWappalyzer,
		CmdBin:        "masscan",
	}
	if req.CmdBin == "nmap" {
		config.CmdBin = "nmap"
	}
	if config.Port == "" {
		config.Port = conf.GlobalWorkerConfig().Portscan.Port
	}
	if config.Rate == 0 {
		config.Rate = conf.GlobalWorkerConfig().Portscan.Rate
	}
	if config.Tech == "" {
		config.Target = conf.GlobalWorkerConfig().Portscan.Tech
	}
	// config.OrgId 为int，默认为0
	// db.Organization.OrgId为指针，默认nil
	if *config.OrgId == 0 {
		config.OrgId = nil
	}
	configJSON, err := json.Marshal(config)
	if err != nil {
		logging.RuntimeLog.Errorf("start portscan fail:%s", err.Error())
		return "", err
	}
	taskId, err = serverapi.NewTask("portscan", string(configJSON))
	if err != nil {
		logging.RuntimeLog.Errorf("start portscan fail:%s", err.Error())
		return "", err
	}
	return taskId, nil
}

// doBatchScan 探测+端口扫描
func (c *TaskController) doBatchScan(target string, port string, req portscanRequestParam) (taskId string, err error) {
	config := portscan.Config{
		Target:        target,
		ExcludeTarget: req.ExcludeIP,
		Port:          port,
		OrgId:         &req.OrgId,
		Rate:          req.Rate,
		IsPing:        req.IsPing,
		Tech:          req.NmapTech,
		IsIpLocation:  req.IsIPLocation,
		IsHttpx:       req.IsHttpx,
		IsWhatWeb:     req.IsWhatweb,
		IsScreenshot:  req.IsScreenshot,
		IsWappalyzer:  req.IsWappalyzer,
		CmdBin:        "masscan",
	}
	if req.CmdBin == "nmap" {
		config.CmdBin = "nmap"
	}
	if config.Port == "" {
		config.Port = "80,443,8080|" + conf.GlobalWorkerConfig().Portscan.Port
	}
	if config.Rate == 0 {
		config.Rate = conf.GlobalWorkerConfig().Portscan.Rate
	}
	if config.Tech == "" {
		config.Target = conf.GlobalWorkerConfig().Portscan.Tech
	}
	// config.OrgId 为int，默认为0
	// db.Organization.OrgId为指针，默认nil
	if *config.OrgId == 0 {
		config.OrgId = nil
	}
	configJSON, err := json.Marshal(config)
	if err != nil {
		logging.RuntimeLog.Errorf("start portscan fail:%s", err.Error())
		return "", err
	}
	taskId, err = serverapi.NewTask("batchscan", string(configJSON))
	if err != nil {
		logging.RuntimeLog.Errorf("start batchscan fail:%s", err.Error())
		return "", err
	}
	return taskId, nil
}

// doFofa FOFA搜索
func (c *TaskController) doFofa(target string, orgId *int, iplocation bool) (taskId string, err error) {
	config := onlineapi.FofaConfig{Target: target, OrgId: orgId, IsIPLocation: iplocation}
	// config.OrgId 为int，默认为0
	// db.Organization.OrgId为指针，默认nil
	if *config.OrgId == 0 {
		config.OrgId = nil
	}
	configJSON, err := json.Marshal(config)
	if err != nil {
		logging.RuntimeLog.Errorf("start fofa fail:%s", err.Error())
		return "", err
	}
	taskId, err = serverapi.NewTask("fofa", string(configJSON))
	if err != nil {
		logging.RuntimeLog.Errorf("start iplocation fail:%s", err.Error())
		return "", err
	}
	return taskId, nil
}

// doICPQuery ICP备案信息查询
func (c *TaskController) doICPQuery(target string) (taskId string, err error) {
	config := onlineapi.ICPQueryConfig{Target: target}
	configJSON, err := json.Marshal(config)
	if err != nil {
		logging.RuntimeLog.Errorf("start icpquery fail:%s", err.Error())
		return "", err
	}
	taskId, err = serverapi.NewTask("icpquery", string(configJSON))
	if err != nil {
		logging.RuntimeLog.Errorf("start iplocation fail:%s", err.Error())
		return "", err
	}
	return taskId, nil
}

// doIPLocation IP归属地
func (c *TaskController) doIPLocation(target string, orgId *int) (taskId string, err error) {
	config := custom.Config{Target: target, OrgId: orgId}
	// config.OrgId 为int，默认为0
	// db.Organization.OrgId为指针，默认nil
	if *config.OrgId == 0 {
		config.OrgId = nil
	}
	configJSON, err := json.Marshal(config)
	if err != nil {
		logging.RuntimeLog.Errorf("start portscan fail:%s", err.Error())
		return "", err
	}
	taskId, err = serverapi.NewTask("iplocation", string(configJSON))
	if err != nil {
		logging.RuntimeLog.Errorf("start iplocation fail:%s", err.Error())
		return "", err
	}
	return taskId, nil
}

//validateRequestParam 校验请求的参数
func (c *TaskController) validateRequestParam(req *taskRequestParam) {
	if req.Length <= 0 {
		req.Length = 50
	}
	if req.Start < 0 {
		req.Start = 0
	}
}

// getSearchMap 根据查询参数生成查询条件
func (c *TaskController) getSearchMap(req taskRequestParam) (searchMap map[string]interface{}) {
	searchMap = make(map[string]interface{})

	if req.Name != "" {
		searchMap["task_name"] = req.Name
	}
	if req.State != "" {
		searchMap["state"] = req.State
	}
	if req.KwArgs != "" {
		searchMap["kwargs"] = req.KwArgs
	}
	if req.Worker != "" {
		searchMap["worker"] = req.Worker
	}
	return
}

// getTaskListData 获取列显示的数据
func (c *TaskController) getTaskListData(req taskRequestParam) (resp DataTableResponseData) {
	vul := db.Task{}
	searchMap := c.getSearchMap(req)
	startPage := req.Start/req.Length + 1
	results, total := vul.Gets(searchMap, startPage, req.Length)
	for i, taskRow := range results {
		t := TaskListData{}
		t.Id = taskRow.Id
		t.Index = req.Start + i + 1
		t.TaskId = taskRow.TaskId
		t.TaskName = taskRow.TaskName
		t.Worker = taskRow.Worker
		t.State = taskRow.State
		t.Result = getResultMsg(taskRow.Result)
		t.KwArgs = getTargetFromKwArgs(taskRow.KwArgs)
		if taskRow.StartedTime != nil {
			t.StartedTime = FormatDateTime(*taskRow.StartedTime)
		}
		if taskRow.ReceivedTime != nil {
			t.ReceivedTime = FormatDateTime(*taskRow.ReceivedTime)
		}
		t.Runtime = formatRuntime(&taskRow)

		resp.Data = append(resp.Data, t)
	}
	resp.Draw = req.Draw
	resp.RecordsTotal = total
	resp.RecordsFiltered = total
	if resp.Data == nil {
		resp.Data = make([]interface{}, 0)
	}
	return
}

// getTaskInfo 获取一个任务的详情
func getTaskInfo(taskId string) (r TaskInfo) {
	task := db.Task{TaskId: taskId}
	if !task.GetByTaskId() {
		return
	}
	r.Id = task.Id
	r.TaskId = task.TaskId
	r.TaskName = task.TaskName
	r.Worker = task.Worker
	r.Result = task.Result
	r.State = task.State
	r.KwArgs = task.KwArgs
	if task.StartedTime != nil {
		r.StartedTime = FormatDateTime(*task.StartedTime)
	}
	if task.ReceivedTime != nil {
		r.ReceivedTime = FormatDateTime(*task.ReceivedTime)
	}
	if task.RetriedTime != nil {
		r.RetriedTime = FormatDateTime(*task.RetriedTime)
	}
	if task.RevokedTime != nil {
		r.RevokedTime = FormatDateTime(*task.RevokedTime)
	}
	if task.FailedTime != nil {
		r.FailedTime = FormatDateTime(*task.FailedTime)
	}
	if task.SucceededTime != nil {
		r.SucceededTime = FormatDateTime(*task.SucceededTime)
	}
	r.Runtime = formatRuntime(&task)
	r.CreateTime = FormatDateTime(task.CreateDatetime)
	r.UpdateTime = FormatDateTime(task.UpdateDatetime)

	return
}

// formatRuntime 计算任务运行时间
func formatRuntime(t *db.Task) (runtime string) {
	var endTime *time.Time
	if t.SucceededTime != nil {
		endTime = t.SucceededTime
	} else if t.FailedTime != nil {
		endTime = t.FailedTime
	} else {
		return
	}
	var startedTime time.Time
	if t.StartedTime != nil {
		startedTime = *t.StartedTime
	} else if t.ReceivedTime != nil {
		startedTime = *t.ReceivedTime
	} else {
		return
	}
	runtime = endTime.Sub(startedTime).Truncate(time.Second).String()

	return
}

// getTargetFromKwArgs 从经过JSON序列化的参数中单独提取出target
func getTargetFromKwArgs(kwargs string) (target string) {
	const displayedLength = 100
	//{"target":"192.168.120.0/24","executeTarget":"...
	targetReg := regexp.MustCompile(`^{"target":"(.*?)"`)
	targetArray := targetReg.FindStringSubmatch(kwargs)
	if targetArray != nil {
		target = targetArray[1]
	} else {
		target = kwargs
	}
	if len(target) > displayedLength {
		return fmt.Sprintf("%s...", target[:displayedLength])
	}
	return
}

// getResultMsg 从经过JSON反序列化的结果中提取出结果的消息
func getResultMsg(resultJSON string) (msg string) {
	var result ampq.TaskResult
	err := json.Unmarshal([]byte(resultJSON), &result)
	if err != nil {
		return resultJSON
	}
	return result.Msg
}

// getDomainFLD 提取域名的FLD
func getDomainFLD(target string) (fldDomain []string) {
	domains := make(map[string]struct{})
	tld := domainscan.NewTldExtract()
	for _, t := range strings.Split(target, "\n") {
		domain := strings.TrimSpace(t)
		fld := tld.ExtractFLD(domain)
		if fld == "" {
			continue
		}
		if _, ok := domains[fld]; !ok {
			domains[fld] = struct{}{}
		}
	}
	fldDomain = utils.SetToSlice(domains)
	return
}
