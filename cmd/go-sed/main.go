package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/waywardcode/sed"
)


var noPrint bool
var sedFile string
type evalStrings []string
var evalProg evalStrings

func (es *evalStrings) String() string {
  return strings.Join(*es," ; ")
}

func (es *evalStrings) Set(v string) error {
  *es = append(*es,v)
  return nil
}

func init() {
	flag.BoolVar(&noPrint, "n", false, "do not automatically print lines")
	flag.BoolVar(&noPrint, "silent", false, "do not automatically print lines")
	flag.BoolVar(&noPrint, "quiet", false, "do not automatically print lines")

	flag.Var(&evalProg, "e", "a string to evaluate as the program")
	flag.Var(&evalProg, "expression", "a string to evaluate as the program")

	flag.StringVar(&sedFile, "f", "", "a file to read as the program")
	flag.StringVar(&sedFile, "file", "", "a file to read as the program")
}

func compileScript(args *[]string) (*sed.Engine, error) {
	var program io.Reader

	// STEP ONE: Find the script
	switch {
	case len(evalProg) > 0:
		program = strings.NewReader(evalProg.String())
		if sedFile != "" {
			return nil, fmt.Errorf("Cannot specify both an expression and a program file!")
		}
	case sedFile != "":
		fl, err := os.Open(sedFile)
		if err != nil {
			return nil, fmt.Errorf("Error opening %s: %v", sedFile, err)
		}
		defer fl.Close()
		program = fl
	case len(*args) > 0:
		// no -e or -f given, so the first argument is taken as the script to run
		program = strings.NewReader((*args)[0])
		*args = (*args)[1:]
	}

	// STEP TWO: compile the program
	var compiler func(io.Reader) (*sed.Engine, error)
	if noPrint {
		compiler = sed.NewQuiet
	} else {
		compiler = sed.New
	}
	return compiler(program)
}

func main() {
	flag.Parse()
	args := flag.Args()
	var err error

	// Find and compile the script
	engine, err := compileScript(&args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}

	if len(args) == 0 {
		err = engine.Run(os.Stdin, os.Stdout)
	} else {
		for _, fname := range args {
			fl, err := os.Open(fname)
			if err != nil {
				break
			}

			err = engine.Run(fl, os.Stdout)

			fl.Close()
			if err != nil {
				break
			}
		}
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
}
