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
	os.Stdout = nil

	for t:=1; t <= 16 ; t++ {
		value = true
		tests := gol.Params{

			ImageWidth: 512, ImageHeight: 512,
			
		}
		tests.Threads = t
		name := fmt.Sprintf("%dx%dx%d-%d", tests.ImageWidth, tests.ImageHeight, turnNum, t)

		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if value == true {
					events := make(chan gol.Event)
					gol.Run(tests, events, nil)

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
