package main

import (
	"os"
	"strconv"
	"strings"
	"testing"

	"uk.ac.bris.cs/gameoflife/gol"
)

func Benchmark(b *testing.B) {
	os.Stdout = nil // Disable all program output apart from benchmark results
	tests := []gol.Params{
		{ImageWidth: 16, ImageHeight: 16, Turns: 10000000000000},
		{ImageWidth: 64, ImageHeight: 64, Turns: 10000000000000},
		{ImageWidth: 512, ImageHeight: 512, Turns: 10000000000000},
	}
	for _, test := range tests {
		test.Threads = 1

		name := strings.Join([]string{strconv.Itoa(test.ImageWidth), strconv.Itoa(test.ImageHeight), strconv.Itoa(test.Turns), strconv.Itoa(test.Threads)}, "x")
		b.Run(name, func(b *testing.B) {

			events := make(chan gol.Event)
			gol.Run(test, events, nil)
		})
		test.Threads = 16
		name = strings.Join([]string{strconv.Itoa(test.ImageWidth), strconv.Itoa(test.ImageHeight), strconv.Itoa(test.Turns), strconv.Itoa(test.Threads)}, "x")
		b.Run(name, func(b *testing.B) {

			events := make(chan gol.Event)
			gol.Run(test, events, nil)
		})
	}
}
