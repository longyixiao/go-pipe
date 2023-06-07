//go:build darwin || linux
// +build darwin linux

package pipe

import (
	"net"
	"time"
)

func newPipeConn(address PipeAddr) (*pipeConn, error) {
	conn, err := net.DialUnix("unix", nil, &net.UnixAddr{Name: address.String(), Net: "unix"})
	if err != nil {
		return nil, err
	}
	return &pipeConn{
		conn: conn,
	}, nil
}

// PipeConn是命名管道连接的net.Conne接口的实现。
type pipeConn struct {
	addr PipeAddr

	conn *net.UnixConn
}

// 读取内容上锁
func (c *pipeConn) Read(b []byte) (int, error) {
	return c.conn.Read(b)
}

// 写入文件上锁
func (c *pipeConn) Write(b []byte) (int, error) {
	return c.conn.Write(b)
}

func (c *pipeConn) Close() (err error) {
	return c.conn.Close()
}
func (c *pipeConn) LocalAddr() net.Addr {
	return c.addr

}
func (c *pipeConn) RemoteAddr() net.Addr {
	return c.addr

}
func (c *pipeConn) SetDeadline(t time.Time) error {
	c.SetReadDeadline(t)
	c.SetWriteDeadline(t)
	return nil
}
func (c *pipeConn) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}
func (c *pipeConn) SetWriteDeadline(t time.Time) error {
	return c.conn.SetWriteDeadline(t)
}
