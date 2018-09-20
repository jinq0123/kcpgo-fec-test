# kcpgo-fec-test

Test [kcp-go](https://github.com/xtaci/kcp-go) FEC

Derived from kcp-go kcp_test.go adding FEC.

To plot with [gnuplot](http://www.gnuplot.info/):
```
gnuplot --persist -e "plot 'normal_sorted.txt' with lines, 'fast_sorted.txt' with lines, 'normal-fec_sorted.txt' with lines, 'fast-fec_sorted.txt' with lines"
```
Or:
```
kcpgo-fec-test && cd output && gnuplot --persist -e "plot 'normal_sorted.txt' with lines, 'fast_sorted.txt' with lines, 'normal-fec_sorted.txt' with lines, 'fast-fec_sorted.txt' with lines" && cd ..
```
Or run [run_and_plot.bat](run_and_plot.bat).

Output example:
![output.png](output.png)

Add more modes in `main.go`:
```go
var modes = []Mode{
	// Mode{"default", 0, 10, 0, 0},
	Mode{"normal", 0, 10, 0, 1, false},
	Mode{"fast", 1, 10, 2, 1, false},
	Mode{"normal-fec", 0, 10, 0, 1, true},
	Mode{"fast-fec", 1, 10, 2, 1, true},
}
```

Change loss rate or rtt in `consts.go`:
```go
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
```