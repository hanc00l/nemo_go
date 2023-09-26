package utils

import (
	"crypto/tls"
	"errors"
	"fmt"
	"golang.org/x/net/proxy"
	"net"
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

func WrapperTcpWithTimeout(network, address string, timeout time.Duration) (net.Conn, error) {
	d := &net.Dialer{Timeout: timeout}
	return WrapperTCP(network, address, d)
}

func WrapperTCP(network, address string, forward *net.Dialer) (net.Conn, error) {
	//get conn
	var conn net.Conn
	if Socks5Proxy == "" {
		var err error
		conn, err = forward.Dial(network, address)
		if err != nil {
			return nil, err
		}
	} else {
		dailer, err := Socks5Dailer(forward)
		if err != nil {
			return nil, err
		}
		conn, err = dailer.Dial(network, address)
		if err != nil {
			return nil, err
		}
	}
	return conn, nil

}

func Socks5Dailer(forward *net.Dialer) (proxy.Dialer, error) {
	u, err := url.Parse(Socks5Proxy)
	if err != nil {
		return nil, err
	}
	if strings.ToLower(u.Scheme) != "socks5" {
		return nil, errors.New("Only support socks5")
	}
	address := u.Host
	var auth proxy.Auth
	var dailer proxy.Dialer
	if u.User.String() != "" {
		auth = proxy.Auth{}
		auth.User = u.User.Username()
		password, _ := u.User.Password()
		auth.Password = password
		dailer, err = proxy.SOCKS5("tcp", address, &auth, forward)
	} else {
		dailer, err = proxy.SOCKS5("tcp", address, nil, forward)
	}

	if err != nil {
		return nil, err
	}
	return dailer, nil
}

// GetProtocol 检测URL协议
func GetProtocol(host string, Timeout int64) (protocol string) {
	protocol = "http"
	//如果端口是80或443,跳过Protocol判断
	//if strings.HasSuffix(host, ":80") || !strings.Contains(host, ":") {
	//	return
	//} else
	if strings.HasSuffix(host, ":443") {
		protocol = "https"
		return
	}

	socksconn, err := WrapperTcpWithTimeout("tcp", host, time.Duration(Timeout)*time.Second)
	if err != nil {
		return
	}
	conn := tls.Client(socksconn, &tls.Config{InsecureSkipVerify: true})
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
