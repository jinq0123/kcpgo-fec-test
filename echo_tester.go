package main

import (
	"fmt"
	"time"
)

type EchoTester struct {
	mode Mode

	clt *EchoClient
	svr *EchoServer

	// 模拟网络
	vnet *LatencySimulator

	buffer []byte
}

func NewEchoTester(mode Mode) *EchoTester {
	e := &EchoTester{
		mode: mode,

		vnet: NewLatencySimulator(),

		buffer: make([]byte, 2000),
	}

	e.clt = NewEchoClient(mode, e.sendC2S)
	e.svr = NewEchoServer(mode, e.sendS2C)
	return e
}

func (e *EchoTester) Run() {
	start := iclock()
	for e.clt.pongCount < maxCount {
		time.Sleep(1 * time.Millisecond)
		e.tickMs()
	}

	total := iclock() - start
	mode := e.mode
	fmt.Printf("NoDelay(%d,%d,%d,%d) mode result (%dms):\n",
		mode.nodelay, mode.interval, mode.resend, mode.nc, total)
	fmt.Printf("avgrtt=%d e.maxrtt=%d\n", int(e.clt.sumrtt/uint32(e.clt.pongCount)), e.clt.maxrtt)
	e.clt.SaveRtt()
}

// tickMs ticks every ms.
func (e *EchoTester) tickMs() {
	e.recvFromVNet()
	e.clt.TickMs()
	e.svr.TickMs()
}

func (e *EchoTester) sendC2S(buf []byte, size int) {
	e.vnet.SendC2S(buf[:size])
}

func (e *EchoTester) sendS2C(buf []byte, size int) {
	e.vnet.SendS2C(buf[:size])
}

func (e *EchoTester) recvFromVNet() {
	// Recv on server side.
	for {
		hr := e.vnet.RecvOnSvrSide(e.buffer)
		if hr < 0 {
			break
		}
		// 如果 p2收到udp，则作为下层协议输入到kcp2
		e.svr.Input(e.buffer[:hr])
	}

	// Recv on client side.
	for {
		hr := e.vnet.RecvOnCltSide(e.buffer)
		if hr < 0 {
			break
		}
		// 如果 p1收到udp，则作为下层协议输入到kcp1
		e.clt.Input(e.buffer[:hr])
	}
}
