package socks5forward

import (
	"context"
	"fmt"
	"github.com/hanc00l/nemo_go/v3/pkg/conf"
	"github.com/hanc00l/nemo_go/v3/pkg/logging"
	"github.com/things-go/go-socks5"
	"github.com/things-go/go-socks5/bufferpool"
	"github.com/things-go/go-socks5/statute"
	"golang.org/x/net/proxy"
	"io"
	"net"
	"net/url"
	"strings"
	"time"
)

type TCPDial func(ctx context.Context, net_, addr string) (net.Conn, error)
type UserConnectHandle func(ctx context.Context, writer io.Writer, request *socks5.Request) error
type closeWriter interface {
	CloseWrite() error
}

var bufferPool = bufferpool.NewPool(32 * 1024)

// DefaultTCPDial is the default dial function used by the server
func DefaultTCPDial() TCPDial {
	return func(ctx context.Context, net_, addr string) (net.Conn, error) {
		return net.Dial(net_, addr)
	}
}

// UpstreamSocks5TCPDial 转发到上游socks5代理
func UpstreamSocks5TCPDial(proxyServer string) TCPDial {
	uri, err := url.Parse(proxyServer)
	if err == nil {
		return func(ctx context.Context, net_, addr string) (net.Conn, error) {
			forward := &net.Dialer{Timeout: time.Second * 2}
			dial, err := proxy.FromURL(uri, forward)
			if err != nil {
				return nil, err
			}
			conn, err := dial.Dial(net_, addr)
			if err != nil {
				return nil, err
			}
			return conn, nil
		}
	} else {
		logging.RuntimeLog.Warning("获取代理配置失败或禁用了代理功能，代理被跳过")
		logging.CLILog.Warning("get proxy config fail or disabled by worker,skip proxy!")
		return DefaultTCPDial()
	}
}

// handleConnect is used to handle a connect command
func handleConnect(ctx context.Context, writer io.Writer, request *socks5.Request) error {
	// Attempt to connect
	var dial TCPDial
	// 获取配置文件中的代理配置
	if proxy := conf.GetProxyConfig(); proxy != "" {
		dial = UpstreamSocks5TCPDial(proxy)
	} else {
		dial = DefaultTCPDial()
	}
	target, err := dial(ctx, "tcp", request.DestAddr.String())
	if err != nil {
		msg := err.Error()
		resp := statute.RepHostUnreachable
		if strings.Contains(msg, "refused") {
			resp = statute.RepConnectionRefused
		} else if strings.Contains(msg, "network is unreachable") {
			resp = statute.RepNetworkUnreachable
		}
		if err := socks5.SendReply(writer, resp, nil); err != nil {
			return fmt.Errorf("failed to send reply, %v", err)
		}
		return fmt.Errorf("connect to %v failed, %v", request.RawDestAddr, err)
	}
	defer target.Close()

	// Send success
	if err = socks5.SendReply(writer, statute.RepSuccess, target.LocalAddr()); err != nil {
		return fmt.Errorf("failed to send reply, %v", err)
	}

	// Start proxying
	errCh := make(chan error, 2)
	go func() { errCh <- Proxy(target, request.Reader) }()
	go func() { errCh <- Proxy(writer, target) }()
	// Wait
	for i := 0; i < 2; i++ {
		e := <-errCh
		if e != nil {
			// return from this function closes target (and conn).
			return e
		}
	}
	return nil
}

func Proxy(dst io.Writer, src io.Reader) error {
	buf := bufferPool.Get()
	defer bufferPool.Put(buf)
	_, err := io.CopyBuffer(dst, src, buf[:cap(buf)])
	if tcpConn, ok := dst.(closeWriter); ok {
		tcpConn.CloseWrite() //nolint: errcheck
	}
	return err
}

// StartSocks5Forward 启动socks5转发
func StartSocks5Forward(forwardSocks5Addr string, listenFail chan struct{}) {
	// Create a SOCKS5 server
	server := socks5.NewServer(
		socks5.WithLogger(logging.CLILog),
		socks5.WithBufferPool(bufferPool),
		socks5.WithConnectHandle(handleConnect))

	logging.CLILog.Infof("start socks5 forward server: %v", forwardSocks5Addr)
	if err := server.ListenAndServe("tcp", forwardSocks5Addr); err != nil {
		listenFail <- struct{}{}

		logging.RuntimeLog.Warningf("socks5 forward server: %v", err)
		logging.CLILog.Warningf("socks5 forward server: %v", err)
	}
}
