@echo Run kcpgo-fec-test.exe... It takes 10s.
kcpgo-fec-test.exe
cd output
gnuplot -p -e "plot 'normal_sorted.txt' with lines, 'normal-fec_sorted.txt' with lines, 'fast_sorted.txt' with lines, 'fast-fec_sorted.txt' with lines"
cd ..
pause
