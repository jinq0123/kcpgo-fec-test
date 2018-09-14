package main

import "time"

type MsClock uint32

func iclock() MsClock {
	return MsClock((time.Now().UnixNano() / 1000000) & 0xffffffff)
}
