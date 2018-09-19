package main

import (
	"fmt"
	"time"
)

type EchoTester struct {
	mode Mode

	clt *EchoClient
	svr *EchoServer

	// virtual net
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

// Run sends ping and calculates the round-trip time.
func (e *EchoTester) Run() {
	start := iclock()
	for e.clt.pongCount < maxCount {
		time.Sleep(1 * time.Millisecond)
		e.tickMs()
	}

	total := iclock() - start
	mode := e.mode
	var fec string
	if mode.fec {
		fec = ",fec"
	}
	fmt.Printf("%s mode(%d,%d,%d,%d%s) result (%dms):\n",
		mode.name, mode.nodelay, mode.interval, mode.resend, mode.nc, fec, total)
	fmt.Printf("avgrtt=%d e.maxrtt=%d\n", int(e.clt.sumrtt/uint32(e.clt.pongCount)), e.clt.maxrtt)
	e.printFecRecoveredCount()
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
		// Input to server's kcp/fec.
		e.svr.Input(e.buffer[:hr])
	}

	// Recv on client side.
	for {
		hr := e.vnet.RecvOnCltSide(e.buffer)
		if hr < 0 {
			break
		}
		// Input to client's kcp/fec.
		e.clt.Input(e.buffer[:hr])
	}
}

func (e *EchoTester) printFecRecoveredCount() {
	if !e.mode.fec {
		return
	}
	fmt.Printf("FEC recovered: server=%d, client=%d\n", e.svr.fecRecovered, e.clt.fecRecovered)
}
