package stubs

var CallInitial = "NextStateOperation.InitialState"
var CallReturn = "NextStateOperation.FinalState"
var CallAlive = "NextStateOperation.Alive"
var CallDoKeypresses = "NextStateOperation.DoKeypresses"
var CalculateNextState = "Worker.CalculateNextState"
var GetAddress = "NextStateOperation.GetAddress"

type Response struct {
	AliveCells int
	Message    [][]byte
	Turn       int
	Done       bool
}

type Request struct {
	Message  [][]byte
	Threads  int
	Turns    int
	Keypress rune
	Pause    bool
}

type ReqWorker struct {
	World  [][]byte
	StartY int
	EndY   int
}

type ResWorker struct {
	World [][]byte
}

type ReqAddress struct {
	WorkerAddress string
}

type ResAddress struct {
	// WorkerAddress string
}
