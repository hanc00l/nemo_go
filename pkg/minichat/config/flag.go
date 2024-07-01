package config

import (
	"flag"
	"github.com/hanc00l/nemo_go/pkg/minichat/util"
)

var IsNotDelFileDir = false
var ChatPath = ""
var LoadHistory = true
var MaxHistoryMessage = 1000

var RsaPrivateKey = ""
var RsaPublicKey = ""

func FlagParse() {
	flag.BoolVar(&IsNotDelFileDir, "mnd", false, "聊天结束后默认删除聊天目录,-nd则不删除")
	flag.StringVar(&ChatPath, "mc", "", "设置聊天首页路由地址,防止泄露,不配置则默认无")
	flag.BoolVar(&LoadHistory, "lh", true, "自动加载历史聊天记录,默认加载")
	flag.IntVar(&MaxHistoryMessage, "mh", 1000, "设置最大历史消息数量,默认1000")

	_, publicKey, privateKey := util.GenerateRSAKey(2048)
	RsaPrivateKey = string(privateKey)
	RsaPublicKey = string(publicKey)
	flag.Parse()
}
