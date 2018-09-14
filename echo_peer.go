package main

import kcp "github.com/xtaci/kcp-go"

// Common part of EchoServer and EchoClient
type EchoPeer struct {
	// KCP ARQ protocol
	kcp *kcp.KCP

	// FEC codec
	fec *FecCodec

	buffer []byte
}

// output_callback is a prototype for NewKCP()
type OutputCallback = func(buf []byte, size int)

func NewEchoPeer(mode Mode, output OutputCallback) *EchoPeer {
	e := &EchoPeer{
		buffer: make([]byte, 2000),
	}

	// 创建两个端点的 kcp对象，第一个参数 conv是会话编号，同一个会话需要相同
	// 最后一个是输出函数
	e.kcp = kcp.NewKCP(0x11223344, output)
	// 配置窗口大小：平均延迟200ms，每20ms发送一个包，
	// 而考虑到丢包重发，设置最大收发窗口为128
	e.kcp.WndSize(1280, 1280)
	// nodelay, interval, resend, nc int
	e.kcp.NoDelay(mode.nodelay, mode.interval, mode.resend, mode.nc)
	return e
}
