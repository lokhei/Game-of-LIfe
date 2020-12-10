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

var quit bool
var world [][]byte
var bottom []byte
var top []byte
var turn int
var newWorld [][]byte
var AliveCells []util.Cell

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
func calculateNeighbours(height, width, x, y int, bottom, top []byte, world [][]byte) int {
	neighbours := 0
	for i := -1; i <= 1; i++ {
		for j := -1; j <= 1; j++ {
			if i != 0 || j != 0 { //not [y][x]
				if y == 0 && i == -1 {
					if bottom[mod(x+j, width)] == alive {
						neighbours++
					}
				} else if y == height-1 && i == 1 {
					if top[mod(x+j, width)] == alive {
						neighbours++
					}
				} else {
					if world[mod(y+i, height)][mod(x+j, width)] == alive {
						neighbours++
					}
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

func computeTurns(height, width, totalTurns, currentTurn int, world [][]byte, top, bottom []byte) {
	newWorld = make([][]byte, height)
	for i := range newWorld {
		newWorld[i] = make([]byte, width)
	}
	turn = currentTurn
	AliveCells = make([]util.Cell, 0)
	//loop through turns
	//after each turn, send back world[1] and world[-2]
	//and retreive world[0] and world[-1]
	//sets cells to dead or alive according to num of neighbours
	// for turn := currentTurn; turn <= totalTurns; turn++ {
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			neighbours := calculateNeighbours(height, width, x, y, bottom, top, world)
			if world[y][x] == alive {
				if neighbours == 2 || neighbours == 3 {
					newWorld[y][x] = alive
					AliveCells = append(AliveCells, util.Cell{X: x, Y: y})

				} else {
					newWorld[y][x] = dead
				}
			} else {
				if neighbours == 3 {
					newWorld[y][x] = alive
					AliveCells = append(AliveCells, util.Cell{X: x, Y: y})
				} else {
					newWorld[y][x] = dead
				}
			}
			if quit {
				// res.World = world
				return
			}
		}
	}
	bottom = newWorld[0]
	top = newWorld[height-1]
	turn++
	// }
}

//CalculateNextState takes the current state of the world and completes one evolution of the world. It then returns the result.
func (w *Worker) CalculateNextState(req stubs.ReqWorker, res *stubs.ResWorker) (err error) {

	// fmt.Println(req.CurrentTurn)

	// fmt.Println(len(req.World[0]))

	width := len(req.World[0])
	// fmt.Println(req.CurrentTurn)
	height := len(req.World)
	world := req.World
	top = req.Top
	bottom = req.Bottom
	totalTurns := req.Turns
	currentTurn := req.CurrentTurn

	computeTurns(height, width, totalTurns, currentTurn, world, top, bottom)

	// res.Bottom = newWorld[0]

	// res.Top = newWorld[height-1]

	// turn++
	fmt.Println(turn, totalTurns)
	if turn == totalTurns {
		res.CurrentTurn = turn
		res.World = newWorld
		return
	}
	if turn == currentTurn+1 {
		res.Bottom = bottom
		res.Top = top
		res.CurrentTurn = turn

	}

	// if turn != totalTurns {
	// 	return
	// }
	// res.World = newWorld
	return
}

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

	rpc.Accept(listener)
	listener.Close()

	flag.Parse()

}
