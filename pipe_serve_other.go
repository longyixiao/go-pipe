//go:build darwin || linux
// +build darwin linux

package pipe

import (
	"net"
	"os"
)

func createPipe(address PipeAddr, config *PipeConfig, first bool) (*net.UnixListener, error) {
	// 创建新命名管道
	listener, err := net.ListenUnix("unix", &net.UnixAddr{Name: address.String(), Net: "unix"})
	if err != nil {
		return nil, err
	}
	//是否支持跨权访问
	if config.IsCrossAuthority {
		// 设置命名管道的权限
		if err := os.Chmod(address.String(), 0666); err != nil {
			return nil, err
		}
	}

	return listener, nil
}

func newPipeListener(address PipeAddr, c *PipeConfig) (*pipeListener, error) {
	os.Remove(address.String())
	listener, err := createPipe(address, c, true)
	if err != nil {
		return nil, err
	}
	return &pipeListener{
		addr:     address,
		listener: listener,
	}, nil
}

// pipeListener 是一个命名的管道监听器
type pipeListener struct {
	addr PipeAddr

	listener *net.UnixListener
}

// 等待客户端连接
func (l *pipeListener) Accept() (net.Conn, error) {
	return l.listener.Accept()
}

// 关闭监听
func (l *pipeListener) Close() error {
	return l.listener.Close()
}

// 返回监听地址
func (l *pipeListener) Addr() net.Addr { return l.addr }
