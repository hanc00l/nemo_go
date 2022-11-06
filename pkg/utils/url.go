package utils

import (
	"crypto/tls"
	"errors"
	"fmt"
	"golang.org/x/net/proxy"
	"net"
	"net/url"
	"regexp"
	"strings"
	"time"
)

var Socks5Proxy string

// HostStrip 将http://a.b.c:80/这种url去除不相关的字符，返回主机名
func HostStrip(u string) string {
	p, err := url.Parse(u)
	if err == nil && p.Host != "" {
		return p.Hostname()
	}
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
	if strings.HasSuffix(host, ":80") || !strings.Contains(host, ":") {
		return
	} else if strings.HasSuffix(host, ":443") {
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
