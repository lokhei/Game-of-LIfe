package main

import (
	"flag"
	"log"
	"net"
	"net/rpc"
	"os"
	"time"

	"uk.ac.bris.cs/gameoflife/stubs"
	"uk.ac.bris.cs/gameoflife/util"
)

var FinalWorld [][]byte
var CurrentWorld [][]byte
var AliveCells int
var Currentturn int
var done bool
var key bool
var pause bool
var Waddress []string
var CAddress string
var quit bool

// var alive bool

const alive = 255
const dead = 0

func mod(x, m int) int {
	return (x + m) % m
}

type NextStateOperation struct{}

// Distributor divides the work between workers and interacts with other goroutines.
func distributor(world [][]byte, turns, threads int) {
	done = false
	height := len(world)
	width := len(world[0])
	rem := mod(height, len(Waddress))
	splitThreads := height / len(Waddress)
	AliveChannels := make(chan []util.Cell, 0)

	//dials into client.go
	client, err := rpc.Dial("tcp", CAddress)
	// defer client.Close()

	if err != nil {
		log.Fatal("Dial error:", err)
	}

	//don't want this
	for turn := Currentturn; turn <= turns; turn++ {
		// aliveCellList := make([]util.Cell, 0)

		if turn > 0 {

			for pause {

			}

			workerChannels := make([]chan [][]byte, len(Waddress))
			for i := range workerChannels {

				workerChannels[i] = make(chan [][]byte)
				startY := i*splitThreads + rem
				endY := (i+1)*splitThreads + rem

				if i < rem {
					startY = i * (splitThreads + 1)
					endY = (i + 1) * (splitThreads + 1)
				}
				//pass in subworld
				//pass in turns

				go CallWorker(world, startY, endY, workerChannels[i], AliveChannels, Waddress[i])
				//receive the edge rows and send off to respective workers
			}

			//only append subworlds if required to send back to client
			tempWorld := make([][]byte, 0)
			for i := range workerChannels { // collects the resulting parts into a single 2D slice
				workerResults := <-workerChannels[i]
				tempWorld = append(tempWorld, workerResults...)
			}
			Currentturn++
			world = tempWorld
			CurrentWorld = world
		}
		AliveCells = 0
		for h := 0; h < height; h++ {
			for w := 0; w < width; w++ {
				if world[h][w] == alive {
					AliveCells++

				}
			}
		}

		//
		select {
		case alive := <-AliveChannels:
			SdlReq := stubs.SDLReq{Alive: alive, Turn: turn}
			SdlRes := new(stubs.SDLRes)
			client.Call(stubs.SdlEvent, SdlReq, SdlRes)
		default:
		}
	}
	client.Close()

	FinalWorld = world
	done = true
}

//InitialState : Initial state of the world
func (s *NextStateOperation) InitialState(req stubs.Request, res *stubs.Response) (err error) {

	World := req.Message
	Turn := req.Turns
	// quit = false
	Threads := req.Threads
	if key {
		World = CurrentWorld
	} else {
		Currentturn = 0
	}
	go distributor(World, Turn, Threads)
	return
}

//FinalState : Final state of the world
func (s *NextStateOperation) FinalState(req stubs.Request, res *stubs.Response) (err error) {
	for done == false {
		//
	}
	res.Message = FinalWorld
	return
}

//Alive : Return current World + Turn for counting alive cells
func (s *NextStateOperation) Alive(req stubs.Request, res *stubs.Response) (err error) {
	res.Turn = Currentturn
	res.AliveCells = AliveCells
	return
}

// DoKeypresses : function for keypresses
func (s *NextStateOperation) DoKeypresses(req stubs.Request, res *stubs.Response) (err error) {
	res.Turn = Currentturn
	res.Message = CurrentWorld
	key = true
	pause = req.Pause
	return
}

//Quit closes all instances
func (s *NextStateOperation) Quit(req stubs.Request, res *stubs.Response) (err error) {
	quit = true
	return
}

//GetAddress gets address of worker node
func (s *NextStateOperation) GetAddress(req stubs.ReqAddress, res *stubs.ResAddress) (err error) {
	Waddress = append(Waddress, req.WorkerAddress)
	return
}

//GetCAddress gets address of client
func (s *NextStateOperation) GetCAddress(req stubs.ReqAddress, res *stubs.ResAddress) (err error) {
	CAddress = req.WorkerAddress
	return
}

//CallWorker creates connection to worker node
func CallWorker(world [][]byte, startingY, endingY int, workerChannels chan<- [][]byte, AliveChannels chan<- []util.Cell, address string) (err error) {
	//connects to worker

	worker, err := rpc.Dial("tcp", address)
	if err != nil {
		log.Fatal("Dial error:", err)
		return err
	}
	if quit {

		worker.Call(stubs.QuitW, stubs.ReqWorker{}, new(stubs.ResWorker))
		time.Sleep(10 * time.Millisecond)
		worker.Close()
		os.Exit(0)
	}

	request := stubs.ReqWorker{World: world, StartY: startingY, EndY: endingY}
	response := new(stubs.ResWorker)

	worker.Call(stubs.CalculateNextState, request, response)

	workerChannels <- response.World
	AliveChannels <- response.Alive
	worker.Close()

	return
}

//to connect to client
func main() {
	//its own connection
	pAddr := flag.String("port", ":8030", "Port to listen on")
	flag.Parse()
	rpc.Register(&NextStateOperation{})
	listener, err := net.Listen("tcp", *pAddr)
	if err != nil {
		log.Fatal("listen error:", err)
	}

	rpc.Accept(listener)
	listener.Close()

}
