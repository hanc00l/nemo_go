package core

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/hanc00l/nemo_go/v3/pkg/conf"
	"github.com/hanc00l/nemo_go/v3/pkg/logging"
	"github.com/hanc00l/nemo_go/v3/pkg/utils"
	"github.com/redis/go-redis/v9"
	"io"
	"net"
	"strings"
	"time"
)

var redisOptions *redis.Options

func getRedisConfig() {
	redisConfig := conf.GlobalServerConfig().Redis
	redisOptions = &redis.Options{
		Addr:     fmt.Sprintf("%s:%d", redisConfig.Host, redisConfig.Port),
		Password: redisConfig.Password,
		DB:       0,
	}
	//redisOptions.Password = redisConfig.Password
}

type RedisLock struct {
	key        string
	value      string
	expiration time.Duration
	redisCli   *redis.Client
	ctx        context.Context
}

type RedisReverseServer struct {
	RedisAddr  string
	ListenAddr string
	AuthPass   string
}
type RedisProxyServer struct {
	ReverseServerAddr string
	LocalListenAddr   string
	AuthPass          string
}

func GetRedisClient() (*redis.Client, error) {
	if redisOptions == nil {
		getRedisConfig()
	}
	rdb := redis.NewClient(redisOptions)
	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}
	return rdb, nil
}

func CloseRedisClient(client *redis.Client) error {
	return client.Close()
}

func NewRedisLock(key string, expiration time.Duration, redisCli *redis.Client) *RedisLock {
	return &RedisLock{
		key:        key,
		value:      uuid.New().String(),
		expiration: expiration,
		redisCli:   redisCli,
		ctx:        context.Background(),
	}
}

func NewRedisReverseServer(redisAddr string, listenAddr string, authPass string) *RedisReverseServer {
	return &RedisReverseServer{
		RedisAddr:  redisAddr,
		ListenAddr: listenAddr,
		AuthPass:   authPass,
	}
}

func NewRedisProxyServer(reverseServerAddr string, localListenAddr string, authPass string) *RedisProxyServer {
	return &RedisProxyServer{
		ReverseServerAddr: reverseServerAddr,
		LocalListenAddr:   localListenAddr,
		AuthPass:          authPass,
	}
}

func (rl *RedisLock) TryLock() (bool, error) {
	luaScript := `
    if redis.call("setNx", KEYS[1], ARGV[1]) then
        if redis.call("expire", KEYS[1], ARGV[2]) then
            return 1
        else
            redis.call("del", KEYS[1])
            return 0
        end
    end
    return 0
    `
	result, err := rl.redisCli.Eval(rl.ctx, luaScript, []string{rl.key}, rl.value, int(rl.expiration.Seconds())).Result()
	if err != nil {
		return false, err
	}
	if result == int64(1) {
		return true, nil
	}
	return false, nil
}

func (rl *RedisLock) Unlock() error {
	script := `
    if redis.call("get", KEYS[1]) == ARGV[1] then
        return redis.call("del", KEYS[1])
    else
        return 0
    end
    `
	_, err := rl.redisCli.Eval(rl.ctx, script, []string{rl.key}, rl.value).Result()
	return err
}

// 密码验证函数
func (rr *RedisReverseServer) authenticate(salt, password string) bool {
	authPass := utils.MD5(rr.AuthPass + salt)
	return password == authPass
}

// 处理TCP连接
func (rr *RedisReverseServer) handleConnection(localConn net.Conn) {
	defer localConn.Close()
	// 发送salt
	salt := utils.GetRandomString2(16)
	localConn.Write([]byte(salt + "\n"))

	// 读取客户端发送的密码
	reader := bufio.NewReader(localConn)
	password, err := reader.ReadString('\n')
	if err != nil {
		logging.CLILog.Errorf("Failed to read password: %v", err)
		return
	}
	// 验证密码
	if !rr.authenticate(salt, strings.TrimSpace(password)) {
		logging.CLILog.Errorf("Failed to authenticate")
		localConn.Write([]byte("-ERR invalid password\n"))
		return
	}

	// 密码验证通过，连接到内网Redis服务器
	redisConn, err := net.Dial("tcp", rr.RedisAddr)
	if err != nil {
		logging.CLILog.Errorf("Failed to connect to Redis server: %v", err)
		return
	}
	//logging.CLILog.Infof("Connected to Redis server")
	defer redisConn.Close()

	// 发送成功响应给客户端
	localConn.Write([]byte("+OK\n"))

	// 数据转发
	go io.Copy(redisConn, localConn)
	io.Copy(localConn, redisConn)
}

func (rr *RedisReverseServer) Start() {
	// 边界代理监听地址
	listener, err := net.Listen("tcp", rr.ListenAddr)
	if err != nil {
		logging.RuntimeLog.Errorf("Failed to listen on %s:%v", rr.ListenAddr, err)
		return
	}
	defer listener.Close()
	logging.RuntimeLog.Infof("Redis reverse proxy server listening on %s", rr.ListenAddr)
	logging.CLILog.Infof("Redis reverse proxy server listening on %s", rr.ListenAddr)

	for {
		localConn, err := listener.Accept()
		if err != nil {
			logging.CLILog.Infof("Failed to accept connection: %v", err)
			continue
		}

		// 处理每个连接
		go rr.handleConnection(localConn)
	}
}

// 密码验证函数
func (rp *RedisProxyServer) authenticate(salt string) string {
	authPass := utils.MD5(rp.AuthPass + salt)
	return authPass
}

// 处理TCP连接
func (rp *RedisProxyServer) handleConnection(localConn net.Conn) {
	defer localConn.Close()
	// 连接到边界上的反向代理服务器
	reverseProxyConn, err := net.Dial("tcp", rp.ReverseServerAddr)
	if err != nil {
		logging.CLILog.Errorf("Failed to connect to reverse proxy %s: %v", rp.ReverseServerAddr, err)
		return
	}
	//logging.CLILog.Infof("Connected to reverse proxy")

	defer reverseProxyConn.Close()
	// 读取salt
	reader := bufio.NewReader(reverseProxyConn)
	salt, err := reader.ReadString('\n')
	if err != nil {
		logging.CLILog.Errorf("Failed to read password: %v", err)
		return
	}
	// 发送密码给边界上的反向代理服务器
	authPass := rp.authenticate(strings.TrimSpace(salt))
	reverseProxyConn.Write([]byte(authPass + "\n"))
	// 读取验证响应
	msg, err := reader.ReadString('\n')
	if err != nil {
		logging.CLILog.Errorf("Failed to read auth response: %v", err)
		return
	}
	if msg != "+OK\n" {
		logging.CLILog.Errorf("Invalid auth response: %s", msg)
		return
	}
	logging.CLILog.Infof("Successfully authenticated with reverse proxy")
	//logging.CLILog.Infof("Starting data forwarding")
	// 数据转发
	go io.Copy(reverseProxyConn, localConn)
	io.Copy(localConn, reverseProxyConn)
}

func (rp *RedisProxyServer) Start() {
	// 客户端代理监听地址
	listener, err := net.Listen("tcp", rp.LocalListenAddr)
	if err != nil {
		logging.CLILog.Errorf("Failed to listen on %v: %v", rp.LocalListenAddr, err)
		return
	}
	defer listener.Close()
	logging.RuntimeLog.Infof("Redis client proxy server listening on %s", rp.LocalListenAddr)
	logging.CLILog.Infof("Redis client proxy server listening on %s", rp.LocalListenAddr)

	for {
		localConn, err := listener.Accept()
		if err != nil {
			logging.CLILog.Errorf("Failed to accept connection: %v", err)
			continue
		}

		// 处理每个连接
		go rp.handleConnection(localConn)
	}
}

func storeTimeToRedis(client *redis.Client, key string, t time.Time) error {
	// 使用json.Marshal序列化时间
	jsonBytes, err := json.Marshal(t)
	if err != nil {
		return err
	}
	_, err = client.Set(context.Background(), key, string(jsonBytes), 0).Result()
	return err
}

func getTimeFromRedis(client *redis.Client, key string) (time.Time, error) {
	result, err := client.Get(context.Background(), key).Result()
	if err != nil {
		return time.Time{}, err
	}
	var t time.Time
	// 将JSON格式的字符串反序列化回time.Time对象
	err = json.Unmarshal([]byte(result), &t)
	return t, err
}

// SetWorkerStatusToRedis 将 WorkerAliveStatus 存储到 Redis 中
func SetWorkerStatusToRedis(client *redis.Client, workerID string, status *WorkerStatus) error {
	//for workerID, status := range WorkerAliveStatus {
	// 将 WorkerStatus 序列化为 JSON
	jsonData, err := json.Marshal(status)
	if err != nil {
		return fmt.Errorf("序列化 WorkerStatus 失败: %v", err)
	}

	// 将 JSON 数据存储到 Redis 的哈希结构中
	err = client.HSet(context.Background(), "worker_alive_status", workerID, jsonData).Err()
	if err != nil {
		return fmt.Errorf("存储到 Redis 失败: %v", err)
	}
	//}
	return nil
}

// GetWorkerStatusFromRedis 从 Redis 中读取指定 workerID 的 WorkerStatus
func GetWorkerStatusFromRedis(client *redis.Client, workerID string) (*WorkerStatus, error) {
	// 从 Redis 的哈希结构中获取指定 workerID 的值
	jsonData, err := client.HGet(context.Background(), "worker_alive_status", workerID).Result()
	if errors.Is(err, redis.Nil) {
		return nil, err
	} else if err != nil {
		return nil, fmt.Errorf("从 Redis 读取 workerID '%s' 失败: %v", workerID, err)
	}

	// 反序列化 JSON 数据为 WorkerStatus
	var status WorkerStatus
	err = json.Unmarshal([]byte(jsonData), &status)
	if err != nil {
		return nil, fmt.Errorf("反序列化 WorkerStatus 失败: %v", err)
	}

	return &status, nil
}

// DeleteWorkerStatusFromRedis 从 Redis 中删除指定 workerID 的记录
func DeleteWorkerStatusFromRedis(client *redis.Client, workerID string) error {
	// 使用 HDEL 方法删除哈希结构中的指定字段
	result, err := client.HDel(context.Background(), "worker_alive_status", workerID).Result()
	if err != nil {
		return fmt.Errorf("从 Redis 删除 workerID '%s' 失败: %v", workerID, err)
	}

	// 如果返回的删除记录数为 0，说明键不存在
	if result == 0 {
		return fmt.Errorf("workerID '%s' 不存在", workerID)
	}
	//
	//fmt.Printf("成功删除 workerID '%s' 的记录\n", workerID)
	return nil
}

// LoadWorkerStatusFromRedis 从 Redis 中读取 WorkerAliveStatus
func LoadWorkerStatusFromRedis(client *redis.Client) (map[string]*WorkerStatus, error) {
	workerAliveStatus := make(map[string]*WorkerStatus)
	// 从 Redis 中获取所有 worker 的状态
	result, err := client.HGetAll(context.Background(), "worker_alive_status").Result()
	if err != nil {
		return nil, fmt.Errorf("从 Redis 读取失败: %v", err)
	}
	// 遍历结果并反序列化
	for workerID, jsonData := range result {
		var status WorkerStatus
		err := json.Unmarshal([]byte(jsonData), &status)
		if err != nil {
			return nil, fmt.Errorf("反序列化 WorkerStatus 失败: %v", err)
		}
		workerAliveStatus[workerID] = &status
	}
	return workerAliveStatus, nil
}
