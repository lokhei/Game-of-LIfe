package main

import (
	"os"
	"testing"

	"uk.ac.bris.cs/gameoflife/gol"
)

// Benchmark applies the filter to the ship.png b.N times.
// The time taken is carefully measured by go.
// The b.N  repetition is needed because benchmark results are not always constant.
// func Benchmark(b *testing.B) {
// 	os.Stdout = nil // Disable all program output apart from benchmark results
// 	b.Run("Gol.run benchmark", func(b *testing.B) {
// 		p := gol.Params{
// 			p.Threads := 1
// 			p.Turns := 10
// 			p.ImageHeight := 16
// 			p.ImageWidth := 16
// 		}
// 		keyPresses := make(chan rune, 10)
// 		events := make(chan gol.Event, 1000)
// 		for i := 0; i < b.N; i++ {
// 			p.Threads = 10
// 			gol.Run(p, events, keyPresses)
// 		}
// 	})
// }

func Benchmark(b *testing.B) {
	os.Stdout = nil // Disable all program output apart from benchmark results
	tests := []gol.Params{
		{ImageWidth: 16, ImageHeight: 16},
		{ImageWidth: 64, ImageHeight: 64},
		{ImageWidth: 512, ImageHeight: 512},
	}
	b.Run("benchmark", func(b *testing.B) {
		// for i := 0; i < b.N; i++ {
		for _, p := range tests {
			for threads := 1; threads <= 16; threads++ {
				p.Threads = threads
				events := make(chan gol.Event)
				gol.Run(p, events, nil)
			}
		}
	})
}
