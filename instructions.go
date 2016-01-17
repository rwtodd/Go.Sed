package main

import (
	"fmt"
	"io"
)

// mapping: sed --> cmd

//   n    -->   print ; fill_next     (or if -n is on:  fill_next)
//   d    -->   branch(0)
//   x    -->   swap
//   p    -->   print
//   b tgt -->  branch(tgt)   (just 'b' is a branch to 1.. since 0 is always fill_next)

//   conditional:
//   - numeric
//   - regexp
//   - EOF ('$')
//   - range_conditional

//  program 'pgm' is transformed into:
//      fill_next ; { pgm } ; print ; branch(0)
//  ... or if (-n) is on:
//      fill_next ; { pgm } ; branch(0)

// ---------------------------------------------------
type cmd_swap struct{}

func (_ cmd_swap) run(e *engine) error {
	e.pat, e.hold = e.hold, e.pat
	e.ip++
	return nil
}

// ---------------------------------------------------
type cmd_branch struct {
	target int
}

func (b *cmd_branch) run(e *engine) error {
	e.ip = b.target
	return nil
}

// ---------------------------------------------------
type cmd_print struct{}

func (_ cmd_print) run(e *engine) error {
	e.ip++
	_, err := e.output.WriteString(e.pat)
	return err
}

// ---------------------------------------------------
type cmd_lineno struct{}

func (_ cmd_lineno) run(e *engine) error {
	e.ip++
	var lineno = fmt.Sprintf("%d\n", e.lineno)
	_, err := e.output.WriteString(lineno)
	return err
}

// ---------------------------------------------------
type cmd_fillnext struct{}

func (_ cmd_fillnext) run(e *engine) error {
	if e.lastl {
		return io.EOF
	}

	e.ip++
	e.pat = e.nxtl
	e.lineno++

	var err error
	e.nxtl, err = e.input.ReadString('\n')
	if err == io.EOF {
		if len(e.nxtl) == 0 {
			e.lastl = true
		}
		err = nil
	}

	return err
}
