package main

import (
	"container/list"
	"math/rand"
)

type LatencySimulator struct {
	lostrate       int
	rttmin, rttmax int
	c2s            *list.List // DelayTunnel
	s2c            *list.List // DelayTunnel
}

// lostrate: 单向丢包率，百分比
// rttmin：rtt最小值
// rttmax：rtt最大值
func NewLatencySimulator(lostrate, rttmin, rttmax int) *LatencySimulator {
	p := &LatencySimulator{}

	p.c2s = list.New()
	p.s2c = list.New()

	p.lostrate = lostrate
	if rttmin > rttmax {
		rttmin, rttmax = rttmax, rttmin
	}
	p.rttmin = rttmin
	p.rttmax = rttmax

	return p
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
	if rand.Intn(100) < p.lostrate {
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
		delay += rand.Int() % (p.rttmax - p.rttmin)
	}
	return MsClock(delay / 2)
}
