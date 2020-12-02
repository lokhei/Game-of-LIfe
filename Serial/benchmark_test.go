package main

import (
	"os"
	"strconv"
	"strings"
	"testing"

	"uk.ac.bris.cs/gameoflife/gol"
)

func BenchmarkThreads(b *testing.B) {

	os.Stdout = nil
	test := gol.Params{
		ImageWidth:  512,
		ImageHeight: 512,
		Turns:       100,
	}

	for thread := 1; thread <= 10; thread++ {
		test.Threads = thread
		name := strings.Join([]string{strconv.Itoa(test.ImageHeight), strconv.Itoa(test.ImageWidth), strconv.Itoa(test.Turns), strconv.Itoa(thread)}, "x")

		b.Run(name, func(b *testing.B) {
			events := make(chan gol.Event)
			gol.Run(test, events, nil)
		})

	}
}
