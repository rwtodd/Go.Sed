package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"syscall"

	"github.com/rwtodd/Go.Sed/sed"
)

var noPrint bool
var sedFile string

type evalStrings []string

var evalProg evalStrings

var inplace bool

func (es *evalStrings) String() string {
	return strings.Join(*es, " ; ")
}

func (es *evalStrings) Set(v string) error {
	*es = append(*es, v)
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

	flag.BoolVar(&inplace, "i", false, "change file(s) inplace")
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
	default:
		// we didn't get anything valid...
		flag.Usage()
		return nil, fmt.Errorf("No sed program given (-e or -f args).")
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
		fmt.Fprintf(os.Stderr, "script compile failed: %s\n", err)
		os.Exit(1)
	}

	if len(args) == 0 {
		_, err = io.Copy(os.Stdout, engine.Wrap(os.Stdin))
		if err != nil {
			fmt.Fprintf(os.Stderr, "engine failed: %s\n", err)
			os.Exit(2)
		}
	} else {
		for _, filename := range args {
			var inputFile *os.File
			inputFile, err = os.Open(filename)
			if err != nil {
				fmt.Fprintf(os.Stderr, "open input file '%s' failed: %s\n", filename, err)
				os.Exit(3)
			}

			var (
				target   io.Writer = os.Stdout
				tempFile *os.File
			)
			if inplace {
				tempFilename := fmt.Sprintf("%s-*", path.Base(filename))
				tempFile, err = ioutil.TempFile(path.Dir(filename), tempFilename)
				if err != nil {
					fmt.Fprintf(os.Stderr, "failed to create temporary file: %s\n", err)
					os.Exit(4)
				}
				target = tempFile
			}

			_, err = io.Copy(target, engine.Wrap(inputFile))
			if err != nil {
				fmt.Fprintf(os.Stderr, "engine failed on file '%s': %s\n", filename, err)
				os.Exit(5)
			}

			inputFile.Close()
			if err != nil {
				fmt.Fprintf(os.Stderr, "closing input file '%s' failed: %s\n", filename, err)
				os.Exit(6)
			}

			if inplace {
				stat, err := os.Stat(filename)
				if err != nil {
					fmt.Fprintf(os.Stderr, "stat of '%s' failed: %s\n", filename, err)
					os.Exit(8)
				}

				err = tempFile.Chmod(stat.Mode())
				if err != nil {
					fmt.Fprintf(os.Stderr, "set mode of '%s' failed: %s\n", tempFile.Name(), err)
					os.Exit(9)
				}

				tempFile.Close()
				if err != nil {
					fmt.Fprintf(os.Stderr, "closing temporary file '%s' failed: %s\n", tempFile.Name(), err)
					os.Exit(7)
				}

				tmp := stat.Sys()
				statSys, ok := tmp.(*syscall.Stat_t)
				if ok {
					err = os.Chown(tempFile.Name(), int(statSys.Uid), int(statSys.Gid))
					if err != nil {
						// errors might be platform related, just warn
						fmt.Fprintf(os.Stderr, "failed to set UID/GID on tempfile '%s': %s\n", tempFile.Name(), err)
					}
				}

				err = os.Rename(tempFile.Name(), filename)
				if err != nil {
					fmt.Fprintf(os.Stderr, "renaming tempfile '%s' to %s failed: %s\n", tempFile.Name(), filename, err)
					os.Exit(10)
				}
			}
		}
	}
}
