package stubs

var Nextworld = "NextStateOperation.Distributor"

type Response struct {
	Message [][]uint8
}

type Request struct {
	Message [][]uint8
	Threads int
	Turns   int
}
