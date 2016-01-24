package main

import (
	"fmt"
	"io"
	"strings"
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
func cmd_swap(e *engine) error {
	e.pat, e.hold = e.hold, e.pat
	e.ip++
	return nil
}

// ---------------------------------------------------
func cmd_hold(e *engine) error {
	e.hold = e.pat
	e.ip++
	return nil
}

// ---------------------------------------------------
func cmd_holdapp(e *engine) error {
	var lines = make([]string, 0, 2)

	if e.hold != nil {
		lines = append(lines, *e.hold)
	}
	if e.pat != nil {
		lines = append(lines, *e.pat)
	}

	newhold := strings.Join(lines, "\n")
	e.hold = &newhold

	e.ip++
	return nil
}

// ---------------------------------------------------
// newBranch generates branch instructions with specific
// targets
func cmd_newBranch(target int) instruction {
	return func(e *engine) error {
		e.ip = target
		return nil
	}
}

// ---------------------------------------------------
func cmd_print(e *engine) (err error) {
	e.ip++

	// FIXME check if real sed puts a newline when pattern space is empty
	//   like   " g ; p "
	if e.pat != nil {
		_, err = e.output.WriteString(*e.pat)
		if err == nil {
			err = e.output.WriteByte('\n')
		}
	}
	return err
}

// ---------------------------------------------------
func cmd_lineno(e *engine) error {
	e.ip++
	var lineno = fmt.Sprintf("%d\n", e.lineno)
	_, err := e.output.WriteString(lineno)
	return err
}

// ---------------------------------------------------
func cmd_fillnext(e *engine) error {
	if e.lastl {
		return io.EOF
	}

	e.ip++

	patstring := e.nxtl // make a copy
	e.pat = &patstring
	e.lineno++

	var prefix = true
	var err error
	var line []byte

	var lines []string

	for prefix {
		line, prefix, err = e.input.ReadLine()
		if err != nil {
			break
		}
		buf := make([]byte, len(line))
		copy(buf, line)
		lines = append(lines, string(buf))
	}

	e.nxtl = strings.Join(lines, "")

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

// --------------------------------------------------
type cmd_twocond struct {
	start    condition // the condition that begines the block
	end      condition // the condition that ends the block
	metloc   int       // where to jump if the condition is met
	unmetloc int       // where to jump if the condition is not met
	isOn     bool      // are we active already?
	offFrom  int       // if we saw the end condition, what line was it on?
}

func newTwoCond(c1 condition, c2 condition, metloc int, unmetloc int) *cmd_twocond {
	return &cmd_twocond{c1, c2, metloc, unmetloc, false, 0}
}

// isLastLine is here to support multi-line "c\" commands.
// The command needs to know when it's the end of the
// section so it can do the replacement.
func (c *cmd_twocond) isLastLine(e *engine) bool {
	return c.isOn && (c.offFrom == e.lineno)
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
		if c.end.isMet(e) {
			c.offFrom = e.lineno
		}
		e.ip = c.metloc
	}
	return nil
}

// --------------------------------------------------
type cmd_change struct {
	guard *cmd_twocond
	text  string
}

func (c *cmd_change) run(e *engine) error {
	e.ip = 0 // go to the the next cycle

	var err error
	if (c.guard == nil) || c.guard.isLastLine(e) {
		_, err = e.output.WriteString(c.text)
	}
	return err
}
