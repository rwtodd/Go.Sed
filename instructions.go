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
type cmd_hold struct{}

func (_ cmd_hold) run(e *engine) error {
	e.hold = e.pat
	e.ip++
	return nil
}

// ---------------------------------------------------
type cmd_holdapp struct{}

func (_ cmd_holdapp) run(e *engine) error {
	// FIXME make this more performant one day
	if len(e.hold) > 0 {
		e.hold += "\n"
	}
	e.hold += e.pat
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
	if err == nil {
		err = e.output.WriteByte('\n')
	}
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

	e.nxtl = ""

	var prefix = true
	var err error
	var line []byte

	for prefix {
		line, prefix, err = e.input.ReadLine()
		if err != nil {
			break
		}
		buf := make([]byte, len(line))
		copy(buf, line)
		e.nxtl += string(buf)
	}

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
	cond     condition // the condition to check
	metloc   int       // where to jump if the condition is met
	unmetloc int       // where to jump if the condition is not met
}

func (c *cmd_simplecond) run(e *engine) error {
	if c.cond.isMet(e) {
		e.ip = c.metloc
	} else {
		e.ip = c.unmetloc
	}
	return nil
}

func (c *cmd_simplecond) invert() {
	c.metloc, c.unmetloc = c.unmetloc, c.metloc
}

// --------------------------------------------------
type cmd_twocond struct {
	start    condition // the condition that begines the block
	end      condition // the condition that ends the block
	metloc   int       // where to jump if the condition is met
	unmetloc int       // where to jump if the condition is not met
	isOn     bool      // are we active already?
	offFrom  int       // if we say the end condition, what line was it on?
}

func newTwoCond(c1 condition, c2 condition, metloc int, unmetloc int) instruction {
	return &cmd_twocond{c1, c2, metloc, unmetloc, false, 0}
}

func (c *cmd_twocond) run(e *engine) error {
	if c.isOn && (c.offFrom > 0) && (c.offFrom < e.lineno) {
		c.isOn = false
		c.offFrom = 0
	}

	if !c.isOn {
		if c.start.isMet(e) {
			e.ip = c.metloc
			c.isOn = true
		} else {
			e.ip = c.unmetloc
		}
	} else {
		if (c.offFrom == e.lineno) || (c.end.isMet(e)) {
			c.offFrom = e.lineno
		}
		e.ip = c.metloc
	}
	return nil
}

func (c *cmd_twocond) invert() {
	c.metloc, c.unmetloc = c.unmetloc, c.metloc
}
