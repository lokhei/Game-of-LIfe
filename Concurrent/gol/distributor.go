package gol

import (
	"strconv"
	"strings"

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

	c.ioCommand <- ioInput
	c.filename <- strings.Join([]string{strconv.Itoa(p.ImageWidth), strconv.Itoa(p.ImageHeight)}, "x")

	// TODO: Create a 2D slice to store the world.
	world := make([][]byte, p.ImageHeight)
	for i := range world {
		world[i] = make([]byte, p.ImageWidth)
	}

	for i := range world {
		for j := range world {
			world[i][j] = <-c.input
		}
	}

	// TODO: For all initially alive cells send a CellFlipped Event.
	// TODO: Execute all turns of the Game of Life.
	// TODO: Send correct Events when required, e.g. CellFlipped, TurnComplete and FinalTurnComplete.
	//		 See event.go for a list of all events.
	splitThreads := p.ImageHeight / p.Threads

	turn := 0
	for turn = 0; turn <= p.Turns; turn++ {
		if turn > 0 {
			workerChannels := make([]chan [][]byte, p.Threads)

			newWorld := make([][]byte, p.ImageWidth)

			for i := range workerChannels {
				workerChannels[i] = make(chan [][]byte)                                                  //make individual channels
				go worker(p, i*splitThreads, (i+1)*splitThreads, splitThreads, world, workerChannels[i]) //start 4 workers

			}

			for i := range workerChannels { // collects the resulting parts into a single 2D slice
				workerResults := <-workerChannels[i]
				newWorld = append(newWorld, workerResults...)
			}

			// for y := range world {
			// 	for x := range world {
			// 		world[y][x] = newWorld[y][x]
			// 	}
			// }
			world = newWorld
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

	for y := 0; y < p.ImageHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			c.output <- world[y][x]
		}
	}

	// Make sure that the Io has finished any output before exiting.
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle

	c.events <- StateChange{turn, Quitting}
	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)
}

func worker(p Params, startY, endY, splitThreads int, world [][]byte, out chan<- [][]byte) {
	world = calculateNextState(p, world, startY, endY, splitThreads)
	out <- world

}
