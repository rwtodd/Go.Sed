package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

var noPrint bool
var evalProg string
var sedFile string

func init() {
	flag.BoolVar(&noPrint, "n", false, "do not automatically print lines")
	flag.BoolVar(&noPrint, "silent", false, "do not automatically print lines")
	flag.BoolVar(&noPrint, "quiet", false, "do not automatically print lines")

	flag.StringVar(&evalProg, "e", "", "a string to evaluate as the program")
	flag.StringVar(&evalProg, "expression", "", "a string to evaluate as the program")

	flag.StringVar(&sedFile, "f", "", "a file to read as the program")
	flag.StringVar(&sedFile, "file", "", "a file to read as the program")
}

func compileScript(args *[]string) ([]instruction, error) {
	var program io.RuneScanner

	// STEP ONE: Find the script
	switch {
	case evalProg != "":
		program = strings.NewReader(evalProg)
		if sedFile != "" {
			return nil, fmt.Errorf("Cannot specify both an expression and a program file!")
		}
	case sedFile != "":
		fl, err := os.Open(sedFile)
		if err != nil {
			return nil, fmt.Errorf("Error opening %s: %v", sedFile, err)
		}
		defer fl.Close()
		program = bufio.NewReader(fl)
	case len(*args) > 0:
		// no -e or -f given, so the first argument is taken as the script to run
		program = strings.NewReader((*args)[0])
		*args = (*args)[1:]
	}

	// STEP TWO:  Lex/Parse/Compile the script
	ch := make(chan *token, 128)
	go lex(program, ch)

	return parse(ch)
}

func main() {
	flag.Parse()
	args := flag.Args()
	var err error

	// Find and compile the script
	instructions, err := compileScript(&args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}

	// Now, run the script against the input
	output := bufio.NewWriter(os.Stdout)

	if len(args) == 0 {
		eng := engine{input: bufio.NewReader(os.Stdin),
			output: output,
			ins:    instructions}
		err = run(&eng)
	} else {
		for _, fname := range args {
			fl, err := os.Open(fname)
			if err != nil {
				break
			}

			eng := engine{input: bufio.NewReader(fl),
				output: output,
				ins:    instructions}
			err = run(&eng)

			fl.Close()
		}
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
}
