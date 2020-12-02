package main

import (
	"os"
	// "strconv"
	// "strings"
	"fmt"
	"testing"

	"uk.ac.bris.cs/gameoflife/gol"
)

const turnNum = 100

func Benchmark(b *testing.B) {
	tests := gol.Params{
		ImageWidth: 16, ImageHeight: 16,
	}

	os.Stdout = nil
	for tests.Threads = 1; tests.Threads <= 16; tests.Threads++ {
		name := fmt.Sprintf("%dx%dx%d-%d", tests.ImageWidth, tests.ImageHeight, turnNum, tests.Threads)

		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				gol.Run(tests, nil, nil)
			}
		})
	}
}
