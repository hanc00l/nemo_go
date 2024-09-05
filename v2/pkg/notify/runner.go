package notify

import (
	"github.com/hanc00l/nemo_go/v2/pkg/conf"
	"github.com/hanc00l/nemo_go/v2/pkg/logging"
	"sync"
)

// Sender 各个API的消息通知调用handler
type Sender interface {
	Send(token string, message string) (err error)
}

// Send 根据server的配置，调用各个接口handler发送消息通知
func Send(message string) {
	// 采用多线程同时发送模式
	swg := sync.WaitGroup{}
	for senderName, config := range conf.GlobalServerConfig().Notify {
		if len(config.Token) == 0 {
			continue
		}
		var sender Sender
		switch senderName {
		case "serverchan":
			sender = new(ServerChan)
		case "dingtalk":
			sender = new(DingTalk)
		case "feishu":
			sender = new(Feishu)
		default:
			msg := "invalid notify sender"
			logging.RuntimeLog.Error(msg)
			logging.CLILog.Error(msg)
			continue
		}
		//send message
		swg.Add(1)
		go func(s Sender, token string) {
			defer swg.Done()
			if err := s.Send(token, message); err != nil {
				logging.CLILog.Error(err)
				logging.RuntimeLog.Error(err)
			}
		}(sender, config.Token)
	}
	swg.Wait()
}
