package core

import (
	"sync"
	"time"
)

const (
	EnvServiceHost = "SERVICE_HOST"
	EnvServicePort = "SERVICE_PORT"
	EnvServiceAuth = "SERVICE_AUTH"
)

type ServiceOptions struct {
	ServiceHost string `long:"service" description:"Service host" json:"service_host" form:"service_host"`
	ServicePort int    `long:"port" description:"Service port" default:"5001" json:"service_port" form:"service_port"`
	ServiceAuth string `long:"auth" description:"Service auth" json:"service_auth" form:"service_auth"`
}

type WorkerTaskOption struct {
	Concurrency       int                 `short:"c" long:"concurrency" description:"Number of concurrent workers" default:"2" json:"concurrency" form:"concurrency"`
	WorkerPerformance int                 `short:"p" long:"worker-performance" description:"worker performance,default is autodetect (0:autodetect, 1:high, 2:normal)" default:"0" json:"worker_performance" form:"worker_performance"`
	WorkerRunTaskMode string              `short:"m" long:"worker-run-task-mode" description:"worker run task mode; 0: all, 1:active, 2:finger, 3:passive, 4:pocscan, 5:custom; run multiple mode separated by \",\"" default:"0" json:"worker_run_task_mode" form:"worker_run_task_mode"`
	WorkerTopic       map[string]struct{} `json:"-"`
}

type WorkerOption struct {
	ServiceOptions   `group:"services"`
	WorkerTaskOption `group:"worker-tasks"`
	ConfigFile       string `short:"f" long:"config-file" description:"config file" json:"default_config_file" form:"default_config_file"`
	NoProxy          bool   `long:"no-proxy" description:"disable proxy configuration,include socks5 proxy and socks5forward" json:"no_proxy" form:"no_proxy"`
	NoRedisProxy     bool   `long:"no-redis-proxy" description:"disable redis proxy configuration" json:"no_redis_proxy" form:"no_redis_proxy"`
	IpV6Support      bool   `long:"ipv6" description:"support ipv6 portscan" json:"ipv6" form:"ipv6"`
}

type WebOption struct {
	TLSCertFile string `long:"tls_cert_file" description:"TLS certificate file" default:"server.crt" form:"tls_cert_file"`
	TLSKeyFile  string `long:"tls_key_file" description:"TLS key file" default:"server.key" form:"tls_key_file"`
}

type ServerOption struct {
	Web         bool `long:"web" description:"web service"`
	Cron        bool `long:"cron" description:"cron service"`
	Service     bool `long:"service" description:"rpc service"`
	RedisTunnel bool `long:"redis-tunnel" description:"redis tunnel service"`
	WebOption   `group:"web-option"`
}

type TaskResult struct {
	Status string `json:"status"`
	Msg    string `json:"msg"`
}

type WorkerStatus struct {
	sync.Mutex `json:"-"`
	// worker's task status
	WorkerName         string    `json:"worker_name"`
	WorkerTopics       string    `json:"worker_topic"`
	CreateTime         time.Time `json:"create_time"`
	UpdateTime         time.Time `json:"update_time"`
	TaskExecutedNumber int       `json:"task_number"`
	TaskStartedNumber  int       `json:"started_number"`
	// worker's run status
	ManualReloadFlag       bool   `json:"manual_reload_flag"`
	ManualFileSyncFlag     bool   `json:"manual_file_sync_flag"`
	ManualUpdateOptionFlag bool   `json:"manual_update_daemon_option"`
	CPULoad                string `json:"cpu_load"`
	MemUsed                string `json:"mem_used"`
	// worker's option
	WorkerRunOption    []byte `json:"worker_run_option"`    //worker当前运行的启动参数
	WorkerUpdateOption []byte `json:"worker_update_option"` //worker需要更新的启动参数
	// daemon option
	IsDaemonProcess        bool      `json:"is_daemon_process"`
	WorkerDaemonUpdateTime time.Time `json:"worker_daemon_update_time"`
}
