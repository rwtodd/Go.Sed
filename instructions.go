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

// --------------------------------------------------
type cmd_simplecond struct {
	cond condition // the condition to check
	loc  int       // where to jump if the condition is not met
}

func (c *cmd_simplecond) run(e *engine) error {
	if c.cond.isMet(e) {
		e.ip++
	} else {
		e.ip = c.loc
	}
	return nil
}

// --------------------------------------------------
type cmd_twocond struct {
	start   condition // the condition that begines the block
	end     condition // the condition that ends the block
	loc     int       // where to jump if the condition is not met
	isOn    bool      // are we active already?
	offFrom int       // if we say the end condition, what line was it on?
}

func newTwoCond(c1 condition, c2 condition, loc int) instruction {
	return &cmd_twocond{c1, c2, loc, false, 0}
}

func (c *cmd_twocond) run(e *engine) error {
	if c.isOn && (c.offFrom > 0) && (c.offFrom < e.lineno) {
		c.isOn = false
		c.offFrom = 0
	}

	if !c.isOn {
		if c.start.isMet(e) {
			e.ip++
			c.isOn = true
		} else {
			e.ip = c.loc
		}
	} else {
		if (c.offFrom == e.lineno) || (c.end.isMet(e)) {
			c.offFrom = e.lineno
		}
		e.ip++
	}
	return nil
}
