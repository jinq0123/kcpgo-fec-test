package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/xtaci/kcp-go"
)

type EchoTester struct {
	mode Mode

	// KCP ARQ protocol
	kcpClt *kcp.KCP // Client side KCP
	kcpSvr *kcp.KCP // Server side KCP

	// FEC codec
	fec *FecCodec

	// 模拟网络
	vnet *LatencySimulator

	nextPingTime MsClock
	pingIndex    uint32
	buffer       []byte
	pingCount    uint32

	sumrtt uint32
	maxrtt uint32
}

func NewEchoTester(mode Mode) *EchoTester {
	e := &EchoTester{
		mode: mode,
		// 创建模拟网络：丢包率10%，Rtt 60ms~125ms
		vnet: NewLatencySimulator(10, 60, 125),

		buffer: make([]byte, 2000),
	}

	// 创建两个端点的 kcp对象，第一个参数 conv是会话编号，同一个会话需要相同
	// 最后一个是 user参数，用来传递标识
	e.kcpClt = kcp.NewKCP(0x11223344, e.sendC2S)
	e.kcpSvr = kcp.NewKCP(0x11223344, e.sendS2C)

	return e
}

func (e *EchoTester) Run() {
	current := iclock()
	e.nextPingTime = current + 20

	// 配置窗口大小：平均延迟200ms，每20ms发送一个包，
	// 而考虑到丢包重发，设置最大收发窗口为128
	e.kcpClt.WndSize(1280, 1280)
	e.kcpSvr.WndSize(1280, 1280)
	// nodelay, interval, resend, nc int
	mode := e.mode
	e.kcpClt.NoDelay(mode.nodelay, mode.interval, mode.resend, mode.nc)
	e.kcpSvr.NoDelay(mode.nodelay, mode.interval, mode.resend, mode.nc)

	ts1 := iclock()

	for e.pingCount < 500 {
		time.Sleep(1 * time.Millisecond)
		e.tickMs()
	}

	ts1 = iclock() - ts1
	fmt.Printf("NoDelay(%d,%d,%d,%d) mode result (%dms):\n",
		mode.nodelay, mode.interval, mode.resend, mode.nc, ts1)
	fmt.Printf("avgrtt=%d e.maxrtt=%d\n", int(e.sumrtt/uint32(e.pingCount)), e.maxrtt)
}

// tickMs ticks every ms.
func (e *EchoTester) tickMs() {
	current := iclock()
	e.kcpClt.Update()
	e.kcpSvr.Update()

	// 每隔 20ms，kcp1发送数据
	for ; current >= e.nextPingTime; e.nextPingTime += 20 {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.LittleEndian, uint32(e.pingIndex))
		e.pingIndex++
		binary.Write(buf, binary.LittleEndian, uint32(current))
		// 发送上层协议包
		e.kcpClt.Send(buf.Bytes())
		//println("now", iclock())
	}

	// 处理虚拟网络：检测是否有udp包从p1->p2
	for {
		hr := e.vnet.recv(1, e.buffer, 2000)
		if hr < 0 {
			break
		}
		// 如果 p2收到udp，则作为下层协议输入到kcp2
		e.kcpSvr.Input(e.buffer[:hr], true, false)
	}

	// 处理虚拟网络：检测是否有udp包从p2->p1
	for {
		hr := e.vnet.recv(0, e.buffer, 2000)
		if hr < 0 {
			break
		}
		// 如果 p1收到udp，则作为下层协议输入到kcp1
		e.kcpClt.Input(e.buffer[:hr], true, false)
		//println("@@@@", hr, r)
	}

	// kcp2接收到任何包都返回回去
	for {
		hr := int32(e.kcpSvr.Recv(e.buffer[:10]))
		// 没有收到包就退出
		if hr < 0 {
			break
		}
		// 如果收到包就回射
		buf := bytes.NewReader(e.buffer)
		var sn uint32
		binary.Read(buf, binary.LittleEndian, &sn)
		e.kcpSvr.Send(e.buffer[:hr])
	}

	// kcp1收到kcp2的回射数据
	for {
		hr := int32(e.kcpClt.Recv(e.buffer[:10]))
		buf := bytes.NewReader(e.buffer)
		// 没有收到包就退出
		if hr < 0 {
			break
		}
		var sn uint32
		var ts, rtt uint32
		binary.Read(buf, binary.LittleEndian, &sn)
		binary.Read(buf, binary.LittleEndian, &ts)
		rtt = uint32(current) - ts

		if sn != e.pingCount {
			// 如果收到的包不连续
			println("ERROR sn ", sn, "<-> count ", e.pingCount)
			return
		}

		e.sumrtt += rtt
		e.pingCount++
		if rtt > e.maxrtt {
			e.maxrtt = rtt
		}

		//println("[RECV] mode=", mode, " sn=", sn, " rtt=", rtt)
	}
}

func (e *EchoTester) sendC2S(buf []byte, size int) {
	e.vnet.send(0, buf, size)
}

func (e *EchoTester) sendS2C(buf []byte, size int) {
	e.vnet.send(1, buf, size)
}
