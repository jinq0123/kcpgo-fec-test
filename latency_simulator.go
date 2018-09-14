package main

import (
	"container/list"
	"math/rand"
)

type LatencySimulator struct {
	lostrate, rttmin, rttmax int
	c2s                      *list.List // DelayTunnel
	s2c                      *list.List // DelayTunnel
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
	p.send(true, data)
}

func (p *LatencySimulator) SendS2C(data []byte) {
	p.send(false, data)
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
		it = p.s2c.Front()
		if p.s2c.Len() == 0 {
			return -1
		}
	} else {
		it = p.c2s.Front()
		if p.c2s.Len() == 0 {
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
		p.s2c.Remove(it)
	} else {
		p.c2s.Remove(it)
	}
	maxsize = pkt.size()
	copy(data, pkt.data[:maxsize])
	return int32(maxsize)
}

func (p *LatencySimulator) getRandDelay() MsClock {
	delay := p.rttmin
	if p.rttmax != p.rttmin {
		delay += rand.Int() % (p.rttmax - p.rttmin)
	}
	return MsClock(delay / 2)
}
