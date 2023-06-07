package common

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	modkernel32 = syscall.NewLazyDLL("kernel32.dll")

	procWaitNamedPipeW      = modkernel32.NewProc("WaitNamedPipeW")
	procDisconnectNamedPipe = modkernel32.NewProc("DisconnectNamedPipe")
)

func DisconnectNamedPipe(handle windows.Handle) (err error) {

	r1, _, e1 := syscall.SyscallN(procDisconnectNamedPipe.Addr(), uintptr(handle))
	if r1 == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func WaitNamedPipe(name *uint16, timeout uint32) (err error) {
	r1, _, e1 := syscall.SyscallN(procWaitNamedPipeW.Addr(), uintptr(unsafe.Pointer(name)), uintptr(timeout))
	if r1 == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}
