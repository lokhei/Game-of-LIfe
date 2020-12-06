package main

import (
	"flag"
	"log"
	"net"
	"net/rpc"

	"uk.ac.bris.cs/gameoflife/stubs"
)

const alive = 255
const dead = 0

func mod(x, m int) int {
	return (x + m) % m
}

//helper function that attempts to determine this process' IP address.
func getOutboundIP() string {
	conn, _ := net.Dial("udp", "8.8.8.8:80")
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

//CalculateNextState takes the current state of the world and completes one evolution of the world. It then returns the result.
func (w *Worker) CalculateNextState(req stubs.ReqWorker, res *stubs.ResWorker) (err error) {
	//makes a new world
	startY := req.StartY
	endY := req.EndY
	width := len(req.World[0])
	height := len(req.World)
	world := req.World

	newWorld := make([][]byte, endY-startY)
	for i := range newWorld {
		newWorld[i] = make([]byte, width)
	}
	//sets cells to dead or alive according to num of neighbours
	for y := startY; y < endY; y++ {
		for x := 0; x < width; x++ {
			neighbours := calculateNeighbours(height, width, x, y, world)
			if world[y][x] == alive {
				if neighbours == 2 || neighbours == 3 {
					newWorld[y-startY][x] = alive
				} else {
					newWorld[y-startY][x] = dead
				}
			} else {
				if neighbours == 3 {
					newWorld[y-startY][x] = alive
				} else {
					newWorld[y-startY][x] = dead
				}
			}
		}
	}
	res.World = newWorld
	return
}

func main() {
	//pAddr - works as server
	pAddr := flag.String("port", ":8050", "Port to listen on")
	logicAddr := flag.String("logic", "127.0.0.1:8030", "Address of logic instance")
	flag.Parse()
	client, err := rpc.Dial("tcp", *logicAddr)

	// rand.Seed(time.Now().UnixNano())
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
