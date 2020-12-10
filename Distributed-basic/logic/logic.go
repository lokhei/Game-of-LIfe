package main

import (
	"flag"
	"fmt"
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
var Alive []bool
var pause bool
var Waddress []string
var TopC [][]byte
var BottomC [][]byte
var Current []int

// var CAddress string
var quit bool

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
	// width := len(world[0])
	rem := mod(height, len(Waddress))
	splitThreads := height / len(Waddress)
	AliveChannels := make(chan []util.Cell, 0)

	workerChannels := make([]chan [][]byte, len(Waddress))
	startY := make([]int, len(Waddress))
	endY := make([]int, len(Waddress))
	bottom := make([]byte, len(Waddress))
	top := make([]byte, len(Waddress))
	TopC = make([][]byte, len(Waddress))
	BottomC = make([][]byte, len(Waddress))

	Current = make([]int, len(Waddress))

	for i := range workerChannels {
		startY[i] = i*splitThreads + rem
		endY[i] = (i+1)*splitThreads + rem

		if i < rem {
			startY[i] = i * (splitThreads + 1)
			endY[i] = (i + 1) * (splitThreads + 1)
		}

		workerChannels[i] = make(chan [][]byte)

	}
	if turns == 0 {
		done = true
		FinalWorld = world
	}

	if turns > 0 {

		for pause {

		}

		for i := range workerChannels {
			if startY[i] == 0 {
				bottom = world[height-1]
			} else {
				bottom = world[startY[i]-1]
			}
			if endY[i] == height {
				top = world[0]
			} else {
				top = world[endY[i]]

			}
			Current[i] = 0
			subworld := world[startY[i]:endY[i]]
			go CallWorker(subworld, bottom, top, turns, Current[i], i, workerChannels[i], AliveChannels, Waddress[i])

		}

		// world = tempWorld
		// CurrentWorld = world
	}
	for {
		for i := 0; i < len(Current)-1; i++ {
			if Current[i] != Current[i+1] {
				break
			}

		}
		Currentturn++
		if Currentturn == turns {

			reconstruct(workerChannels)
		}
		break
	}

	// client.Close()

}

//InitialState : Initial state of the world
func (s *NextStateOperation) InitialState(req stubs.Request, res *stubs.Response) (err error) {

	World := req.Message
	Turn := req.Turns
	// quit = false
	Threads := req.Threads
	Alive = make([]bool, len(Waddress))
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

//Alive : Return number of alive cells + Turn
func (s *NextStateOperation) Alive(req stubs.Request, res *stubs.Response) (err error) {
	AliveCells = 0
	all := false
	for i := range Alive {
		Alive[i] = true
	}
	for {
		for i := range Alive {
			if Alive[i] == false {
				all = true
			} else {
				all = false
				break
			}
		}
		break
	}
	if all == true {
		res.Turn = Currentturn
		res.AliveCells = AliveCells
	}
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

//CallWorker creates connection to worker node
func CallWorker(subworld [][]byte, bottom, top []byte, turn, current, workernum int, workerChannels chan<- [][]byte, AliveChannels chan<- []util.Cell, address string) (err error) {
	//connects to worker

	worker, err := rpc.Dial("tcp", address)
	if err != nil {
		log.Fatal("Dial error:", err)
		return err
	}

	//calls worker to close worker node
	if quit {
		worker.Call(stubs.QuitW, stubs.ReqWorker{}, new(stubs.ResWorker))
		time.Sleep(10 * time.Millisecond)
		worker.Close()
		os.Exit(0)
	}

	//initial turn
	var request stubs.ReqWorker
	var response *stubs.ResWorker

	for current < turn {
		if current == 0 {
			request = stubs.ReqWorker{World: subworld, Top: top, Bottom: bottom, Turns: turn, CurrentTurn: current, Alive: Alive[workernum]}
		} else if current < turn && len(Waddress) == 1 {
			request = stubs.ReqWorker{Top: BottomC[0], Bottom: TopC[0], Turns: turn, CurrentTurn: current, Alive: Alive[workernum]}

		} else if current < turn && workernum == 0 {
			request = stubs.ReqWorker{Top: BottomC[workernum+1], Bottom: TopC[len(Waddress)-1], Turns: turn, CurrentTurn: current, Alive: Alive[workernum]}

		} else if current < turn && workernum == len(Waddress)-1 {
			request = stubs.ReqWorker{Top: BottomC[0], Bottom: TopC[workernum-1], Turns: turn, CurrentTurn: current, Alive: Alive[workernum]}

		} else {
			request = stubs.ReqWorker{Top: BottomC[workernum+1], Bottom: TopC[workernum-1], Turns: turn, CurrentTurn: current, Alive: Alive[workernum]}

		}
		response = new(stubs.ResWorker)
		worker.Call(stubs.CalculateNextState, request, response)
		BottomC[workernum] = response.Bottom
		TopC[workernum] = response.Top
		current = response.CurrentTurn
		Current[workernum] = current
		fmt.Println(current)
		if Alive[workernum] == true {
			AliveCells += len(response.Alive)
			Alive[workernum] = false
		}

	}
	worker.Close()
	fmt.Println(current, turn, Currentturn)
	if current == turn {
		fmt.Println(response.World)
		workerChannels <- response.World
		AliveChannels <- response.Alive
	}

	return
}

func reconstruct(workerChannels []chan [][]byte) {
	tempWorld := make([][]byte, 0)
	for i := range workerChannels { // collects the resulting parts into a single 2D slice
		workerResults := <-workerChannels[i]
		tempWorld = append(tempWorld, workerResults...)
	}
	FinalWorld = tempWorld
	fmt.Println(len(FinalWorld))
	done = true
}

func calculateAliveCells(world [][]byte, height, width int) int {
	AliveCells = 0
	for h := 0; h < height; h++ {
		for w := 0; w < width; w++ {
			if world[h][w] == alive {
				AliveCells++
			}
		}
	}
	return AliveCells
}

//to connect to client
func main() {
	//open port
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
