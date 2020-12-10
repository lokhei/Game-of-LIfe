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

var quit bool

const alive = 255
const dead = 0

func mod(x, m int) int {

	if x == m {
		return 0
	} else if x == -1 {
		return m - 1
	} else {
		return x
	}
}

//helper function that attempts to determine this process' IP address.
func getOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal("listen error:", err)
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr).IP.String()
	return localAddr
}

type Worker struct{}

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

//Quitw closes workers
func (w *Worker) QuitW(req stubs.ReqWorker, res *stubs.ResWorker) (err error) {
	quit = true
	time.Sleep(10 * time.Millisecond)

	os.Exit(0)
	return
}

//CalculateNextState takes the current state of the world and completes one evolution of the world. It then returns the result.
func (w *Worker) CalculateNextState(req stubs.ReqWorker, res *stubs.ResWorker) (err error) {
	//makes a new world
	startY := req.StartY
	endY := req.EndY
	width := len(req.World[0])
	height := len(req.World)
	world := req.World
	// res.Alive :

	newWorld := make([][]byte, endY-startY)
	for i := range newWorld {
		newWorld[i] = make([]byte, width)
	}

	//loop through turns
	//after each turn, send back world[1] and world[-2]
	//and retreive world[0] and world[-1]
	//sets cells to dead or alive according to num of neighbours
	for y := startY; y < endY; y++ {
		for x := 0; x < width; x++ {
			neighbours := calculateNeighbours(height, width, x, y, world)
			if world[y][x] == alive {
				if neighbours == 2 || neighbours == 3 {
					newWorld[y-startY][x] = alive
				} else {
					newWorld[y-startY][x] = dead
					res.Alive = append(res.Alive, util.Cell{x, y})
				}
			} else {
				if neighbours == 3 {
					newWorld[y-startY][x] = alive
					res.Alive = append(res.Alive, util.Cell{x, y})
				} else {
					newWorld[y-startY][x] = dead
				}
			}
			if quit {
				res.World = world
				return
			}
		}
	}

	res.World = newWorld
	return
}

//if called by worker, return whole subworld

func main() {
	//pAddr - works as server
	pAddr := flag.String("port", ":8050", "Port to listen on")
	logicAddr := flag.String("logic", "127.0.0.1:8030", "Address of logic instance")
	flag.Parse()
	client, err := rpc.Dial("tcp", *logicAddr)

	rpc.Register(&Worker{})
	listener, err := net.Listen("tcp", *pAddr)
	if err != nil {
		log.Fatal("listen error:", err)
	}

	status := new(stubs.ResAddress)
	client.Call(stubs.GetAddress, stubs.ReqAddress{WorkerAddress: getOutboundIP() + *pAddr}, status)

	client.Close()
	// fmt.Println("client close")

	rpc.Accept(listener)
	listener.Close()
	// fmt.Println("listener worker close")

	flag.Parse()

}
