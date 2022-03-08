package conf

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
	"path/filepath"
)

const (
	Release = "Release" //正式运行模式
	Debug   = "Debug"   //开发模式
)

// RunMode 运行模式
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
	if workerConfig == nil {
		workerConfig = new(Worker)
		err := workerConfig.ReloadConfig()
		if err != nil {
			fmt.Println("Load Worker config fail!")
			os.Exit(0)
		}
	}
	return workerConfig
}

type Server struct {
	Web      Web      `yaml:"web"`
	Rpc      RPC      `yaml:"rpc"`
	Database Database `yaml:"database"`
	Rabbitmq Rabbitmq `yaml:"rabbitmq"`
	Task     Task     `yaml:"task"`
}

type Worker struct {
	Rpc        RPC        `yaml:"rpc"`
	Rabbitmq   Rabbitmq   `yaml:"rabbitmq"`
	API        API        `yaml:"api"`
	Portscan   Portscan   `yaml:"portscan"`
	Domainscan Domainscan `yaml:"domainscan"`
	Pocscan    Pocscan    `yaml:"pocscan"`
}

type Web struct {
	Host           string `yaml:"host"`
	Port           int    `yaml:"port"`
	Username       string `yaml:"username"`
	Password       string `yaml:"password"`
	ScreenshotPath string `yaml:"screenshotPath"`
	TaskResultPath string `yaml:"taskresultPath"`
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

type Task struct {
	IpSliceNumber   int `yaml:"ipSliceNumber"`
	PortSliceNumber int `yaml:"portSliceNumber"`
}

type API struct {
	Fofa   APIKey `yaml:"fofa"`
	ICP    APIKey `yaml:"icp"`
	Quake  APIKey `yaml:"quake"`
	Hunter APIKey `yaml:"hunter"`
}

type APIKey struct {
	Name string `yaml:"name"`
	Key  string `yaml:"key"`
}

type Portscan struct {
	IsPing bool   `yaml:"ping"`
	Port   string `yaml:"port"`
	Rate   int    `yaml:"rate"`
	Tech   string `yaml:"tech"`
	Cmdbin string `yaml:"cmdbin"`
}

type Pocscan struct {
	Xray struct {
		PocPath       string `yaml:"pocPath"`
		LatestVersion string `yaml:"latest"`
	} `yaml:"xray"`
	Pocsuite struct {
		PocPath string `yaml:"pocPath"`
		Threads int    `yaml:"threads"`
	} `yaml:"pocsuite"`
}

type Domainscan struct {
	Resolver       string `yaml:"resolver"`
	Wordlist       string `yaml:"wordlist"`
	MassdnsThreads int    `yaml:"massdnsThreads"`
}

// WriteConfig 写配置到yaml文件中
func (config *Server) WriteConfig() error {
	content, err := yaml.Marshal(config)
	if err != nil {
		fmt.Println(err)
		return err
	}
	err = os.WriteFile(filepath.Join(GetRootPath(), "conf/server.yml"), content, 0666)
	if err != nil {
		fmt.Println(err)
	}
	return err
}

// ReloadConfig 从yaml文件中加载配置
func (config *Server) ReloadConfig() error {
	fileContent, err := os.ReadFile(filepath.Join(GetRootPath(), "conf/server.yml"))
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
	fileContent, err := os.ReadFile(filepath.Join(GetRootPath(), "conf/worker.yml"))
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

// GetRootPath 获取运行时系统的root位置，解决调试时无法使用相对位置的困扰
func GetRootPath() string {
	if RunMode == Debug {
		return "/Users/user/golang/nemo_go"
	}
	return "."
}
