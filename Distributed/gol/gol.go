package gol

import (
	"flag"
)

// Params provides the details of how to run the Game of Life and which image to load.
type Params struct {
	Turns       int
	Threads     int
	ImageWidth  int
	ImageHeight int
}

// Run starts the processing of Game of Life. It should initialise channels and goroutines.
func Run(p Params, events chan<- Event, keyPresses <-chan rune) {

	var server string

	if flag.Lookup("server") == nil {
		serverTemp := flag.String("server", "127.0.0.1:8030", "IP:port string to connect to as server")
		flag.Parse()
		server = *serverTemp
	} else {
		server = flag.Lookup("server").Value.(flag.Getter).Get().(string)
	}

	ioCommand := make(chan ioCommand)
	ioIdle := make(chan bool)
	filename := make(chan string)
	input := make(chan uint8)
	output := make(chan uint8)

	ioChannels := ioChannels{
		command:  ioCommand,
		idle:     ioIdle,
		filename: filename,
		output:   output,
		input:    input,
	}
	go startIo(p, ioChannels)

	go makeCall(keyPresses, server, events, p, filename, input, output, ioCommand, ioIdle)

}
