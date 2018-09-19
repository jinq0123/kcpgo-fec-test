# kcpgo-fec-test

Test [kcp-go](https://github.com/xtaci/kcp-go) FEC

Derived from kcp-go kcp_test.go adding FEC.

To plot:
```
gnuplot -p -e "plot 'normal_sorted.txt' with lines, 'normal-fec_sorted.txt' with lines, 'fast_sorted.txt' with lines, 'fast-fec_sorted.txt' with lines"
```
Or:
```
kcpgo-fec-test && cd output && gnuplot -p -e "plot 'normal_sorted.txt' with lines, 'normal-fec_sorted.txt' with lines, 'fast_sorted.txt' with lines, 'fast-fec_sorted.txt' with lines" && cd ..
```
