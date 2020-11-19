package gol

import (
	"strconv"
	"strings"
	// "fmt"
	"uk.ac.bris.cs/gameoflife/util"
)

type distributorChannels struct {
	events    chan<- Event
	ioCommand chan<- ioCommand
	ioIdle    <-chan bool
	filename  chan<- string
	input     <-chan uint8
	output    chan<- uint8
}

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p Params, c distributorChannels) {
	// TODO: Create a 2D slice to store the world.
	// TODO: For all initially alive cells send a CellFlipped Event.
	// TODO: Execute all turns of the Game of Life.
	// TODO: Send correct Events when required, e.g. CellFlipped, TurnComplete and FinalTurnComplete.
	//		 See event.go for a list of all events.

	c.ioCommand <- ioInput
	c.filename <- strings.Join([]string{strconv.Itoa(p.ImageWidth), strconv.Itoa(p.ImageHeight)}, "x")

	world := make([][]uint8, p.ImageHeight)
	for i := range world {
		world[i] = make([]uint8, p.ImageWidth)
	}

	for i := range world {
		for j := range world {
			world[i][j] = <-c.input
		}
	} 


	turn := 0
	for turn = 0; turn <= p.Turns; turn++ {
		if turn > 0 {
			workerChannels := make([]chan [][]uint8, p.Threads) 
			splitThreads := p.ImageHeight / p.Threads //16/2 = 8
			for i := range workerChannels {
				workerChannels[i] = make(chan [][]uint8)

				go worker(p, i*splitThreads, (i+1)*splitThreads, 0, p.ImageWidth, world, workerChannels[i])
				
			}
			
			tempWorld := make([][]uint8, 0)
			for i := range workerChannels { // collects the resulting parts into a single 2D slice
				workerResults := <-workerChannels[i] 
				tempWorld = append(tempWorld, workerResults...)
			}

			for i := range tempWorld{
				for j := range tempWorld{
					world[i][j] = tempWorld[i][j]
				}
			}

		}

		c.events <- TurnComplete{CompletedTurns: turn}

		for y := 0; y < p.ImageHeight; y++ {
			for x := 0; x < p.ImageWidth; x++ {
				if world[y][x] == alive {
					c.events <- CellFlipped{CompletedTurns: turn, Cell: util.Cell{X: x, Y: y}}
				}
			}
		}
		if turn == p.Turns {
			c.events <- FinalTurnComplete{CompletedTurns: turn, Alive: calculateAliveCells(p, world)}
		}
	}

	c.ioCommand <- ioOutput
	c.filename <- strings.Join([]string{strconv.Itoa(p.ImageWidth), strconv.Itoa(p.ImageHeight)}, "x")

	for i := range world {
		for j := range world {
			c.output <- world[i][j]
		}
	}

	// Make sure that the Io has finished any output before exiting.
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle

	c.events <- StateChange{turn, Quitting}
	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)
}

func worker(p Params, startY, endY, startX, endX int, world [][]uint8, out chan<- [][]uint8) {
	newData := calculateNextState(p, world, startY, endY, startX, endX)
	subslice := newData[startY:endY]

	out <- subslice

}
