# go-pipe
go语言进入程序间通信库
windows平台借鉴于https://github.com/natefinch/npipe

在该模块的基础上增加了跨权管道通信

linux平台和macos平台运用标准net库中ListenUnix完成，在标准库的基础上新增了跨权管道通信

### 用法：
The Dial function connects a client to a named pipe:


	conn, err := npipe.Dial(`\\.\pipe\mypipename`)
	if err != nil {
		<handle error>
	}
	fmt.Fprintf(conn, "Hi server!\n")
	msg, err := bufio.NewReader(conn).ReadString('\n')
	...

The Listen function creates servers:


	ln, err := npipe.Listen(`\\.\pipe\mypipename`)
	if err != nil {
		// handle error
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			// handle error
			continue
		}
		go handleConnection(conn)
	}
