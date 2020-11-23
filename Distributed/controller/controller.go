package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/rpc"
	"os"

	"uk.ac.bris.cs/gameoflife/stubs"
)

// Params provides the details of how to run the Game of Life and which image to load.
type Params struct {
	Turns       int
	Threads     int
	ImageWidth  int
	ImageHeight int
}

//

// Run starts the processing of Game of Life. It should initialise channels and goroutines.
func Run(p Params, events chan<- Event, keyPresses <-chan rune) {
	//////////////////
	// ioCommand := make(chan ioCommand)
	// ioIdle := make(chan bool)
	// filename := make(chan string)
	// input := make(chan uint8)
	// output := make(chan uint8)

	// distributorChannels := distributorChannels{
	// 	events,
	// 	ioCommand,
	// 	ioIdle,
	// 	filename,
	// 	input,
	// 	output,
	// 	keyPresses,
	// }
	// go distributor(p, distributorChannels)

	// ioChannels := ioChannels{
	// 	command:  ioCommand,
	// 	idle:     ioIdle,
	// 	filename: filename,
	// 	output:   output,
	// 	input:    input,
	// }
	// go startIo(p, ioChannels)
	////////////////////////

	// periodicChan := make(chan bool)
	// 	go ticker(periodicChan)

	// select {
	// case key := <-keyPresses:
	// 	if key == 's' {
	// 		printBoard(p, c, world, turn)

	// 	} else if key == 'q' {
	// 		printBoard(p, c, world, turn)
	// 		events <- StateChange{CompletedTurns: turn, NewState: Quitting}
	// 		close(events)
	// 		return

	// 	} else if key == 'p' {
	// 		fmt.Println(turn)
	// 		c.events <- StateChange{CompletedTurns: turn, NewState: Paused}
	// 		for {
	// 			tempKey := <-c.keyPresses
	// 			if tempKey == 'p' {
	// 				fmt.Println("Continuing")
	// 				c.events <- StateChange{CompletedTurns: turn, NewState: Executing}
	// 				break
	// 			}
	// 		}
	// 	}

	// case <-periodicChan:
	// 	c.events <- AliveCellsCount{CompletedTurns: turn, CellsCount: len(calculateAliveCells(p, world))}
	// default:
	// }

}

// func ticker(aliveChan chan bool) {
// 	for {
// 		time.Sleep(2 * time.Second)
// 		aliveChan <- true
// 	}
// }

func makeCall(client rpc.Client, message string) {
	request := stubs.Request{Message: message}
	response := new(stubs.Response)
	client.Call(stubs.Nextworld, request, response)
	fmt.Println("Responded: " + response.Message)
}

func main() {
	server := flag.String("server", "127.0.0.1:8030", "IP:port string to connect to as server")
	flag.Parse()
	client, _ := rpc.Dial("tcp", *server)
	defer client.Close()

	file, _ := os.Open("wordlist")
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		t := scanner.Text()
		fmt.Println("Called: " + t)
		makeCall(*client, t)
	}

}
