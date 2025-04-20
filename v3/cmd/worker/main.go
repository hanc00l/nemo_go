package main

import (
	"encoding/json"
	"fmt"
	"github.com/hanc00l/nemo_go/v3/pkg/conf"
	"github.com/hanc00l/nemo_go/v3/pkg/core"
	"github.com/hanc00l/nemo_go/v3/pkg/logging"
	"github.com/hanc00l/nemo_go/v3/pkg/socks5forward"
	"github.com/hanc00l/nemo_go/v3/pkg/task/workerapi"
	"github.com/hanc00l/nemo_go/v3/pkg/utils"
	//_ "net/http/pprof"
	"time"
)

/*
	WorkerOptions: 定义worker的启动配置项
    WorkerConfig: 定义Worker的运行参数，除了RPC的配置由WorkerOptions指定，其它参数都由worker在启动时通过RPC从server获取

	Worker的运行参数由两种方式指定：
        一是-f指定本地配置文件。
        二是如果没有-f，只需要指定RPC的配置后，worker则从server通过RPC获取配置；指定RPC的参数有两种方式：命令行，或者环境变量。

	服务端的运行参数包括server.yaml和worker.yaml，其中server.yaml定义了server的配置，worker.yaml定义了worker的运行参数，供worker启动时获取。
*/

// keepAlive worker与server的心跳与同步
func keepAlive() {
	workerapi.WStatus.Lock()
	workerapi.WStatus.WorkerRunOption, _ = json.Marshal(core.WorkerRunOption)
	workerapi.WStatus.Unlock()
	time.Sleep(10 * time.Second)
	for {
		workerapi.WStatus.Lock()
		if !core.DoKeepAlive(&workerapi.WStatus) {
			logging.RuntimeLog.Errorf("keep alive fail")
			logging.CLILog.Error("keep alive fail")
		}
		workerapi.WStatus.Unlock()
		time.Sleep(30 * time.Second)
	}
}

func initWorkerStatus() {
	workerapi.WStatus.WorkerName = core.GetWorkerNameBySelf()
	workerapi.WStatus.CreateTime = time.Now()
	workerapi.WStatus.UpdateTime = time.Now()
	workerapi.WStatus.WorkerTopics = utils.SetToString(core.WorkerRunOption.WorkerTopic)
}

func startWorker() {
	for mode := range core.WorkerRunOption.WorkerTopic {
		go func(topicName string, concurrency int) {
			err := workerapi.StartWorker(topicName, concurrency)
			if err != nil {
				logging.CLILog.Error(err.Error())
				logging.RuntimeLog.Fatal(err.Error())
			}
		}(mode, core.WorkerRunOption.Concurrency)
		time.Sleep(1 * time.Second)
	}
}

// startSocks5Forward 启动socks5代理转发
func startSocks5Forward() {
	localPort := 5010
	chListenFail := make(chan struct{})
	go func() {
		for {
			conf.Socks5ForwardAddr = fmt.Sprintf("127.0.0.1:%d", localPort)
			go socks5forward.StartSocks5Forward(conf.Socks5ForwardAddr, chListenFail)
			select {
			case <-chListenFail:
				localPort++
			}
		}
	}()
}

func startRedisLocalProxy() {
	// 将redis通过本地端口转发到远程redis tunnel，在验证通过后才会真正转发到远程redis，避免server端直接暴露redis端口
	// reverseProxyTunnel 提供认证的redis通道，由server提供
	reverseProxyTunnel := fmt.Sprintf("%s:%d", conf.GlobalWorkerConfig().Service.Host, conf.GlobalWorkerConfig().RedisTunnel.Port)
	// localRedisListenAddr 本地redis代理的监听地址，由mq使用
	localRedisListenAddr := fmt.Sprintf("%s:%d", conf.GlobalWorkerConfig().Redis.Host, conf.GlobalWorkerConfig().Redis.Port)
	proxy := core.NewRedisProxyServer(reverseProxyTunnel, localRedisListenAddr, conf.GlobalWorkerConfig().RedisTunnel.AuthKey)
	go proxy.Start()
}

func main() {
	//pprof
	//if conf.RunMode == conf.Debug {
	//	go func() {
	//		log.Println(http.ListenAndServe("localhost:6060", nil))
	//	}()
	//}
	core.WorkerRunOption = core.PrepareWorkerOptions()
	if core.WorkerRunOption == nil {
		return
	}
	conf.NoProxyByCmd = core.WorkerRunOption.NoProxy

	go keepAlive()
	go core.StartSaveRuntimeLog(core.GetWorkerNameBySelf())
	core.CheckWorkerPerformance(core.WorkerRunOption.WorkerPerformance)
	initWorkerStatus()
	if !core.WorkerRunOption.NoProxy {
		startSocks5Forward()
	}
	if !core.WorkerRunOption.NoRedisProxy {
		startRedisLocalProxy()
	}
	startWorker()
	core.SetupCloseHandler()
}
