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

const alive = 255
const dead = 0

func mod(x, m int) int {
	return (x + m) % m
}

//calculates number of neighbours of cell
func calculateNeighbours(p Params, x, y int, world [][]byte) int {
	neighbours := 0
	for i := -1; i <= 1; i++ {
		for j := -1; j <= 1; j++ {
			if i != 0 || j != 0 { //not [y][x]
				if world[mod(y+i, p.ImageHeight)][mod(x+j, p.ImageWidth)] == alive {
					neighbours++
				}
			}
		}
	}
	return neighbours
}

//takes the current state of the world and completes one evolution of the world. It then returns the result.
func calculateNextState(p Params, subworld [][]byte, c distributorChannels, startY, endY, turns int) [][]byte {

	newWorld := make([][]byte, endY-startY)
	for i := range newWorld {
		newWorld[i] = make([]byte, p.ImageWidth)
	}
	//sets cells to dead or alive according to num of neighbours
	for y := startY; y < endY; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			neighbours := calculateNeighbours(p, x, y, subworld)
			if subworld[y][x] == alive {
				if neighbours == 2 || neighbours == 3 {
					newWorld[y-startY][x] = alive
				} else {
					newWorld[y-startY][x] = dead
					c.events <- CellFlipped{CompletedTurns: turns, Cell: util.Cell{X: x, Y: y}}

				}
			} else {
				if neighbours == 3 {
					newWorld[y-startY][x] = alive
					c.events <- CellFlipped{CompletedTurns: turns, Cell: util.Cell{X: x, Y: y}}

				} else {
					newWorld[y-startY][x] = dead
				}
			}
		}
	}
	return newWorld
}

//takes the world as input and returns the (x, y) coordinates of all the cells that are alive.
func calculateAliveCells(p Params, world [][]uint8) []util.Cell {
	aliveCells := []util.Cell{}

	for y := 0; y < p.ImageHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			if world[y][x] == alive {
				aliveCells = append(aliveCells, util.Cell{X: x, Y: y})
			}
		}
	}

	return aliveCells
}

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p Params, c distributorChannels) {
	// TODO: Send correct Events when required, e.g. CellFlipped, TurnComplete and FinalTurnComplete.
	//		 See event.go for a list of all events.

	c.ioCommand <- ioInput
	c.filename <- strings.Join([]string{strconv.Itoa(p.ImageWidth), strconv.Itoa(p.ImageHeight)}, "x")

	// Create the 2D slice to store the world
	world := make([][]byte, p.ImageHeight)
	for i := range world {
		world[i] = make([]byte, p.ImageWidth)
		for j := range world {
			world[i][j] = <-c.input
			if world[i][j] == alive {
				c.events <- CellFlipped{CompletedTurns: 0, Cell: util.Cell{X: j, Y: i}}
			}
		}
	}
	initialWorld := world

	rem := p.ImageHeight % p.Threads
	splitThreads := p.ImageHeight / p.Threads

	workerChannels := make([]chan [][]byte, p.Threads)
	for i := range workerChannels {
		workerChannels[i] = make(chan [][]byte)
	}

	periodicChan := make(chan bool)
	go ticker(periodicChan)

	// Execute all turns of the Game of Life.
	turn := 0
	for turn = 0; turn <= p.Turns; turn++ {
		if turn > 0 {
			for i := range workerChannels {
				start := i*splitThreads + rem
				end := (i+1)*splitThreads + rem
				if i < rem {

					start = i * (splitThreads + 1)
					end = (i + 1) * (splitThreads + 1)

				}

				go worker(p, start, end, turn, world, c, workerChannels[i])

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
						key = <-c.keyPresses
						if key == 'p' {
							fmt.Println("Continuing")
							c.events <- StateChange{CompletedTurns: turn, NewState: Executing}
							break
						} else if key == 's' {
							printBoard(p, c, world, turn)

						} else if key == 'q' {
							printBoard(p, c, world, turn)
							c.events <- StateChange{CompletedTurns: turn, NewState: Quitting}
							close(c.events)
							return
						} else if key == 'k' {
							fmt.Println("Resetting World")
							turn = 0
							world = initialWorld
							c.events <- AliveCellsCount{CompletedTurns: turn, CellsCount: len(calculateAliveCells(p, world))}

						}
					}

				} else if key == 'k' {
					fmt.Println("Resetting World")
					turn = 0
					world = initialWorld
					c.events <- AliveCellsCount{CompletedTurns: turn, CellsCount: len(calculateAliveCells(p, world))}

				}

			case <-periodicChan:
				c.events <- AliveCellsCount{CompletedTurns: turn, CellsCount: len(calculateAliveCells(p, world))}
			default:
			}

		}

		c.events <- TurnComplete{CompletedTurns: turn}

		if turn == p.Turns {
			c.events <- FinalTurnComplete{CompletedTurns: turn, Alive: calculateAliveCells(p, world)}

		}
	}

	printBoard(p, c, world, p.Turns)

	c.events <- StateChange{turn, Quitting}
	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)
}

func worker(p Params, startY, endY, turn int, world [][]byte, c distributorChannels, out chan<- [][]byte) {
	newData := calculateNextState(p, world, c, startY, endY, turn)
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
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle
}
