package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {

	fl, _ := os.Open("program.sed")
	program := bufio.NewReader(fl)

	ch := make(chan *token)
	go lex(program, ch)

	instructions, err := parse(ch)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		return
	}

	eng := engine{input: bufio.NewReader(os.Stdin),
		output: bufio.NewWriter(os.Stdout),
		ins:    instructions}
	err = run(&eng)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
	}
}
