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
	"strings"
	"sync"
	"time"
)

// Service RPC服务
type Service struct{}

// ScanResultArgs IP与域名扫描结果请求参数
type ScanResultArgs struct {
	TaskID              string
	IPConfig            *portscan.Config
	DomainConfig        *domainscan.Config
	IPResult            map[string]*portscan.IPResult
	DomainResult        map[string]*domainscan.DomainResult
	VulnerabilityResult []pocscan.Result
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
	TaskName   string
	ConfigJSON string
	TaskID     string
}

// globalXClient 全局的RPC连接（长连接方式）
var globalXClient client.XClient

// 数据库操作的同步锁
var saveIPMutex sync.RWMutex
var saveDomainMutex sync.RWMutex

// NewXClient 获取一个RPC连接
func NewXClient() client.XClient {
	if globalXClient == nil {
		host := conf.GlobalWorkerConfig().Rpc.Host
		if conf.RunMode == conf.Debug || host == "0.0.0.0" {
			host = "127.0.0.1"
		}
		d, _ := client.NewPeer2PeerDiscovery(fmt.Sprintf("tcp@%s:%d", host, conf.GlobalWorkerConfig().Rpc.Port), "")
		globalXClient = client.NewXClient("Service", client.Failtry, client.RandomSelect, d, client.DefaultOption)
		globalXClient.Auth(conf.GlobalWorkerConfig().Rpc.AuthKey)
	}
	return globalXClient
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

	*replay = strings.Join(msg, ",")
	return nil
}

// SaveScreenshotResult 保存Screenshot的结果到Server
func (s *Service) SaveScreenshotResult(ctx context.Context, args *[]fingerprint.ScreenshotFileInfo, replay *string) error {
	ss := fingerprint.NewScreenShot()
	//检查保存结果的路径
	screenshotPath := path.Join(conf.GlobalServerConfig().Web.WebFiles, "screenshot")
	if !utils.MakePath(screenshotPath) {
		logging.RuntimeLog.Error("创建保存screenshot的目录失败！")
		return errors.New("创建保存screenshot的目录失败！")
	}
	*replay = ss.SaveFile(screenshotPath, *args)
	return nil
}

// SaveIconImageResult 保存IconImage结果到Server
func (s *Service) SaveIconImageResult(ctx context.Context, args *[]fingerprint.IconHashInfo, replay *string) error {
	iconImagePath := path.Join(conf.GlobalServerConfig().Web.WebFiles, "iconimage")
	if !utils.MakePath(iconImagePath) {
		*replay = "fail to mkdir"
		logging.RuntimeLog.Error("创建任务保存Image的目录失败！")
		return errors.New("创建任务保存Image的目录失败！")
	}
	hash := fingerprint.NewIconHash()
	*replay = hash.SaveFile(iconImagePath, *args)

	return nil
}

// SaveVulnerabilityResult 保存漏洞结果
func (s *Service) SaveVulnerabilityResult(ctx context.Context, args *ScanResultArgs, replay *string) error {
	*replay = pocscan.SaveResult(args.VulnerabilityResult)
	if len(args.VulnerabilityResult) > 0 {
		saveTaskResult(args.TaskID, args.VulnerabilityResult)
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
	task := &db.Task{TaskId: *args}
	if !task.GetByTaskId() {
		return nil
	}
	replay.IsExist = true
	if task.State == ampq.REVOKED {
		replay.IsRevoked = true
	}
	replay.TaskID = task.TaskId
	replay.State = task.State
	replay.Worker = task.Worker
	replay.Result = task.Result

	return nil
}

// UpdateTask 更新任务状态到数据库中
func (s *Service) UpdateTask(ctx context.Context, args *TaskStatusArgs, replay *bool) error {
	taskCheck := &db.Task{TaskId: args.TaskID}
	if !taskCheck.GetByTaskId() {
		logging.RuntimeLog.Errorf("task not exists: %s", args.TaskID)
		return nil
	}
	dt := time.Now()
	task := &db.Task{
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
		logging.RuntimeLog.Errorf("Update task:%s,state:%s fail !", args.TaskID, args.State)
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
	taskId, err := serverapi.NewTask(args.TaskName, args.ConfigJSON, "")
	if err != nil {
		return err
	}
	replay = &taskId

	return nil
}

// LoadOpenedPort 读取指定IP已开放的全部端口
func (s *Service) LoadOpenedPort(ctx context.Context, args *string, replay *string) error {
	var resultIPAndPort []string

	ips := strings.Split(*args, ",")
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
			ipDb := db.Ip{}
			portDb := db.Port{}
			ipDb.IpName = ipOneByOne
			// 如果数据库中无IP记录
			if ipDb.GetByIp() == false {
				continue
			}
			portDb.IpId = ipDb.Id
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

// saveTaskResult 将任务结果保存到本地文件
func saveTaskResult(taskID string, result interface{}) {
	if taskID == "" {
		logging.RuntimeLog.Error("任务ID为空！")
		return
	}
	//检查保存结果的路径
	resultPath := path.Join(conf.GlobalServerConfig().Web.WebFiles, "taskresult")
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
