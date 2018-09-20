package main

import (
	"os"
	"sync"
)

var modes = []Mode{
	// Mode{"default", 0, 10, 0, 0, false},
	Mode{"normal", 0, 10, 0, 1, false},
	Mode{"fast", 1, 10, 2, 1, false},
	Mode{"normal-fec", 0, 10, 0, 1, true},
	Mode{"fast-fec", 1, 10, 2, 1, true},
}

func main() {
	os.Mkdir(outputDir, 0777)
	var wg sync.WaitGroup
	for _, mode := range modes {
		wg.Add(1)
		go func(mode Mode) {
			defer wg.Done()
			et := NewEchoTester(mode)
			et.Run()
		}(mode)
	}
	wg.Wait()
}
