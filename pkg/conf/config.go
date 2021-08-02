package conf

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
	"path/filepath"
)

const (
	Release = "Release" //正式运行模式
	Debug   = "Debug"     //开发模式
)

// RunMode 运行模式
//var RunMode = Release

var RunMode = Debug

// Nemo 系统运行全局配置参数
var Nemo Config

func init() {
	err := Nemo.ReloadConfig()
	if err != nil {
		fmt.Println("Load nemo config fail!")
		os.Exit(0)
	}
}

type Config struct {
	Web        Web        `yaml:"web"`
	Database   Database   `yaml:"database"`
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
	EncryptKey     string `yaml:"encryptKey"`
	ScreenshotPath string `yaml:"screenshotPath"`
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

type API struct {
	Fofa APIKey `yaml:"fofa"`
	ICP  APIKey `yaml:"icp"`
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

// ReloadConfig 从yaml文件中加载配置
func (config *Config) ReloadConfig() error {
	fileContent, err := os.ReadFile(filepath.Join(GetRootPath(), "conf/config.yml"))
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
func (config *Config) WriteConfig() error {
	content, err := yaml.Marshal(config)
	if err != nil {
		fmt.Println(err)
		return err
	}
	err = os.WriteFile(filepath.Join(GetRootPath(), "conf/config.yml"), content, 0666)
	if err != nil {
		fmt.Println(err)
	}
	return err
}

// GetRootPath 获取运行时系统的root位置，解决调试时无法使用相对位置的困扰
func GetRootPath() string {
	if RunMode == Debug {
		return "/Users/user/10post-exploit/3.golang/nemo_go"
	}
	return "."
}
