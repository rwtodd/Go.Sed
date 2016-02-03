// Package sed implements the classic UNIX sed language in pure Go.
// The interface is very simple: a user compiles a program into an
// execution engine by calling New or NewQuiet. Then, the engine
// is run from an input to an output via Run().
//
// All classic sed commands are supported, but since the package
// uses Go's regexp package for the regular expressions, the syntax
// for regexps will not be the same as a typical UNIX sed.  In other
// words, instead of:  s|ab\(c*\)d|\1|g  you would say: s|ab(c*)d|$1|g.
// So this is a Go-flavored sed, rather than a drop-in replacement for
// a UNIX sed.  Depending on your tastes, you will either consider this
// an improvement or completely brain-dead.
package sed

import (
	"bufio"
	"io"
)

// Engine is the VM state for the sed program. It is the main type that
// users of the go-sed library will interact with.
type Engine struct {
	nxtl     string        // the next line
	pat      string        // the pattern space, possibly nil
	hold     string        // the hold buffer,   possibly nil
	appl     *string       // any lines we've been asked to 'a\'ppend, usually nil
	lastl    bool          // true if it's the last line
	ins      []instruction // the instruction stream
	ip       int           // the current locaiton in the instruction stream
	input    *bufio.Reader // the input stream
	output   *bufio.Writer // the output stream
	lineno   int           // current line number
	modified bool          // have we modified the pattern space?
}

// a sed instruction is mostly a function transforming an engine
type instruction func(*Engine) error

// makeEngine is the logic behine the New and NewQuiet public functions.
// It lexes and parses the program, and makes a new Engine out of it.
func makeEngine(program *bufio.Reader, isQuiet bool) (*Engine, error) {
	ch := make(chan *token, 128)
	errch := make(chan error, 1)
	go lex(program, ch, errch)

	instructions, parseErr := parse(ch, isQuiet)
	var err = <-errch // look for lexing errors first...
	if err == nil {
		// if there were no lex errors, look for a parsing error
		err = parseErr
	}

	if err != nil {
		// we had an error, so return a nil engine
		return nil, err
	}

	return &Engine{ins: instructions}, nil
}

// New creates a new sed engine from a program.  The program is executed
// via the Run method. If the provided program has any errors, the returned
// engine will be 'nil' and the error will be returned.  Otherwise, the returned
// error will be nil.
func New(program *bufio.Reader) (*Engine, error) {
	return makeEngine(program, false)
}

// NewQuiet creates a new sed engine from a program.  It behaves exactly as
// New(), except it produces an engine that doesn't print lines by defualt. This
// is the classic '-n' sed behaviour.
func NewQuiet(program *bufio.Reader) (*Engine, error) {
	return makeEngine(program, true)
}

// Run executes the program embodied by the Engine on the given
// input, with output going to the given output. Both must be
// buffered, which should not be a big inconvenience since bufio
// can wrap any io.Reader or io.Writer.  Still, perhaps in a later
// release, more convenience functions can be added.
//
// Any errors encountered during the run will be returned to the caller.
func (e *Engine) Run(input *bufio.Reader, output *bufio.Writer) error {
	var err error

	// prime the engine by resetting the internal flags and filling nxtl...
	*e = Engine{ins: e.ins, input: input, output: output}
	err = cmd_fillNext(e)

	// roll back the IP and lineno
	e.ip = 0
	e.lineno = 0

	// run the program
	for err == nil {
		err = e.ins[e.ip](e)
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
