package main

import (
	"fmt"
	"os"
	"testing"

	"uk.ac.bris.cs/gameoflife/gol"
)

const turnNum = 1000
var value bool

func Benchmark(b *testing.B) {
	tests := []gol.Params{
		{ImageWidth: 512, ImageHeight: 512, Threads: 1},
		{ImageWidth: 512, ImageHeight: 512, Threads: 5},
		{ImageWidth: 512, ImageHeight: 512, Threads: 10},
		{ImageWidth: 512, ImageHeight: 512, Threads: 16},
	}

	for _, t := range tests {
		value = true
		os.Stdout = nil
		t.Turns = turnNum
		name := fmt.Sprintf("%dx%dx%d-%d", t.ImageWidth, t.ImageHeight, turnNum, t.Threads)

		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if value == true {
					events := make(chan gol.Event)
					gol.Run(t, events, nil)

					value = false

					for event := range events {
						switch e := event.(type) {
						case gol.FinalTurnComplete:
							fmt.Println("type ", e)
							value = true

						}
					}
				}
			}
		})
	}
}

// package main

// import (
// 	"fmt"
// 	"testing"

// 	"uk.ac.bris.cs/gameoflife/gol"
// )

// // TestGol tests 16x16, 64x64 and 512x512 images on 0, 1 and 100 turns using 1-16 worker threads.
// func BenchmarkTests(b *testing.B) {
// 	tests := []gol.Params{
// 		{ImageWidth: 16, ImageHeight: 16},
// 		{ImageWidth: 64, ImageHeight: 64},
// 		{ImageWidth: 512, ImageHeight: 512},
// 	}
// 	for _, p := range tests {
// 		for _, turns := range []int{0, 1, 100} {
// 			p.Turns = turns
// 			for threads := 1; threads <= 16; threads++ {
// 				p.Threads = threads
// 				testName := fmt.Sprintf("%dx%dx%d-%d", p.ImageWidth, p.ImageHeight, p.Turns, p.Threads)
// 				b.Run(testName, func(b *testing.B) {
// 					for i := 0; i < b.N; i++ {
// 						events := make(chan gol.Event)
// 						gol.Run(p, events, nil)
// 						// var cells []util.Cell
// 						for event := range events {
// 							switch e := event.(type) {
// 							case gol.FinalTurnComplete:
// 								cells := e.Alive
// 								break
// 							}
// 						}
// 					}
// 				})
// 			}
// 		}
// 	}
// }
