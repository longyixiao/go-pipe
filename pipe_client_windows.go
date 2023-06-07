package pipe

import (
	"io"
	"net"
	"syscall"
	"time"

	"github.com/longyixiao/go-pipe/common"
	"golang.org/x/sys/windows"
)

var (
	nmpwait_wait_forever uint32 = 0xFFFFFFFF
)

// isPipeNotReady checks the error to see if it indicates the pipe is not ready
func isPipeNotReady(err error) bool {

	return err == syscall.ERROR_FILE_NOT_FOUND || err == error_pipe_busy
}

func dial(address PipeAddr, timeout uint32) (*pipeConn, error) {
	if err := common.WaitNamedPipe(address.StringToUTF16Ptr(), timeout); err != nil {
		if err == error_bad_pathname {
			return nil, common.BadAddr(address.String())
		}
		return nil, err
	}
	handle, err := windows.CreateFile(address.StringToUTF16Ptr(), syscall.GENERIC_READ|syscall.GENERIC_WRITE,
		uint32(syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE), nil, syscall.OPEN_EXISTING,
		syscall.FILE_FLAG_OVERLAPPED, 0)
	if err != nil {
		return nil, err
	}
	return &pipeConn{handle: handle, addr: address}, nil
}

func newPipeConn(address PipeAddr) (*pipeConn, error) {
	for {

		conn, err := dial(address, nmpwait_wait_forever)
		if err == nil {
			return conn, nil
		}
		if isPipeNotReady(err) {
			<-time.After(100 * time.Millisecond)
			continue
		}
		return nil, err
	}
}

// PipeConn是命名管道连接的net.Conne接口的实现。
type pipeConn struct {
	handle windows.Handle
	addr   PipeAddr

	// these aren't actually used yet
	readDeadline  *time.Time
	writeDeadline *time.Time
}
type iodata struct {
	n   uint32
	err error
}

// 完成请求
func (c *pipeConn) completeRequest(data iodata, deadline *time.Time, overlapped *windows.Overlapped) (int, error) {
	if data.err == error_io_incomplete || data.err == syscall.ERROR_IO_PENDING {
		var timer <-chan time.Time
		if deadline != nil {
			if timeDiff := deadline.Sub(time.Now()); timeDiff > 0 {
				timer = time.After(timeDiff)
			}
		}
		done := make(chan iodata)
		go func() {
			n, err := waitForCompletion(c.handle, overlapped)
			done <- iodata{n, err}
		}()
		select {
		case data = <-done:
		case <-timer:
			windows.CancelIoEx(c.handle, overlapped)
			data = iodata{0, common.TimeOut(c.addr.String())}
		}
	}
	if data.err == syscall.ERROR_BROKEN_PIPE {
		data.err = io.EOF
	}
	return int(data.n), data.err
}
func (c *pipeConn) Read(b []byte) (int, error) {
	overlapped, err := newOverlapped()
	if err != nil {
		return 0, err
	}
	defer windows.CloseHandle(overlapped.HEvent)
	var n uint32
	err = windows.ReadFile(c.handle, b, &n, overlapped)
	return c.completeRequest(iodata{n, err}, c.readDeadline, overlapped)
}
func (c *pipeConn) Write(b []byte) (int, error) {
	overlapped, err := newOverlapped()
	if err != nil {
		return 0, err
	}
	defer windows.CloseHandle(overlapped.HEvent)
	var n uint32
	err = windows.WriteFile(c.handle, b, &n, overlapped)
	return c.completeRequest(iodata{n, err}, c.writeDeadline, overlapped)
}

func (c *pipeConn) Close() (err error) {
	return windows.CloseHandle(c.handle)
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
	c.readDeadline = &t
	return nil
}
func (c *pipeConn) SetWriteDeadline(t time.Time) error {
	c.writeDeadline = &t
	return nil
}
