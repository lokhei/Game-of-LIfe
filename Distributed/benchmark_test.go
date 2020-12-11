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

		{ImageWidth: 16, ImageHeight: 16, Threads: 2},

		{ImageWidth: 64, ImageHeight: 64, Threads: 2},

		{ImageWidth: 128, ImageHeight: 128, Threads: 2},

		{ImageWidth: 256, ImageHeight: 256, Threads: 2},

		{ImageWidth: 512, ImageHeight: 512, Threads: 2},
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
