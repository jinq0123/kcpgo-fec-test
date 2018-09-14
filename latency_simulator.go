package main

import (
	"container/list"
	"math/rand"
)

type LatencySimulator struct {
	lostrate, rttmin, rttmax int
	p12                      *list.List // DelayTunnel
	p21                      *list.List // DelayTunnel
}

// lostrate: 往返一周丢包率的百分比，默认 10%
// rttmin：rtt最小值，默认 60
// rttmax：rtt最大值，默认 125
func NewLatencySimulator(lostrate, rttmin, rttmax int) *LatencySimulator {
	p := &LatencySimulator{}

	p.p12 = list.New()
	p.p21 = list.New()
	p.lostrate = lostrate / 2 // 上面数据是往返丢包率，单程除以2
	p.rttmin = rttmin / 2
	p.rttmax = rttmax / 2

	return p
}

func (p *LatencySimulator) SendC2S(data []byte) {
	p.send(0, data, len(data))
}

func (p *LatencySimulator) SendS2C(data []byte) {
	p.send(1, data, len(data))
}

// 发送数据
// peer - 端点0/1，从0发送，从1接收；从1发送从0接收
func (p *LatencySimulator) send(peer int, data []byte, size int) int {
	if rand.Intn(100) < p.lostrate {
		return 0
	}
	pkt := NewDelayPacket(data[:size])
	delay := p.rttmin
	if p.rttmax > p.rttmin {
		delay += rand.Int() % (p.rttmax - p.rttmin)
	}
	pkt.ts = iclock() + MsClock(delay)
	if peer == 0 {
		p.p12.PushBack(pkt)
	} else {
		p.p21.PushBack(pkt)
	}
	return 1
}

func (p *LatencySimulator) RecvOnCltSide(data []byte) int32 {
	return p.recv(0, data, len(data))
}

func (p *LatencySimulator) RecvOnSvrSide(data []byte) int32 {
	return p.recv(1, data, len(data))
}

// 接收数据
func (p *LatencySimulator) recv(peer int, data []byte, maxsize int) int32 {
	var it *list.Element
	if peer == 0 {
		it = p.p21.Front()
		if p.p21.Len() == 0 {
			return -1
		}
	} else {
		it = p.p12.Front()
		if p.p12.Len() == 0 {
			return -1
		}
	}
	pkt := it.Value.(*DelayPacket)
	if iclock() < pkt.ts {
		return -2
	}
	if maxsize < pkt.size() {
		return -3
	}
	if peer == 0 {
		p.p21.Remove(it)
	} else {
		p.p12.Remove(it)
	}
	maxsize = pkt.size()
	copy(data, pkt.data[:maxsize])
	return int32(maxsize)
}
