package gol

import (
	"flag"
	"fmt"
	"net/rpc"
	"strconv"
	"strings"
	"time"

	"uk.ac.bris.cs/gameoflife/stubs"
	"uk.ac.bris.cs/gameoflife/util"
)

// Params provides the details of how to run the Game of Life and which image to load.
type Params struct {
	Turns       int
	Threads     int
	ImageWidth  int
	ImageHeight int
}

func makeCall(client rpc.Client, world [][]uint8, events chan<- Event, keyPresses <-chan rune, p Params, filename chan<- string, output chan<- uint8, ioCommand chan<- ioCommand, ioIdle <-chan bool) {

	request := stubs.Request{Message: world, Threads: p.Threads, Turns: p.Turns}
	response := new(stubs.Response)
	client.Call(stubs.Nextworld, request, response)

	periodicChan := make(chan bool)
	go ticker(periodicChan)

	select {
	case key := <-keyPresses:
		if key == 's' {
			printBoard(p, response.Message, filename, output, ioCommand, ioIdle, events)

		} else if key == 'q' {
			printBoard(p, response.Message, filename, output, ioCommand, ioIdle, events)
			events <- StateChange{CompletedTurns: p.Turns, NewState: Quitting}
			close(events)
			return

		} else if key == 'p' {
			fmt.Println(p.Turns)
			events <- StateChange{CompletedTurns: p.Turns, NewState: Paused}
			for {
				tempKey := <-keyPresses
				if tempKey == 'p' {
					fmt.Println("Continuing")
					events <- StateChange{CompletedTurns: p.Turns, NewState: Executing}
					break
				}
			}
		}

	case <-periodicChan:
		events <- AliveCellsCount{CompletedTurns: p.Turns, CellsCount: len(calculateAliveCells(p, world))}
	default:
	}

	printBoard(p, response.Message, filename, output, ioCommand, ioIdle, events)
}

// Run starts the processing of Game of Life. It should initialise channels and goroutines.
func Run(p Params, events chan<- Event, keyPresses <-chan rune) {

	server := flag.String("server", "127.0.0.1:8030", "IP:port string to connect to as server")
	flag.Parse()
	client, _ := rpc.Dial("tcp", *server)
	defer client.Close()

	ioCommand := make(chan ioCommand)
	ioIdle := make(chan bool)
	filename := make(chan string)
	input := make(chan uint8)
	output := make(chan uint8)

	ioChannels := ioChannels{
		command:  ioCommand,
		idle:     ioIdle,
		filename: filename,
		output:   output,
		input:    input,
	}
	go startIo(p, ioChannels)

	ioCommand <- ioInput
	filename <- strings.Join([]string{strconv.Itoa(p.ImageWidth), strconv.Itoa(p.ImageHeight)}, "x")

	world := make([][]byte, p.ImageHeight)
	for i := range world {
		world[i] = make([]byte, p.ImageWidth)
	}

	for i := range world {
		for j := range world {
			world[i][j] = <-input
		}
	}

	makeCall(*client, world, events, keyPresses, p, filename, output, ioCommand, ioIdle)

}

func ticker(aliveChan chan bool) {
	for {
		time.Sleep(2 * time.Second)
		aliveChan <- true
	}
}

func printBoard(p Params, world [][]byte, filename chan<- string, output chan<- uint8, ioCommand chan<- ioCommand, IoIdle <-chan bool, events chan<- Event) {
	ioCommand <- ioOutput
	fileName := strings.Join([]string{strconv.Itoa(p.ImageWidth), strconv.Itoa(p.ImageHeight), strconv.Itoa(p.Turns)}, "x")
	filename <- fileName

	for y := range world {
		for x := range world {
			output <- world[y][x]
		}
	}
	events <- ImageOutputComplete{CompletedTurns: p.Turns, Filename: fileName}
	// Make sure that the Io has finished any output before exiting.
	ioCommand <- ioCheckIdle
	<-IoIdle
}

// takes the world as input and returns the (x, y) coordinates of all the cells that are alive.
func calculateAliveCells(p Params, world [][]uint8) []util.Cell {
	aliveCells := []util.Cell{}

	for y := 0; y < p.ImageHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			if world[y][x] == 255 {
				aliveCells = append(aliveCells, util.Cell{X: x, Y: y})
			}
		}
	}

	return aliveCells
}
