package gol
import (
	"strconv"
	"strings"
)

// Params provides the details of how to run the Game of Life and which image to load.
type Params struct {
	Turns       int
	Threads     int
	ImageWidth  int
	ImageHeight int
}

type cell struct {
	x, y int
}

const alive = 255
const dead = 0

func mod(x, m int) int {
	return (x + m) % m
}

func gameOfLife(p Params, initialWorld [][]byte) [][]byte {

	world := initialWorld

	return world
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
func calculateNextState(p Params, world [][]byte) [][]byte {
	//makes a new world
	newWorld := make([][]byte, p.ImageHeight)
	for i := range newWorld {
		newWorld[i] = make([]byte, p.ImageWidth)
	}
	//sets cells to dead or alive according to num of neighbours
	for y := 0; y < p.ImageHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			neighbours := calculateNeighbours(p, x, y, world)
			if world[y][x] == alive {
				if neighbours == 2 || neighbours == 3 {
					newWorld[y][x] = alive
				} else {
					newWorld[y][x] = dead
				}
			} else {
				if neighbours == 3 {
					newWorld[y][x] = alive
				} else {
					newWorld[y][x] = dead
				}
			}
		}
	}
	return newWorld
}

//takes the world as input and returns the (x, y) coordinates of all the cells that are alive.
func calculateAliveCells(p Params, world [][]byte) []cell {
	aliveCells := []cell{}

	for y := 0; y < p.ImageHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			if world[y][x] == alive {
				aliveCells = append(aliveCells, cell{x: x, y: y})
			}
		}
	}

	return aliveCells
}

// Run starts the processing of Game of Life. It should initialise channels and goroutines.
func Run(p Params, events chan<- Event, keyPresses <-chan rune) {
	
	ioCommand := make(chan ioCommand)
	ioIdle := make(chan bool)
	filename := make (chan string)
	input := make(chan uint8)
	output := make(chan uint8)
	

	distributorChannels := distributorChannels{
		events,
		ioCommand,
		ioIdle,
		filename,
		input,
	}
	go distributor(p, distributorChannels)

	ioChannels := ioChannels{
		command:  ioCommand,
		idle:     ioIdle,
		filename: filename,
		output:   output,
		input:    input,
	}
	go startIo(p, ioChannels)
}
