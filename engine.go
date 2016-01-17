package main

import (
	"bufio"
	"io"
)

// engine is the main program state
type engine struct {
	nxtl   string        // the next line
	lastl  bool          // true if it's the last line
	pat    string        // the pattern space
	hold   string        // the hold buffer
	ins    []instruction // the instruction stream
	ip     int           // the current locaiton in the instruction stream
	input  *bufio.Reader // the input stream
	output *bufio.Writer // the output stream
	lineno int           // current line number
}

// a sed instruction is mostly a function transforming an engine
type instruction interface {
	run(e *engine) error
}

// Run executes the instructions until we hit an error.  The most
// common "error" will be io.EOF, which we will translate to nil
func run(e *engine) error {
	var err error

	// prime the engine by filling nxtl... roll back the IP and lineno
	err = cmd_fillnext{}.run(e)
	e.ip = 0
	e.lineno = 0

	for err == nil {
		err = e.ins[e.ip].run(e)
	}

	if err == io.EOF {
		err = nil
	}

	ferr := e.output.Flush() // attempt to flush output
	if ferr != nil && err == nil {
		err = ferr
	}

	return err
}
