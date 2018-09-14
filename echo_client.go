package main

import (
	"bytes"
	"encoding/binary"
)

type EchoClient struct {
	*EchoPeer

	nextPingTime MsClock
	pingIndex    uint32
	pingCount    uint32

	sumrtt uint32
	maxrtt uint32
}

func NewEchoClient(mode Mode, output OutputCallback) *EchoClient {
	return &EchoClient{
		EchoPeer: NewEchoPeer(mode, output),
	}
}

func (e *EchoClient) TickMs() {
	e.kcp.Update()

	// 每隔 20ms，kcp1发送数据
	e.tryToPing()

	e.recvPong()
}

func (e *EchoClient) tryToPing() {
	current := iclock()
	if 0 == e.nextPingTime {
		e.nextPingTime = current
	}

	for ; current >= e.nextPingTime; e.nextPingTime += 20 {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.LittleEndian, uint32(e.pingIndex))
		e.pingIndex++
		binary.Write(buf, binary.LittleEndian, uint32(current))
		// 发送上层协议包
		e.kcp.Send(buf.Bytes())
	}
}

func (e *EchoClient) recvPong() {
	current := iclock()
	// kcpClt收到Server的回射数据
	for {
		hr := int32(e.kcp.Recv(e.buffer[:10]))
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
