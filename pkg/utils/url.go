package utils

import (
	"crypto/tls"
	"net"
	"net/url"
	"regexp"
	"strings"
	"time"
)

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

	conn, err := tls.DialWithDialer(&net.Dialer{Timeout: time.Duration(Timeout) * time.Second}, "tcp", host, &tls.Config{InsecureSkipVerify: true})
	defer func() {
		if conn != nil {
			conn.Close()
		}
	}()

	if err == nil || strings.Contains(err.Error(), "handshake failure") {
		protocol = "https"
	}
	return protocol
}
