package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"sort"
)

type EchoClient struct {
	*EchoPeer

	nextPingTime MsClock
	pingIndex    uint32
	pongCount    uint32

	sumrtt uint32
	maxrtt uint32

	modeName string
	aRtt     []int
}

func NewEchoClient(mode Mode, output OutputCallback) *EchoClient {
	return &EchoClient{
		EchoPeer: NewEchoPeer(mode, output),
		modeName: mode.name,
		aRtt:     make([]int, maxCount),
	}
}

func (e *EchoClient) TickMs() {
	e.kcp.Update()

	// Ping repeatedly
	e.tryToPing()

	e.recvPong()
}

func (e *EchoClient) tryToPing() {
	current := iclock()
	if 0 == e.nextPingTime {
		e.nextPingTime = current
	}

	for ; current >= e.nextPingTime; e.nextPingTime += pingIntervalMs {
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
	for e.pongCount < maxCount {
		buffer := make([]byte, 10)
		hr := int32(e.kcp.Recv(buffer))
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
		e.aRtt[e.pongCount] = int(rtt)

		if sn != e.pongCount {
			// 如果收到的包不连续
			println("ERROR sn ", sn, "<-> count ", e.pongCount)
			return
		}

		e.sumrtt += rtt
		e.pongCount++
		if rtt > e.maxrtt {
			e.maxrtt = rtt
		}

		//println("[RECV] mode=", mode, " sn=", sn, " rtt=", rtt)
	}
}

// SaveRtt save aRtt[] to file.
func (e *EchoClient) SaveRtt() {
	saveToFile(e.aRtt, outputDir+"/"+e.modeName+".txt")

	a := make([]int, maxCount)
	copy(a, e.aRtt)
	sort.Ints(a)
	saveToFile(a, outputDir+"/"+e.modeName+"_sorted.txt")
}

func saveToFile(a []int, fileName string) {
	file, err := os.Create(fileName)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	for _, v := range a {
		file.WriteString(fmt.Sprintf("%d\n", v))
	}
}
