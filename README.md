# Game of Life Coursework

## Task Overview

The British mathematician John Horton Conway devised a cellular automaton named ‘The Game of Life’. The game resides on a 2-valued 2D matrix, i.e. a binary image, where the cells can either be ‘alive’ (pixel value 255 - white) or ‘dead’ (pixel value 0 - black). The game evolution is determined by its initial state and requires no further input. Every cell interacts with its eight neighbour pixels: cells that are horizontally, vertically, or diagonally adjacent. At each matrix update in time the following transitions may occur to create the next evolution of the domain:

- any live cell with fewer than two live neighbours dies
- any live cell with two or three live neighbours is unaffected
- any live cell with more than three live neighbours dies
- any dead cell with exactly three live neighbours becomes alive

Consider the image to be on a closed domain (pixels on the top row are connected to pixels at the bottom row, pixels on the right are connected to pixels on the left and vice versa). A user can only interact with the Game of Life by creating an initial configuration and observing how it evolves. Note that evolving such complex, deterministic systems is an important application of scientific computing, often making use of parallel architectures and concurrent programs running on large computing farms.

Our task was to design and implement programs which simulate the Game of Life on an image matrix:

- **Parallel Implementation:**  write code to evolve Game of Life using multiple worker goroutines on a single machine. 
- **Distributed Implementation:** create an implementation that uses a number of AWS nodes to cooperatively calculate the new state of the Game of Life board, and communicate state between machines over a network. 

## Usage

To run tests:
`go test -v`

To run program:
`go run .`

### Controls for parallel  implementation

- If `s` is pressed, generate a PGM file with the current state of the board.
- If `q` is pressed, generate a PGM file with the current state of the board and then terminate the program. Your program should *not* continue to execute all turns set in `gol.Params.Turns`.
- If `p` is pressed, pause the processing and print the current turn that is being processed. If `p` is pressed again resume the processing and print `"Continuing"`. It is *not* necessary for `q` and `s` to work while the execution is paused.


### Controls for distributed  implementation

- If `s` is pressed, the controller should generate a PGM file with the current state
  of the board.
- If `q` is pressed, close the controller client program without causing an
  error on the GoL server. A new controller should be able to take over
interaction with the GoL engine.
- If `p` is pressed, pause the processing *on the AWS node* and have the
  *controller* print the current turn that is being processed. If `p` is pressed
again resume the processing and have the controller print `"Continuing"`. It is
*not* necessary for `q` and `s` to work while the execution is paused.
- If `k` is pressed, all components of the distributed system are shut down cleanly, and the
system outputs a PGM image of the latest state.
