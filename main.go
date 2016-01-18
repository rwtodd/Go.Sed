package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	eng := engine{input: bufio.NewReader(os.Stdin), output: bufio.NewWriter(os.Stdout)}

	re1, _ := newRECondition("import")
	re2, _ := newRECondition(`^\)`)

	eng.ins = append(eng.ins,
		cmd_fillnext{},
		&cmd_simplecond{eofcond{}, 2, 3}, // $ d
		&cmd_branch{0},
		newTwoCond(numbercond(8), numbercond(11), 4, 6), // 8,11 {
		cmd_lineno{},                                    //     =
		cmd_print{},                                     //     p  }
		newTwoCond(re1, re2, 7, 8),                      //  /import/,/^)/ {
		&cmd_branch{0},                                  //        d }
		cmd_print{},
		&cmd_branch{0})
	err := run(&eng)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
	}
}
