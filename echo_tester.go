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
	kcpC2S *kcp.KCP // Client -> Server
	kcpS2C *kcp.KCP // Server -> Client

	// FEC codec
	fecDecoder *fecDecoder
	fecEncoder *fecEncoder

	// 模拟网络
	vnet *LatencySimulator
}

func NewEchoTester(mode Mode) *EchoTester {
	e := &EchoTester{
		mode: mode,
		// 创建模拟网络：丢包率10%，Rtt 60ms~125ms
		vnet: NewLatencySimulator(10, 60, 125),
	}

	// 创建两个端点的 kcp对象，第一个参数 conv是会话编号，同一个会话需要相同
	// 最后一个是 user参数，用来传递标识
	e.kcpC2S = kcp.NewKCP(0x11223344, e.sendC2S)
	e.kcpS2C = kcp.NewKCP(0x11223344, e.sendS2C)

	return e
}

func (e *EchoTester) Run() {
	current := uint32(iclock())
	slap := current + 20
	index := 0
	next := 0
	var sumrtt uint32
	count := 0
	maxrtt := 0

	// 配置窗口大小：平均延迟200ms，每20ms发送一个包，
	// 而考虑到丢包重发，设置最大收发窗口为128
	e.kcpC2S.WndSize(1280, 1280)
	e.kcpS2C.WndSize(1280, 1280)
	// nodelay, interval, resend, nc int
	mode := e.mode
	e.kcpC2S.NoDelay(mode.nodelay, mode.interval, mode.resend, mode.nc)
	e.kcpS2C.NoDelay(mode.nodelay, mode.interval, mode.resend, mode.nc)

	buffer := make([]byte, 2000)
	var hr int32

	ts1 := iclock()

	for {
		time.Sleep(1 * time.Millisecond)
		current = uint32(iclock())
		e.kcpC2S.Update()
		e.kcpS2C.Update()

		// 每隔 20ms，kcp1发送数据
		for ; current >= slap; slap += 20 {
			buf := new(bytes.Buffer)
			binary.Write(buf, binary.LittleEndian, uint32(index))
			index++
			binary.Write(buf, binary.LittleEndian, uint32(current))
			// 发送上层协议包
			e.kcpC2S.Send(buf.Bytes())
			//println("now", iclock())
		}

		// 处理虚拟网络：检测是否有udp包从p1->p2
		for {
			hr = e.vnet.recv(1, buffer, 2000)
			if hr < 0 {
				break
			}
			// 如果 p2收到udp，则作为下层协议输入到kcp2
			e.kcpS2C.Input(buffer[:hr], true, false)
		}

		// 处理虚拟网络：检测是否有udp包从p2->p1
		for {
			hr = e.vnet.recv(0, buffer, 2000)
			if hr < 0 {
				break
			}
			// 如果 p1收到udp，则作为下层协议输入到kcp1
			e.kcpC2S.Input(buffer[:hr], true, false)
			//println("@@@@", hr, r)
		}

		// kcp2接收到任何包都返回回去
		for {
			hr = int32(e.kcpS2C.Recv(buffer[:10]))
			// 没有收到包就退出
			if hr < 0 {
				break
			}
			// 如果收到包就回射
			buf := bytes.NewReader(buffer)
			var sn uint32
			binary.Read(buf, binary.LittleEndian, &sn)
			e.kcpS2C.Send(buffer[:hr])
		}

		// kcp1收到kcp2的回射数据
		for {
			hr = int32(e.kcpC2S.Recv(buffer[:10]))
			buf := bytes.NewReader(buffer)
			// 没有收到包就退出
			if hr < 0 {
				break
			}
			var sn uint32
			var ts, rtt uint32
			binary.Read(buf, binary.LittleEndian, &sn)
			binary.Read(buf, binary.LittleEndian, &ts)
			rtt = uint32(current) - ts

			if sn != uint32(next) {
				// 如果收到的包不连续
				//for i:=0;i<8 ;i++ {
				//println("---", i, buffer[i])
				//}
				println("ERROR sn ", count, "<->", next, sn)
				return
			}

			next++
			sumrtt += rtt
			count++
			if rtt > uint32(maxrtt) {
				maxrtt = int(rtt)
			}

			//println("[RECV] mode=", mode, " sn=", sn, " rtt=", rtt)
		}

		if next > 500 {
			break
		}
	}

	ts1 = iclock() - ts1

	fmt.Printf("NoDelay(%d,%d,%d,%d) mode result (%dms):\n",
		mode.nodelay, mode.interval, mode.resend, mode.nc, ts1)
	fmt.Printf("avgrtt=%d maxrtt=%d\n", int(sumrtt/uint32(count)), maxrtt)
}

func (e *EchoTester) sendC2S(buf []byte, size int) {
	e.vnet.send(0, buf, size)
}

func (e *EchoTester) sendS2C(buf []byte, size int) {
	e.vnet.send(1, buf, size)
}
