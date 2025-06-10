package core

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hanc00l/nemo_go/v3/pkg/conf"
	"github.com/hanc00l/nemo_go/v3/pkg/db"
	"github.com/hanc00l/nemo_go/v3/pkg/logging"
	"github.com/hanc00l/nemo_go/v3/pkg/resource"
	"github.com/hanc00l/nemo_go/v3/pkg/task/custom"
	"github.com/hanc00l/nemo_go/v3/pkg/task/execute"
	"github.com/hanc00l/nemo_go/v3/pkg/utils"
	"github.com/redis/go-redis/v9"
	"github.com/smallnest/rpcx/client"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Service RPC服务
type Service struct{}

// TaskStatusArgs 任务状态请求与返回参数
type TaskStatusArgs struct {
	WorkspaceId string
	TaskID      string
	IsExist     bool
	IsFinished  bool
	State       string
	Worker      string
	Result      string
}

type CheckTaskArgs struct {
	TaskID     string
	MainTaskID string
}

// NewTaskArgs 新建任务请求与返回参数
type NewTaskArgs struct {
	TaskName      string
	ConfigJSON    string
	MainTaskID    string
	LastRunTaskId string
}
type TaskAssetDocumentResultArgs struct {
	WorkspaceId string
	MainTaskId  string
	Result      []db.AssetDocument
}

type ScreenShotResultArgs struct {
	WorkspaceId    string
	Scheme         string
	Host           string
	Port           string
	ScreenshotByte []byte
}

type VulResultArgs struct {
	WorkspaceId string
	MainTaskId  string
	Result      []db.VulDocument
}

type RuntimeLogArgs struct {
	Source     string
	LogMessage []byte
}

type RequestResourceArgs struct {
	Category string
	Name     string
}

type ResourceResultArgs struct {
	Path  string
	Hash  string
	Bytes []byte
}

type AssetSaveResultResp struct {
	AssetTotal  int `json:"assetTotal,omitempty"`
	AssetNew    int `json:"assetNew,omitempty"`
	AssetUpdate int `json:"assetUpdate,omitempty"`
	HostTotal   int `json:"hostTotal,omitempty"`
	HostNew     int `json:"hostNew,omitempty"`
	HostUpdate  int `json:"hostUpdate,omitempty"`
	ScreenShot  int `json:"screenshot,omitempty"`
	VulTotal    int `json:"vulTotal,omitempty"`
	VulNew      int `json:"vulNew,omitempty"`
	VulUpdate   int `json:"vulUpdate,omitempty"`
}

var (
	// globalXClient 全局的RPC连接（长连接方式）
	globalXClient      client.XClient
	globalXClientMutex sync.Mutex
	TLSCertFile        string
	TLSKeyFile         string
)

func (r *AssetSaveResultResp) String() string {
	var sb strings.Builder
	space := ""
	if r.AssetTotal > 0 {
		sb.WriteString(fmt.Sprintf("资产: %d", r.AssetTotal))
		space = " "
	}
	if r.AssetNew > 0 || r.AssetUpdate > 0 {
		sb.WriteString("(")
		if r.AssetNew > 0 {
			sb.WriteString(fmt.Sprintf("新增%d", r.AssetNew))
		}
		if r.AssetUpdate > 0 {
			sb.WriteString(fmt.Sprintf("更新%d", r.AssetUpdate))
		}
		sb.WriteString(")")
	}
	if r.HostTotal > 0 {
		sb.WriteString(fmt.Sprintf("%s主机: %d", space, r.HostTotal))
		space = " "
		if r.HostNew > 0 || r.HostUpdate > 0 {
			sb.WriteString("(")
			if r.HostNew > 0 {
				sb.WriteString(fmt.Sprintf("新增%d", r.HostNew))
			}
			if r.HostUpdate > 0 {
				sb.WriteString(fmt.Sprintf("更新%d", r.HostUpdate))
			}
			sb.WriteString(")")
		}
	}
	if r.ScreenShot > 0 {
		sb.WriteString(fmt.Sprintf("%s截图:%d", space, r.ScreenShot))
		space = " "
	}
	if r.VulTotal > 0 {
		sb.WriteString(fmt.Sprintf("%s漏洞: %d", space, r.VulTotal))
		space = " "
		if r.VulNew > 0 || r.VulUpdate > 0 {
			sb.WriteString("(")
			if r.VulNew > 0 {
				sb.WriteString(fmt.Sprintf("新增%d", r.VulNew))
			}
			if r.VulUpdate > 0 {
				sb.WriteString(fmt.Sprintf("更新%d", r.VulUpdate))
			}
			sb.WriteString(")")
		}
	}

	return sb.String()
}

// CallXClient RPC远程调用
func CallXClient(serviceMethod string, args interface{}, reply interface{}) error {
	globalXClientMutex.Lock()
	defer globalXClientMutex.Unlock()

	if globalXClient == nil {
		host := conf.GlobalWorkerConfig().Service.Host
		if conf.RunMode == conf.Debug || host == "0.0.0.0" {
			host = "127.0.0.1"
		}
		option := client.DefaultOption
		option.TLSConfig = &tls.Config{InsecureSkipVerify: true}
		d, _ := client.NewPeer2PeerDiscovery(fmt.Sprintf("tcp@%s:%d", host, conf.GlobalWorkerConfig().Service.Port), "")
		globalXClient = client.NewXClient("Service", client.Failtry, client.RandomSelect, d, option)
		globalXClient.Auth(conf.GlobalWorkerConfig().Service.AuthKey)
	}

	return globalXClient.Call(context.Background(), serviceMethod, args, reply)
}

// CallImageServiceXClient 调用图像服务的RPC远程调用
func CallImageServiceXClient(serviceMethod string, args interface{}, reply interface{}) error {
	option := client.DefaultOption
	option.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	d, _ := client.NewPeer2PeerDiscovery(fmt.Sprintf("tcp@%s:%d", conf.GlobalServerConfig().ImageService.Host, conf.GlobalServerConfig().ImageService.Port), "")
	xc := client.NewXClient("Service", client.Failtry, client.RandomSelect, d, option)
	xc.Auth(conf.GlobalServerConfig().ImageService.AuthKey)

	return xc.Call(context.Background(), serviceMethod, args, reply)
}

// CheckTask 检查任务在数据库中的状态：任务是否存在、是否被取消，任务状态、结果
func (s *Service) CheckTask(ctx context.Context, args *CheckTaskArgs, replay *TaskStatusArgs) error {
	mongoClient, err := db.GetClient()
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		return err
	}
	defer db.CloseClient(mongoClient)
	etDoc, err := db.NewExecutorTask(mongoClient).GetByTaskId(args.TaskID)
	if err != nil {
		return nil
	}
	if len(args.MainTaskID) > 0 {
		_, err = db.NewMainTask(mongoClient).GetByTaskId(args.MainTaskID)
		if err != nil {
			return nil
		}
	}
	replay.IsExist = true
	if etDoc.EndTime != nil {
		replay.IsFinished = true
	}
	replay.TaskID = etDoc.TaskId
	replay.State = etDoc.Status
	replay.Worker = etDoc.Worker
	replay.Result = etDoc.Result

	return nil
}

// UpdateTask 更新任务状态到数据库中
func (s *Service) UpdateTask(ctx context.Context, args *TaskStatusArgs, replay *bool) error {
	// 更新executorTask的状态
	mongoClient, err := db.GetClient()
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		return err
	}
	defer db.CloseClient(mongoClient)

	executorTask := db.NewExecutorTask(mongoClient)
	executorTaskDoc, err := executorTask.GetByTaskId(args.TaskID)
	if err != nil {
		return err
	}
	dt := time.Now()
	update := bson.M{db.Status: args.State, "worker": args.Worker, "result": args.Result}
	if args.State == STARTED {
		update[db.StartTime] = dt
	} else if args.State == SUCCESS || args.State == FAILURE {
		update[db.EndTime] = dt
	}
	isSuccess, err := executorTask.Update(executorTaskDoc.Id.Hex(), update)
	if err != nil {
		*replay = false
		if !errors.Is(err, mongo.ErrNoDocuments) {
			logging.RuntimeLog.Errorf("更新任务状态失败:%s,state:%s,err:%v !", args.TaskID, args.State, err)
			return err
		}
		return nil
	}
	if !isSuccess {
		logging.RuntimeLog.Errorf("更新任务状态失败:%s,state:%s fail !", args.TaskID, args.State)
		*replay = false
		return errors.New("更新任务状态失败")
	}
	*replay = true

	return nil
}

func (s *Service) NewTask(ctx context.Context, args *execute.ExecutorTaskInfo, replay *bool) error {
	err := newExecutorTask(*args)
	if err != nil {
		*replay = false
		return err
	}
	*replay = true
	return nil
}

func (s *Service) SaveTaskResult(ctx context.Context, args *TaskAssetDocumentResultArgs, replay *string) error {
	blc := NewBlacklist()
	blc.LoadBlacklist(args.WorkspaceId)
	mongoClient, err := db.GetClient()
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		return err
	}
	defer db.CloseClient(mongoClient)
	ipl4 := custom.NewIPv4Location(args.WorkspaceId)
	ipl6, _ := custom.NewIPv6Location()

	taskAsset := db.NewAsset(args.WorkspaceId, db.TaskAsset, args.MainTaskId, mongoClient)
	var newAsset, updateAsset, blacklist []string
	for _, doc := range args.Result {
		if blc.IsHostBlocked(doc.Host) {
			blacklist = append(blacklist, doc.Host)
			continue
		}
		for k, ipv4 := range doc.Ip.IpV4 {
			if ipv4.Location == "" {
				doc.Ip.IpV4[k].Location = ipl4.Find(ipv4.IPName)
			}
		}
		for k, ipv6 := range doc.Ip.IpV6 {
			if ipv6.Location == "" {
				doc.Ip.IpV6[k].Location = ipl6.Find(ipv6.IPName)
			}
		}
		dss, err := taskAsset.InsertOrUpdate(doc)
		if err != nil {
			logging.RuntimeLog.Warningf("保存资产失败:%s err:%v", doc.Authority, err)
			continue
		}
		if !dss.IsSuccess {
			logging.RuntimeLog.Warningf("保存资产失败:%s fail", doc.Authority)
			continue
		}
		if dss.IsNew {
			newAsset = append(newAsset, doc.Authority)
		} else {
			updateAsset = append(updateAsset, doc.Authority)
		}
	}
	if len(blacklist) > 0 {
		logging.RuntimeLog.Warningf("不会保存黑名单资产:%s", strings.Join(blacklist, ","))
	}
	msg := formatNewAndUpdateMessage(newAsset, updateAsset)
	*replay = msg

	return nil
}

// SaveScreenShotResult 保存截图结果到本地文件
/* 保存截图结果到本地文件
   由于v3版本里service与web是可以分开部署的，所以截图结果保存到本地文件的方式需要修改；如果service与web部署在同一台机器，则可以直接保存到web目录下。
   如果service与web部署在不同的机器，则需要通过rpc调用web的接口，将截图结果保存到web目录下。
   保存逻辑判断为：
       如果在server.yml中定义了imageService，检查host，port和auth均不为空的话，则调用imageService的rpc接口，将截图结果保存到web目录下；
       否则，则保存到本地文件。
*/
func (s *Service) SaveScreenShotResult(ctx context.Context, args *[]ScreenShotResultArgs, replay *string) error {
	imageService := conf.GlobalServerConfig().ImageService
	// 调用imageService的rpc接口，保存截图到web目录下
	if imageService.Host != "" && imageService.Port > 0 && imageService.AuthKey != "" {
		logging.RuntimeLog.Info("call image service to save screenshot")
		var rr string
		err := CallImageServiceXClient("UploadScreenShotResult", args, &rr)
		if err != nil {
			return err
		}
		*replay = rr
		return nil
	}
	// 保存截图到本地文件
	var r string
	err := saveScreenShotResult(ctx, args, &r)
	if err != nil {
		return err
	}
	*replay = r
	return nil
}

func (s *Service) UploadScreenShotResult(ctx context.Context, args *[]ScreenShotResultArgs, replay *string) error {
	var r string
	err := saveScreenShotResult(ctx, args, &r)
	if err != nil {
		return err
	}
	*replay = r
	return nil
}

func saveScreenShotResult(ctx context.Context, args *[]ScreenShotResultArgs, replay *string) error {
	thumbnailWidth := 120
	for _, ss := range *args {
		screenshotPath := filepath.Join(conf.GlobalServerConfig().Web.WebFiles, ss.WorkspaceId, "screenshot")
		domainPath := filepath.Join(screenshotPath, ss.Host)
		if !utils.MakePath(screenshotPath) || !utils.MakePath(domainPath) {
			logging.RuntimeLog.Errorf("check upload path fail:%s", domainPath)
			return fmt.Errorf("check upload path fail:%s", domainPath)
		}
		fileName := filepath.Join(domainPath, fmt.Sprintf("%s_%s.png", ss.Port, ss.Scheme))
		err := os.WriteFile(fileName, ss.ScreenshotByte, 0666)
		if err != nil {
			logging.RuntimeLog.Errorf("write file %s fail:%v", fileName, err)
			return err
		}
		//生成缩略图
		fileNameThumbnail := filepath.Join(domainPath, fmt.Sprintf("%s_%s_thumbnail.png", ss.Port, ss.Scheme))
		if !utils.ReSizePicture(fileName, fileNameThumbnail, thumbnailWidth, 0) {
			logging.RuntimeLog.Error("generate thumbnail picture fail")
			return fmt.Errorf("generate thumbnail picature fail")
		}
	}
	var assetResult AssetSaveResultResp
	assetResult.ScreenShot = len(*args)
	*replay = assetResult.String()

	return nil
}

func (s *Service) SaveVulResult(ctx context.Context, args *VulResultArgs, replay *string) error {
	mongoClient, err := db.GetClient()
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		return err
	}
	defer db.CloseClient(mongoClient)
	vul := db.NewVul(args.WorkspaceId, db.TaskVul, mongoClient)
	for _, doc := range args.Result {
		oldDocs, _ := vul.Find(bson.M{"authority": doc.Authority, "task_id": doc.TaskId, "url": doc.Url, "source": doc.Source, "pocfile": doc.PocFile}, 0, 0)
		var isSuccess bool
		if len(oldDocs) == 0 {
			isSuccess, err = vul.Insert(doc)
		} else {
			doc.Id = oldDocs[0].Id
			doc.CreateTime = oldDocs[0].CreateTime
			isSuccess, err = vul.Update(oldDocs[0].Id.Hex(), doc)
		}
		if err != nil {
			logging.RuntimeLog.Warningf("保存漏洞失败:%s fail，err:%v", doc.Authority, err)
			continue
		}
		if !isSuccess {
			logging.RuntimeLog.Warningf("保存漏洞失败:%s fail", doc.Authority)
			continue
		}
	}
	var assetResult AssetSaveResultResp
	assetResult.VulTotal = len(args.Result)
	*replay = assetResult.String()

	return nil
}

func (s *Service) SaveQueryData(ctx context.Context, args *[]db.QueryDocument, replay *string) error {
	mongoClient, err := db.GetClient()
	if err != nil {
		logging.RuntimeLog.Errorf("get mongo client fail:%v", err)
		return err
	}
	defer db.CloseClient(mongoClient)

	q := db.NewQueryData(db.GlobalDatabase, mongoClient)
	for _, doc := range *args {
		oldDoc, _ := q.GetByDomain(doc.Domain, doc.Category)
		if oldDoc == nil {
			_, err = q.Insert(&doc)
			if err != nil {
				logging.RuntimeLog.Errorf(err.Error())
				return err
			}
		} else {
			if oldDoc.Content != doc.Content {
				_, err = q.Update(oldDoc.Id, doc)
				if err != nil {
					logging.RuntimeLog.Errorf(err.Error())
					return err
				}
			}
		}
	}
	*replay = fmt.Sprintf("数据:%d", len(*args))

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
		msg := "运行日志格式化错误"
		replay = &msg
		return err
	}
	doc := db.RuntimeLogDocument{
		Source:  args.Source,
		File:    logMessage.File,
		Func:    logMessage.Func,
		Level:   logMessage.Level,
		Message: logMessage.Message,
	}
	mongoClient, err := db.GetClient()
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		return err
	}
	defer db.CloseClient(mongoClient)

	_, err = db.NewRuntimeLog(mongoClient).Insert(doc)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		msg := "保存运行日志失败"
		replay = &msg
	} else {
		msg := "保存运行日志成功"
		replay = &msg
	}

	return nil
}

func (s *Service) LookupQueryData(ctx context.Context, args *db.QueryDocument, replay *db.QueryDocument) error {
	mongoClient, err := db.GetClient()
	if err != nil {
		logging.RuntimeLog.Errorf("get mongo client fail:%v", err)
		return err
	}
	defer db.CloseClient(mongoClient)

	q := db.NewQueryData(db.GlobalDatabase, mongoClient)
	doc, err := q.GetByDomain(args.Domain, args.Category)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		logging.RuntimeLog.Errorf(err.Error())
		return err
	}
	if doc == nil {
		*replay = db.QueryDocument{}
	}
	return nil
}

func (s *Service) LoadWorkerConfig(ctx context.Context, args *string, replay *conf.Worker) error {
	err := conf.GlobalWorkerConfig().ReloadConfig()
	if err != nil {
		logging.RuntimeLog.Errorf(err.Error())
		return err
	}
	*replay = *conf.GlobalWorkerConfig()

	return err
}

// KeepAlive worker通过RPC，保持与server的心跳与同步
func (s *Service) KeepAlive(ctx context.Context, args *WorkerStatus, replay *string) error {
	if args == nil || args.WorkerName == "" {
		logging.RuntimeLog.Error("no worker name")
		return nil
	}
	rdb, err := GetRedisClient()
	if err != nil {
		logging.RuntimeLog.Errorf("get redis client fail:%v", err)
		return err
	}
	defer rdb.Close()

	workerAliveStatus, err := GetWorkerStatusFromRedis(rdb, args.WorkerName)
	if errors.Is(err, redis.Nil) {
		workerAliveStatus = args
	} else if err != nil {
		logging.RuntimeLog.Errorf(err.Error())
		return err
	} else {
		workerAliveStatus.MemUsed = args.MemUsed
		workerAliveStatus.CPULoad = args.CPULoad
		workerAliveStatus.TaskExecutedNumber = args.TaskExecutedNumber
		workerAliveStatus.TaskStartedNumber = args.TaskStartedNumber
	}
	workerAliveStatus.UpdateTime = time.Now()
	err = SetWorkerStatusToRedis(rdb, args.WorkerName, workerAliveStatus)
	if err != nil {
		logging.RuntimeLog.Errorf(err.Error())
		*replay = "set worker status fail"
		return err
	}
	// 远程关闭worker（针对standalone模式）
	if workerAliveStatus.ManualStopFlag {
		*replay = "stop"
	} else {
		*replay = "ok"
	}

	return nil
}

// KeepDaemonAlive worker的daemon通过RPC，保持与server的心跳与同步
func (s *Service) KeepDaemonAlive(ctx context.Context, args *string, replay *KeepAliveDaemonInfo) error {
	// args -> WorkName
	if args == nil || *args == "" {
		logging.RuntimeLog.Error("no worker name")
		return errors.New("no worker name")
	}
	rdb, err := GetRedisClient()
	if err != nil {
		logging.RuntimeLog.Errorf("get redis client fail:%v", err)
		return err
	}
	defer rdb.Close()

	daemonInfo := KeepAliveDaemonInfo{}
	workerAliveStatus, err := GetWorkerStatusFromRedis(rdb, *args)
	if errors.Is(err, redis.Nil) {
		logging.RuntimeLog.Errorf("worker %s not exist", *args)
		return errors.New("worker not exist")
	} else if err != nil {
		logging.RuntimeLog.Errorf(err.Error())
		return err
	}
	// 远程关闭worker（针对daemon模式）
	daemonInfo.ManualStopFlag = workerAliveStatus.ManualStopFlag
	daemonInfo.ManualReloadFlag = workerAliveStatus.ManualReloadFlag
	daemonInfo.ManualInitEnvFlag = workerAliveStatus.ManualInitEnvFlag
	daemonInfo.ManualConfigAndPocSyncFlag = workerAliveStatus.ManualConfigAndPocSyncFlag
	daemonInfo.ManualUpdateOptionFlag = workerAliveStatus.ManualUpdateOptionFlag
	if workerAliveStatus.ManualUpdateOptionFlag {
		w := WorkerOption{}
		if err = json.Unmarshal(workerAliveStatus.WorkerUpdateOption, &w); err == nil {
			daemonInfo.WorkerRunOption = &w
		}
	}
	// 复位标志
	workerAliveStatus.ManualUpdateOptionFlag = false
	workerAliveStatus.ManualInitEnvFlag = false
	workerAliveStatus.ManualReloadFlag = false
	workerAliveStatus.ManualConfigAndPocSyncFlag = false
	workerAliveStatus.ManualStopFlag = false
	workerAliveStatus.IsDaemonProcess = true
	workerAliveStatus.WorkerDaemonUpdateTime = time.Now()
	err = SetWorkerStatusToRedis(rdb, *args, workerAliveStatus)
	if err != nil {
		logging.RuntimeLog.Errorf(err.Error())
		return err
	}

	*replay = daemonInfo
	return nil
}

func (s *Service) RequestResource(ctx context.Context, args *RequestResourceArgs, replay *ResourceResultArgs) error {
	if args == nil || args.Category == "" || args.Name == "" {
		logging.RuntimeLog.Error("no category or name")
		return errors.New("no category or name")
	}
	var re resource.Resource
	var ok bool
	if re, ok = resource.Resources[args.Category][args.Name]; !ok {
		return fmt.Errorf("resource %s:%s not found", args.Category, args.Name)
	}

	var rr *resource.Resource
	var err error
	if re.Type == resource.ExecuteFile || re.Type == resource.ConfigFile || re.Type == resource.DataFile {
		rr, err = resource.LoadFileResource(args.Category, args.Name)
	} else {
		rr, err = resource.LoadDirResource(args.Category, args.Name)
	}
	if err != nil {
		logging.RuntimeLog.Errorf(err.Error())
		return err
	}
	*replay = ResourceResultArgs{Hash: rr.Hash, Bytes: rr.Bytes, Path: rr.Path}

	return nil
}

func formatNewAndUpdateMessage(newAsset, updateAsset []string) string {
	f := func(asset []string) map[string]string {
		r := make(map[string]string)
		for _, authority := range asset {
			hostPort := strings.Split(authority, ":")
			if len(hostPort) != 2 {
				continue
			}
			r[hostPort[0]] = hostPort[1]
		}
		return r
	}
	hostNew := f(newAsset)
	hostUpdate := f(updateAsset)

	var assetResult AssetSaveResultResp
	assetResult.AssetTotal = len(newAsset) + len(updateAsset)
	assetResult.HostTotal = len(hostNew) + len(hostUpdate)

	return assetResult.String()

}

// checkWorkerStatus 对超过指定时间未同步的的worker，移除相关信息
func checkWorkerStatus() {
	rdb, err := GetRedisClient()
	if err != nil {
		logging.RuntimeLog.Errorf("get redis client fail:%v", err)
		return
	}
	defer rdb.Close()

	workerAliveStatus, err := LoadWorkerStatusFromRedis(rdb)
	if errors.Is(err, redis.Nil) {
		return
	} else if err != nil {
		logging.RuntimeLog.Errorf(err.Error())
		return
	}
	for _, v := range workerAliveStatus {
		//fmt.Println(v)
		// 如果worker资源（cpu、内存）比较低，会在大并发任务的时候导致很长一段时间阻塞同步
		// 因此将worker不存活的时间调整为超过6个小时
		if time.Now().Sub(v.UpdateTime).Hours() > 6 {
			//delete(workerAliveStatus, v.WorkerName)
			if err := DeleteWorkerStatusFromRedis(rdb, v.WorkerName); err != nil {
				logging.RuntimeLog.Errorf(err.Error())
			}
		}
	}
}
