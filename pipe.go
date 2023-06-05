package pipe

import (
	"net"
)

var (
	//默认写入缓冲区大小
	DefaultInputBufferSize int32 = 1024
	//默认读取缓冲区大小
	DefaultOutputBufferSize int32 = 1024
	//默认开启跨权读取
	DefaultIsCrossAuthority bool = true
)

//管道配置对象
type PipeConfig struct {
	//windows平台创建命名管道的配置： PIPE_TYPE_BYTE字节流写入管道  PIPE_TYPE_MESSAGE消息流写入管道
	//是否消息流写入管道 默认为字节流写入管道
	MessageMode bool
	//写入缓冲区大小
	InputBufferSize int32
	//读取缓冲区大小
	OutputBufferSize int32
	//是否开启跨权读取 默认为跨权读取
	IsCrossAuthority bool
}

// pipeAddr 命名管道的地址
type PipeAddr string

//Network返回地址的网络名称“管道”。
func (a PipeAddr) Network() string { return "pipe" }

//字符串返回管道的地址
func (a PipeAddr) String() string {
	return string(a)
}

//开启命名管道监听
func Listen(address PipeAddr, c *PipeConfig) (net.Listener, error) {
	if c == nil {
		c = &PipeConfig{
			InputBufferSize:  DefaultInputBufferSize,
			OutputBufferSize: DefaultOutputBufferSize,
			IsCrossAuthority: DefaultIsCrossAuthority,
		}
	}
	return newPipeListener(address, c)
}

//链接命名管道
func Dial(address PipeAddr) (net.Conn, error) {
	return newPipeConn(address)
}
