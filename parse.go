package main

import "fmt"

// these functions parse the lex'ed tokens (lex.go) and
// build a program for the engine (engine.go) to run.

type waitingBranch struct {
	target *int      // location to store the branch target
	label  string    // the target label
	loc    *location // the original parse location
}

const (
	END_OF_PROGRAM_LABEL = "the end" // has a space... no conflicts with user labels
)

type parseState struct {
	toks       <-chan *token   // our input
	ins        []instruction   // the compiled instructions
	branches   []waitingBranch // references to fix up
	labels     map[string]int  // the label locations
	blockLevel int             // how deeply nested are our blocks?
	err        error           // record any errors we encounter
}

func parse(input <-chan *token) ([]instruction, error) {
	ps := &parseState{toks: input, labels: make(map[string]int)}

	ps.ins = append(ps.ins, cmd_fillnext{})
	parse_toplevel(ps)
	ps.labels[END_OF_PROGRAM_LABEL] = len(ps.ins)
	ps.ins = append(ps.ins, cmd_print{}, &cmd_branch{0})
	parse_resolveBranches(ps)

	return ps.ins, ps.err
}
func parse_resolveBranches(ps *parseState) {
	for _, val := range ps.branches {
		var ok bool
		*val.target, ok = ps.labels[val.label]
		if !ok {
			ps.err = fmt.Errorf("unknown label %s %v", val.label, val.loc)
			break
		}
	}
}

func parse_toplevel(ps *parseState) {
	for tok := range ps.toks {
		switch tok.typ {
		case TOK_CMD:
			compile_cmd(ps, tok)
		case TOK_LABEL:
			compile_label(ps, tok)
		case TOK_DOLLAR:
			compile_cond(ps, eofcond{})
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
		panic("Unhandled comma.")
	case TOK_BANG:
		tok, ok = mustGetToken(ps)
		if !ok {
			return
		}
		sc := &cmd_simplecond{c, 0, len(ps.ins) + 1}
		ps.ins = append(ps.ins, sc)
		compile_block(ps, tok)
		sc.metloc = len(ps.ins)
	default:
		sc := &cmd_simplecond{c, len(ps.ins) + 1, 0}
		ps.ins = append(ps.ins, sc)
		compile_block(ps, tok)
		sc.unmetloc = len(ps.ins)
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
	case 'H':
		ps.ins = append(ps.ins, cmd_holdapp{})
	case 'b':
		b := &cmd_branch{}
		ps.ins = append(ps.ins, b)
		compile_branchTarget(ps, &b.target, cmd)
	case 'd':
		ps.ins = append(ps.ins, &cmd_branch{0})
	case 'h':
		ps.ins = append(ps.ins, cmd_hold{})
	case 'p':
		ps.ins = append(ps.ins, cmd_print{})
	case 't':
		panic("t not supported")
	case 'x':
		ps.ins = append(ps.ins, cmd_swap{})
	}
}

func compile_branchTarget(ps *parseState, tgt *int, cmd *token) {
	label := cmd.args[1]
	if len(label) == 0 {
		label = END_OF_PROGRAM_LABEL
	}

	ps.branches = append(ps.branches, waitingBranch{tgt, label, &cmd.location})
}

func compile_label(ps *parseState, lbl *token) {
	name := lbl.args[0]
	if len(name) == 0 {
		ps.err = fmt.Errorf("Bad label name %v", &lbl.location)
		return
	}

	ps.labels[name] = len(ps.ins) // point to the end of
	// the instruction stream
}
