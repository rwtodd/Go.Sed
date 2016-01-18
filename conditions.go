package main

import (
	"regexp"
)

// conditions are what I'm calling the '1,10' in
// commands ike '1,10 d'.  They are the line numbers,
// regexps, and '$' that you can use to control when
// commands execute.

type condition interface {
	isMet(e *engine) bool
}

// -----------------------------------------------------
type numbercond int // for matching line number conditions

func (n numbercond) isMet(e *engine) bool {
	return e.lineno == int(n)
}

// -----------------------------------------------------
type eofcond struct{} // for matching the condition '$'

func (_ eofcond) isMet(e *engine) bool {
	return e.lastl
}

// -----------------------------------------------------
type regexpcond struct {
	re *regexp.Regexp // for matching regexp conditions
}

func (r *regexpcond) isMet(e *engine) bool {
	return r.re.MatchString(e.pat)
}

func newRECondition(s string) (*regexpcond, error) {
	re, err := regexp.Compile(s)
	return &regexpcond{re}, err
}
