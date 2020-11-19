package gol

import (
	"strconv"
	"strings"
	"time"

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

	world := make([][]byte, p.ImageHeight)
	for i := range world {
		world[i] = make([]byte, p.ImageWidth)
	}

	for i := range world {
		for j := range world {
			world[i][j] = <-c.input
		}
	}

	periodicChan := make(chan bool)
	go ticker(periodicChan)

	turn := 0
	for turn = 0; turn <= p.Turns; turn++ {
		if turn > 0 {
			workerChannels := make([]chan [][]byte, p.Threads)
			splitThreads := p.ImageHeight / p.Threads
			for i := range workerChannels {
				workerChannels[i] = make(chan [][]byte)
				if i == p.Threads-1 {
					rem := mod(p.ImageHeight, p.Threads)
					go worker(p, i*splitThreads, (i+1)*splitThreads+rem, world, workerChannels[i])

				} else {
					go worker(p, i*splitThreads, (i+1)*splitThreads, world, workerChannels[i])
				}

			}

			tempWorld := make([][]byte, 0)
			for i := range workerChannels { // collects the resulting parts into a single 2D slice
				workerResults := <-workerChannels[i]
				tempWorld = append(tempWorld, workerResults...)
			}

			world = tempWorld

			select {
			case <-periodicChan:
				c.events <- AliveCellsCount{CompletedTurns: turn, CellsCount: len(calculateAliveCells(p, world))}
			default:
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

	printBoard(p, c, world)

	c.events <- StateChange{turn, Quitting}
	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)
}

func worker(p Params, startY, endY int, world [][]byte, out chan<- [][]uint8) {
	newData := calculateNextState(p, world, startY, endY)
	out <- newData

}

func ticker(aliveChan chan bool) {
	for {
		time.Sleep(2 * time.Second)
		aliveChan <- true
	}
}

func printBoard(p Params, c distributorChannels, world [][]byte) {
	c.ioCommand <- ioOutput
	c.filename <- strings.Join([]string{strconv.Itoa(p.ImageWidth), strconv.Itoa(p.ImageHeight), strconv.Itoa(p.Turns)}, "x")

	for y := range world {
		for x := range world {
			c.output <- world[y][x]
		}
	}

	// Make sure that the Io has finished any output before exiting.
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle
}
