package main

import (
	"flag"
	"log"
	"math/rand"
	"net"
	"net/rpc"
	"time"

	// "fmt"
	"uk.ac.bris.cs/gameoflife/stubs"
)

type NextStateOperation struct{}

// Distributor divides the work between workers and interacts with other goroutines.
func (s *NextStateOperation) Distributor(req stubs.Request, res *stubs.Response) (err error) {

	height := len(req.Message)
	width := len(req.Message[0])
	rem := mod(height, req.Threads)
	splitThreads := height / req.Threads

	world := req.Message
	// periodicChan := make(chan bool)
	// go ticker(periodicChan)

	ticker := time.NewTicker(500 * time.Millisecond)
	done := make(chan bool)
	res.Turns = 0
	for turns := 0; turns <= req.Turns; turns++ {
		if turns > 0 {
			res.Turns++
			workerChannels := make([]chan [][]byte, req.Threads)
			for i := range workerChannels {
				workerChannels[i] = make(chan [][]byte)

				startY := i*splitThreads + rem
				endY := (i+1)*splitThreads + rem

				if i < rem {
					startY = i * (splitThreads + 1)
					endY = (i + 1) * (splitThreads + 1)
				}

				go worker(height, width, startY, endY, world, workerChannels[i])

			}

			tempWorld := make([][]byte, 0)
			for i := range workerChannels { // collects the resulting parts into a single 2D slice
				workerResults := <-workerChannels[i]
				tempWorld = append(tempWorld, workerResults...)
			}
			res.Message = tempWorld
			world = tempWorld

			// select {
			// case <-periodicChan:
			// 	return
			// default:
			// }

			send := false
			go func() {
				for {
					select {
					case <-done:
						return
					case <-ticker.C:
						send = true
					}
				}
			}()
			if send == true {
				return

			}
		}

	}
	res.Message = world
	return
}

func worker(height, width, startY, endY int, world [][]byte, out chan<- [][]uint8) {
	newData := calculateNextState(height, width, startY, endY, world)
	out <- newData
}

func ticker(aliveChan chan bool) {
	for {
		time.Sleep(2 * time.Second)
		aliveChan <- true
	}
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
	pAddr := flag.String("port", ":8030", "Port to listen on")
	flag.Parse()
	rand.Seed(time.Now().UnixNano())
	rpc.Register(&NextStateOperation{})
	listener, err := net.Listen("tcp", *pAddr)
	if err != nil {
		log.Fatal("listen error:", err)
	}

	defer listener.Close()
	rpc.Accept(listener)
}
