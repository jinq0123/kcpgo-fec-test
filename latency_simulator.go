package main

import (
	"container/list"
	"math/rand"
)

type DelayTunnel struct{ *list.List }

type LatencySimulator struct {
	current                  int32
	lostrate, rttmin, rttmax int
	p12                      DelayTunnel
	p21                      DelayTunnel
	r12                      *rand.Rand
	r21                      *rand.Rand
}

// lostrate: 往返一周丢包率的百分比，默认 10%
// rttmin：rtt最小值，默认 60
// rttmax：rtt最大值，默认 125
func NewLatencySimulator(lostrate, rttmin, rttmax int) *LatencySimulator {
	p := &LatencySimulator{}

	p.r12 = rand.New(rand.NewSource(9))
	p.r21 = rand.New(rand.NewSource(99))
	p.p12 = DelayTunnel{list.New()}
	p.p21 = DelayTunnel{list.New()}
	p.current = iclock()
	p.lostrate = lostrate / 2 // 上面数据是往返丢包率，单程除以2
	p.rttmin = rttmin / 2
	p.rttmax = rttmax / 2

	return p
}

// 发送数据
// peer - 端点0/1，从0发送，从1接收；从1发送从0接收
func (p *LatencySimulator) send(peer int, data []byte, size int) int {
	rnd := 0
	if peer == 0 {
		rnd = p.r12.Intn(100)
	} else {
		rnd = p.r21.Intn(100)
	}
	//println("!!!!!!!!!!!!!!!!!!!!", rnd, p.lostrate, peer)
	if rnd < p.lostrate {
		return 0
	}
	pkt := NewDelayPacket(data[:size])
	p.current = iclock()
	delay := p.rttmin
	if p.rttmax > p.rttmin {
		delay += rand.Int() % (p.rttmax - p.rttmin)
	}
	pkt.ts = p.current + int32(delay)
	if peer == 0 {
		p.p12.PushBack(pkt)
	} else {
		p.p21.PushBack(pkt)
	}
	return 1
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
	p.current = iclock()
	if p.current < pkt.ts {
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
