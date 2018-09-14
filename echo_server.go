package main

import (
	"bytes"
	"encoding/binary"
)

type EchoServer struct {
	*EchoPeer
}

func NewEchoServer(mode Mode, output OutputCallback) *EchoServer {
	return &EchoServer{
		EchoPeer: NewEchoPeer(mode, output),
	}
}

func (e *EchoServer) TickMs() {
	e.kcp.Update()
	e.echo()
}

func (e *EchoServer) echo() {
	// kcpSvr 接收到任何包都返回回去
	for {
		hr := int32(e.kcp.Recv(e.buffer[:10]))
		// 没有收到包就退出
		if hr < 0 {
			break
		}
		// 如果收到包就回射
		buf := bytes.NewReader(e.buffer)
		var sn uint32
		binary.Read(buf, binary.LittleEndian, &sn)
		e.kcp.Send(e.buffer[:hr])
	}
}
