package main

// This file has the functionality for substitution.
// It's the most complicated function, so I didn't want
// to mix it in with the other instructions in instructions.go.

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

type substitute struct {
	pattern     *regexp.Regexp // the pattern to match
	replacement string         // the template for replacements
	which       []int          // which patterns to replace
	pflag       bool           // do we print upon replacement?
}

func (s *substitute) run(e *engine) (err error) {
	e.ip++
	matches := s.pattern.FindAllStringSubmatchIndex(e.pat, -1)
	if matches == nil {
		return
	}

	if len(s.which) > 0 {
		// filter down the replacement list
		filtered := make([][]int, 0, len(s.which))

		for _, which := range s.which {
			if which >= len(matches) {
				break
			}
			filtered = append(filtered, matches[which])
		}

		if len(filtered) == 0 {
			return
		}
		matches = filtered
	}

	if len(matches) > 0 {
		e.pat = subst_replaceAll(e.pat, s, matches)
		if s.pflag {
			err = cmd_print(e)
			e.ip-- // roll back ip from the print command
		}
	}

	return
}

func subst_replaceAll(src string, subst *substitute, indexes [][]int) string {
	var substrings []string
	endpt := 0 // where we left off in the src string
	for _, idx := range indexes {
		exp := string(subst.pattern.ExpandString(nil, subst.replacement, src, idx))
		substrings = append(substrings, src[endpt:idx[0]], exp)
		endpt = idx[1]
	}
	substrings = append(substrings, src[endpt:])

	return strings.Join(substrings, "")
}

func newSubstitution(pattern string, replacement string, mods string) (instruction, error) {
	rx, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	command := &substitute{pattern: rx, replacement: replacement}
	var gflag = false

	for _, char := range mods {
		switch char {
		case 'p':
			command.pflag = true
		case 'g':
			gflag = true
		case '1', '2', '3', '4', '5', '6', '7', '8', '9':
			command.which = append(command.which, int(char-'1'))
		default:
			err = fmt.Errorf("Bad regexp modifier <%v>", char)
		}
		if err != nil {
			break
		}
	}

	// if it's not a global replacement, and they
	// didn't specify numbers, we only replace the first
	// match.
	if !gflag && (len(command.which) == 0) {
		command.which = append(command.which, 0)
	}

	// make sure any specified indices are sorted...
	sort.Ints(command.which)

	return command.run, err
}
