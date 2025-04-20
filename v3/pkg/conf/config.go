package conf

import (
	"fmt"
	"gopkg.in/yaml.v3"
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
	HighPerformance   = "High"
	NormalPerformance = "Normal"
)

// WorkerPerformanceMode worker默认的性能模式为Normal
var WorkerPerformanceMode = NormalPerformance

// NoProxyByCmd 不使用proxy，由worker启动时命令行指定
var NoProxyByCmd bool
var Ipv6Support bool

// Socks5ForwardAddr chrome浏览器的socks5代理地址转发地址，支持用户名和密码认证（默认chrome不支持带认证的socks5代理）
var Socks5ForwardAddr string
var ServerDefaultConfigFile = "conf/server.yml"
var WorkerDefaultConfigFile = "conf/worker.yml"
var WorkerConfigReloadMutex sync.Mutex                        // worker读配置文件同步锁
var LocalTimeLocation, _ = time.LoadLocation("Asia/Shanghai") // 本地时区
// RunMode 运行模式：正式运行请使用Release模式，Debug模式只用于开发调试过程
var RunMode = Release

//var RunMode = Debug

// Nemo 系统运行全局配置参数
var serverConfig *Server
var workerConfig *Worker

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
	if workerConfig == nil && WorkerDefaultConfigFile != "" {
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
	Web          Web         `yaml:"web"`          // web
	Service      Service     `yaml:"service"`      // rpc service
	ImageService Service     `yaml:"imageService"` // 图片上传到web的rpc service配置，一般由web服务监听rpc service，在其它提供service的地址配置
	Database     Database    `yaml:"database"`     // 数据库配置
	Redis        Redis       `yaml:"redis"`        // redis配置
	RedisTunnel  RedisTunnel `yaml:"redisTunnel"`  // 提供redis tunnel服务的配置
}

type Worker struct {
	Service     Service     `yaml:"service"`     // 需连接的rpc service
	Redis       Redis       `yaml:"redis"`       // 供mq使用的redis地址；由worker通过本地redis代理、通过tunnel连接到实际的redis
	RedisTunnel RedisTunnel `yaml:"redisTunnel"` // 由worker通过redis tunnel连接到service提供的redis tunnel服务的地址
	API         API         `yaml:"api"`
	LLMAPI      LLMAPI      `yaml:"llmapi"`
	Proxy       Proxy       `yaml:"proxy"`
}

type Web struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	WebFiles string `yaml:"webfiles"`
}
type Service struct {
	Host    string `yaml:"host"`
	Port    int    `yaml:"port"`
	AuthKey string `yaml:"authKey"`
}

type Database struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Dbname   string `yaml:"name,omitempty"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type Redis struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
}

type RedisTunnel struct {
	Port    int    `yaml:"port"`
	AuthKey string `yaml:"authKey"`
}

type API struct {
	SearchPageSize   int    `yaml:"searchPageSize" json:"searchPageSize"`
	SearchLimitCount int    `yaml:"searchLimitCount" json:"searchLimitCount"`
	Fofa             APIKey `yaml:"fofa" json:"fofa"`
	Quake            APIKey `yaml:"quake" json:"quake"`
	Hunter           APIKey `yaml:"hunter" json:"hunter"`
	ICP              APIKey `yaml:"icp" json:"icp"`
	ICPPlus          APIKey `yaml:"icpPlus" json:"icpPlus"`
}

type LLMAPI struct {
	Kimi     APIToken `yaml:"kimi" json:"kimi"`
	DeepSeek APIToken `yaml:"deepseek" json:"deepseek"`
	Qwen     APIToken `yaml:"qwen" json:"qwen"`
}

type APIKey struct {
	Key string `yaml:"key" json:"key"`
}
type APIToken struct {
	API   string `yaml:"api" json:"api"`
	Model string `yaml:"model" json:"model"`
	Token string `yaml:"token" json:"token"`
}

type Proxy struct {
	Host []string `yaml:"host" json:"host"`
}

// WriteConfig 写配置到yaml文件中
func (config *Server) WriteConfig() error {
	content, err := yaml.Marshal(config)
	if err != nil {
		fmt.Println(err)
		return err
	}
	err = os.WriteFile(filepath.Join(GetRootPath(), ServerDefaultConfigFile), content, 0666)
	if err != nil {
		fmt.Println(err)
	}
	return err
}

// ReloadConfig 从yaml文件中加载配置
func (config *Server) ReloadConfig() error {
	fileContent, err := os.ReadFile(filepath.Join(GetRootPath(), ServerDefaultConfigFile))
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

func (config *Worker) GetWorkerConfig() *Worker {
	return workerConfig
}

func (config *Worker) SetWorkerConfig(c *Worker) {
	workerConfig = c
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
		return "/Users/user/21GolandProjects/nemo_go_v3"
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
