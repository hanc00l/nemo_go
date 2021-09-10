package comm

import (
	"context"
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
	"github.com/smallnest/rpcx/client"
	"strings"
	"time"
)

// Service RPC服务
type Service struct{}

// ScanResultArgs IP与域名扫描结果请求参数
type ScanResultArgs struct {
	IPConfig     *portscan.Config
	DomainConfig *domainscan.Config
	IPResult     map[string]*portscan.IPResult
	DomainResult map[string]*domainscan.DomainResult
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

// globalXClient 全局的RPC连接（长连接方式）
var globalXClient client.XClient

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
		msg = append(msg, r.SaveResult(*args.IPConfig))
	}
	if args.DomainConfig != nil && args.DomainResult != nil {
		r := domainscan.Result{
			DomainResult: args.DomainResult,
		}
		msg = append(msg, r.SaveResult(*args.DomainConfig))
	}

	*replay = strings.Join(msg, ",")
	return nil
}

// SaveScreenshotResult 保存Screenshot的结果到Server
func (s *Service) SaveScreenshotResult(ctx context.Context, args *[]fingerprint.ScreenshotFileInfo, replay *string) error {
	ss := fingerprint.NewScreenShot()
	*replay = ss.SaveFile(conf.GlobalServerConfig().Web.ScreenshotPath, *args)
	return nil
}

// SaveVulnerabilityResult 保存漏洞结果
func (s *Service) SaveVulnerabilityResult(ctx context.Context, args *[]pocscan.Result, replay *string) error {
	*replay = pocscan.SaveResult(*args)
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
