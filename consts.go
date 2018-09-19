package main

const (
	outputDir = "output" // to save rtt data

	// ping test
	maxCount       = 500
	pingIntervalMs = 20

	// 模拟网络：丢包率10%，Rtt 60ms~125ms
	rtLossRate = 10  // round-trip loss rate, 0..100
	rttMin     = 60  // min round-trip time in Ms
	rttMax     = 125 // max round-trip time in Ms

	// Fec
	dataShards   = 2
	parityShards = 1
)
