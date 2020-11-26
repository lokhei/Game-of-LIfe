package stubs

// import "fmt"

// type Event interface {
// 	fmt.Stringer
// 	GetCompletedTurns() int
// }

var Nextworld = "NextStateOperation.Distributor"

type Response struct {
	Message [][]uint8
}

type Request struct {
	Message [][]uint8
	Threads int
	Turns   int
	// Events  chan<- Event
}
