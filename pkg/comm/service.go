package comm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/db"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/ampq"
	"github.com/hanc00l/nemo_go/pkg/task/domainscan"
	"github.com/hanc00l/nemo_go/pkg/task/fingerprint"
	"github.com/hanc00l/nemo_go/pkg/task/onlineapi"
	"github.com/hanc00l/nemo_go/pkg/task/pocscan"
	"github.com/hanc00l/nemo_go/pkg/task/portscan"
	"github.com/hanc00l/nemo_go/pkg/task/serverapi"
	"github.com/hanc00l/nemo_go/pkg/utils"
	whoisparser "github.com/likexian/whois-parser"
	"github.com/smallnest/rpcx/client"
	"github.com/tidwall/pretty"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Service RPC服务
type Service struct{}

// ScanResultArgs IP与域名扫描结果请求参数
type ScanResultArgs struct {
	TaskID              string
	MainTaskId          string
	IPConfig            *portscan.Config
	DomainConfig        *domainscan.Config
	IPResult            map[string]*portscan.IPResult
	DomainResult        map[string]*domainscan.DomainResult
	VulnerabilityResult []pocscan.Result
}

// ScreenshotResultArgs screenshot结果请求参数
type ScreenshotResultArgs struct {
	MainTaskId  string
	WorkspaceId int
	FileInfo    []fingerprint.ScreenshotFileInfo
}

// IconHashResultArgs icconhash保存的结果值
type IconHashResultArgs struct {
	WorkspaceId  int
	IconHashInfo []fingerprint.IconHashInfo
}

// TaskStatusArgs 任务状态请求与返回参数
type TaskStatusArgs struct {
	TaskID    string
	IsExist   bool
	IsRevoked bool
	State     string
	Worker    string
	Result    string
}

// NewTaskArgs 新建任务请求与返回参数
type NewTaskArgs struct {
	TaskName      string
	ConfigJSON    string
	MainTaskID    string
	LastRunTaskId string
}

type LoadIPOpenedPortArgs struct {
	WorkspaceId int
	Target      string
}

type LoadDomainOpenedPortArgs struct {
	WorkspaceId int
	Target      map[string]struct{}
}

type MainTaskResultMap struct {
	IPResult         map[string]map[int]interface{}
	DomainResult     map[string]interface{}
	VulResult        map[string]map[string]interface{}
	ScreenShotResult int
	IPNew            int
	PortNew          int
	DomainNew        int
	VulnerabilityNew int
}

type RuntimeLogArgs struct {
	Source     string
	LogMessage []byte
}

var (
	// globalXClient 全局的RPC连接（长连接方式）
	globalXClient      client.XClient
	globalXClientMutex sync.Mutex
	// 数据库操作的同步锁
	saveIPMutex     sync.RWMutex
	saveDomainMutex sync.RWMutex
	// MainTaskResult 缓存汇总各个子任务、保存任务的结果
	MainTaskResult      map[string]MainTaskResultMap
	MainTaskResultMutex sync.Mutex
)

// CallXClient RPC远程调用
func CallXClient(serviceMethod string, args interface{}, reply interface{}) error {
	globalXClientMutex.Lock()
	defer globalXClientMutex.Unlock()

	if globalXClient == nil {
		host := conf.GlobalWorkerConfig().Rpc.Host
		if conf.RunMode == conf.Debug || host == "0.0.0.0" {
			host = "127.0.0.1"
		}
		d, _ := client.NewPeer2PeerDiscovery(fmt.Sprintf("tcp@%s:%d", host, conf.GlobalWorkerConfig().Rpc.Port), "")
		globalXClient = client.NewXClient("Service", client.Failtry, client.RandomSelect, d, client.DefaultOption)
		globalXClient.Auth(conf.GlobalWorkerConfig().Rpc.AuthKey)
	}

	return globalXClient.Call(context.Background(), serviceMethod, args, reply)
}

// SaveScanResult 保存IP与域名的扫描结果
func (s *Service) SaveScanResult(ctx context.Context, args *ScanResultArgs, replay *string) error {
	var msg []string
	if args.IPConfig != nil && args.IPResult != nil {
		r := portscan.Result{
			IPResult: args.IPResult,
		}

		saveIPMutex.Lock()
		msg = append(msg, r.SaveResult(*args.IPConfig))
		saveIPMutex.Unlock()

		if len(args.IPResult) > 0 {
			saveTaskResult(args.TaskID, args.IPResult)
		}
	}
	if args.DomainConfig != nil && args.DomainResult != nil {
		r := domainscan.Result{
			DomainResult: args.DomainResult,
		}

		saveDomainMutex.Lock()
		msg = append(msg, r.SaveResult(*args.DomainConfig))
		saveDomainMutex.Unlock()

		if len(args.DomainResult) > 0 {
			saveTaskResult(args.TaskID, args.DomainResult)
		}
	}
	saveMainTaskResult(args.MainTaskId, args.IPResult, args.DomainResult, args.VulnerabilityResult, 0)
	*replay = strings.Join(msg, ",")
	saveMainTaskNewResult(args.MainTaskId, *replay)

	return nil
}

// SaveScreenshotResult 保存Screenshot的结果到Server
func (s *Service) SaveScreenshotResult(ctx context.Context, args *ScreenshotResultArgs, replay *string) error {
	ss := fingerprint.NewScreenShot()
	//检查保存结果的路径
	workspace := db.Workspace{Id: args.WorkspaceId}
	if workspace.Get() == false {
		logging.RuntimeLog.Error("workspace error")
	}
	screenshotPath := path.Join(conf.GlobalServerConfig().Web.WebFiles, workspace.WorkspaceGUID, "screenshot")
	if !utils.MakePath(screenshotPath) {
		logging.RuntimeLog.Error("创建保存screenshot的目录失败！")
		return errors.New("创建保存screenshot的目录失败！")
	}
	count := ss.SaveFile(screenshotPath, args.FileInfo)
	saveMainTaskResult(args.MainTaskId, nil, nil, nil, count)
	*replay = fmt.Sprintf("screenshot:%d", count)
	return nil
}

// SaveIconImageResult 保存IconImage结果到Server
func (s *Service) SaveIconImageResult(ctx context.Context, args *IconHashResultArgs, replay *string) error {
	workspace := db.Workspace{Id: args.WorkspaceId}
	if workspace.Get() == false {
		logging.RuntimeLog.Error("workspace error")
	}
	iconImagePath := path.Join(conf.GlobalServerConfig().Web.WebFiles, workspace.WorkspaceGUID, "iconimage")
	if !utils.MakePath(iconImagePath) {
		*replay = "fail to mkdir"
		logging.RuntimeLog.Error("创建任务保存Image的目录失败！")
		return errors.New("创建任务保存Image的目录失败！")
	}
	hash := fingerprint.NewIconHash()
	*replay = hash.SaveFile(iconImagePath, args.IconHashInfo)

	return nil
}

// SaveVulnerabilityResult 保存漏洞结果
func (s *Service) SaveVulnerabilityResult(ctx context.Context, args *ScanResultArgs, replay *string) error {
	*replay = pocscan.SaveResult(args.VulnerabilityResult)
	if len(args.VulnerabilityResult) > 0 {
		saveTaskResult(args.TaskID, args.VulnerabilityResult)
		saveMainTaskResult(args.MainTaskId, nil, nil, args.VulnerabilityResult, 0)
		saveMainTaskNewResult(args.MainTaskId, *replay)
	}
	return nil
}

// SaveICPResult 保存ICP查询结果到服务器的查询缓存文件中
func (s *Service) SaveICPResult(ctx context.Context, args *map[string]*onlineapi.ICPInfo, replay *string) error {
	if *args == nil || len(*args) <= 0 {
		*replay = "icp:0"
		return nil
	}
	icp := onlineapi.NewICPQuery(onlineapi.ICPQueryConfig{})
	for k, v := range *args {
		icp.ICPMap[k] = v
	}
	if icp.SaveLocalICPInfo() {
		*replay = fmt.Sprintf("icp:%d", len(*args))
	} else {
		logging.RuntimeLog.Error("save icp fail")
		*replay = "save icp fail"
	}

	return nil
}

// SaveWhoisResult 保存whois查询结果到服务器的查询缓存文件中
func (s *Service) SaveWhoisResult(ctx context.Context, args *map[string]*whoisparser.WhoisInfo, replay *string) error {
	if *args == nil || len(*args) <= 0 {
		*replay = "whois:0"
		return nil
	}
	w := onlineapi.NewWhois(onlineapi.WhoisQueryConfig{})
	for k, v := range *args {
		w.WhoisMap[k] = v
	}
	if w.SaveLocalWhoisInfo() {
		*replay = fmt.Sprintf("whois:%d", len(*args))
	} else {
		logging.RuntimeLog.Error("save whois fail")
		*replay = "save whois fail"
	}

	return nil
}

// CheckTask 检查任务在数据库中的状态：任务是否存在、是否被取消，任务状态、结果
func (s *Service) CheckTask(ctx context.Context, args *string, replay *TaskStatusArgs) error {
	taskRun := &db.TaskRun{TaskId: *args}
	if !taskRun.GetByTaskId() {
		logging.RuntimeLog.Warningf("task not exists: %s", taskRun.TaskId)
		return nil
	}
	taskMain := &db.TaskMain{TaskId: taskRun.MainTaskId}
	if !taskMain.GetByTaskId() {
		logging.RuntimeLog.Warningf("mainTask not exists: %s", taskMain.TaskId)
		return nil
	}
	replay.IsExist = true
	if taskRun.State == ampq.REVOKED || taskMain.State == ampq.REVOKED {
		replay.IsRevoked = true
	}
	replay.TaskID = taskRun.TaskId
	replay.State = taskRun.State
	replay.Worker = taskRun.Worker
	replay.Result = taskRun.Result

	return nil
}

// UpdateTask 更新任务状态到数据库中
func (s *Service) UpdateTask(ctx context.Context, args *TaskStatusArgs, replay *bool) error {
	taskCheck := &db.TaskRun{TaskId: args.TaskID}
	if !taskCheck.GetByTaskId() {
		logging.RuntimeLog.Warningf("task not exists: %s", args.TaskID)
		return nil
	}
	dt := time.Now()
	task := &db.TaskRun{
		TaskId: args.TaskID,
		State:  args.State,
		Worker: args.Worker,
		Result: args.Result,
	}
	switch args.State {
	case ampq.SUCCESS:
		task.SucceededTime = &dt
	case ampq.FAILURE:
		task.FailedTime = &dt
	case ampq.REVOKED:
		task.RevokedTime = &dt
	case ampq.STARTED:
		task.StartedTime = &dt
	case ampq.RETRY:
		task.RetriedTime = &dt
	case ampq.RECEIVED:
		task.ReceivedTime = &dt
	}
	if task.SaveOrUpdate() {
		*replay = true
	} else {
		logging.RuntimeLog.Errorf("update task:%s,state:%s fail !", args.TaskID, args.State)
	}
	return nil
}

// KeepAlive worker通过RPC，保持与server的心跳与同步
func (s *Service) KeepAlive(ctx context.Context, args *KeepAliveInfo, replay *map[string]string) error {
	if args.WorkerStatus.WorkerName == "" {
		logging.RuntimeLog.Error("no worker name")
		return nil
	}
	WorkerStatusMutex.Lock()
	WorkerStatus[args.WorkerStatus.WorkerName] = &args.WorkerStatus
	WorkerStatus[args.WorkerStatus.WorkerName].UpdateTime = time.Now()
	WorkerStatusMutex.Unlock()

	*replay = newKeepAliveResponseInfo(args.CustomFiles)
	return nil
}

// KeepDaemonAlive worker的daemon通过RPC，保持与server的心跳与同步
func (s *Service) KeepDaemonAlive(ctx context.Context, args *string, replay *WorkerDaemonManualInfo) error {
	// args -> WorkName
	wdm := WorkerDaemonManualInfo{}
	if *args == "" {
		logging.RuntimeLog.Error("no worker name")
		*replay = wdm
		return nil
	}
	WorkerStatusMutex.Lock()
	if _, ok := WorkerStatus[*args]; ok {
		wdm.ManualReloadFlag = WorkerStatus[*args].ManualReloadFlag
		wdm.ManualFileSyncFlag = WorkerStatus[*args].ManualFileSyncFlag
		//
		WorkerStatus[*args].WorkerDaemonUpdateTime = time.Now()
		WorkerStatus[*args].ManualFileSyncFlag = false //重置文件同步请求标志
	}
	WorkerStatusMutex.Unlock()

	*replay = wdm
	return nil
}

// NewTask 创建一个新执行任务
func (s *Service) NewTask(ctx context.Context, args *NewTaskArgs, replay *string) error {
	if args.TaskName == "" || args.ConfigJSON == "" {
		msg := "taskName or configJSON is empty!"
		replay = &msg
		return errors.New(msg)
	}
	taskId, err := serverapi.NewRunTask(args.TaskName, args.ConfigJSON, args.MainTaskID, args.LastRunTaskId)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return err
	}
	replay = &taskId

	return nil
}

// LoadOpenedPort 读取指定IP已开放的全部端口
func (s *Service) LoadOpenedPort(ctx context.Context, args *LoadIPOpenedPortArgs, replay *string) error {
	var resultIPAndPort []string

	ips := strings.Split(args.Target, ",")
	for _, ip := range ips {
		// 如果不是有效的IP（可能是域名）
		if utils.CheckIPV4(ip) == false && utils.CheckIPV4Subnet(ip) == false {
			continue
		}
		//解析原始输入，可能是ip，也可能是ip/掩码
		ipAllByParse := utils.ParseIP(ip)
		for _, ipOneByOne := range ipAllByParse {
			//Fix Bug：
			//每次重新初始化数据库对象
			ipDb := db.Ip{IpName: ipOneByOne, WorkspaceId: args.WorkspaceId}
			// 如果数据库中无IP记录
			if ipDb.GetByIp() == false {
				continue
			}
			portDb := db.Port{IpId: ipDb.Id}
			ports := portDb.GetsByIPId()
			// 如果该IP无已扫描到的开放端口
			if len(ports) == 0 {
				continue
			}
			for _, port := range ports {
				resultIPAndPort = append(resultIPAndPort, fmt.Sprintf("%s:%d", ipOneByOne, port.PortNum))
			}
		}
	}

	*replay = strings.Join(resultIPAndPort, ",")
	return nil
}

// LoadDomainOpenedPort 获取域名关联的IP的端口
func (s *Service) LoadDomainOpenedPort(ctx context.Context, args *LoadDomainOpenedPortArgs, replay *map[string]map[int]struct{}) error {
	result := make(map[string]map[int]struct{})

	domainIP := make(map[string]map[string]struct{})
	//获取域名关联的所有IP
	for domainName := range args.Target {
		domain := db.Domain{DomainName: domainName, WorkspaceId: args.WorkspaceId}
		if !domain.GetByDomain() {
			continue
		}
		domainIP[domainName] = make(map[string]struct{})
		domainAttr := db.DomainAttr{RelatedId: domain.Id}
		domainAttrData := domainAttr.GetsByRelatedId()
		for _, da := range domainAttrData {
			if da.Tag == "A" {
				domainIP[domainName][da.Content] = struct{}{}
			}
		}
	}
	//获取IP的开放的端口
	for domainName, ips := range domainIP {
		ports := make(map[int]struct{})
		for ip := range ips {
			ipDb := db.Ip{IpName: ip, WorkspaceId: args.WorkspaceId}
			if ipDb.GetByIp() {
				portDb := db.Port{IpId: ipDb.Id}
				pts := portDb.GetsByIPId()
				for _, p := range pts {
					ports[p.PortNum] = struct{}{}
				}
			}
		}
		result[domainName] = ports
	}
	*replay = result
	return nil
}

// SaveRuntimeLog 保存RuntimeLog
func (s *Service) SaveRuntimeLog(ctx context.Context, args *RuntimeLogArgs, replay *string) error {
	if len(args.Source) == 0 || len(args.LogMessage) == 0 {
		msg := "null source or message"
		replay = &msg
		return errors.New(msg)
	}
	logMessage := logging.RuntimeLogMessage{}
	err := json.Unmarshal(args.LogMessage, &logMessage)
	if err != nil {
		msg := "runtimelog message error"
		replay = &msg
		return err
	}
	rtlog := db.RuntimeLog{
		Source: args.Source,
		File:   logMessage.File,
		Func:   logMessage.Func,
		Level:  logMessage.Level,
	}
	if len(logMessage.Message) > 500 {
		rtlog.Message = logMessage.Message[:500]
	} else {
		rtlog.Message = logMessage.Message
	}
	if rtlog.Add() {
		msg := "save success"
		replay = &msg
	} else {
		msg := "save fail"
		replay = &msg
	}

	return nil
}

// getWorkspaceGUIDByRunTaskId 根据runtask获取workspace的GUID
func getWorkspaceGUIDByRunTaskId(taskId string) string {
	runTask := db.TaskRun{TaskId: taskId}
	if runTask.GetByTaskId() {
		workspace := db.Workspace{Id: runTask.WorkspaceId}
		if workspace.Get() {
			return workspace.WorkspaceGUID
		}
	}
	return ""
}

// saveTaskResult 将任务结果保存到本地文件
func saveTaskResult(taskID string, result interface{}) {
	if taskID == "" {
		logging.RuntimeLog.Error("任务ID为空！")
		return
	}
	workspaceGUID := getWorkspaceGUIDByRunTaskId(taskID)
	if workspaceGUID == "" {
		logging.RuntimeLog.Error("workspace GUID为空！")
		return
	}
	//检查保存结果的路径
	resultPath := path.Join(conf.GlobalServerConfig().Web.WebFiles, workspaceGUID, "taskresult")
	if !utils.MakePath(resultPath) {
		logging.RuntimeLog.Error("创建任务保存结果的目录失败！")
		return
	}
	content, err := json.Marshal(result)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return
	}
	fileName := filepath.Join(resultPath, fmt.Sprintf("%s.json", taskID))
	//读原来的任务保存结果
	//主要是针对FOFA这种有IP同时也有Domain结果，防止覆盖
	oldContent, err := os.ReadFile(fileName)
	if err == nil {
		var buff bytes.Buffer
		buff.Write([]byte("["))
		buff.Write(oldContent)
		buff.Write([]byte(","))
		buff.Write(pretty.Pretty(content))
		buff.Write([]byte("]"))
		err = os.WriteFile(fileName, buff.Bytes(), 0666)
	} else {
		err = os.WriteFile(fileName, pretty.Pretty(content), 0666)
	}
	if err != nil {
		logging.RuntimeLog.Error(err)
	}
}

// saveMainTaskResult 保存runtask的任务结果到maintask的缓存中
func saveMainTaskResult(taskId string, ipResult map[string]*portscan.IPResult, domainResult map[string]*domainscan.DomainResult, vulResult []pocscan.Result, screenshotResult int) {
	MainTaskResultMutex.Lock()
	defer MainTaskResultMutex.Unlock()

	if _, ok := MainTaskResult[taskId]; !ok {
		return
	}
	taskObj := MainTaskResult[taskId]
	if ipResult != nil {
		for ip, ipr := range ipResult {
			if _, ok := taskObj.IPResult[ip]; !ok {
				taskObj.IPResult[ip] = make(map[int]interface{})
			}
			for port := range ipr.Ports {
				taskObj.IPResult[ip][port] = struct{}{}
			}
		}
	}
	if domainResult != nil {
		for domain := range domainResult {
			if _, ok := taskObj.DomainResult[domain]; !ok {
				taskObj.DomainResult[domain] = struct{}{}
			}
		}
	}
	if vulResult != nil {
		for _, poc := range vulResult {
			if _, ok := taskObj.VulResult[poc.Target]; !ok {
				taskObj.VulResult[poc.Target] = make(map[string]interface{})
			}
			taskObj.VulResult[poc.Target][poc.PocFile] = struct{}{}
		}
	}
	taskObj.ScreenShotResult += screenshotResult

	MainTaskResult[taskId] = taskObj
	return
}

// saveMainTaskNewResult 解析并保存任务结果中新增的资产数量
func saveMainTaskNewResult(mainTaskId, msg string) {
	MainTaskResultMutex.Lock()
	defer MainTaskResultMutex.Unlock()

	if _, ok := MainTaskResult[mainTaskId]; !ok {
		return
	}
	taskObj := MainTaskResult[mainTaskId]
	allResult := strings.Split(msg, ",")
	for _, result := range allResult {
		kv := strings.Split(result, ":")
		switch kv[0] {
		case "ipNew":
			if v, err := strconv.Atoi(kv[1]); err == nil {
				taskObj.IPNew += v
			}
		case "portNew":
			if v, err := strconv.Atoi(kv[1]); err == nil {
				taskObj.PortNew += v
			}
		case "domainNew":
			if v, err := strconv.Atoi(kv[1]); err == nil {
				taskObj.DomainNew += v
			}
		case "vulnerabilityNew":
			if v, err := strconv.Atoi(kv[1]); err == nil {
				taskObj.VulnerabilityNew += v
			}
		}
	}

	MainTaskResult[mainTaskId] = taskObj
	return
}
