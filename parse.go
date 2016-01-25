package main

import (
	"fmt"
	"regexp"
	"strconv"
)

// these functions parse the lex'ed tokens (lex.go) and
// build a program for the engine (engine.go) to run.

var zeroBranch = cmd_newBranch(0)

type waitingBranch struct {
	ip     int       // address of the branch to fix up
	label  string    // the target label
	letter rune      // 'b' or 't' branch
	loc    *location // the original parse location
}

const (
	END_OF_PROGRAM_LABEL = "the end" // has a space... no conflicts with user labels
)

type parseState struct {
	toks       <-chan *token          // our input
	ins        []instruction          // the compiled instructions
	branches   []waitingBranch        // references to fix up
	b_labels   map[string]instruction // named b branch labels
	t_labels   map[string]instruction // named t branch labels
	blockLevel int                    // how deeply nested are our blocks?
	err        error                  // record any errors we encounter
}

func parse(input <-chan *token) ([]instruction, error) {
	ps := &parseState{toks: input, b_labels: make(map[string]instruction), t_labels: make(map[string]instruction)}

	ps.ins = append(ps.ins, cmd_fillNext)
	parse_toplevel(ps)
	if ps.err == nil && ps.blockLevel > 0 {
		ps.err = fmt.Errorf("It looks like you are missing a closing brace!")
	}

	// if the parsing failed in some way, just give up now
	if ps.err != nil {
		return nil, ps.err
	}

	ps.b_labels[END_OF_PROGRAM_LABEL] = cmd_newBranch(len(ps.ins))
	ps.t_labels[END_OF_PROGRAM_LABEL] = cmd_newChangedBranch(len(ps.ins))
	if !noPrint {
		ps.ins = append(ps.ins, cmd_print)
	}
	ps.ins = append(ps.ins, zeroBranch)
	parse_resolveBranches(ps)

	return ps.ins, ps.err
}

func parse_resolveBranches(ps *parseState) {
	waiting := ps.branches
	for idx := range waiting {
		var (
			ins instruction
			ok  bool
		)
		if waiting[idx].letter == 'b' {
			ins, ok = ps.b_labels[waiting[idx].label]
		} else {
			ins, ok = ps.t_labels[waiting[idx].label]
		}
		if !ok {
			ps.err = fmt.Errorf("unknown label %s %v", waiting[idx].label, waiting[idx].loc)
			break
		}
		ps.ins[waiting[idx].ip] = ins
	}
}

func parse_toplevel(ps *parseState) {
	for tok := range ps.toks {
		switch tok.typ {
		case TOK_CMD:
			compile_cmd(ps, tok)
		case TOK_LABEL:
			compile_label(ps, tok)
		case TOK_NUM:
			n, err := strconv.Atoi(tok.args[0])
			if err != nil {
				ps.err = fmt.Errorf("Bad number <%s> %v", tok.args[0], &tok.location)
				break
			}
			compile_cond(ps, numbercond(n))
		case TOK_DOLLAR:
			compile_cond(ps, eofcond{})
		case TOK_RX:
			rx, err := regexp.Compile(tok.args[0])
			if err != nil {
				ps.err = fmt.Errorf("Bad regexp %v", &tok.location)
				break
			}
			compile_cond(ps, &regexpcond{rx})
		case TOK_EOL:
			// top level empty lines are OK
		case TOK_RBRACE:
			if ps.blockLevel == 0 {
				ps.err = fmt.Errorf("Unexpected brace %v", &tok.location)
			}
			ps.blockLevel--
			return
		default:
			ps.err = fmt.Errorf("Unexpected token %v", &tok.location)
		}
		if ps.err != nil {
			break
		}
	}
}

func mustGetToken(ps *parseState) (t *token, ok bool) {
	t, ok = <-ps.toks
	if !ok {
		ps.err = fmt.Errorf("Unexpected end of script!")
	}
	return
}

// compile_cond operates when we see a condition. It looks for
// a closing condition and an inverter '!'
func compile_cond(ps *parseState, c condition) {
	tok, ok := mustGetToken(ps)
	if !ok {
		return
	}

	switch tok.typ {
	case TOK_COMMA:
		compile_twocond(ps, c)
	case TOK_BANG:
		tok, ok = mustGetToken(ps)
		if !ok {
			return
		}
		sc := &cmd_simplecond{c, 0, len(ps.ins) + 1}
		ps.ins = append(ps.ins, sc.run)
		compile_block(ps, tok)
		sc.metloc = len(ps.ins)
	default:
		sc := &cmd_simplecond{c, len(ps.ins) + 1, 0}
		ps.ins = append(ps.ins, sc.run)
		compile_block(ps, tok)
		sc.unmetloc = len(ps.ins)
	}
}

// compile_twocond operates when we have a comma-separated
// pair of conditions, and we are expecting to read the second
// condition next.
func compile_twocond(ps *parseState, c1 condition) {
	tok, ok := mustGetToken(ps)
	if !ok {
		return
	}

	var c2 condition

	switch tok.typ {
	case TOK_NUM:
		n, err := strconv.Atoi(tok.args[0])
		if err != nil {
			ps.err = fmt.Errorf("Bad number <%s> %v", tok.args[0], &tok.location)
			break
		}
		c2 = numbercond(n)
	case TOK_DOLLAR:
		c2 = eofcond{}
	case TOK_RX:
		rx, err := regexp.Compile(tok.args[0])
		if err != nil {
			ps.err = fmt.Errorf("Bad regexp %v", &tok.location)
			break
		}
		c2 = &regexpcond{rx}
	default:
		ps.err = fmt.Errorf("Expected a second condition after comma %v", &tok.location)
	}

	if ps.err != nil {
		return
	}

	// now, we need to get the next token to determine if we're inverting
	// the condition...
	tok, ok = mustGetToken(ps)
	if !ok {
		return
	}

	switch tok.typ {
	case TOK_BANG:
		tok, ok = mustGetToken(ps)
		if !ok {
			return
		}
		tc := newTwoCond(c1, c2, 0, len(ps.ins)+1)
		ps.ins = append(ps.ins, tc.run)
		compile_block(ps, tok)
		tc.metloc = len(ps.ins)
	case TOK_CHANGE:
		// special case for 2-condition change command...
		// it has to be able to talk to the condition
		// to know when it's the last line of the change
		tc := newTwoCond(c1, c2, len(ps.ins)+1, 0)
		ps.ins = append(ps.ins, tc.run, cmd_newChanger(tok.args[0], tc))
		tc.unmetloc = len(ps.ins)
	default:
		tc := newTwoCond(c1, c2, len(ps.ins)+1, 0)
		ps.ins = append(ps.ins, tc.run)
		compile_block(ps, tok)
		tc.unmetloc = len(ps.ins)
	}
}

// compile_block parses a top-level block if it gets a
// LBRACE, or parses a single CMD as a block otherwise.
// Anything other than LBRACE or CMD is not allowed here.
func compile_block(ps *parseState, cmd *token) {
	switch cmd.typ {
	case TOK_LBRACE:
		ps.blockLevel++
		parse_toplevel(ps)
	case TOK_CMD, TOK_CHANGE:
		compile_cmd(ps, cmd)
	default:
		ps.err = fmt.Errorf("Unexpected token %v", &cmd.location)
	}
}

// compile_cmd compiles the individual sed commands
// into instructions.
func compile_cmd(ps *parseState, cmd *token) {
	switch cmd.letter {
	case '=':
		ps.ins = append(ps.ins, cmd_lineno)
	case 'D':
		ps.ins = append(ps.ins, cmd_deleteFirstLine)
	case 'G':
		ps.ins = append(ps.ins, cmd_getapp)
	case 'H':
		ps.ins = append(ps.ins, cmd_holdapp)
	case 'N':
		ps.ins = append(ps.ins, cmd_fillNextAppend)
	case 'P':
		ps.ins = append(ps.ins, cmd_printFirstLine)
	case 'a':
		ps.ins = append(ps.ins, cmd_newAppender(cmd.args[0]))
	case 'b', 't':
		compile_branchTarget(ps, len(ps.ins), cmd)
		ps.ins = append(ps.ins, zeroBranch) // placeholder
	case 'c':
		ps.ins = append(ps.ins, cmd_newChanger(cmd.args[0], nil))
	case 'd':
		ps.ins = append(ps.ins, zeroBranch)
	case 'g':
		ps.ins = append(ps.ins, cmd_get)
	case 'h':
		ps.ins = append(ps.ins, cmd_hold)
	case 'i':
		ps.ins = append(ps.ins, cmd_newInserter(cmd.args[0]))
	case 'n':
		if !noPrint {
			ps.ins = append(ps.ins, cmd_print)
		}
		ps.ins = append(ps.ins, cmd_fillNext)
	case 'p':
		ps.ins = append(ps.ins, cmd_print)
	case 'q':
		if !noPrint {
			ps.ins = append(ps.ins, cmd_print)
		}
		ps.ins = append(ps.ins, cmd_quit)
	case 's':
		subst, err := newSubstitution(cmd.args[0], cmd.args[1], cmd.args[2])
		if err != nil {
			ps.err = fmt.Errorf("Substitution error: %s %v", err.Error(), &cmd.location)
			break
		}
		ps.ins = append(ps.ins, subst)
	case 'x':
		ps.ins = append(ps.ins, cmd_swap)
	}
}

func compile_branchTarget(ps *parseState, ip int, cmd *token) {
	label := cmd.args[0]
	if len(label) == 0 {
		label = END_OF_PROGRAM_LABEL
	}

	ps.branches = append(ps.branches, waitingBranch{ip, label, cmd.letter, &cmd.location})
}

func compile_label(ps *parseState, lbl *token) {
	name := lbl.args[0]
	if len(name) == 0 {
		ps.err = fmt.Errorf("Bad label name %v", &lbl.location)
		return
	}

	// store a branch instruction to jump to the current location.
	// They will be inserted into the instruction stream in
	// the parse_resolveBranches function.
	ps.b_labels[name] = cmd_newBranch(len(ps.ins))
	ps.t_labels[name] = cmd_newChangedBranch(len(ps.ins))
}
