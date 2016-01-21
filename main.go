package main

import (
	"fmt"
	"strings"
)

func main() {

	var program = `
# a test program
21,5 {
   /hel$/ p
   /^one/,/39.3/ {
          s|ab(cd*)|am$1|g  # a big substitution
          G
   }
   b   
   d
:printIt
   n;p
}

g  # do it!
`

	ch := make(chan *token)
	fmt.Printf("%s\n", program)
	go lex(strings.NewReader(program), ch)

	for tok := range ch {
		fmt.Printf("%+v\n", tok)
	}

	//  old test is here:  hard-code the program and run the engine...
	// 	eng := engine{input: bufio.NewReader(os.Stdin), output: bufio.NewWriter(os.Stdout)}
	//
	// 	re1, _ := newRECondition("import")
	// 	re2, _ := newRECondition(`^\)`)
	//
	// 	eng.ins = append(eng.ins,
	// 		cmd_fillnext{},
	// 		&cmd_simplecond{eofcond{}, 2, 3}, // $ d
	// 		&cmd_branch{0},
	// 		newTwoCond(numbercond(8), numbercond(11), 4, 6), // 8,11 {
	// 		cmd_lineno{},                                    //     =
	// 		cmd_print{},                                     //     p  }
	// 		newTwoCond(re1, re2, 7, 8),                      //  /import/,/^)/ {
	// 		&cmd_branch{0},                                  //        d }
	// 		cmd_print{},
	// 		&cmd_branch{0})
	// 	err := run(&eng)
	//
	// 	if err != nil {
	// 		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
	// 	}
}
