package utils

import (
	"crypto/tls"
	"fmt"
	"github.com/hanc00l/nemo_go/v2/pkg/conf"
	"github.com/hanc00l/nemo_go/v2/pkg/logging"
	"golang.org/x/net/proxy"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var Socks5Proxy string

// ParseHost 将http://a.b.c:80/这种url去除不相关的字符，返回主机名
func ParseHost(u string) string {
	// url.Parse必须是完整的schema://host:port格式才能解析，比如http://example.org:8080
	// url.Parse支持对ipv6完整URL的解析
	// 返回的hostname为example.org（不带端口）
	p, err := url.Parse(u)
	if err == nil && p.Host != "" {
		return p.Hostname()
	}

	if _, host, _ := ParseHostUrl(u); len(host) > 0 {
		return host
	}
	// 其它格式处理：
	host := strings.ReplaceAll(u, "https://", "")
	host = strings.ReplaceAll(host, "http://", "")
	host = strings.ReplaceAll(host, "/", "")
	host = strings.Split(host, ":")[0]

	return host
}

// ParseHostPort 将http://a.b.c:80/这种url去除不相关的字符，返回主机名,端口号
func ParseHostPort(u string) (string, int) {
	// url.Parse必须是完整的schema://host:port格式才能解析，比如http://example.org:8080
	// url.Parse支持对ipv6完整URL的解析
	// 返回的hostname为example.org（不带端口）
	p, err := url.Parse(u)
	if err == nil && p.Host != "" {
		port, _ := strconv.Atoi(p.Port())
		return p.Hostname(), port
	}

	if _, host, port := ParseHostUrl(u); len(host) > 0 && port > 0 {
		return host, port
	}
	// 其它格式处理：
	url := strings.ReplaceAll(u, "https://", "")
	url = strings.ReplaceAll(url, "http://", "")
	url = strings.ReplaceAll(url, "/", "")
	datas := strings.Split(url, ":")
	var host string
	var port int
	if len(datas) >= 1 {
		host = datas[0]
		if len(datas) >= 2 {
			port, _ = strconv.Atoi(datas[1])
		} else {
			if strings.HasPrefix(u, "https://") {
				port = 443
			} else {
				port = 80
			}
		}
	}
	return host, port
}

func CheckDomain(domain string) bool {
	domainPattern := `^(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z0-9][a-z0-9-]{0,61}[a-z0-9]$`
	reg := regexp.MustCompile(strings.TrimSpace(domainPattern))
	return reg.MatchString(domain)
}

// GetFaviconSuffixUrl 获取favicon文件的后缀名称
func GetFaviconSuffixUrl(u string) string {
	p, err := url.Parse(u)
	if err != nil {
		return ""
	}
	suffixes := strings.Split(p.Path, ".")
	if len(suffixes) < 2 {
		return ""
	}
	fileSuffix := strings.ToLower(suffixes[len(suffixes)-1])
	if !in(fileSuffix, []string{"ico", "gif", "jpg", "jpeg", "gif", "png", "bmp"}) {
		return ""
	}
	return fileSuffix
}

func in(target string, strArray []string) bool {
	for _, element := range strArray {
		if target == element {
			return true
		}
	}
	return false
}

func WrapperTCP(network, address string, timeout time.Duration) (net.Conn, error) {
	d := &net.Dialer{Timeout: timeout}
	return WrapperTCPWithSocks5(network, address, d)
}

func WrapperTCPWithSocks5(network, address string, forward *net.Dialer) (net.Conn, error) {
	if proxyServer := conf.GetProxyConfig(); proxyServer != "" {
		uri, err := url.Parse(proxyServer)
		if err == nil {
			dial, err := proxy.FromURL(uri, forward)
			if err != nil {
				return nil, err
			}
			conn, err := dial.Dial(network, address)
			if err != nil {
				return nil, err
			}
			return conn, nil
		}
	}
	return forward.Dial(network, address)
}

// GetProtocol 检测URL协议
func GetProtocol(host string, Timeout int64) (protocol string) {
	protocol = "http"
	if strings.HasSuffix(host, ":443") {
		protocol = "https"
		return
	}
	tcpConn, err := WrapperTCP("tcp", host, time.Duration(Timeout)*time.Second)
	if err != nil {
		return
	}
	conn := tls.Client(tcpConn, &tls.Config{InsecureSkipVerify: true})
	defer func() {
		if conn != nil {
			defer func() {
				if err := recover(); err != nil {
					fmt.Println(err)
				}
			}()
			conn.Close()
		}
	}()
	conn.SetDeadline(time.Now().Add(time.Duration(Timeout) * time.Second))
	err = conn.Handshake()
	if err == nil || strings.Contains(err.Error(), "handshake failure") {
		protocol = "https"
	}
	return protocol
}

// ParseHostUrl ipv4/v6地址格式识别的通用函数，用于识别ipv4,ipv4:port,ipv6,[ipv6],[ipv6]:port的形式
func ParseHostUrl(u string) (isIpv6 bool, ip string, port int) {
	// ipv4
	if CheckIPV4(u) {
		ip = u
		return
	}
	// ipv6地址 2400:dd01:103a:4041::10
	if CheckIPV6(u) {
		isIpv6 = true
		ip = u
		return
	}
	//ipv6地址：[2400:dd01:103a:4041::101]:443
	if strings.Index(u, "[") >= 0 && strings.Index(u, "]") >= 0 {
		ipv6Re := regexp.MustCompile(`\[(.*?)\]`)
		m := ipv6Re.FindStringSubmatch(u)
		if len(m) == 2 && CheckIPV6(m[1]) {
			isIpv6 = true
			ip = m[1]
			portRe := regexp.MustCompile(`\[.*?\]:(\d{1,5})`)
			n := portRe.FindStringSubmatch(u)
			if len(n) == 2 {
				port, _ = strconv.Atoi(n[1])
			}
			return
		}
	}
	// ipv4:port  192.168.1.1:443
	if strings.Index(u, ":") > 0 {
		ipp := strings.Split(u, ":")
		if len(ipp) == 2 {
			if CheckIPV4(ipp[0]) {
				ip = ipp[0]
				port, _ = strconv.Atoi(ipp[1])
			}
			return
		}
	}
	return
}

// FormatHostUrl 将ipv4/v6生成url格式，ipv6生成url时，必须增加[]
func FormatHostUrl(protocol, host string, port int) string {
	var h string
	if CheckIPV6(host) {
		h = fmt.Sprintf("[%s]", host)
	} else {
		h = host
	}
	var hostPort string
	if port > 0 {
		hostPort = fmt.Sprintf("%s:%d", h, port)
	} else {
		hostPort = h
	}
	if len(protocol) > 0 {
		return fmt.Sprintf("%s://%s", protocol, hostPort)
	} else {
		return hostPort
	}
}

// GetProxyHttpClient 获取代理的http client
func GetProxyHttpClient(isProxy bool) *http.Client {
	var transport *http.Transport
	transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	if isProxy {
		if proxy := conf.GetProxyConfig(); proxy != "" {
			proxyURL, parseErr := url.Parse(proxy)
			if parseErr != nil {
				logging.RuntimeLog.Warningf("proxy config fail:%v,skip proxy!", parseErr)
				logging.CLILog.Warningf("proxy config fail:%v,skip proxy!", parseErr)
			} else {
				transport = &http.Transport{
					Proxy:           http.ProxyURL(proxyURL),
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				}
			}
		} else {
			logging.RuntimeLog.Warning("get proxy config fail or disabled by worker,skip proxy!")
			logging.CLILog.Warning("get proxy config fail or disabled by worker,skip proxy!")
		}
	}
	httpClient := &http.Client{
		Transport: transport,
		Timeout:   3 * time.Second,
	}
	return httpClient
}

// FindDomain 从字符串中提取域名
func FindDomain(content string) []string {
	domainPattern := `(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z0-9][a-z0-9-]{0,61}[a-z0-9]`
	r := regexp.MustCompile(strings.TrimSpace(domainPattern))

	allResult := r.FindAllString(content, -1)
	//去重：
	ips := make(map[string]struct{})
	for _, ip := range allResult {
		if _, existed := ips[ip]; !existed {
			ips[ip] = struct{}{}
		}
	}

	return SetToSlice(ips)
}
