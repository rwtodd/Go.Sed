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
	END_OF_PROGRAM_LABEL = "the end"
)

type parseState struct {
	toks      <-chan *token   // our input
	ins       []instruction   // the compiled instructions
	branches  []waitingBranch // references to fix up
	labels    map[string]int  // the label locations
	condstack []*int          // stack of condition locations to fix up
	err       error           // record any errors we encounter
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
		case TOK_EOL:
			// top level empty lines are OK
		}
		if ps.err != nil {
			break
		}
	}
}

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
