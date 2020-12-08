package stubs

import (
	"uk.ac.bris.cs/gameoflife/util"
)

var CallInitial = "NextStateOperation.InitialState"
var CallReturn = "NextStateOperation.FinalState"
var CallAlive = "NextStateOperation.Alive"
var CallDoKeypresses = "NextStateOperation.DoKeypresses"
var Quit = "NextStateOperation.Quit"
var GetAddress = "NextStateOperation.GetAddress"
var GetCAddress = "NextStateOperation.GetCAddress"

var CalculateNextState = "Worker.CalculateNextState"
var QuitW = "Worker.QuitW"

var SdlEvent = "Sdl.SdlEvent"

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
	Reset    bool
}

type ReqWorker struct {
	World  [][]byte
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
	// ErrorMessage string
}

type SDLRes struct {
}

type SDLReq struct {
	Alive []util.Cell
	Turn  int
}
