package gol

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"uk.ac.bris.cs/gameoflife/util"
)

type distributorChannels struct {
	events     chan<- Event
	ioCommand  chan<- ioCommand
	ioIdle     <-chan bool
	filename   chan<- string
	input      <-chan uint8
	output     chan<- uint8
	keyPresses <-chan rune
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

	rem := mod(p.ImageHeight, p.Threads)
	splitThreads := p.ImageHeight / p.Threads

	turn := 0
	for turn = 0; turn <= p.Turns; turn++ {
		if turn > 0 {
			workerChannels := make([]chan [][]byte, p.Threads)
			for i := range workerChannels {
				workerChannels[i] = make(chan [][]byte)

				startY := i*splitThreads + rem
				endY := (i+1)*splitThreads + rem
				if i < rem {

					startY = i * (splitThreads + 1)
					endY = (i + 1) * (splitThreads + 1)

				}
				size := endY - startY + 2

				subworld := make([][]byte, size)
				for k := range subworld {
					subworld[k] = make([]byte, p.ImageWidth)
				}
				if i != 0 && i != p.Threads-1 { // not equal to first or last worker
					go worker(p, startY, endY, world[startY-1:endY+1], workerChannels[i])
					continue

				} else if i == 0 { //first worker
					subworld[0] = world[p.ImageHeight-1]
					if p.Threads == 1 {
						subworld[size-1] = world[0]
						size--

					}
					for j := 1; j < size; j++ {
						subworld[j] = world[j-1]

					}
				} else { //last worker

					for j := 0; j < size-1; j++ {
						subworld[j] = world[startY+j-1]
					}
					subworld[size-1] = world[0]
				}

				go worker(p, startY, endY, subworld, workerChannels[i])

			}

			tempWorld := make([][]byte, 0)
			for i := range workerChannels { // collects the resulting parts into a single 2D slice
				workerResults := <-workerChannels[i]
				tempWorld = append(tempWorld, workerResults...)
			}
			world = tempWorld

			select {
			case key := <-c.keyPresses:
				if key == 's' {
					printBoard(p, c, world, turn)

				} else if key == 'q' {
					printBoard(p, c, world, turn)
					c.events <- StateChange{CompletedTurns: turn, NewState: Quitting}
					close(c.events)
					return

				} else if key == 'p' {
					fmt.Println(turn)
					c.events <- StateChange{CompletedTurns: turn, NewState: Paused}
					for {
						tempKey := <-c.keyPresses
						if tempKey == 'p' {
							fmt.Println("Continuing")
							c.events <- StateChange{CompletedTurns: turn, NewState: Executing}
							break
						}
					}
				}

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

	printBoard(p, c, world, p.Turns)

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

func printBoard(p Params, c distributorChannels, world [][]byte, turn int) {
	c.ioCommand <- ioOutput
	fileName := strings.Join([]string{strconv.Itoa(p.ImageWidth), strconv.Itoa(p.ImageHeight), strconv.Itoa(turn)}, "x")
	c.filename <- fileName

	for y := range world {
		for x := range world {
			c.output <- world[y][x]
		}
	}
	c.events <- ImageOutputComplete{CompletedTurns: turn, Filename: fileName}
	// Make sure that the Io has finished any output before exiting.
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle
}
