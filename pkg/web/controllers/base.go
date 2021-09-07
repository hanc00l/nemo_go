package controllers

import (
	"bytes"
	"github.com/beego/beego/v2/server/web"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"net/http"
	"strings"
	"sync"
	"time"
)

type BaseController struct {
	web.Controller
}

const (
	Success = "success"
	Fail    = "fail"
)

// StatusResponseData JSON的状态响应
type StatusResponseData struct {
	Status string `json:"status"`
	Msg    string `json:"msg"`
}

// DatableRequestParam DataTable列表的请求数据的公共部份
type DatableRequestParam struct {
	Draw   int `form:"draw"`
	Start  int `form:"start"`
	Length int `form:"length"`
}

// GlobalSessionData 用于保存查询过程中的一些参数
type GlobalSessionData struct {
	IpAddressIp     string `json:"ip_address_ip"`
	IpAddressDomain string `json:"ip_address_domain"`
	DomainAddress   string `json:"domain_address"`
	Port            string `json:"port"`
	OrgId           string `json:"session_org_id"`
}

// DataTableResponseData DataTable列表的返回数据
type DataTableResponseData struct {
	Draw            int           `json:"draw"`
	RecordsTotal    int           `json:"recordsTotal"`
	RecordsFiltered int           `json:"recordsFiltered"`
	Data            []interface{} `json:"data"`
}

// OnlineUserInfo 在线用户
type OnlineUserInfo struct {
	IP           string
	LoginTime    time.Time
	UpdateTime   time.Time
	UpdateNumber int64
}

// OnlineUserMutex 在线用户Mutex
var OnlineUserMutex sync.Mutex
// OnlineUser 在线用户信息
var OnlineUser = make(map[string]*OnlineUserInfo)

func (c *BaseController) GetGlobalSessionData() GlobalSessionData {
	data := GlobalSessionData{
		IpAddressIp:     c.getSessionData("ip_address_ip", ""),
		IpAddressDomain: c.getSessionData("ip_address_domain", ""),
		DomainAddress:   c.getSessionData("domain_address", ""),
		Port:            c.getSessionData("port", ""),
		OrgId:           c.getSessionData("session_org_id", ""),
	}
	return data
}

// FailedStatus 返回fail结果的JSON格式消息
func (c *BaseController) FailedStatus(msg string) {
	c.Data["json"] = StatusResponseData{Status: Fail, Msg: msg}
}

// SucceededStatus 返回success结果的JSON格式消息
func (c *BaseController) SucceededStatus(msg string) {
	c.Data["json"] = StatusResponseData{Status: Success, Msg: msg}
}

// MakeStatusResponse 生成默认的状态返回JSON
func (c *BaseController) MakeStatusResponse(isSuccess bool) {
	if isSuccess {
		c.SucceededStatus("")
	} else {
		c.FailedStatus("")
	}
}

// setSessionData 设置一个session数据
func (c *BaseController) setSessionData(key, value string) {
	err := c.SetSession(key, value)
	if err != nil {
		logging.RuntimeLog.Error(err)
	}
}

// getSessionData 读取一个session数据
func (c *BaseController) getSessionData(key, defaultValue string) string {
	v := c.GetSession(key)
	if v == nil {
		return defaultValue
	}
	return v.(string)
}

// writeByteContent 输出byte的返回数据
func (c *BaseController) writeByteContent(out []byte) {
	rw := c.Ctx.ResponseWriter
	rw.Header().Set("Content-Type", "application/octet-stream")
	rw.WriteHeader(http.StatusOK)
	http.ServeContent(rw, c.Ctx.Request, "", time.Now(), bytes.NewReader(out))
}

// FormatDateTime 日期统一格式化
func FormatDateTime(dt time.Time) string {
	return dt.Format("2006-01-02 15:04:05")
}

// UpdateOnlineUser 更新在线用户IP和时间
func (c *BaseController) UpdateOnlineUser() {
	ip := strings.Split(c.Ctx.Request.RemoteAddr, ":")[0]

	OnlineUserMutex.Lock()
	defer OnlineUserMutex.Unlock()

	if _, ok := OnlineUser[ip]; ok {
		OnlineUser[ip].UpdateTime = time.Now()
		OnlineUser[ip].UpdateNumber++
	} else {
		OnlineUser[ip] = &OnlineUserInfo{
			IP:           ip,
			LoginTime:    time.Now(),
			UpdateTime:   time.Now(),
			UpdateNumber: 1,
		}
	}
}

// DeleteOnlineUser 用户注销或过期，清除在线用户
func (c *BaseController) DeleteOnlineUser() {
	ip := strings.Split(c.Ctx.Request.RemoteAddr, ":")[0]
	OnlineUserMutex.Lock()
	defer OnlineUserMutex.Unlock()

	if _, ok := OnlineUser[ip]; ok {
		delete(OnlineUser, ip)
	}
}
