package main

import (
	"flag"
	"log"
	"net"
	"net/rpc"
	"os"
	"time"

	"uk.ac.bris.cs/gameoflife/stubs"
)

var FinalWorld [][]byte
var CurrentWorld [][]byte
var AliveCells int
var Currentturn int
var done bool
var key bool
var pause bool
var Waddress []string
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
	width := len(world[0])
	rem := mod(height, len(Waddress))
	splitThreads := height / len(Waddress)

	for turn := Currentturn; turn <= turns; turn++ {
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
				go CallWorker(world, startY, endY, workerChannels[i], Waddress[i])
				if quit == true {
					os.Exit(0)
				}
			}

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
	}
	FinalWorld = world
	done = true
}

//InitialState : Initial state of the world
func (s *NextStateOperation) InitialState(req stubs.Request, res *stubs.Response) (err error) {
	// fmt.Println("Gamestate initialised")
	World := req.Message
	Turn := req.Turns
	quit = false
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
	// fmt.Println("Final Gamestate returned")
	for done == false {
		//
	}
	res.Message = FinalWorld
	return
}

//Alive : Return current World + Turn for counting alive cells
func (s *NextStateOperation) Alive(req stubs.Request, res *stubs.Response) (err error) {
	// fmt.Println("Return num of alive cells")
	res.Turn = Currentturn
	res.AliveCells = AliveCells
	return
}

// DoKeypresses : function for keypresses
func (s *NextStateOperation) DoKeypresses(req stubs.Request, res *stubs.Response) (err error) {
	// fmt.Println("Return num of alive cells")
	res.Turn = Currentturn
	res.Message = CurrentWorld
	key = true
	pause = req.Pause
	return
}

//Quit closes all instances
func (s *NextStateOperation) Quit(req stubs.Request, res *stubs.Response) (err error) {
	quit = true
	time.Sleep(2 * time.Second)
	os.Exit(0)
	return
}

//GetAddress gets address of worker node
func (s *NextStateOperation) GetAddress(req stubs.ReqAddress, res *stubs.ResAddress) (err error) {
	Waddress = append(Waddress, req.WorkerAddress)
	return
}

//CallWorker creates connection to worker node
func CallWorker(world [][]byte, startingY, endingY int, workerChannels chan<- [][]byte, address string) (err error) {
	worker, err := rpc.Dial("tcp", address)
	if err != nil {
		log.Fatal("Dial error:", err)
		return err
	}

	go func() {
		for {
			if quit {
				request := stubs.ReqWorker{}
				response := new(stubs.ResWorker)
				worker.Call(stubs.QuitW, request, response)
				worker.Close()
			}
		}
	}()

	request := stubs.ReqWorker{World: world, StartY: startingY, EndY: endingY}
	response := new(stubs.ResWorker)

	worker.Call(stubs.CalculateNextState, request, response)

	workerChannels <- response.World
	worker.Close()
	// fmt.Println("worker close")

	return
}

//to connect to gol
func main() {
	pAddr := flag.String("port", ":8030", "Port to listen on")
	flag.Parse()
	// rand.Seed(time.Now().UnixNano())
	rpc.Register(&NextStateOperation{})
	listener, err := net.Listen("tcp", *pAddr)
	if err != nil {
		log.Fatal("listen error:", err)
	}

	rpc.Accept(listener)
	listener.Close()

}
