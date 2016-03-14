package sed // import "go.waywardcode.com/sed"

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

func cmd_quit(e *Engine) error {
	return io.EOF
}

// ---------------------------------------------------
func cmd_swap(e *Engine) error {
	e.pat, e.hold = e.hold, e.pat
	e.ip++
	return nil
}

// ---------------------------------------------------
func cmd_get(e *Engine) error {
	e.pat = e.hold
	e.ip++
	return nil
}

// ---------------------------------------------------
func cmd_hold(e *Engine) error {
	e.hold = e.pat
	e.ip++
	return nil
}

// ---------------------------------------------------
func cmd_getapp(e *Engine) error {
	e.pat = strings.Join([]string{e.pat, e.hold}, "\n")
	e.ip++
	return nil
}

// ---------------------------------------------------
func cmd_holdapp(e *Engine) error {
	e.hold = strings.Join([]string{e.hold, e.pat}, "\n")
	e.ip++
	return nil
}

// ---------------------------------------------------
// newBranch generates branch instructions with specific
// targets
func cmd_newBranch(target int) instruction {
	return func(e *Engine) error {
		e.ip = target
		return nil
	}
}

// ---------------------------------------------------
// newChangedBranch generates branch instructions with specific
// targets that only trigger on modified pattern spaces
func cmd_newChangedBranch(target int) instruction {
	return func(e *Engine) error {
		if e.modified {
			e.ip = target
			e.modified = false
		} else {
			e.ip++
		}
		return nil
	}
}

// ---------------------------------------------------
func cmd_print(e *Engine) (err error) {
	e.ip++

	_, err = e.output.WriteString(e.pat)
	if err == nil {
		err = e.output.WriteByte('\n')
	}
	return err
}

// ---------------------------------------------------
func cmd_printFirstLine(e *Engine) (err error) {
	e.ip++

	idx := strings.IndexRune(e.pat, '\n')

	if idx == -1 {
		idx = len(e.pat)
	}

	_, err = e.output.WriteString(e.pat[:idx])
	if err == nil {
		err = e.output.WriteByte('\n')
	}
	return err
}

// ---------------------------------------------------
func cmd_deleteFirstLine(e *Engine) (err error) {
	idx := strings.IndexRune(e.pat, '\n')

	if idx == -1 {
		e.pat = ""
		e.ip = 0 // go back and fillNext
	} else {
		e.pat = e.pat[idx+1:]
		e.ip = 1 // restart, but skip filling
	}

	return nil
}

// ---------------------------------------------------
func cmd_lineno(e *Engine) error {
	e.ip++
	var lineno = fmt.Sprintf("%d\n", e.lineno)
	_, err := e.output.WriteString(lineno)
	return err
}

// ---------------------------------------------------
func cmd_fillNext(e *Engine) error {
	var err error

	// first, put out any stored-up 'a\'ppended text:
	if e.appl != nil {
		_, err = e.output.WriteString(*e.appl)
		e.appl = nil
		if err != nil {
			return err
		}
	}

	// just return if we're at EOF
	if e.lastl {
		return io.EOF
	}

	// otherwise, copy nxtl to the pattern space and
	// refill.
	e.ip++

	e.pat = e.nxtl
	e.lineno++
	e.modified = false

	var prefix = true
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

func cmd_fillNextAppend(e *Engine) error {
	var lines = make([]string, 2)
	lines[0] = e.pat
	err := cmd_fillNext(e) // increments e.ip, so we don't
	lines[1] = e.pat
	e.pat = strings.Join(lines, "\n")
	return err
}

// --------------------------------------------------

type cmd_simplecond struct {
	cond     condition // the condition to check
	metloc   int       // where to jump if the condition is met
	unmetloc int       // where to jump if the condition is not met
}

func (c *cmd_simplecond) run(e *Engine) error {
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
func (c *cmd_twocond) isLastLine(e *Engine) bool {
	return c.isOn && (c.offFrom == e.lineno)
}

func (c *cmd_twocond) run(e *Engine) error {
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
func cmd_newChanger(text string, guard *cmd_twocond) instruction {
	return func(e *Engine) error {
		e.ip = 0 // go to the the next cycle

		var err error
		if (guard == nil) || guard.isLastLine(e) {
			_, err = e.output.WriteString(text)
		}
		return err
	}
}

// --------------------------------------------------
func cmd_newAppender(text string) instruction {
	return func(e *Engine) error {
		e.ip++
		if e.appl == nil {
			e.appl = &text
		} else {
			var newstr = *e.appl + text
			e.appl = &newstr
		}
		return nil
	}
}

// --------------------------------------------------
func cmd_newInserter(text string) instruction {
	return func(e *Engine) error {
		e.ip++
		_, err := e.output.WriteString(text)
		return err
	}
}

// --------------------------------------------------
// The 'r' command is basically and 'a\' with the contents
// of a file. I implement it literally that way below.
func cmd_newReader(filename string) (instruction, error) {
	bytes, err := ioutil.ReadFile(filename)
	return cmd_newAppender(string(bytes)), err
}

// --------------------------------------------------
// The 'w' command appends the current pattern space
// to the named file.  In this implementation, it opens
// the file for appending, writes the file, and then
// closes the file.  This appears to be consistent with
// what OS X sed does.
func cmd_newWriter(filename string) instruction {
	return func(e *Engine) error {
		e.ip++
		f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err == nil {
			defer f.Close()
			_, err = f.WriteString(e.pat)
		}
		if err == nil {
			_, err = f.WriteString("\n")
		}
		return err
	}
}
