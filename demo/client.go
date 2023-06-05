package main

import (
	"bufio"
	"fmt"
	"runtime"
	"time"

	"github.com/longyixiao/go-pipe"
)

var pipeFile = "\\\\.\\pipe\\mypipe"

// 初始化函数
func init() {
	if runtime.GOOS != "windows" {
		pipeFile = "/tmp/mypipe"
	}
}

func ExampleDial(i int) {
	conn, err := pipe.Dial(pipe.PipeAddr(pipeFile))
	if err != nil {
		fmt.Println(err)
		return
	}
	go func() {
		for {

			r := bufio.NewReader(conn)
			msg, err := r.ReadString('\n')
			if err != nil {
				// handle eror
				return
			}
			recvMsg := fmt.Sprintf("client [%d] recv server msg: %s", i, msg)
			fmt.Println(recvMsg)

		}
	}()

	go func() {
		for {
			sendMsg := fmt.Sprintf("hi server, i'm client %d", i)
			if _, err := fmt.Fprintln(conn, sendMsg); err != nil {
				// handle error
				return
			}
			time.Sleep(1 * time.Second)
		}
	}()
}

func main() {
	fmt.Println("开始连接管道")
	for i := 0; i < 10; i += 1 {
		go ExampleDial(i)
	}
	select {}
}
