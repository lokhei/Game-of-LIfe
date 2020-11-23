package main

import (
	"flag"
	"math/rand"
	"net"
	"net/rpc"
	"time"

	"uk.ac.bris.cs/gameoflife/stubs"
)

// type distributorChannels struct {
// 	events     chan<- Event
// 	ioCommand  chan<- ioCommand
// 	ioIdle     <-chan bool
// 	filename   chan<- string
// 	input      <-chan uint8
// 	output     chan<- uint8
// 	keyPresses <-chan rune
// }

type NextStateOperation struct{}

// Distributor divides the work between workers and interacts with other goroutines.
func (s *NextStateOperation) Distributor(req stubs.Request, res *stubs.Response)(err error) {

	height := len(req.Message)
	width := len(req.Message[0])

	rem := mod(height, req.Threads)
	splitThreads := height / req.Threads

	turn := 0
	for turn = 0; turn <= req.Turns; turn++ {
		if turn > 0 {
			workerChannels := make([]chan [][]byte, req.Threads)
			for i := range workerChannels {
				workerChannels[i] = make(chan [][]byte)

				startY := i*splitThreads + rem
				endY := (i+1)*splitThreads + rem

				if i < rem {
					startY = i * (splitThreads + 1)
					endY = (i + 1) * (splitThreads + 1)
				}

				go worker(height, width, startY, endY, req.Message, workerChannels[i])

			}

			tempWorld := make([][]byte, 0)
			for i := range workerChannels { // collects the resulting parts into a single 2D slice
				workerResults := <-workerChannels[i]
				tempWorld = append(tempWorld, workerResults...)
			}
			res.Message = tempWorld

		}
		

		// c.events <- TurnComplete{CompletedTurns: turn}
		// for y := 0; y < height; y++ {
		// 	for x := 0; x < width; x++ {
		// 		if world[y][x] == alive {
		// 			c.events <- CellFlipped{CompletedTurns: turn, Cell: util.Cell{X: x, Y: y}}
		// 		}
		// 	}
		// }
		// if turn == turnNum {
		// 	c.events <- FinalTurnComplete{CompletedTurns: turn, Alive: calculateAliveCells(p, world)}

		// }
	}

	return
	// c.events <- StateChange{turn, Quitting}
	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	// close(c.events)
}


func worker(height, width, startY, endY int, world [][]byte, out chan<- [][]uint8) {
	newData := calculateNextState(height, width, startY, endY, world)
	out <- newData

}

////////////
const alive = 255
const dead = 0

func mod(x, m int) int {
	return (x + m) % m
}

//calculates number of neighbours of cell
func calculateNeighbours(height, width, x, y int, world [][]byte) int {
	neighbours := 0
	for i := -1; i <= 1; i++ {
		for j := -1; j <= 1; j++ {
			if i != 0 || j != 0 { //not [y][x]
				if world[mod(y+i, height)][mod(x+j, width)] == alive {
					neighbours++
				}
			}
		}
	}
	return neighbours
}

//takes the current state of the world and completes one evolution of the world. It then returns the result.
func calculateNextState(height, width, startY, endY int, world [][]byte) [][]byte {
	//makes a new world
	newWorld := make([][]uint8, endY-startY)
	for i := range newWorld {
		newWorld[i] = make([]uint8, width)
	}
	//sets cells to dead or alive according to num of neighbours
	for y := startY; y < endY; y++ {
		for x := 0; x < width; x++ {
			neighbours := calculateNeighbours(height, width, x, y, world)
			if world[y][x] == alive {
				if neighbours == 2 || neighbours == 3 {
					newWorld[y-startY][x] = alive
				} else {
					newWorld[y-startY][x] = dead
				}
			} else {
				if neighbours == 3 {
					newWorld[y-startY][x] = alive
				} else {
					newWorld[y-startY][x] = dead
				}
			}
		}
	}
	return newWorld
}



func main() {
	pAddr := flag.String("port", "8030", "Port to listen on")
	flag.Parse()
	rand.Seed(time.Now().UnixNano())
	rpc.Register(&NextStateOperation{})
	listener, _ := net.Listen("tcp", ":"+*pAddr)
	defer listener.Close()
	rpc.Accept(listener)
}
