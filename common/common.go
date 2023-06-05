package common

import "fmt"

var ErrClosed = PipeError{"Pipe has been closed.", false}

type PipeError struct {
	msg     string
	timeout bool
}

func (e PipeError) Error() string {
	return e.msg
}

func (e PipeError) Timeout() bool {
	return e.timeout
}

func (e PipeError) Temporary() bool {
	return false
}

func BadAddr(addr string) PipeError {
	return PipeError{fmt.Sprintf("Invalid pipe address '%s'.", addr), false}
}
func TimeOut(addr string) PipeError {
	return PipeError{fmt.Sprintf("Pipe IO timed out waiting for '%s'", addr), true}
}
