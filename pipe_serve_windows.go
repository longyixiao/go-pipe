package pipe

import (
	"errors"
	"log"
	"net"
	"sync"
	"syscall"
	"unsafe"

	"github.com/longyixiao/go-pipe/common"
	"golang.org/x/sys/windows"
)

var (
	error_no_data        syscall.Errno = 0xE8
	error_pipe_connected syscall.Errno = 0x217
	error_pipe_busy      syscall.Errno = 0xE7
	error_sem_timeout    syscall.Errno = 0x79

	error_bad_pathname syscall.Errno = 0xA1
	error_invalid_name syscall.Errno = 0x7B

	error_io_incomplete syscall.Errno = 0x3e4
)

// 将字符串转成UTF16的指针
func (a PipeAddr) StringToUTF16Ptr() *uint16 {
	return windows.StringToUTF16Ptr(a.String())
}

func (a PipeAddr) UTF16PtrFromString() (*uint16, error) {
	return windows.UTF16PtrFromString(a.String())
}

// 创建异步io对象
func newOverlapped() (*windows.Overlapped, error) {
	event, err := windows.CreateEvent(nil, 1, 1, nil)
	if err != nil {
		return nil, err
	}
	return &windows.Overlapped{HEvent: event}, nil
}

// 创建命名管道
func createPipe(address PipeAddr, config *PipeConfig, first bool) (windows.Handle, error) {
	var sa *windows.SecurityAttributes = nil
	//是否支持跨权读取
	if config.IsCrossAuthority {
		// 创建一个安全描述符对象
		secDesc, err := windows.NewSecurityDescriptor()
		if err != nil {
			log.Println("Error initializing security descriptor:", err)
			return 0, err
		}
		err = secDesc.SetDACL(nil, true, false)
		if err != nil {
			return 0, err
		}
		sa = &windows.SecurityAttributes{
			Length:             uint32(unsafe.Sizeof(windows.SecurityAttributes{})),
			SecurityDescriptor: secDesc,
			InheritHandle:      0,
		}
	}
	//PIPE_ACCESS_DUPLEX:双向管道,服务器和客户端进程都可以从管道读取和写入管道
	//FILE_FLAG_OVERLAPPED:启用重叠模式。 如果启用此模式，执行读取、写入和连接操作的函数可能会立即返回，这些操作可能需要很长时间才能完成。
	//此模式允许启动操作的线程在后台执行耗时操作时执行其他操作。
	//例如，在重叠模式下，线程可以处理管道的多个实例上的同时输入和输出 (I/O) 操作，或在同一管道句柄上同时执行读取和写入操作
	mode := uint32(windows.PIPE_ACCESS_DUPLEX | windows.FILE_FLAG_OVERLAPPED)

	//如果尝试使用此标志创建管道的多个实例，则创建第一个实例会成功，但创建下一个实例会失败并 ERROR_ACCESS_DENIED。
	//第二次创建不使用此标志
	if first {
		mode |= windows.FILE_FLAG_FIRST_PIPE_INSTANCE
	}

	pipeMode := uint32(windows.PIPE_TYPE_BYTE)
	//管道模式 字节流和消息流
	if config.MessageMode {
		pipeMode = windows.PIPE_TYPE_MESSAGE
	}

	//pipeMode |= windows.PIPE_WAIT
	// 创建命名管道
	pipe, err := windows.CreateNamedPipe(
		address.StringToUTF16Ptr(),
		mode,
		pipeMode,
		windows.PIPE_UNLIMITED_INSTANCES,
		uint32(config.OutputBufferSize),
		uint32(config.InputBufferSize),
		0,
		sa,
	)
	if err != nil {
		//返回错误信息
		return 0, err
	}

	return pipe, nil
}

// 等待异步IO完成
func waitForCompletion(handle windows.Handle, overlapped *windows.Overlapped) (uint32, error) {
	_, err := windows.WaitForSingleObject(overlapped.HEvent, windows.INFINITE)
	if err != nil {
		return 0, err
	}
	var transferred uint32
	err = windows.GetOverlappedResult(windows.Handle(handle), overlapped, &transferred, true)
	return transferred, err
}

func newPipeListener(address PipeAddr, c *PipeConfig) (*pipeListener, error) {
	h, err := createPipe(address, c, true)
	if err != nil {
		return nil, err
	}
	return &pipeListener{
		addr:   address,
		config: *c,
		handle: h,
	}, nil
}

// pipeListener 是一个命名的管道监听器
type pipeListener struct {
	mu sync.Mutex

	addr   PipeAddr
	handle windows.Handle
	config PipeConfig
	closed bool

	acceptHandle     windows.Handle
	acceptOverlapped *windows.Overlapped
}

func (l *pipeListener) acceptPipe() (*pipeConn, error) {
	if l == nil {
		return nil, errors.New("parameter error")
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.addr == "" || l.closed {
		return nil, errors.New("listening has been turned off or parameter error")
	}
	handle := l.handle
	if handle == 0 {
		var err error
		handle, err = createPipe(l.addr, &l.config, false)
		if err != nil {
			return nil, err
		}
	} else {
		l.handle = 0
	}

	overlapped, err := newOverlapped()
	if err != nil {
		return nil, err
	}
	defer windows.CloseHandle(overlapped.HEvent)
	err = windows.ConnectNamedPipe(handle, overlapped)
	if err == nil || err == error_pipe_connected {
		return &pipeConn{handle: handle, addr: l.addr}, nil
	}
	if err == error_io_incomplete || err == windows.ERROR_IO_PENDING {
		l.acceptOverlapped = overlapped
		l.acceptHandle = handle
		l.mu.Unlock()
		defer func() {
			l.mu.Lock()
			l.acceptOverlapped = nil
			l.acceptHandle = 0
		}()

		_, err = waitForCompletion(handle, overlapped)
	}

	if err == syscall.ERROR_OPERATION_ABORTED {
		return nil, common.ErrClosed
	}
	if err != nil {
		return nil, err
	}
	return &pipeConn{handle: handle, addr: l.addr}, nil
}

// Accept在net.Listener接口中实现了Accept方法；它
// 等待下一个调用并返回一个通用的net.Conn。
func (l *pipeListener) Accept() (net.Conn, error) {
	c, err := l.acceptPipe()
	for err == error_no_data {
		//忽略连接并立即断开连接的客户端。
		c, err = l.acceptPipe()
	}
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (l *pipeListener) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.closed {
		return nil
	}
	l.closed = true
	if l.handle != 0 {
		err := common.DisconnectNamedPipe(l.handle)
		if err != nil {
			return err
		}
		err = windows.CloseHandle(l.handle)
		if err != nil {
			return err
		}
		l.handle = 0
	}
	if l.acceptOverlapped != nil && l.acceptHandle != 0 {
		//取消挂起的IO。此调用不会阻塞，因此可以安全地保留上面的互斥对象。
		if err := windows.CancelIoEx(l.acceptHandle, l.acceptOverlapped); err != nil {
			return err
		}
		err := windows.CloseHandle(l.acceptOverlapped.HEvent)
		if err != nil {
			return err
		}
		l.acceptOverlapped.HEvent = 0
		err = windows.CloseHandle(l.acceptHandle)
		if err != nil {
			return err
		}
		l.acceptHandle = 0
	}
	return nil
}

func (l *pipeListener) Addr() net.Addr { return l.addr }
