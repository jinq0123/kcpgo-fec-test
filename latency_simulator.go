package main

import (
	"container/list"
	"math/rand"
)

type LatencySimulator struct {
	lossRate float32 // one-way loss rate, [0.0..0.5)
	rttmin   MsClock
	rttmax   MsClock
	c2s      *list.List // DelayTunnel
	s2c      *list.List // DelayTunnel
}

// rtLossRate: round-trip loss rate, 往返一周的丢包率，百分比 0..100
// rttmin：rtt最小值
// rttmax：rtt最大值
func NewLatencySimulator(rtLossRate int, rttmin, rttmax MsClock) *LatencySimulator {
	if rttmin > rttmax {
		rttmin, rttmax = rttmax, rttmin
	}
	return &LatencySimulator{
		c2s: list.New(),
		s2c: list.New(),

		lossRate: float32(rtLossRate) / 100.0 / 2.0,
		rttmin:   rttmin,
		rttmax:   rttmax,
	}
}

func (p *LatencySimulator) SendC2S(data []byte) {
	c2s := true
	p.send(c2s, data)
}

func (p *LatencySimulator) SendS2C(data []byte) {
	c2s := true
	p.send(!c2s, data)
}

// 发送数据
func (p *LatencySimulator) send(c2s bool, data []byte) {
	if rand.Float32() < p.lossRate {
		return // packet lost
	}

	pkt := NewDelayPacket(data)
	pkt.ts = iclock() + p.getRandDelay()
	if c2s {
		p.c2s.PushBack(pkt)
	} else {
		p.s2c.PushBack(pkt)
	}
}

func (p *LatencySimulator) RecvOnCltSide(data []byte) int {
	cltSide := true
	return p.recv(cltSide, data)
}

func (p *LatencySimulator) RecvOnSvrSide(data []byte) int {
	cltSide := true
	return p.recv(!cltSide, data)
}

// 接收数据
func (p *LatencySimulator) recv(isCltSide bool, data []byte) int {
	var it *list.Element
	lst := p.c2s
	if isCltSide {
		lst = p.s2c // receive on the client side
	}
	if lst.Len() == 0 {
		return -1
	}

	it = lst.Front()
	pkt := it.Value.(*DelayPacket)
	if iclock() < pkt.ts {
		return -2
	}
	if len(data) < pkt.size() {
		return -3
	}

	lst.Remove(it)
	copy(data, pkt.data)
	return len(pkt.data)
}

func (p *LatencySimulator) getRandDelay() MsClock {
	delay := p.rttmin
	if p.rttmax != p.rttmin {
		delay += MsClock(rand.Int()) % (p.rttmax - p.rttmin)
	}
	return delay / 2
}
