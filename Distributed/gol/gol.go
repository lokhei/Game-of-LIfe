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

func makeCall(keyPresses <-chan rune, server string, events chan<- Event, p Params, filename chan<- string, input <-chan uint8, output chan<- uint8, ioCommand chan<- ioCommand, ioIdle <-chan bool) {

	client, err := rpc.Dial("tcp", server)
	if err != nil {
		fmt.Println("RPC client returned error:")
		fmt.Println(err)
		fmt.Println("stopping connection")
	}
	defer client.Close()

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

	//initial Call
	request := stubs.Request{Message: world, Threads: p.Threads, Turns: p.Turns}
	response := new(stubs.Response)
	client.Call(stubs.CallInitial, request, response)

	//ticker
	ticker := time.NewTicker(2 * time.Second)
	done := make(chan bool)

	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				requestAlive := stubs.Request{}
				responseAlive := new(stubs.Response)
				client.Call(stubs.CallAlive, requestAlive, responseAlive)
				events <- AliveCellsCount{CompletedTurns: responseAlive.Turn, CellsCount: responseAlive.AliveCells}
			case key := <-keyPresses:
				if key == 's' {
					reqKey := stubs.Request{}
					resKey := new(stubs.Response)
					client.Call(stubs.CallDoKeypresses, reqKey, resKey)
					printBoard(p, resKey.Turn, resKey.Message, filename, output, ioCommand, ioIdle, events)
				} else if key == 'q' {
					close(events)
					secReq := stubs.Request{}
					secRes := new(stubs.Response)
					client.Call(stubs.CallDoKeypresses, secReq, secRes)
					// } else if key == 'p' {
					// 	pause = true
					// 	reqKey := stubs.Request{}
					// 	resKey := new(stubs.Response)
					// 	client.Call(stubs.KeypressPause, reqKey, resKey)

				}
				// } else if key == 'p' {
				// 	pause = true
				// 	reqKey := stubs.Request{}
				// 	resKey := new(stubs.Response)
				// 	client.Call(stubs.KeypressPause, reqKey, resKey)

				// } else if key == 'p' {
				// 	fmt.Println(turn)
				// 	c.events <- StateChange{CompletedTurns: turn, NewState: Paused}
				// 	for {
				// 		tempKey := <-c.keyPresses
				// 		if tempKey == 'p' {
				// 			fmt.Println("Continuing")
				// 			c.events <- StateChange{CompletedTurns: turn, NewState: Executing}
				// 			break
				// 		}
				// 	}
				// }

			}

		}
	}()

	//final message
	requestFinal := stubs.Request{}
	responseFinal := new(stubs.Response)
	client.Call(stubs.CallReturn, requestFinal, responseFinal)
	returnedworld := responseFinal.Message

	ticker.Stop()

	events <- FinalTurnComplete{p.Turns, calculateAliveCells(returnedworld)}
	printBoard(p, p.Turns, returnedworld, filename, output, ioCommand, ioIdle, events)
	events <- StateChange{p.Turns, Quitting}
	close(events)

}

// Run starts the processing of Game of Life. It should initialise channels and goroutines.
func Run(p Params, events chan<- Event, keyPresses <-chan rune) {

	var server string

	if flag.Lookup("server") == nil {
		serverTemp := flag.String("server", "127.0.0.1:8030", "IP:port string to connect to as server")
		flag.Parse()
		server = *serverTemp
	} else {
		server = flag.Lookup("server").Value.(flag.Getter).Get().(string)
	}

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

	go makeCall(keyPresses, server, events, p, filename, input, output, ioCommand, ioIdle)

}

func printBoard(p Params, turn int, world [][]byte, filename chan<- string, output chan<- uint8, ioCommand chan<- ioCommand, IoIdle <-chan bool, events chan<- Event) {
	ioCommand <- ioOutput
	fileName := strings.Join([]string{strconv.Itoa(p.ImageWidth), strconv.Itoa(p.ImageHeight), strconv.Itoa(turn)}, "x")
	filename <- fileName

	for y := range world {
		for x := range world {
			output <- world[y][x]
		}
	}

	events <- ImageOutputComplete{CompletedTurns: turn, Filename: fileName}
	// Make sure that the Io has finished any output before exiting.
	ioCommand <- ioCheckIdle
	<-IoIdle
}

// takes the world as input and returns the (x, y) coordinates of all the cells that are alive.
func calculateAliveCells(world [][]byte) []util.Cell {
	aliveCells := []util.Cell{}

	for y := 0; y < len(world); y++ {
		for x := 0; x < len(world[0]); x++ {
			if world[y][x] == 255 {
				aliveCells = append(aliveCells, util.Cell{X: x, Y: y})
			}
		}
	}

	return aliveCells
}

// func ticker(aliveChan chan bool) {
// 	for {
// 		time.Sleep(2 * time.Second)
// 		aliveChan <- true
// 	}
// }
