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
	ip    int       // address of the branch to fix up
	label string    // the target label
	loc   *location // the original parse location
}

const (
	END_OF_PROGRAM_LABEL = "the end" // has a space... no conflicts with user labels
)

type parseState struct {
	toks       <-chan *token          // our input
	ins        []instruction          // the compiled instructions
	branches   []waitingBranch        // references to fix up
	labels     map[string]instruction // named label branches
	blockLevel int                    // how deeply nested are our blocks?
	err        error                  // record any errors we encounter
}

func parse(input <-chan *token) ([]instruction, error) {
	ps := &parseState{toks: input, labels: make(map[string]instruction)}

	ps.ins = append(ps.ins, cmd_fillnext)
	parse_toplevel(ps)
	if ps.err == nil && ps.blockLevel > 0 {
		ps.err = fmt.Errorf("It looks like you are missing a closing brace!")
	}

	// if the parsing failed in some way, just give up now
	if ps.err != nil {
		return nil, ps.err
	}

	ps.labels[END_OF_PROGRAM_LABEL] = cmd_newBranch(len(ps.ins))
	ps.ins = append(ps.ins, cmd_print, zeroBranch)
	parse_resolveBranches(ps)

	return ps.ins, ps.err
}

func parse_resolveBranches(ps *parseState) {
	b := ps.branches
	for idx := range b {
		ins, ok := ps.labels[b[idx].label]
		if !ok {
			ps.err = fmt.Errorf("unknown label %s %v", b[idx].label, b[idx].loc)
			break
		}
		ps.ins[b[idx].ip] = ins
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
	case TOK_CMD:
		compile_cmd(ps, cmd)
	default:
		ps.err = fmt.Errorf("Unexpected token %v", &cmd.location)
	}
}

// compile_cmd compiles the individual sed commands
// into instructions.
func compile_cmd(ps *parseState, cmd *token) {
	switch cmd.args[0][0] {
	case '=':
		ps.ins = append(ps.ins, cmd_lineno)
	case 'H':
		ps.ins = append(ps.ins, cmd_holdapp)
	case 'b':
		compile_branchTarget(ps, len(ps.ins), cmd)
		ps.ins = append(ps.ins, zeroBranch) // placeholder
	case 'd':
		ps.ins = append(ps.ins, zeroBranch)
	case 'h':
		ps.ins = append(ps.ins, cmd_hold)
	case 'p':
		ps.ins = append(ps.ins, cmd_print)
	case 's':
		subst, err := newSubstitution(cmd.args[1], cmd.args[2], cmd.args[3])
		if err != nil {
			ps.err = fmt.Errorf("Substitution error: %s %v", err.Error(), &cmd.location)
			break
		}
		ps.ins = append(ps.ins, subst)
	case 't':
		panic("t not supported")
	case 'x':
		ps.ins = append(ps.ins, cmd_swap)
	}
}

func compile_branchTarget(ps *parseState, ip int, cmd *token) {
	label := cmd.args[1]
	if len(label) == 0 {
		label = END_OF_PROGRAM_LABEL
	}

	ps.branches = append(ps.branches, waitingBranch{ip, label, &cmd.location})
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
	ps.labels[name] = cmd_newBranch(len(ps.ins))
}
