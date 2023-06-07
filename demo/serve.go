package main

import (
	"bufio"
	"fmt"
	"net"
	"runtime"

	"github.com/longyixiao/go-pipe"
)

var pipeFile = "\\\\.\\pipe\\mypipe"

// 初始化函数
func init() {
	if runtime.GOOS != "windows" {
		pipeFile = "/tmp/mypipe"
	}
}

func main() {
	fmt.Println("开始管道监听")
	ln, err := pipe.Listen(pipe.PipeAddr(pipeFile), nil)
	if err != nil {
		fmt.Println("pipe.Listen err:", err)
		return
	}
	defer ln.Close()
	for {
		conn, err := ln.Accept()
		if err != nil {
			// handle error
			continue
		}

		// handle connection like any other net.Conn
		go func(conn net.Conn) {
			for {
				r := bufio.NewReader(conn)
				msg, err := r.ReadString('\n')
				if err != nil {
					// handle error
					return
				}
				recvMsg := fmt.Sprintf("server recv msg: %s", msg)
				fmt.Println(recvMsg)
				fmt.Fprintln(conn, msg)
			}
		}(conn)
	}
}
