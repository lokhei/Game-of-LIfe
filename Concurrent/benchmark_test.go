package main

import (
	"os"
	"strconv"
	"strings"
	"testing"

	"uk.ac.bris.cs/gameoflife/gol"
)

const turnNum = 100

func Benchmark(b *testing.B) {
	tests := gol.Params{
		ImageWidth: 16, ImageHeight: 16,
	}

	os.Stdout = nil
	for j := 1; j <= 16; j++ {
		tests.Threads = j
		name := strings.Join([]string{strconv.Itoa(tests.ImageWidth), strconv.Itoa(tests.ImageHeight), strconv.Itoa(turnNum)}, "x")

		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				gol.Run(tests, nil, nil)
			}
		})
	}
}
