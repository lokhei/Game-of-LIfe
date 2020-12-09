package gol

import (
	"fmt"
	"log"
	"net/rpc"
	"os"
	"strconv"
	"strings"
	"time"

	"uk.ac.bris.cs/gameoflife/stubs"
	"uk.ac.bris.cs/gameoflife/util"
)

func makeCall(keyPresses <-chan rune, server string, events chan<- Event, p Params, filename chan<- string, input <-chan uint8, output chan<- uint8, ioCommand chan<- ioCommand, ioIdle <-chan bool) {

	//client is connecting to logic
	client, err := rpc.Dial("tcp", server)
	if err != nil {
		log.Fatal("listen error:", err)
	}

	ioCommand <- ioInput
	filename <- strings.Join([]string{strconv.Itoa(p.ImageWidth), strconv.Itoa(p.ImageHeight)}, "x")

	world := make([][]byte, p.ImageHeight)
	for i := range world {
		world[i] = make([]byte, p.ImageWidth)
		for j := range world {
			world[i][j] = <-input
			if world[i][j] == 255 {

				events <- CellFlipped{0, util.Cell{X: j, Y: i}}

			}
		}
	}
	//initial Call
	request := stubs.Request{Message: world, Threads: p.Threads, Turns: p.Turns}
	response := new(stubs.Response)
	client.Call(stubs.CallInitial, request, response)

	ticker := time.NewTicker(2 * time.Second)
	done := make(chan bool)
	pause := false

	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				if !pause {
					requestAlive := stubs.Request{}
					responseAlive := new(stubs.Response)
					client.Call(stubs.CallAlive, requestAlive, responseAlive)
					events <- AliveCellsCount{CompletedTurns: responseAlive.Turn, CellsCount: responseAlive.AliveCells}
				}
			case key := <-keyPresses:
				if key == 's' {
					reqKey := stubs.Request{}
					resKey := new(stubs.Response)
					client.Call(stubs.CallDoKeypresses, reqKey, resKey)
					printBoard(p, resKey.Turn, resKey.Message, filename, output, ioCommand, ioIdle, events)

				} else if key == 'q' {
					close(events)
					reqKey := stubs.Request{}
					resKey := new(stubs.Response)
					client.Call(stubs.CallDoKeypresses, reqKey, resKey)

				} else if key == 'p' {
					reqKey := stubs.Request{Pause: true}
					pause = true
					resKey := new(stubs.Response)
					client.Call(stubs.CallDoKeypresses, reqKey, resKey)
					events <- StateChange{CompletedTurns: resKey.Turn, NewState: Paused}
					for {
						key = <-keyPresses
						if key == 'p' {
							fmt.Println("Continuing")
							events <- StateChange{CompletedTurns: resKey.Turn, NewState: Executing}

							reqKey = stubs.Request{Pause: false}
							resKey = new(stubs.Response)
							pause = false
							client.Call(stubs.CallDoKeypresses, reqKey, resKey)
							break

						}

					}

				} else if key == 'k' {
					fmt.Println("Exit All")

					reqKey := stubs.Request{}
					resKey := new(stubs.Response)
					client.Call(stubs.Quit, reqKey, resKey)
					os.Exit(0)
				}
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
