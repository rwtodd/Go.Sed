package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	eng := engine{input: bufio.NewReader(os.Stdin), output: bufio.NewWriter(os.Stdout)}

	eng.ins = append(eng.ins,
		cmd_fillnext{},
		&cmd_simplecond{numbercond(5), 3},
		cmd_lineno{},
		newTwoCond(numbercond(8), numbercond(11), 6),
		cmd_lineno{},
		cmd_print{},
		cmd_print{},
		&cmd_branch{0})
	err := run(&eng)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
	}
}
