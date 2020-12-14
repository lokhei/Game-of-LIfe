package stubs

import (
	"uk.ac.bris.cs/gameoflife/util"
)

// stub procedures

var CallInitial = "NextStateOperation.InitialState"
var CallReturn = "NextStateOperation.FinalState"
var CallAlive = "NextStateOperation.Alive"
var CallDoKeypresses = "NextStateOperation.DoKeypresses"
var Quit = "NextStateOperation.Quit"
var GetAddress = "NextStateOperation.GetAddress"

var CalculateNextState = "Worker.CalculateNextState"
var QuitW = "Worker.QuitW"

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
	Top    []byte
	Bottom []byte
	StartY int
	EndY   int
}

type ResWorker struct {
	World [][]byte
	Alive []util.Cell
}

type ReqAddress struct {
	WorkerAddress string
}

type ResAddress struct {
}
