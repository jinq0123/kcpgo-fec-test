package main

import (
	"encoding/binary"

	kcp "github.com/xtaci/kcp-go"
)

// Common part of EchoServer and EchoClient
type EchoPeer struct {
	// KCP ARQ protocol
	kcp *kcp.KCP

	// FEC codec
	fec *FecCodec

	outputToNet OutputCallback
}

// output_callback is a prototype for NewKCP()
type OutputCallback = func(buf []byte, size int)

func NewEchoPeer(mode Mode, output OutputCallback) *EchoPeer {
	e := &EchoPeer{
		outputToNet: output,
	}

	// 创建两个端点的 kcp对象，第一个参数 conv是会话编号，同一个会话需要相同
	// 最后一个是输出函数
	e.kcp = kcp.NewKCP(0x11223344, e.output)
	// 配置窗口大小：平均延迟200ms，每20ms发送一个包，
	// 而考虑到丢包重发，设置最大收发窗口为128
	e.kcp.WndSize(1280, 1280)
	// nodelay, interval, resend, nc int
	e.kcp.NoDelay(mode.nodelay, mode.interval, mode.resend, mode.nc)

	if mode.fec {
		e.fec = NewFecCodec()
	}
	return e
}

// Input input data from net to fec.
func (e *EchoPeer) Input(buf []byte) {
	if e.fec == nil {
		e.kcp.Input(buf, true, false)
		return
	}

	f := e.fec.dec.decodeBytes(buf)
	if f.flag == typeData {
		e.kcp.Input(buf[fecHeaderSizePlus2:], true, false)
		return
	}
	if f.flag != typeFEC {
		return
	}

	recovers := e.fec.dec.decode(f)
	for _, r := range recovers {
		if len(r) < 2 { // must be larger than 2bytes
			continue
		}

		sz := binary.LittleEndian.Uint16(r)
		if int(sz) > len(r) || sz < 2 {
			continue
		}

		e.kcp.Input(r[2:sz], false, false)
	} // for
}

// output is the KCP output callback.
func (e *EchoPeer) output(buf []byte, size int) {
	if e.fec == nil {
		e.outputToNet(buf, size)
		return
	}

	// header extended output buffer
	ext := make([]byte, fecHeaderSizePlus2+size)
	copy(ext[fecHeaderSizePlus2:], buf)

	// FEC encoding
	ecc := e.fec.enc.encode(ext)

	// output data
	e.outputToNet(ext, len(ext))

	// output fec
	for _, v := range ecc {
		e.outputToNet(v, len(v))
	}
}
