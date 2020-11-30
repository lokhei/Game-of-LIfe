package stubs

// import "fmt"

// type Event interface {
// 	fmt.Stringer
// 	GetCompletedTurns() int
// }

var CallInitial = "NextStateOperation.InitialState"
var CallReturn = "NextStateOperation.FinalState"
var CallAlive = "NextStateOperation.Alive"
var CallDoKeypresses = "NextStateOperation.DoKeypresses"


type Response struct {
	AliveCells int
	Message    [][]uint8
	Turn       int
	Done       bool
}

type Request struct {
	Message  [][]uint8
	Threads  int
	Turns    int
	Keypress rune
	// Events  chan<- Event
}
