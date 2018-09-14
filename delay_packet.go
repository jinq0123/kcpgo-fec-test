package main

type DelayPacket struct {
	data []byte
	ts   MsClock // timestamp in ms
}

func NewDelayPacket(src []byte) *DelayPacket {
	data := make([]byte, len(src))
	copy(data, src)
	return &DelayPacket{
		data: data,
	}
}

func (d *DelayPacket) size() int {
	return len(d.data)
}
