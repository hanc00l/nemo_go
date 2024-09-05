package conf

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	Release = "Release" //正式运行模式
	Debug   = "Debug"   //开发模式
)

const (
	HighPerformance      = "High"
	NormalPerformance    = "Normal"
	SyncAssetsTypeIP     = "IP"
	SyncAssetsTypeDomain = "Domain"
	SyncOpNew            = "New"
	syncOpDel            = "Del"
)

// WorkerPerformanceMode worker默认的性能模式为Normal
var WorkerPerformanceMode = NormalPerformance

// NoProxyByCmd 不使用proxy，由worker启动时命令行指定
var NoProxyByCmd bool

// Socks5ForwardAddr chrome浏览器的socks5代理地址转发地址，支持用户名和密码认证（默认chrome不支持带认证的socks5代理）
var Socks5ForwardAddr string
var ServerDefaultConfigfile = "conf/server.yml"
var WorkerDefaultConfigFile = "conf/worker.yml"
var WorkerConfigReloadMutex sync.Mutex // worker读配置文件同步锁

// RunMode 运行模式：正式运行请使用Release模式，Debug模式只用于开发调试过程
var RunMode = Release

//var RunMode = Debug

// Nemo 系统运行全局配置参数
var serverConfig *Server
var workerConfig *Worker

// ElasticSyncAssetsArgs 用于同步资产数据到ElasticSearch的参数
type ElasticSyncAssetsArgs struct {
	Contents       []byte
	SyncOp         string
	SyncAssetsType string
}

// ElasticSyncAssetsChan 用于同步资产数据到ElasticSearch的channel
var ElasticSyncAssetsChan chan ElasticSyncAssetsArgs
var ElasticSyncAssetsChanMax = 100

func GlobalServerConfig() *Server {
	if serverConfig == nil {
		serverConfig = new(Server)
		err := serverConfig.ReloadConfig()
		if err != nil {
			fmt.Println("Load Server config fail!")
			os.Exit(0)
		}
	}
	return serverConfig
}

func GlobalWorkerConfig() *Worker {
	if workerConfig == nil {
		workerConfig = new(Worker)

		WorkerConfigReloadMutex.Lock()
		err := workerConfig.ReloadConfig()
		WorkerConfigReloadMutex.Unlock()

		if err != nil {
			fmt.Println("Load Worker config fail!")
			os.Exit(0)
		}
	}
	return workerConfig
}

type Server struct {
	Web      Web               `yaml:"web"`
	Rpc      RPC               `yaml:"rpc"`
	FileSync RPC               `yaml:"fileSync"`
	WebAPI   WebAPI            `yaml:"api"`
	Database Database          `yaml:"database"`
	Rabbitmq Rabbitmq          `yaml:"rabbitmq"`
	Elastic  ElasticSearch     `yaml:"elastic"`
	Task     Task              `yaml:"task"`
	Notify   map[string]Notify `yaml:"notify"`
	Wiki     Wiki              `yaml:"wiki"`
}

type Worker struct {
	Rpc         RPC         `yaml:"rpc"`
	FileSync    RPC         `yaml:"fileSync"`
	Rabbitmq    Rabbitmq    `yaml:"rabbitmq"`
	API         API         `yaml:"api"`
	Portscan    Portscan    `yaml:"portscan"`
	Fingerprint Fingerprint `yaml:"fingerprint"`
	Domainscan  Domainscan  `yaml:"domainscan"`
	OnlineAPI   OnlineAPI   `yaml:"onlineapi"`
	Pocscan     Pocscan     `yaml:"pocscan"`
	Proxy       Proxy       `yaml:"proxy"`
	Filter      Filter      `yaml:"filter"`
}

type Web struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	WebFiles string `yaml:"webfiles"`
}

type WebAPI struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}
type RPC struct {
	Host    string `yaml:"host"`
	Port    int    `yaml:"port"`
	AuthKey string `yaml:"authKey"`
}

type Database struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Dbname   string `yaml:"name"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type Rabbitmq struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type ElasticSearch struct {
	Url      string `yaml:"url"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type Task struct {
	IpSliceNumber   int `yaml:"ipSliceNumber"`
	PortSliceNumber int `yaml:"portSliceNumber"`
}

type API struct {
	SearchPageSize   int    `yaml:"searchPageSize"`
	SearchLimitCount int    `yaml:"searchLimitCount"`
	Fofa             APIKey `yaml:"fofa"`
	ICP              APIKey `yaml:"icp"`
	Quake            APIKey `yaml:"quake"`
	Hunter           APIKey `yaml:"hunter"`
}

type APIKey struct {
	Key string `yaml:"key"`
}

type Portscan struct {
	IsPing bool   `yaml:"ping"`
	Port   string `yaml:"port"`
	Rate   int    `yaml:"rate"`
	Tech   string `yaml:"tech"`
	Cmdbin string `yaml:"cmdbin"`
}

type Fingerprint struct {
	IsHttpx          bool `yaml:"httpx"`
	IsScreenshot     bool `yaml:"screenshot"`
	IsFingerprintHub bool `yaml:"fingerprinthub"`
	IsIconHash       bool `yaml:"iconhash"`
	IsFingerprintx   bool `yaml:"fingerprintx"`
}

type Pocscan struct {
	Xray struct {
		PocPath string `yaml:"pocPath"`
	} `yaml:"xray"`
	Nuclei struct {
		PocPath string `yaml:"pocPath"`
	} `yaml:"nuclei"`
	Goby struct {
		AuthUser string   `yaml:"authUser"`
		AuthPass string   `yaml:"authPass"`
		API      []string `yaml:"api"`
	} `yaml:"goby"`
}

type Domainscan struct {
	Resolver           string `yaml:"resolver"`
	Wordlist           string `yaml:"wordlist"`
	ProviderConfig     string `yaml:"providerConfig"`
	IsSubDomainFinder  bool   `yaml:"subfinder"`
	IsSubDomainBrute   bool   `yaml:"subdomainBrute"`
	IsSubdomainCrawler bool   `yaml:"subdomainCrawler"`
	IsIgnoreCDN        bool   `yaml:"ignoreCDN"`
	IsIgnoreOutofChina bool   `yaml:"ignoreOutofChina"`
	IsPortScan         bool   `yaml:"portscan"`
	IsWhois            bool   `yaml:"whois"`
	IsICP              bool   `yaml:"icp"`
}

type OnlineAPI struct {
	IsFofa   bool `yaml:"fofa"`
	IsQuake  bool `yaml:"quake"`
	IsHunter bool `yaml:"hunter"`
}

type Notify struct {
	Token string `yaml:"token"`
}

type Proxy struct {
	Host []string `yaml:"host"`
}

type Wiki struct {
	Feishu struct {
		AppId                  string `yaml:"appId"`
		AppSecret              string `yaml:"appSecret"`
		UserAccessRefreshToken string `yaml:"refreshToken"`
	} `yaml:"feishu"`
}

type Filter struct {
	MaxPortPerIp   int    `yaml:"maxPortPerIp"`
	MaxDomainPerIp int    `yaml:"maxDomainPerIp"`
	Title          string `yaml:"title"`
}

// WriteConfig 写配置到yaml文件中
func (config *Server) WriteConfig() error {
	content, err := yaml.Marshal(config)
	if err != nil {
		fmt.Println(err)
		return err
	}
	err = os.WriteFile(filepath.Join(GetRootPath(), ServerDefaultConfigfile), content, 0666)
	if err != nil {
		fmt.Println(err)
	}
	return err
}

// ReloadConfig 从yaml文件中加载配置
func (config *Server) ReloadConfig() error {
	fileContent, err := os.ReadFile(filepath.Join(GetRootPath(), ServerDefaultConfigfile))
	if err != nil {
		fmt.Println(err)
		return err
	}
	err = yaml.Unmarshal(fileContent, config)
	if err != nil {
		fmt.Println(err)
	}
	return err
}

// ReloadConfig 从yaml文件中加载配置
func (config *Worker) ReloadConfig() error {
	fileContent, err := os.ReadFile(filepath.Join(GetRootPath(), WorkerDefaultConfigFile))
	if err != nil {
		fmt.Println(err)
		return err
	}
	err = yaml.Unmarshal(fileContent, config)
	if err != nil {
		fmt.Println(err)
	}
	return err
}

// WriteConfig 写配置到yaml文件中
func (config *Worker) WriteConfig() error {
	content, err := yaml.Marshal(config)
	if err != nil {
		fmt.Println(err)
		return err
	}
	err = os.WriteFile(filepath.Join(GetRootPath(), WorkerDefaultConfigFile), content, 0666)
	if err != nil {
		fmt.Println(err)
	}
	return err
}

// GetProxyConfig 从配置文件中获取一个代理配置参数，多个代理则随机选取一个
func GetProxyConfig() string {
	if NoProxyByCmd {
		return ""
	}
	config := GlobalWorkerConfig()
	if len(config.Proxy.Host) == 0 {
		return ""
	}
	if len(config.Proxy.Host) == 1 {
		return config.Proxy.Host[0]
	}
	if len(config.Proxy.Host) > 1 {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		n := r.Intn(len(config.Proxy.Host))
		return config.Proxy.Host[n]
	}
	return ""
}

// GetRootPath 获取运行时系统的root位置，解决调试时无法使用相对位置的困扰
func GetRootPath() string {
	if RunMode == Debug {
		return "/Users/user/nemo_go/v2"
	}
	return "."
}

// GetAbsRootPath 获取运行时系统的root绝对位置，解决调试时无法使用相对位置的困扰
func GetAbsRootPath() string {
	rootPath, err := filepath.Abs(GetRootPath())
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return rootPath
}
